package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestACounterLastsTheWholeRender checks the results of Shopify Liquid 5.11.0
// for increment and decrement: a counter lasts the whole render, a loop or a
// capture included, starts at 0 whatever assign holds, and is read by a
// lookup unless an assign of the same name hides it.
func TestACounterLastsTheWholeRender(t *testing.T) {
	for _, c := range []struct{ source, want string }{
		{"{% for i in (1..2) %}{% increment c %}{% endfor %}{% increment c %}", "012"},
		{"{% assign c = 5 %}{% increment c %}{{ c }}", "05"},
		{"{% increment c %}{{ c }}{% increment c %}{{ c }}", "0112"},
		{"{% decrement c %}{{ c }}{% increment c %}{{ c }}", "-1-1-10"},
		{"{% increment c %}{% assign c = 9 %}{% increment c %}{{ c }}", "019"},
		{"{% for i in (1..2) %}{% increment c %}{{ c }}{% endfor %}", "0112"},
		{"{% capture x %}{% increment c %}{% endcapture %}{% increment c %}{{ x }}", "10"},
		{"{% liquid\nincrement c\nincrement c\necho c %}", "012"},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			r := renderWith(newEnv(mode, limits{}), c.source, nil)
			if r.parseErr != nil || r.output != c.want || len(r.errors) > 0 {
				t.Errorf("%s: %q gives %q (%v, %v), want %q", mode, c.source, r.output, r.parseErr, r.errors, c.want)
			}
		}
	}
}
