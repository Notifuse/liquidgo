package integration

import "testing"

// TestAComparisonIsRubys checks the results of Shopify Liquid 5.11.0 for <,
// >, <= and >= between values of each kind. As Ruby, a side without these
// methods (nil, a boolean, an array, a hash) makes the condition false, and
// a number compared with a string raises Ruby's ArgumentError, whose message
// names the class of the left side and the right side as Ruby writes it.
func TestAComparisonIsRubys(t *testing.T) {
	data := func() map[string]interface{} {
		return map[string]interface{}{
			"arr":   []interface{}{1, 2},
			"posts": []interface{}{map[string]interface{}{"a": 1}},
			"h":     map[string]interface{}{"a": 1},
			"s":     "abc",
			"n":     2,
			"f":     1.5,
		}
	}
	for _, c := range []struct{ source, want string }{
		{"{% if 1 > 1 %}y{% else %}n{% endif %}{% if 1 <= 1 %}y{% else %}n{% endif %}", "ny"},
		{"{% if 1 > arr %}y{% else %}n{% endif %}{% if 1 <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if 1 > true %}y{% else %}n{% endif %}{% if 1 <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if 1 > nil %}y{% else %}n{% endif %}{% if 1 <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if 1 > h %}y{% else %}n{% endif %}{% if 1 <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if 1 > 'b' %}y{% else %}n{% endif %}{% if 1 <= 'b' %}y{% else %}n{% endif %}", "Liquid error: comparison of Integer with String failedLiquid error: comparison of Integer with String failed"},
		{"{% if arr > 1 %}y{% else %}n{% endif %}{% if arr <= 1 %}y{% else %}n{% endif %}", "nn"},
		{"{% if arr > arr %}y{% else %}n{% endif %}{% if arr <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if arr > true %}y{% else %}n{% endif %}{% if arr <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if arr > nil %}y{% else %}n{% endif %}{% if arr <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if arr > h %}y{% else %}n{% endif %}{% if arr <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if arr > 'b' %}y{% else %}n{% endif %}{% if arr <= 'b' %}y{% else %}n{% endif %}", "nn"},
		{"{% if true > 1 %}y{% else %}n{% endif %}{% if true <= 1 %}y{% else %}n{% endif %}", "nn"},
		{"{% if true > arr %}y{% else %}n{% endif %}{% if true <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if true > true %}y{% else %}n{% endif %}{% if true <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if true > nil %}y{% else %}n{% endif %}{% if true <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if true > h %}y{% else %}n{% endif %}{% if true <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if true > 'b' %}y{% else %}n{% endif %}{% if true <= 'b' %}y{% else %}n{% endif %}", "nn"},
		{"{% if nil > 1 %}y{% else %}n{% endif %}{% if nil <= 1 %}y{% else %}n{% endif %}", "nn"},
		{"{% if nil > arr %}y{% else %}n{% endif %}{% if nil <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if nil > true %}y{% else %}n{% endif %}{% if nil <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if nil > nil %}y{% else %}n{% endif %}{% if nil <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if nil > h %}y{% else %}n{% endif %}{% if nil <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if nil > 'b' %}y{% else %}n{% endif %}{% if nil <= 'b' %}y{% else %}n{% endif %}", "nn"},
		{"{% if h > 1 %}y{% else %}n{% endif %}{% if h <= 1 %}y{% else %}n{% endif %}", "nn"},
		{"{% if h > arr %}y{% else %}n{% endif %}{% if h <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if h > true %}y{% else %}n{% endif %}{% if h <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if h > nil %}y{% else %}n{% endif %}{% if h <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if h > h %}y{% else %}n{% endif %}{% if h <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if h > 'b' %}y{% else %}n{% endif %}{% if h <= 'b' %}y{% else %}n{% endif %}", "nn"},
		{"{% if s > 1 %}y{% else %}n{% endif %}{% if s <= 1 %}y{% else %}n{% endif %}", "Liquid error: comparison of String with 1 failedLiquid error: comparison of String with 1 failed"},
		{"{% if s > arr %}y{% else %}n{% endif %}{% if s <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if s > true %}y{% else %}n{% endif %}{% if s <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if s > nil %}y{% else %}n{% endif %}{% if s <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if s > h %}y{% else %}n{% endif %}{% if s <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if s > 'b' %}y{% else %}n{% endif %}{% if s <= 'b' %}y{% else %}n{% endif %}", "ny"},
		{"{% if '2' > 1 %}y{% else %}n{% endif %}{% if '2' <= 1 %}y{% else %}n{% endif %}", "Liquid error: comparison of String with 1 failedLiquid error: comparison of String with 1 failed"},
		{"{% if '2' > arr %}y{% else %}n{% endif %}{% if '2' <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if '2' > true %}y{% else %}n{% endif %}{% if '2' <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if '2' > nil %}y{% else %}n{% endif %}{% if '2' <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if '2' > h %}y{% else %}n{% endif %}{% if '2' <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if '2' > 'b' %}y{% else %}n{% endif %}{% if '2' <= 'b' %}y{% else %}n{% endif %}", "ny"},
		{"{% if f > 1 %}y{% else %}n{% endif %}{% if f <= 1 %}y{% else %}n{% endif %}", "yn"},
		{"{% if f > arr %}y{% else %}n{% endif %}{% if f <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if f > true %}y{% else %}n{% endif %}{% if f <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if f > nil %}y{% else %}n{% endif %}{% if f <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if f > h %}y{% else %}n{% endif %}{% if f <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if f > 'b' %}y{% else %}n{% endif %}{% if f <= 'b' %}y{% else %}n{% endif %}", "Liquid error: comparison of Float with String failedLiquid error: comparison of Float with String failed"},
		{"{% if posts > 1 %}y{% else %}n{% endif %}{% if posts <= 1 %}y{% else %}n{% endif %}", "nn"},
		{"{% if posts > arr %}y{% else %}n{% endif %}{% if posts <= arr %}y{% else %}n{% endif %}", "nn"},
		{"{% if posts > true %}y{% else %}n{% endif %}{% if posts <= true %}y{% else %}n{% endif %}", "nn"},
		{"{% if posts > nil %}y{% else %}n{% endif %}{% if posts <= nil %}y{% else %}n{% endif %}", "nn"},
		{"{% if posts > h %}y{% else %}n{% endif %}{% if posts <= h %}y{% else %}n{% endif %}", "nn"},
		{"{% if posts > 'b' %}y{% else %}n{% endif %}{% if posts <= 'b' %}y{% else %}n{% endif %}", "nn"},
		{"{% if 1 > f %}y{% else %}n{% endif %}{% if 1 <= f %}y{% else %}n{% endif %}", "ny"},
		{"{% if 1 > s %}y{% else %}n{% endif %}{% if 1 <= s %}y{% else %}n{% endif %}", "Liquid error: comparison of Integer with String failedLiquid error: comparison of Integer with String failed"},
		{"{% if f > f %}y{% else %}n{% endif %}{% if f <= f %}y{% else %}n{% endif %}", "ny"},
		{"{% if f > s %}y{% else %}n{% endif %}{% if f <= s %}y{% else %}n{% endif %}", "Liquid error: comparison of Float with String failedLiquid error: comparison of Float with String failed"},
		{"{% if s > f %}y{% else %}n{% endif %}{% if s <= f %}y{% else %}n{% endif %}", "Liquid error: comparison of String with 1.5 failedLiquid error: comparison of String with 1.5 failed"},
		{"{% if s > s %}y{% else %}n{% endif %}{% if s <= s %}y{% else %}n{% endif %}", "ny"},
		{"{% if '2' > f %}y{% else %}n{% endif %}{% if '2' <= f %}y{% else %}n{% endif %}", "Liquid error: comparison of String with 1.5 failedLiquid error: comparison of String with 1.5 failed"},
		{"{% if '2' > s %}y{% else %}n{% endif %}{% if '2' <= s %}y{% else %}n{% endif %}", "ny"},
		{"{% if n > 1 %}y{% endif %}{% if n < 2.5 %}y{% endif %}{% if s < 'b' %}y{% endif %}{% if 'b' >= s %}y{% endif %}", "yyyy"},
	} {
		r := renderWith(newEnv("lax", limits{}), c.source, data())
		if r.parseErr != nil || r.output != c.want {
			t.Errorf("%q gives %q (%v), want %q", c.source, r.output, r.parseErr, c.want)
		}
	}
}
