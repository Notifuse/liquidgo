package integration

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// boundedCase is a template that a resource limit, or the way the library
// iterates, must stop before it allocates budget bytes.
type boundedCase struct {
	name     string
	source   string
	partials map[string]string
	limits   limits
	budget   uint64
	// Exactly one of the two: the output the render must give, or a text
	// the parse or render error must contain.
	wantOutput string
	wantError  string
}

const (
	memoryLimit = "Memory limits exceeded"
	tooDeep     = "Nesting too deep"
)

// allLimits sets the three resource limits, as an application that renders
// templates written by its users should.
var allLimits = limits{renderLength: 1_000_000, renderScore: 100_000, assignScore: 1_000_000}

// fanOut is a partial that renders itself twice until d reaches depth, then
// writes one character: 2^depth renders for a few bytes of template.
func fanOut(depth int) map[string]string {
	return map[string]string{"p": fmt.Sprintf("{%% assign e = d | plus: 1 %%}{%% if e < %d %%}{%% render 'p', d: e %%}{%% render 'p', d: e %%}{%% endif %%}x", depth+1)}
}

// doubled returns a loop that doubles s, which starts as "x", times times.
func doubled(times int, body string) string {
	return fmt.Sprintf("{%% assign s = 'x' %%}{%% for i in (1..%d) %%}%s{%% endfor %%}{{ s.size }}", times, body)
}

// The ranges are 2,000,000 long, which a materialized range allocates in
// about 200 MB: enough to fail the budget, and small enough not to end the
// test binary while the case is known.
var boundedCases = []boundedCase{
	// Ranges are iterated, not built. As in Shopify, reversed applies to
	// the items that offset and limit keep.
	{name: "range-limit", source: "{% for i in (1..2000000) limit: 1 %}{{ i }}{% endfor %}", budget: 1 << 20, wantOutput: "1"},
	{name: "range-reversed", source: "{% for i in (1..2000000) reversed limit: 1 %}{{ i }}{% endfor %}", budget: 1 << 20, wantOutput: "1"},
	{name: "range-offset", source: "{% for i in (1..2000000) offset: 1999999 %}{{ i }}{% endfor %}", budget: 1 << 20, wantOutput: "2000000"},
	{name: "tablerow-range-limit", source: "{% tablerow i in (1..2000000) limit: 1 %}{{ i }}{% endtablerow %}", budget: 1 << 20,
		wantOutput: "<tr class=\"row1\">\n<td class=\"col1\">1</td></tr>\n"},
	// Every iteration counts in the render score, even with an empty body.
	// The ranges and the output are small, so that the render score is the
	// only guard these cases test.
	{name: "empty-loop", source: "{% for i in (1..2000) %}{% endfor %}", limits: limits{renderScore: 1000}, budget: 1 << 20, wantError: memoryLimit},
	{name: "nested-empty-loops", source: "{% for i in (1..100) %}{% for j in (1..100) %}{% endfor %}{% endfor %}", limits: limits{renderScore: 1000}, budget: 1 << 20, wantError: memoryLimit},
	{name: "empty-tablerow", source: "{% tablerow i in (1..2000) %}{% endtablerow %}", limits: limits{renderScore: 100}, budget: 1 << 20, wantError: memoryLimit},
	// A partial does not start the scores again: 2^12 renders score about
	// 32,000 in all, and about 4 each.
	{name: "partial-fan-out", source: "{% render 'p', d: 0 %}", partials: fanOut(12), limits: limits{renderScore: 1000}, budget: 10 << 20, wantError: memoryLimit},
	// A string or an array is refused before it is built: 26 doublings
	// build 64 MB.
	{name: "append-doubled", source: doubled(26, "{% assign s = s | append: s %}"), limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "prepend-doubled", source: doubled(26, "{% assign s = s | prepend: s %}"), limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "capture-doubled", source: doubled(26, "{% capture s %}{{ s }}{{ s }}{% endcapture %}"), limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "replace-chain", source: "{{ 'x'" + rep(" | replace: 'x', 'xx'", 26) + " | size }}", limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "replace-amplified", source: fmt.Sprintf("{%% assign s = '%s' %%}{%% assign t = s | replace: 'x', s %%}{{ t.size }}", rep("x", 8192)),
		limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "join-amplified", source: fmt.Sprintf("{%% assign parts = '%s' | split: ',' %%}{%% assign t = parts | join: '%s' %%}{{ t.size }}", rep("x,", 8192), rep("y", 8192)),
		limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "split-large", source: fmt.Sprintf("{%% assign parts = '%s' | split: '' %%}{{ parts.size }}", rep("x", 4_000_000)),
		limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	{name: "concat-doubled", source: fmt.Sprintf("{%% assign parts = '%s' | split: ',' %%}{%% for i in (1..12) %%}{%% assign parts = parts | concat: parts %%}{%% endfor %%}{{ parts.size }}", rep("x,", 1000)),
		limits: allLimits, budget: 50 << 20, wantError: memoryLimit},
	// Each write of a tag is checked against the length limit, not only
	// the whole tag once it has rendered.
	{name: "tablerow-length", source: "{% tablerow i in (1..5000) %}xxxxxxxxxx{% endtablerow %}", limits: limits{renderLength: 1000}, budget: 2 << 20, wantError: memoryLimit},
	// Nesting is refused while parsing, before it is deep enough to cost
	// time quadratic in it.
	{name: "brackets-100", source: "{{ " + rep("a[", 100) + "0" + rep("]", 100) + " }}", budget: 50 << 20, wantOutput: ""},
	{name: "brackets-101", source: "{{ " + rep("a[", 101) + "0" + rep("]", 101) + " }}", budget: 50 << 20, wantError: tooDeep},
	{name: "liquid-101", source: "{% liquid " + rep("liquid ", 101) + "echo 1 %}", budget: 50 << 20, wantError: tooDeep},
}

// knownUnbounded lists the cases known to fail, with the reason. A listed
// case that passes fails the test until its line is removed.
var knownUnbounded = map[string]string{
	"range-limit":        "a range is built as an array before the loop",
	"range-reversed":     "a range is built as an array before the loop",
	"range-offset":       "a range is built as an array before the loop",
	"empty-loop":         "an iteration with an empty body adds nothing to the render score",
	"nested-empty-loops": "an iteration with an empty body adds nothing to the render score",
	"empty-tablerow":     "an iteration with an empty body adds nothing to the render score",
	"partial-fan-out":    "a partial starts the resource scores again, and keeps the limit error to itself",
	"replace-chain":      "a filter result in an output is not counted",
	"replace-amplified":  "a filter result is counted only once it is built",
	"join-amplified":     "a filter result is counted only once it is built",
	"split-large":        "a filter result is counted only once it is built",
	"brackets-101":       "expressions have no nesting limit",
	"liquid-101":         "liquid tags have no nesting limit",
}

// TestHostileTemplatesAreBounded renders each case and checks its result,
// and that the render allocated less than its budget.
func TestHostileTemplatesAreBounded(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range boundedCases {
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			env := newEnv("", c.limits)
			if c.partials != nil {
				env.SetFileSystem(NewStubFileSystem(c.partials))
			}
			var r result
			runtime.GC()
			alloc := allocated(func() {
				requireNoPanicEscapes(t, func() { r = renderWith(env, c.source, hostileData()) })
			})
			var problems []string
			if alloc > c.budget {
				problems = append(problems, fmt.Sprintf("allocated %s, over the budget of %s", mb(alloc), mb(c.budget)))
			}
			if c.wantError != "" && !r.hasError(c.wantError) && !strings.Contains(r.output, c.wantError) {
				problems = append(problems, fmt.Sprintf("got output %q and errors %v, want the error %q", clip(r.output), r.errors, c.wantError))
			}
			if c.wantError == "" && (r.output != c.wantOutput || r.parseErr != nil || len(r.errors) > 0) {
				problems = append(problems, fmt.Sprintf("got output %q, parse error %v and errors %v, want %q", clip(r.output), r.parseErr, r.errors, c.wantOutput))
			}
			reason, known := knownUnbounded[c.name]
			switch {
			case len(problems) > 0 && known:
				t.Logf("known (%s): %s", reason, strings.Join(problems, "; "))
			case len(problems) > 0:
				t.Error(strings.Join(problems, "; "))
			case known:
				t.Errorf("knownUnbounded lists %s, which now passes (%s allocated): remove it", c.name, mb(alloc))
			default:
				t.Logf("%s allocated", mb(alloc))
			}
		})
	}
	for name := range knownUnbounded {
		if !seen[name] {
			t.Errorf("knownUnbounded lists %s, which is not a case", name)
		}
	}
}

// clip shortens an output for a test message.
func clip(s string) string {
	if len(s) > 80 {
		return s[:80] + "…"
	}
	return s
}

// cacheRun makes the names of each run of TestParseCachesAreBounded new, so
// that -count=2 does not find them in the caches.
var cacheRun int

// TestParseCachesAreBounded parses and renders templates with distinct
// variable names, and evaluates as many expressions with Context.Get, and
// checks that the heap does not keep them once they are unreachable. The
// template lookups have no brackets, which the variable lookup cache leaves
// out; the expressions have one, so that only the expression cache keeps them.
func TestParseCachesAreBounded(t *testing.T) {
	cacheRun++
	env := newEnv("", limits{})
	heap := func() int64 {
		runtime.GC()
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return int64(m.HeapAlloc)
	}
	for _, c := range []struct {
		name, known string
		fill        func(i int)
	}{
		{"variable lookups", "the variable lookup cache is global and never evicted", func(i int) {
			// About 100 KB of distinct lookups.
			var b strings.Builder
			for j := range 4000 {
				fmt.Fprintf(&b, "{{ r%d_v%d_%d.k }}", cacheRun, i, j)
			}
			requireNoPanicEscapes(t, func() { renderWith(env, b.String(), nil) })
		}},
		{"Context.Get", "the expression cache is global and never evicted", func(i int) {
			ctx := newContext(env)
			for j := range 4000 {
				ctx.Get(fmt.Sprintf("r%d_w%d_%d.k[0]", cacheRun, i, j))
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			before := heap()
			for i := range 20 {
				c.fill(i)
			}
			retained := heap() - before
			const budget = 5 << 20
			switch {
			case retained > budget && c.known != "":
				t.Logf("known (%s): %s retained", c.known, mb(uint64(retained)))
			case retained > budget:
				t.Errorf("%s retained after GC, over the budget of %s", mb(uint64(retained)), mb(budget))
			case c.known != "":
				t.Errorf("known as %q, but only %s is retained: clear it", c.known, mb(uint64(max(retained, 0))))
			default:
				t.Logf("%s retained", mb(uint64(max(retained, 0))))
			}
		})
	}
}
