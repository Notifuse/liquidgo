package integration

import (
	"fmt"
	"testing"
	"time"
)

// scalingFamily builds a template of size n from one shape of input: a large
// but ordinary template, an author's mistake, or a hostile template. The test
// renders it at n and 4n.
type scalingFamily struct {
	name  string
	n     int
	build func(n int) string
}

// A loop that appends to a string with assign is not a family: as in
// Shopify, append copies the string, so the loop is quadratic by nature, and
// assign_score_limit bounds it.
//
// The sizes are chosen so that a render at 4n fits in the memory of a CI
// runner on a quadratic or exponential implementation, so that a listed
// family fails slowly instead of ending the test binary.
var scalingFamilies = []scalingFamily{
	// Output written by loops and tags.
	{"for-output", 5000, func(n int) string { return fmt.Sprintf("{%% for i in (1..%d) %%}x{%% endfor %%}", n) }},
	{"flat-variables", 500, func(n int) string { return rep("{{ long }}", n) }},
	{"tablerow", 1000, func(n int) string { return fmt.Sprintf("{%% tablerow i in (1..%d) %%}{{ i }}{%% endtablerow %%}", n) }},
	{"cycle", 5000, func(n int) string {
		return fmt.Sprintf("{%% for i in (1..%d) %%}{%% cycle 'a', 'b' %%}{%% endfor %%}", n)
	}},
	{"capture-linear", 5000, func(n int) string {
		return fmt.Sprintf("{%% capture s %%}{%% for i in (1..%d) %%}x{%% endfor %%}{%% endcapture %%}{{ s.size }}", n)
	}},
	{"for-nested2", 500, func(n int) string {
		return fmt.Sprintf("{%% for i in (1..%d) %%}{%% for j in (1..10) %%}x{%% endfor %%}{%% endfor %%}", n)
	}},
	// One item per level, so that the work is linear in the depth when
	// each block renders its body once.
	{"nested-tablerow", 3, func(n int) string {
		return rep("{% tablerow i in (1..1) %}", n) + "x" + rep("{% endtablerow %}", n)
	}},
	{"nested-ifchanged", 3, func(n int) string { return rep("{% ifchanged %}", n) + "x" + rep("{% endifchanged %}", n) }},
	// Nesting, under the depth limit of blocks at 4n.
	{"nested-if", 20, func(n int) string { return rep("{% if true %}", n) + "x" + rep("{% endif %}", n) }},
	{"nested-for", 20, func(n int) string { return rep("{% for i in (1..1) %}", n) + "x" + rep("{% endfor %}", n) }},
	// Parsing.
	{"nested-brackets", 2000, func(n int) string { return "{{ " + rep("a[", n) + "0" + rep("]", n) + " }}" }},
	{"nested-paren", 2000, func(n int) string { return "{{ " + rep("(", n) + "1..2" + rep(")", n) + " }}" }},
	{"nested-liquid", 500, func(n int) string { return "{% liquid " + rep("liquid ", n) + "echo 1 %}" }},
	{"raw-body", 2500, func(n int) string { return "{% raw %}" + rep("{{ x }} {% y %}", n) + "{% endraw %}" }},
	{"unclosed-raw", 2500, func(n int) string { return rep("{% raw %}", n) }},
	{"comment-unclosed", 5000, func(n int) string { return rep("{% comment %}", n) }},
	{"unclosed-var", 5000, func(n int) string { return rep("{{ a ", n) }},
	{"unclosed-tag", 5000, func(n int) string { return rep("{% if a ", n) }},
	{"many-tags", 5000, func(n int) string { return rep("{% assign a = 1 %}", n) }},
	{"liquid-tag-lines", 5000, func(n int) string { return "{% liquid\n" + rep("assign a = 1\n", n) + "%}" }},
	{"long-filter-chain", 2000, func(n int) string { return "{{ 'x'" + rep(" | upcase", n) + " }}" }},
	// Filters on a string whose length grows with n.
	{"slice-loop", 2000, func(n int) string {
		return fmt.Sprintf("{%% assign s = '%s' %%}{%% for i in (1..%d) %%}{{ s | slice: 1, 1 }}{%% endfor %%}", rep("a", n), n)
	}},
	{"truncate-loop", 2000, func(n int) string {
		return fmt.Sprintf("{%% assign s = '%s' %%}{%% for i in (1..%d) %%}{{ s | truncate: 4 }}{%% endfor %%}", rep("a", n), n)
	}},
	{"split-sort-uniq", 5000, func(n int) string {
		return fmt.Sprintf("{%% assign a = '%s' | split: '' %%}{%% assign b = a | sort | uniq | join: ',' %%}{{ b.size }}", rep("ab", n))
	}},
	{"strip-html", 5000, func(n int) string { return fmt.Sprintf("{{ '%s' | strip_html | size }}", rep("<a", n)) }},
	{"truncatewords", 5000, func(n int) string { return fmt.Sprintf("{{ '%s' | truncatewords: 1000000 | size }}", rep("a ", n)) }},
}

// knownSuperLinear lists the families known to cost more than linear time,
// with the reason. A listed family that becomes linear fails the test until
// its line is removed.
var knownSuperLinear = map[string]string{
	"for-output":       "the output is built by string concatenation",
	"flat-variables":   "the output is built by string concatenation",
	"tablerow":         "the output is built by string concatenation",
	"cycle":            "the output is built by string concatenation",
	"capture-linear":   "the output is built by string concatenation",
	"for-nested2":      "the output is built by string concatenation",
	"slice-loop":       "slice copies the string to runes, and the output is built by string concatenation",
	"truncate-loop":    "truncate copies the string to runes, and the output is built by string concatenation",
	"nested-tablerow":  "a block renders its body more than once",
	"nested-ifchanged": "a block renders its body more than once",
	"nested-brackets":  "nested lookups are parsed again at each level",
	"nested-liquid":    "nested liquid tags are parsed again at each level",
	"raw-body":         "the raw body is built by string concatenation",
	"unclosed-raw":     "the raw body is built by string concatenation",
}

// TestRenderScalesLinearly parses and renders each family at n and 4n and
// checks that 4 times the input takes less than 9 times as long (16 times for
// a quadratic cost), a ratio that does not depend on the machine or -race.
func TestRenderScalesLinearly(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test")
	}
	env := newEnv("", emailLimits)
	measure := func(t *testing.T, source string) time.Duration {
		best := time.Duration(0)
		for range 3 {
			start := time.Now()
			var r result
			requireNoPanicEscapes(t, func() { r = renderWith(env, source, hostileData()) })
			if err := r.internalError(); err != nil {
				t.Fatalf("panic in the render: %v", err)
			}
			if d := time.Since(start); best == 0 || d < best {
				best = d
			}
		}
		return best
	}
	seen := map[string]bool{}
	for _, f := range scalingFamilies {
		seen[f.name] = true
		t.Run(f.name, func(t *testing.T) {
			if _, known := knownSuperLinear[f.name]; known && raceEnabled {
				// A known family only shows that it is still slow,
				// which the run without the race detector checks.
				t.Skip("known super-linear family, measured without -race")
			}
			small, large := f.build(f.n), f.build(4*f.n)
			measure(t, small) // warm up
			base, got := measure(t, small), measure(t, large)
			ratio := float64(got) / float64(max(base, time.Microsecond))
			// Below a few milliseconds the timer and the GC dominate.
			superLinear := ratio > 9 && got > 20*time.Millisecond
			reason, known := knownSuperLinear[f.name]
			switch {
			case superLinear && known:
				t.Logf("known (%s): %d bytes in %v, %d bytes in %v (x%.1f)", reason, len(small), base, len(large), got, ratio)
			case superLinear:
				t.Errorf("%d bytes in %v, %d bytes in %v: x%.1f for x4 input", len(small), base, len(large), got, ratio)
			case known && ratio > 5 && got > 20*time.Millisecond:
				// A loaded machine brings a quadratic ratio of 16 down
				// to here; a linear family is about 4.
				t.Logf("known (%s), inconclusive: x%.1f", reason, ratio)
			case known:
				t.Errorf("knownSuperLinear lists %s, which is now linear (x%.1f): remove it", f.name, ratio)
			default:
				t.Logf("%d bytes in %v, %d bytes in %v (x%.1f)", len(small), base, len(large), got, ratio)
			}
		})
	}
	for name := range knownSuperLinear {
		if !seen[name] {
			t.Errorf("knownSuperLinear lists %s, which is not a family", name)
		}
	}
}
