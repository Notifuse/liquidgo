package integration

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
	"strings"
	"testing"
)

// TestARangeIsACollectionForTheFilters checks the results of Shopify Liquid
// 5.11.0 for a range given to the array filters: its integers, as Ruby
// iterates a Range. first and last give the bounds of the range even when it
// is empty, as Range#first and Range#last.
func TestARangeIsACollectionForTheFilters(t *testing.T) {
	data := func() map[string]interface{} {
		return map[string]interface{}{"strs": []interface{}{"x", "y"}}
	}
	for _, c := range []struct{ source, want string }{
		{`{{ (1..3) | join }}|{{ (-2..2) | join: "," }}|{{ (5..1) | join }}`, "1 2 3|-2,-1,0,1,2|"},
		{`{{ (1..3) | reverse }}|{{ (1..3) | sort | join: "," }}|{{ (1..4) | sum }}|{{ (1..3) | uniq | size }}`, "321|1,2,3|10|3"},
		{`{{ (1..3) | compact | last }}|{{ (1..2) | concat: strs | join: "," }}`, "3|1,2,x,y"},
		{`{{ (5..1) | first }}|{{ (5..1) | last }}|{{ (5..1) | size }}|{{ (1..0) | reverse | size }}`, "5|1|0|0"},
		{`{% assign r = (2..4) %}{{ r | size }}{{ r | first }}{{ r | last }}{{ r | join: "" }}`, "324234"},
		{`{{ (1..3) | sort | first }}|{{ (1..3) | sort_natural | last }}`, "1|3"},
		{`{{ (1..3) | join: ',' }}|{{ (1..3) | size }}|{{ (1..3) | last }}|{{ (1..3) | first }}`, "1,2,3|3|3|1"},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			r := renderWith(newEnv(mode, limits{}), c.source, data())
			if r.parseErr != nil || r.output != c.want || len(r.errors) > 0 {
				t.Errorf("%s: %q gives %q (%v, %v), want %q", mode, c.source, r.output, r.parseErr, r.errors, c.want)
			}
		}
	}
}

// TestALargeRangeStaysWithinTheLimits checks that a range given to a filter
// does not build more than the limits allow: join stops at the render length
// limit, as Ruby's MemoryError, and a filter that builds an array of the
// integers raises at the assign score limit, as the other filters that build
// an array. Without these limits, a range of more than 1,048,576 integers
// raises too, where Ruby would build it. size, first and last read the
// bounds without building anything.
func TestALargeRangeStaysWithinTheLimits(t *testing.T) {
	const memory = "Liquid error: Memory limits exceeded"
	for _, c := range []struct {
		source string
		l      limits
		want   string
	}{
		{"a{{ (1..100000000) | join }}b", limits{renderLength: 50}, memory},
		{"a{{ (1..100000000) | sort | first }}b", limits{assignScore: 50}, memory},
		{"a{{ (1..100000000) | size }}b{{ (1..100000000) | last }}", limits{renderLength: 50, assignScore: 5}, "a100000000b100000000"},
		{"a{{ (1..2147483647) | size }}", limits{assignScore: 5}, "a2147483647"},
		// Without the matching limit, a filter builds at most 1,048,576
		// integers of a range, and join writes at most as many
		{"a{{ (1..300000000) | reverse | size }}", limits{}, memory},
		{"a{{ (1..300000000) | sort | first }}", limits{renderLength: 10000000}, memory},
		{"a{{ (1..2147483647) | join }}", limits{assignScore: 50}, memory},
		{"a{{ (1..2147483647) | map: 'x' | size }}", limits{}, memory},
	} {
		var r result
		used := allocated(func() { r = renderWith(newEnv("lax", c.l), c.source, nil) })
		if !strings.Contains(r.output, c.want) {
			t.Errorf("%q gives %q, want %q", c.source, r.output, c.want)
		}
		if used > 8<<20 {
			t.Errorf("%q allocated %s", c.source, mb(used))
		}
	}
}

// TestARangeUpToTheCeilingIsBuilt checks that a filter still builds a range
// of 1,048,576 integers without limits.
func TestARangeUpToTheCeilingIsBuilt(t *testing.T) {
	r := renderWith(newEnv("lax", limits{}), "a{{ (1..1048576) | reverse | first }}", nil)
	if r.output != "a1048576" {
		t.Errorf("a range of 1,048,576 integers gives %q", r.output)
	}
}
