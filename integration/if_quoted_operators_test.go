package integration

import "testing"

// and, or, | and parentheses inside a quoted string are part of the string, not of the
// condition (Ruby's If::ExpressionsAndOperators scans quoted fragments whole).
func TestIfQuotedOperators(t *testing.T) {
	tests := []struct {
		name     string
		template string
		assigns  map[string]interface{}
		expected string
	}{
		{"and inside quotes", `{% if v == "black and white" %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "black and white"}, "Y"},
		{"or inside quotes, missing variable", `{% if w == "x or 1" %}Y{% else %}N{% endif %}`, nil, "N"},
		{"or inside single quotes", `{% if v == 'a or b' %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "a or b"}, "Y"},
		{"country name", `{% if c == "Trinidad and Tobago" %}Y{% else %}N{% endif %}`, map[string]interface{}{"c": "Trinidad and Tobago"}, "Y"},
		{"spaces kept", `{% if v == "a  b" %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "a  b"}, "Y"},
		{"contains a quoted operator", `{% if v contains " and " %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "Trinidad and Tobago"}, "Y"},
		{"pipe inside quotes", `{% if v == "a|b" %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "a|b"}, "Y"},
		{"parentheses inside quotes", `{% if v == "(x)" %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "(x)"}, "Y"},
		{"real and after a quoted and", `{% if v == "a and b" and n == 1 %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "a and b", "n": 1}, "Y"},
		{"real or after a quoted or", `{% if v == "a or b" or n == 1 %}Y{% else %}N{% endif %}`, map[string]interface{}{"v": "x", "n": 2}, "N"},
		{"right to left", `{% if true or false and false %}Y{% else %}N{% endif %}`, nil, "Y"},
		{"word containing and", `{% if brand == "x" or android %}Y{% else %}N{% endif %}`, map[string]interface{}{"android": true}, "Y"},
		{"elsif with quoted and", `{% if false %}A{% elsif v == "b and c" %}B{% endif %}`, map[string]interface{}{"v": "b and c"}, "B"},
		{"unless with quoted or", `{% unless v == "x or y" %}U{% endunless %}`, map[string]interface{}{"v": "x"}, "U"},
	}
	for _, tt := range tests {
		for _, mode := range []string{"lax", "strict"} {
			t.Run(tt.name+"/"+mode, func(t *testing.T) {
				assertTemplateResult(t, tt.expected, tt.template, tt.assigns, TemplateResultOptions{ErrorMode: mode})
			})
		}
	}
}
