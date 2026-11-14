package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// customEnvWith returns the environment of the custom tags with limits l.
func customEnvWith(l limits) *liquid.Environment {
	env := customTagEnvironment()
	set := map[string]interface{}{}
	if l.renderLength > 0 {
		set["render_length_limit"] = l.renderLength
	}
	if l.assignScore > 0 {
		set["assign_score_limit"] = l.assignScore
	}
	env.SetDefaultResourceLimits(set)
	return env
}

// TestACaptureChecksTheTagsInIt checks that a tag that renders into a string
// of its own inside a capture is checked against the assign score, as the
// capture is, and not against the render length limit.
func TestACaptureChecksTheTagsInIt(t *testing.T) {
	short := renderWith(customEnvWith(limits{renderLength: 5}),
		"{% capture c %}{% blockbuffer %}xxxxxxxxxx{% endblockbuffer %}{% endcapture %}ok", nil)
	if short.output != "ok" || len(short.errors) > 0 {
		t.Errorf("a capture longer than the render length limit renders %q (%v), want %q", short.output, short.errors, "ok")
	}

	env := customEnvWith(limits{assignScore: 1000})
	source := "{% capture c %}{% blockbuffer %}{% for i in (1..1000) %}{% for j in (1..300) %}xxxxxxxxxx{% endfor %}{% endfor %}{% endblockbuffer %}{% endcapture %}"
	var r result
	if alloc := allocated(func() { r = renderWith(env, source, nil) }); alloc > 2<<20 {
		t.Errorf("a tag in a capture allocated %s before the assign score stopped it", mb(alloc))
	}
	if !r.hasError("Memory limits exceeded") {
		t.Errorf("a tag in a capture renders %q (%v), want the limit error", clip(r.output), r.errors)
	}

	counted := renderWith(customEnvWith(limits{assignScore: 12}),
		"{% capture c %}{% blockbuffer %}xxxxxxxxxx{% endblockbuffer %}{% endcapture %}{{ c }}", nil)
	if counted.output != "<xxxxxxxxxx>" || len(counted.errors) > 0 {
		t.Errorf("a tag in a capture is counted once: renders %q (%v), want %q", counted.output, counted.errors, "<xxxxxxxxxx>")
	}
}

// TestIfchangedCountsOnlyTheBodyItKeeps checks that a body that ifchanged
// does not write counts in no limit, as in Ruby.
func TestIfchangedCountsOnlyTheBodyItKeeps(t *testing.T) {
	for _, c := range []struct {
		source string
		limits limits
		want   string
	}{
		{"{% for i in (1..2) %}{% ifchanged %}xxxxxxxxxx{% endifchanged %}{% endfor %}", limits{renderLength: 15}, "xxxxxxxxxx"},
		{"{% capture c %}{% for i in (1..5) %}{% ifchanged %}xxxxxxxxxx{% endifchanged %}{% endfor %}{% endcapture %}{{ c }}", limits{assignScore: 15}, "xxxxxxxxxx"},
	} {
		if r := renderWith(newEnv("", c.limits), c.source, nil); r.output != c.want || len(r.errors) > 0 {
			t.Errorf("%q renders %q (%v), want %q", c.source, r.output, r.errors, c.want)
		}
	}
}

// TestTheErrorsOfARenderedPartialAreReported checks that an error written in
// the output of a partial of render is in the errors of the template, as
// Ruby's subcontexts share the errors.
func TestTheErrorsOfARenderedPartialAreReported(t *testing.T) {
	env := panicEnvironment(limits{})
	env.SetFileSystem(NewStubFileSystem(map[string]string{"p": "x{% crash %}y", "r": "{% render 'r' %}"}))
	for _, c := range []struct{ source, want, wantError string }{
		{"a{% render 'p' %}b", "axLiquid error: internalyb", "internal"},
		{"{% assign two = '1,2' | split: ',' %}a{% render 'p' for two %}b", "axLiquid error: internalyxLiquid error: internalyb", "internal"},
		{"a{% render 'r' %}b", "aLiquid error: Nesting too deepb", "Nesting too deep"},
	} {
		r := renderWith(env, c.source, nil)
		if r.output != c.want {
			t.Errorf("%q renders %q, want %q", c.source, r.output, c.want)
		}
		if len(r.errors) == 0 || !strings.Contains(r.errors[0].Error(), c.wantError) {
			t.Errorf("%q: errors %v, want one with %q", c.source, r.errors, c.wantError)
		}
	}
}

// TestFilterResultsStayWithinTheRenderLength checks that each filter of a
// chain is checked, so that a chain whose filters each grow their input is
// stopped before its result is built.
func TestFilterResultsStayWithinTheRenderLength(t *testing.T) {
	source := "{{ 'xxxxxxxxxxxxxxxxxxxxxxxx'" + strings.Repeat(" | base64_encode", 60) + " | size }}"
	var r result
	if alloc := allocated(func() { r = renderWith(newEnv("", allLimits), source, nil) }); alloc > 20<<20 {
		t.Errorf("a chain of base64_encode allocated %s", mb(alloc))
	}
	if !r.hasError("Memory limits exceeded") && !strings.Contains(r.output, "Memory limits exceeded") {
		t.Errorf("a chain of base64_encode renders %q (%v), want the limit error", clip(r.output), r.errors)
	}
}

// TestArrayFiltersCheckOnlyTheirResult checks that split and concat refuse an
// array with more items than the assign score leaves, without counting the
// arrays they build in a loop, and that split does not count the trailing
// empty strings it drops.
func TestArrayFiltersCheckOnlyTheirResult(t *testing.T) {
	loop := "{% assign kl = 'k,l' | split: ',' %}{% for i in (1..150) %}{{ 'a,b,c,d,e,f,g,h,i,j' | split: ',' | concat: kl | size }}{% endfor %}"
	if r := renderWith(newEnv("", limits{assignScore: 1000}), loop, nil); r.output != strings.Repeat("12", 150) || len(r.errors) > 0 {
		t.Errorf("150 splits under an assign score limit of 1000 render %q (%v)", clip(r.output), r.errors)
	}
	trailing := "{{ 'a" + strings.Repeat(",", 40) + "' | split: ',' | size }}"
	if r := renderWith(newEnv("", limits{assignScore: 10}), trailing, nil); r.output != "1" || len(r.errors) > 0 {
		t.Errorf("%q renders %q (%v), want %q", trailing, r.output, r.errors, "1")
	}
	over := fmt.Sprintf("{{ '%s' | split: ',' | size }}", strings.Repeat("a,", 11))
	if r := renderWith(newEnv("", limits{assignScore: 10}), over, nil); !r.hasError("Memory limits exceeded") && !strings.Contains(r.output, "Memory limits exceeded") {
		t.Errorf("%q renders %q (%v), want the limit error", over, r.output, r.errors)
	}
}
