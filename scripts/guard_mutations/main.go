// Command guard_mutations removes each guard of liquidgo, one at a time, and
// checks that the tests fail without it: a guard that no test isolates can be
// lost without anyone noticing (category 11 of the security check).
//
// Run from the root of the repository:
//
//	go run ./scripts/guard_mutations            # every guard
//	go run ./scripts/guard_mutations NAME...    # the guards whose name contains NAME
//
// Each mutation replaces a piece of code that must occur exactly once in its
// file; a guard that moved fails the run, so that the list is kept up to
// date. The mutated file is given to go test with -overlay: nothing is
// written in the repository.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// guard is one guard and the mutation that removes it: old replaced by new
// in file, and, for guards that cover each other, the edits of extraEdits.
type guard struct {
	name string
	file string
	old  string
	new  string
}

// edit is one more replacement of a mutation.
type edit struct{ file, old, new string }

// extraEdits are the other replacements of a mutation, by guard name.
var extraEdits = map[string][]edit{
	// the early return and the clamp each bound the precision of round
	"round precision": {{"liquid/math_filters.go",
		"case d.Cmp(big.NewInt(maxRoundDigits)) > 0:\n\t\treturn maxRoundDigits + 1",
		"case false:\n\t\treturn maxRoundDigits + 1"}},
}

var guards = []guard{
	// Resource limits
	{"render score limit", "liquid/resource_limits.go",
		"if rl.renderScoreLimit != nil && rl.renderScore > *rl.renderScoreLimit {",
		"if false && rl.renderScoreLimit != nil && rl.renderScore > *rl.renderScoreLimit {"},
	{"assign score limit", "liquid/resource_limits.go",
		"if rl.assignScoreLimit != nil && rl.assignScore > *rl.assignScoreLimit {",
		"if false && rl.assignScoreLimit != nil && rl.assignScore > *rl.assignScoreLimit {"},
	{"render length of a write", "liquid/output_buffer.go",
		"if rl.renderLengthLimit != nil && size > *rl.renderLengthLimit {",
		"if false && rl.renderLengthLimit != nil && size > *rl.renderLengthLimit {"},
	{"assign size of a scratch buffer", "liquid/output_buffer.go",
		"if rl.assignScoreLimit != nil && size > *rl.assignScoreLimit {",
		"if false && rl.assignScoreLimit != nil && size > *rl.assignScoreLimit {"},
	{"captured writes count as assigns", "liquid/output_buffer.go",
		"case countAssign:\n\t\t\trl.IncrementAssignScore(len(s))",
		"case countAssign:\n\t\t\t_ = rl"},
	{"render score of a block body", "liquid/block_body.go",
		"rl.IncrementRenderScore(max(1, len(bb.nodelist)))",
		"_ = rl"},
	{"assign score of an assign", "liquid/tags/assign.go",
		"rl.IncrementAssignScore(score)",
		"_ = score"},
	{"render length of a filter result", "liquid/context.go",
		"if limit := c.resourceLimits.renderLengthLimit; limit != nil && len(str) > *limit {",
		"if limit := c.resourceLimits.renderLengthLimit; false && limit != nil && len(str) > *limit {"},
	{"range items a filter builds", "liquid/filter_limits.go",
		"\tif n > maxRangeItems {\n\t\tif rl := sf.limits(); rl != nil {",
		"\tif false {\n\t\tif rl := sf.limits(); rl != nil {"},
	{"range items join writes", "liquid/array_filters.go",
		"n := r.Len()\n\tif n > maxRangeItems {",
		"n := r.Len()\n\tif false {"},
	{"round precision", "liquid/math_filters.go",
		"if digits > maxRoundDigits {",
		"if false {"},
	{"partials an environment keeps", "liquid/partial_cache.go",
		"for c.order.Len() > 0 && (c.order.Len() > maxEntries || c.bytes > maxBytes) {",
		"for false {"},
	{"a partial too large is not kept", "liquid/partial_cache.go",
		"if maxEntries <= 0 || maxBytes <= 0 || len(source) > maxBytes/8 {",
		"if maxEntries <= 0 || maxBytes <= 0 {"},
	{"date directive width", "liquid/strftime.go",
		"oversized = err != nil || width > maxStrftimeWidth",
		"oversized = err != nil"},

	// Depth
	{"block depth at parse", "liquid/block.go",
		"if parseContext.Depth() >= blockMaxDepth {",
		"if false {"},
	{"liquid tag depth at parse", "liquid/block_body.go",
		"if parseContext.Depth() >= blockMaxDepth {",
		"if false {"},
	{"scope depth", "liquid/context.go",
		"return c.baseScopeDepth+len(c.scopes) > blockMaxDepth",
		"return false"},
	{"expression depth", "liquid/parser.go",
		"if p.depth > blockMaxDepth+1 {",
		"if false {"},
	{"bracket depth of a lookup", "liquid/variable_lookup.go",
		"if bracketDepth(markup) > blockMaxDepth {",
		"if false {"},
	{"equality depth", "liquid/equality.go",
		"if depth > maxEqualityDepth {",
		"if false {"},

	// Stopping a render
	{"cancellation", "liquid/context.go",
		"\tif c == nil || c.done == nil {\n\t\treturn\n\t}\n\tselect {",
		"\tif true {\n\t\treturn\n\t}\n\tselect {"},
	{"a Go panic ends the render", "liquid/rescue.go",
		"\tif p := fatalPanic(context, r); p != nil {\n\t\tpanic(p)\n\t}\n",
		""},

	// What a template reads of Go values
	{"a missing key of a drop is undefined", "liquid/drop.go",
		"if ctx != nil && ctx.StrictVariables() {",
		"if false {"},
	{"the error of a drop is raised", "liquid/drop.go",
		"\tif err != nil {\n\t\tpanic(asLiquidError(err))\n\t}\n",
		"\tif err != nil {\n\t\treturn nil\n\t}\n"},
	{"array filters pass the render to a drop", "liquid/array_filters.go",
		"\tif g, ok := asDrop(item); ok {\n\t\treturn dropGet(ctx, g, property)\n\t}",
		"\tif g, ok := asDrop(item); ok {\n\t\treturn dropGet(nil, g, property)\n\t}"},
	{"a channel or a func writes no address", "liquid/utils.go",
		"\t\tif hasNoText(obj) {\n\t\t\treturn \"\"\n\t\t}\n",
		""},
	{"a nil pointer is nil", "liquid/variable_lookup.go",
		"\tif isNilPointer(obj) {\n\t\treturn nil\n\t}\n\treturn ToLiquid(obj)",
		"\treturn ToLiquid(obj)"},
	{"a reflect.Value writes its String", "liquid/utils.go",
		"\t\tif s, ok := obj.(fmt.Stringer); ok {\n\t\t\treturn s.String()\n\t\t}\n\t\tif items",
		"\t\tif _, ok := obj.(fmt.Stringer); ok {\n\t\t\treturn fmt.Sprintf(\"%v\", obj)\n\t\t}\n\t\tif items"},
	{"an expression is stripped as Ruby strips it", "liquid/expression.go",
		"return strings.Trim(s, \" \\t\\n\\v\\f\\r\\x00\")",
		"return strings.TrimSpace(s)"},
	{"a NUL byte is not the end of a variable", "liquid/tokenizer.go",
		"\t\tfor okA && byteA != closeCurley && byteA != openCurley {\n\t\t\tbyteA, okA = t.ss.next()\n\t\t}\n\n\t\tif !okA {",
		"\t\tfor okA && byteA != 0 && byteA != closeCurley && byteA != openCurley {\n\t\t\tbyteA, okA = t.ss.next()\n\t\t}\n\n\t\tif !okA || byteA == 0 {"},
	{"a plain struct writes nothing", "liquid/utils.go",
		"if isPlainStruct(obj) {\n\t\t\treturn \"\"\n\t\t}",
		""},
	{"a plain struct inspects as nil", "liquid/utils.go",
		"if isPlainStruct(obj) {\n\t\t\treturn \"nil\"\n\t\t}",
		""},
	{"a struct value is a drop by its pointer", "liquid/drop.go",
		"if t.Kind() != reflect.Struct || !reflect.PointerTo(t).Implements(liquidGetterType) {",
		"if t.Kind() != reflect.Struct || true {"},
	{"only a standard filter shares its arguments", "liquid/context.go",
		"\t\tif c.filters[i].name == name {\n\t\t\treturn c.filters[i].fn, false\n\t\t}",
		"\t\tif c.filters[i].name == name {\n\t\t\treturn c.filters[i].fn, true\n\t\t}"},
	{"an environment filter does not share its arguments", "liquid/context.go",
		"\t\tif fn, ok := c.environment.filters[name]; ok {\n\t\t\treturn fn, false\n\t\t}",
		"\t\tif fn, ok := c.environment.filters[name]; ok {\n\t\t\treturn fn, true\n\t\t}"},
	{"an unknown filter calls nothing", "liquid/context.go",
		"\t} else {\n\t\tresult = obj\n\t}",
		"\t} else if g, ok := asDrop(obj); ok {\n\t\tresult = dropGet(c, g, method)\n\t} else {\n\t\tresult = obj\n\t}"},

	// Escaping
	{"output escape", "liquid/variable.go",
		"writeValue(escapingWriter{output}, value, nil)",
		"WriteValue(output, value)"},

	// The imports of the library are checked by TestTheLibraryImportsNoSystemAccess,
	// which reads the sources on disk: an overlay does not reach it.

	// Files
	{"template name check", "liquid/file_system.go",
		"if !TemplateNameRegex.MatchString(templatePath) {",
		"if false {"},
	{"a resolved file stays under the root", "liquid/file_system.go",
		"if !under(resolved, root) {",
		"if !under(resolved, root) && false {"},
	{"a sibling of the root is not under it", "liquid/file_system.go",
		"return err == nil && rel != \"..\" && !strings.HasPrefix(rel, \"..\"+string(filepath.Separator)) && !filepath.IsAbs(rel)",
		"return strings.HasPrefix(path, root) || (err == nil && rel == \"\")"},
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	tmp, err := os.MkdirTemp("", "guard-mutations-")
	if err != nil {
		fail(err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	bad := 0
	for i, g := range guards {
		if !selected(g.name, os.Args[1:]) {
			continue
		}
		verdict := run(root, tmp, i, g)
		fmt.Printf("%-45s %s\n", g.name, verdict)
		if !strings.HasPrefix(verdict, "killed") {
			bad++
		}
	}
	if bad > 0 {
		fmt.Printf("\n%d guard(s) not isolated by a test\n", bad)
		os.Exit(1)
	}
}

// run applies the mutation of g and runs the tests, and says what happened.
func run(root, tmp string, i int, g guard) string {
	files := map[string][]byte{}
	for _, e := range append([]edit{{g.file, g.old, g.new}}, extraEdits[g.name]...) {
		path := filepath.Join(root, e.file)
		src, ok := files[path]
		if !ok {
			read, err := os.ReadFile(path)
			if err != nil {
				return "error: " + err.Error()
			}
			src = read
		}
		if n := bytes.Count(src, []byte(e.old)); n != 1 {
			return fmt.Sprintf("not found: the code occurs %d times in %s, update the guard", n, e.file)
		}
		files[path] = bytes.Replace(src, []byte(e.old), []byte(e.new), 1)
	}
	replace := map[string]string{}
	n := 0
	for path, src := range files {
		n++
		mutated := filepath.Join(tmp, fmt.Sprintf("%d_%d_%s", i, n, filepath.Base(path)))
		if err := os.WriteFile(mutated, src, 0o644); err != nil {
			return "error: " + err.Error()
		}
		replace[path] = mutated
	}
	overlay, _ := json.Marshal(map[string]map[string]string{"Replace": replace})
	overlayFile := filepath.Join(tmp, fmt.Sprintf("%d.json", i))
	if err := os.WriteFile(overlayFile, overlay, 0o644); err != nil {
		return "error: " + err.Error()
	}
	cmd := exec.Command("go", "test", "-overlay", overlayFile, "-count=1", "-timeout", "120s", "./liquid/...", "./integration")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	switch {
	case err == nil:
		return "SURVIVED: no test fails without this guard"
	case bytes.Contains(out, []byte("[build failed]")) || bytes.Contains(out, []byte("[setup failed]")):
		return "error: the mutation does not compile\n" + string(out)
	case bytes.Contains(out, []byte("--- FAIL")):
		return "killed"
	case bytes.Contains(out, []byte("panic: test timed out")):
		return "killed (a test hangs)"
	case bytes.Contains(out, []byte("\nFAIL\t")):
		return "killed (the tests crash)"
	}
	return "error: go test failed without a failing test\n" + string(out)
}

func selected(name string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		if strings.Contains(name, f) {
			return true
		}
	}
	return false
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
