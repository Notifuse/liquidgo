package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestTheTemplateAssignsWinOverTheRenderData checks Ruby's lookup order: a
// variable the template assigns or captures is read before the render data
// of the same name.
func TestTheTemplateAssignsWinOverTheRenderData(t *testing.T) {
	data := map[string]interface{}{
		"name": "<b>",
		"posts": []interface{}{
			map[string]interface{}{"title": "t1", "public": true},
			map[string]interface{}{"title": "t2", "public": false},
		},
	}
	for _, c := range []struct{ source, want string }{
		{"{% assign name = name | escape %}{{ name }}", "&lt;b&gt;"},
		{"{% capture name %}[{{ name }}]{% endcapture %}{{ name }}", "[<b>]"},
		{"{% assign posts = posts | where: 'public' %}{{ posts.size }}{% for p in posts %}{{ p.title }}{% endfor %}", "1t1"},
		{"{% for i in (1..2) %}{% assign name = i %}{% endfor %}{{ name }}", "2"},
		{"{{ name }}", "<b>"},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			if r := renderWith(newEnv(mode, limits{}), c.source, data); r.output != c.want || r.parseErr != nil || len(r.errors) > 0 {
				t.Errorf("%s: %q renders %q (%v %v), want %q", mode, c.source, r.output, r.parseErr, r.errors, c.want)
			}
		}
	}
}
