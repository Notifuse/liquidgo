package integration

import "testing"

// Keyword arguments reach the filter as a trailing map, as in Ruby, and a call with too
// many arguments no longer panics.
func TestFilterKeywordArguments(t *testing.T) {
	assigns := map[string]interface{}{"flag": true, "f": false}
	tests := []struct {
		template string
		expected string
	}{
		{`{{ false | default: "x", allow_false: true }}`, "false"},
		{`{{ f | default: "x", allow_false: flag }}`, "false"},
		{`{{ nil | default: "x", allow_false: true }}`, "x"},
		{`{{ false | default: "x", allow_false: false }}`, "x"},
		{`{{ false | default: "x" }}`, "x"},
		{`{{ "" | default: "x", allow_false: true }}`, "x"},
		{`{% assign v = false | default: "x", allow_false: true %}{{ v }}`, "false"},
		{`{{ "a" | upcase: 1 }}`, "a"},
		{`{{ "a" | upcase: foo: 1 }}`, "a"},
	}
	for _, tt := range tests {
		for _, mode := range []string{"lax", "strict"} {
			t.Run(tt.template+"/"+mode, func(t *testing.T) {
				assertTemplateResult(t, tt.expected, tt.template, assigns, TemplateResultOptions{ErrorMode: mode})
			})
		}
	}
}
