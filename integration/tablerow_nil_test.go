package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestATablerowOverNilRendersNothing checks the results of Shopify Liquid
// 5.11.0 for a tablerow over nil, which renders nothing, and over a value
// that is not a collection, which renders an empty row.
func TestATablerowOverNilRendersNothing(t *testing.T) {
	data := map[string]interface{}{"n": 2, "f": 1.5}
	for _, c := range []struct{ source, want string }{
		{"{% tablerow i in (1.3) %}x{% endtablerow %}", ""},
		{"{% tablerow i in nothing %}x{% endtablerow %}", ""},
		{"{% tablerow i in n %}x{% endtablerow %}", "<tr class=\"row1\">\n</tr>\n"},
		{"{% tablerow i in f %}x{% endtablerow %}", "<tr class=\"row1\">\n</tr>\n"},
		{"{% tablerow i in (1..0) %}x{% endtablerow %}", "<tr class=\"row1\">\n</tr>\n"},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			r := renderWith(newEnv(mode, limits{}), c.source, data)
			if r.parseErr != nil || r.output != c.want {
				t.Errorf("%s: %q gives %q (%v), want %q", mode, c.source, r.output, r.parseErr, c.want)
			}
		}
	}
}
