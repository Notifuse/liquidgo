package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestAnArrayRendersItsItemsAsRuby checks the results of Shopify Liquid
// 5.11.0 for an array in an output, an echo or a cycle: its items one after
// the other, nested arrays included and nil items left out. A filter that
// takes a string still receives the array as Ruby's inspect writes it.
func TestAnArrayRendersItsItemsAsRuby(t *testing.T) {
	data := func() map[string]interface{} {
		return map[string]interface{}{
			"arr":   []interface{}{1, []interface{}{2, nil, []interface{}{3}}, "a", map[string]interface{}{"k": "v"}},
			"strs":  []string{"x", "y"},
			"empty": []interface{}{},
			"nils":  []interface{}{nil, nil},
			"f":     []interface{}{1.5, true},
		}
	}
	for _, c := range []struct{ source, want string }{
		{"{{ arr }}", `123a{"k"=>"v"}`},
		{"{% echo arr %}", `123a{"k"=>"v"}`},
		{"{% liquid\necho strs %}", "xy"},
		{"[{{ empty }}][{{ nils }}]", "[][]"},
		{"{{ f }}", "1.5true"},
		{"{{ strs }}", "xy"},
		{"{% assign a = strs %}{{ a }}", "xy"},
		{"{% capture c %}{{ strs }}{% endcapture %}{{ c | size }}", "2"},
		{"{% cycle arr, 'b' %}", `123a{"k"=>"v"}`},
		{`{{ strs | append: "!" }}`, `["x", "y"]!`},
		{`{{ arr | append: "" }}`, `[1, [2, nil, [3]], "a", {"k"=>"v"}]`},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			r := renderWith(newEnv(mode, limits{}), c.source, data())
			if r.parseErr != nil || r.output != c.want || len(r.errors) > 0 {
				t.Errorf("%s: %q gives %q (%v, %v), want %q", mode, c.source, r.output, r.parseErr, r.errors, c.want)
			}
		}
	}
}

// TestASelfContainingArrayRendersOnce checks that an array that holds itself,
// which Ruby cannot render, renders its other items without recursing
// without end.
func TestASelfContainingArrayRendersOnce(t *testing.T) {
	arr := []interface{}{1, nil, 2}
	arr[1] = arr
	r := renderWith(newEnv("lax", limits{}), "{{ arr }}", map[string]interface{}{"arr": arr})
	if r.output != "12" {
		t.Errorf("a self-containing array gives %q, want %q", r.output, "12")
	}
}
