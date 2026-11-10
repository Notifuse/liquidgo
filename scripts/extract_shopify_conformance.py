#!/usr/bin/env python3
"""Extract conformance cases from the Shopify Liquid Ruby test suite.

Reads the Ruby tests as data (nothing is executed) and writes the cases used by
integration/shopify_conformance_test.go:

  * every `assert_template_result(expected, template, assigns, ...)` written with
    literal values;
  * every `assert_equal(expected, @filters.name(args...))` in
    test/integration/standard_filter_test.rb written with literal values.

The enclosing Ruby context of each call is kept: `with_error_modes(...)` and the
`error_mode:` keyword give the error modes the case runs in, `assert_raises(...)`
marks a case that must fail. A case inside any other block (custom tags, global
filters, loops, time zones...) is dropped, since its meaning depends on code.

Usage:
  python3 scripts/extract_shopify_conformance.py <shopify-liquid-checkout> \
      integration/testdata/shopify_conformance/cases.json

Floats are written as {"__f": 1.5} so that the Go side can tell 1.0 from 1.
"""

import json
import os
import re
import sys


class Skip(Exception):
    """The Ruby expression is not a plain literal."""


ESCAPES = {"n": "\n", "t": "\t", '"': '"', "\\": "\\", "r": "\r", "e": "\x1b", "0": "\0", "#": "#"}


class Parser:
    """A tiny parser for Ruby literals: strings, numbers, symbols, arrays, hashes, heredocs."""

    def __init__(self, src, pos):
        self.s = src
        self.i = pos

    def ws(self):
        while self.i < len(self.s):
            c = self.s[self.i]
            if c in " \t\r\n":
                self.i += 1
            elif c == "#":
                while self.i < len(self.s) and self.s[self.i] != "\n":
                    self.i += 1
            elif self.s.startswith("\\\n", self.i):
                self.i += 2
            else:
                break

    def peek(self):
        self.ws()
        return self.s[self.i] if self.i < len(self.s) else ""

    def value(self):
        self.ws()
        c = self.s[self.i]
        rest = self.s[self.i:]
        if c in "\"'":
            v = self.string()
        elif c == "{":
            v = self.hash()
        elif c == "[":
            v = self.array()
        elif re.match(r":[a-z_]+", rest):
            m = re.match(r":([a-z_]+)", rest)
            self.i += m.end()
            v = m.group(1)
        elif rest.startswith("<<~") or rest.startswith("<<-"):
            v = self.heredoc()
        else:
            m = re.match(r"-?\d+(\.\d+)?(e-?\d+)?", rest)
            after = self.s[self.i + m.end():self.i + m.end() + 1] if m else ""
            if m and not re.match(r"[\w.]", after or " "):
                self.i += m.end()
                t = m.group(0)
                v = {"__f": float(t)} if ("." in t or "e" in t) else int(t)
            else:
                for kw, val in (("true", True), ("false", False), ("nil", None)):
                    if rest.startswith(kw) and not re.match(r"\w", rest[len(kw):len(kw) + 1] or " "):
                        self.i += len(kw)
                        v = val
                        break
                else:
                    raise Skip("expression")
        self.ws()
        if self.i < len(self.s) and self.s[self.i] in ".+*":
            if self.s.startswith(".freeze", self.i):
                self.i += len(".freeze")
                return v
            raise Skip("operator")
        return v

    def heredoc(self):
        m = re.match(r"<<[~-]([A-Z_]+)(\.chomp)?", self.s[self.i:])
        if not m:
            raise Skip("heredoc")
        tag, chomp = m.group(1), bool(m.group(2))
        self.i += m.end()
        j = self.s.index("\n", self.i) + 1
        lines = []
        while True:
            e = self.s.index("\n", j)
            line = self.s[j:e]
            if line.strip() == tag:
                break
            lines.append(line)
            j = e + 1
        if "#{" in "".join(lines):
            raise Skip("interpolation")
        indent = min([len(l) - len(l.lstrip()) for l in lines if l.strip()] or [0])
        text = "\n".join(l[indent:] for l in lines) + "\n"
        text = text.replace("\\\\", "\x00").replace("\\n", "\n").replace("\x00", "\\")
        return text[:-1] if chomp else text

    def string(self):
        q = self.s[self.i]
        self.i += 1
        out = ""
        while True:
            c = self.s[self.i]
            if c == "\\":
                n = self.s[self.i + 1]
                if q == "'":
                    out += n if n in "\\'" else "\\" + n
                elif n in ESCAPES:
                    out += ESCAPES[n]
                else:
                    raise Skip("escape")
                self.i += 2
            elif c == q:
                self.i += 1
                break
            elif q == '"' and self.s.startswith("#{", self.i):
                raise Skip("interpolation")
            else:
                out += c
                self.i += 1
        self.ws()
        if self.i < len(self.s) and self.s[self.i] in "\"'":
            return out + self.string()
        return out

    def hash(self):
        self.i += 1
        d = {}
        while True:
            if self.peek() == "}":
                self.i += 1
                return d
            self.ws()
            m = re.match(r"([a-z_]\w*):(?!:)", self.s[self.i:])
            if m:
                k = m.group(1)
                self.i += m.end()
            else:
                k = self.value()
                if not isinstance(k, str):
                    raise Skip("key")
                self.ws()
                if self.s.startswith("=>", self.i):
                    self.i += 2
                elif self.s[self.i] == ":":
                    self.i += 1
                else:
                    raise Skip("hash")
            d[k] = self.value()
            if self.peek() == ",":
                self.i += 1

    def array(self):
        self.i += 1
        a = []
        while True:
            if self.peek() == "]":
                self.i += 1
                return a
            a.append(self.value())
            if self.peek() == ",":
                self.i += 1


def code_lines(src):
    """The source lines with heredoc bodies blanked, so that they do not count as indentation."""
    lines = src.split("\n")
    out, tag = [], None
    for line in lines:
        if tag is not None:
            out.append("")
            if line.strip() == tag:
                tag = None
            continue
        out.append(line)
        m = re.search(r"<<[~-]?([A-Z_]+)", line)
        if m:
            tag = m.group(1)
    return out


MODE = re.compile(r":(lax|strict2|strict|warn)")


def context(lines, lineno, col_text):
    """Return (modes, raises, test) for the call on 1-based line `lineno`, or raise Skip.

    The enclosing blocks are found by indentation: the Shopify suite is formatted by
    rubocop, so the nearest preceding line with a smaller indent opens the block.
    """
    modes, raises, test = None, "", ""
    # Inline block on the same line, e.g. `assert_raises(SyntaxError) { assert_template_result(...) }`.
    m = re.search(r"assert_raises\(([\w:]+)", col_text)
    if m:
        raises = m.group(1)
    if re.search(r"with_error_modes?\(", col_text):
        modes = MODE.findall(col_text)
    indent = len(lines[lineno - 1]) - len(lines[lineno - 1].lstrip())
    for k in range(lineno - 2, -1, -1):
        line = lines[k]
        if not line.strip():
            continue
        ind = len(line) - len(line.lstrip())
        if ind >= indent:
            continue
        indent = ind
        text = line.strip()
        if text.startswith("def "):
            test = text[4:].split("(")[0].strip()
            continue
        if text.startswith(("class ", "module ")):
            break
        if re.match(r"with_error_modes?\(", text):
            if modes is None:
                modes = MODE.findall(text)
            continue
        m = re.match(r"(?:\w+\s*=\s*)?assert_raises\(([\w:]+)", text)
        if m:
            raises = raises or m.group(1)
            continue
        raise Skip("context: " + text[:60])
    return modes, raises, test


def template_cases(path, rel):
    src = open(path).read()
    lines = code_lines(src)
    for m in re.finditer(r"assert_template_result\s*(\(|\s)", src):
        lineno = src.count("\n", 0, m.start()) + 1
        line_text = lines[lineno - 1]
        if line_text.lstrip().startswith("def "):
            continue
        p = Parser(src, m.end())
        try:
            args, kw = [], {}
            while True:
                p.ws()
                km = re.match(r"([a-z_]+):\s", p.s[p.i:])
                if km:
                    p.i += km.end()
                    kw[km.group(1)] = p.value()
                else:
                    args.append(p.value())
                p.ws()
                if p.i < len(p.s) and p.s[p.i] == ",":
                    p.i += 1
                    continue
                break
            if len(args) < 2 or not isinstance(args[0], str) or not isinstance(args[1], str):
                raise Skip("arguments")
            unknown = set(kw) - {"error_mode", "partials", "render_errors", "message"}
            if unknown:
                raise Skip("keyword " + ",".join(sorted(unknown)))
            modes, raises, test = context(lines, lineno, line_text[: line_text.find("assert_template_result")])
            if "error_mode" in kw:
                modes = [kw["error_mode"]]
            yield {
                "source": "%s:%d" % (rel, lineno),
                "test": test,
                "kind": "template",
                "template": args[1],
                "assigns": args[2] if len(args) > 2 else {},
                "partials": kw.get("partials") or {},
                "render_errors": bool(kw.get("render_errors")),
                "modes": modes,
                "raises": raises,
                "expected": args[0],
            }
        except (Skip, IndexError, ValueError):
            continue


def filter_cases(path, rel):
    src = open(path).read()
    lines = code_lines(src)
    for m in re.finditer(r"assert_equal\s*\(", src):
        lineno = src.count("\n", 0, m.start()) + 1
        p = Parser(src, m.end())
        try:
            expected = p.value()
            p.ws()
            if src[p.i] != ",":
                continue
            p.i += 1
            p.ws()
            fm = re.match(r"@filters\.([a-z_]+)\(", src[p.i:])
            if not fm:
                continue
            name = fm.group(1)
            p.i += fm.end()
            args, kw = [], {}
            while True:
                if p.peek() == ")":
                    p.i += 1
                    break
                p.ws()
                km = re.match(r"([a-z_]+):\s", src[p.i:])
                if km:
                    p.i += km.end()
                    kw[km.group(1)] = p.value()
                else:
                    args.append(p.value())
                if p.peek() == ",":
                    p.i += 1
            if not args:
                continue
            last = args[-1]
            if name == "default" and isinstance(last, dict) and "__f" not in last:
                kw.update(args.pop())
            modes, raises, test = context(lines, lineno, "")
            assigns, params = {"in": args[0]}, []
            for i, a in enumerate(args[1:]):
                assigns["p%d" % i] = a
                params.append("p%d" % i)
            for k, v in kw.items():
                assigns["k_" + k] = v
                params.append("%s: k_%s" % (k, k))
            call = name + ((": " + ", ".join(params)) if params else "")
            yield {
                "source": "%s:%d" % (rel, lineno),
                "test": test,
                "kind": "filter",
                "template": "{%% assign __result = in | %s %%}" % call,
                "assigns": assigns,
                "partials": {},
                "render_errors": False,
                "modes": modes,
                "raises": raises,
                "expected": expected,
            }
        except (Skip, IndexError, ValueError, TypeError):
            continue


def main():
    root, out_path = sys.argv[1], sys.argv[2]
    cases = []
    test_root = os.path.join(root, "test")
    for dirpath, _, files in sorted(os.walk(test_root)):
        for f in sorted(files):
            if not f.endswith(".rb"):
                continue
            path = os.path.join(dirpath, f)
            rel = os.path.relpath(path, root)
            cases.extend(template_cases(path, rel))
            if f == "standard_filter_test.rb":
                cases.extend(filter_cases(path, rel))
    seen = {}
    for c in cases:
        n = seen.get(c["source"], 0)
        seen[c["source"]] = n + 1
        c["id"] = c["source"] if n == 0 else "%s#%d" % (c["source"], n + 1)
    with open(out_path, "w") as fh:
        json.dump(cases, fh, indent=1, ensure_ascii=False, sort_keys=True)
        fh.write("\n")
    print(len(cases), "cases")


if __name__ == "__main__":
    main()
