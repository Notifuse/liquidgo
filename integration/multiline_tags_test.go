package integration

import "testing"

// A tag or a variable may span several lines, as in Ruby (whose regexes use /m).
func TestMultilineTags(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"if on its own lines", "{%\nif x\n%}Y{%\nendif\n%}", "Y"},
		{"newline before the tag end", "{% if x\n%}Y{% endif %}", "Y"},
		{"for over several lines", "0{%\nfor i in (1..3)\n%} {{\ni\n}}{%\nendfor\n%}", "0 1 2 3"},
		{"liquid tag with whitespace control", "{%- liquid\n  assign a = 1\n  echo a\n-%}", "1"},
		{"assign with a filter on the next line", "{% assign s = 'a'\n  | upcase %}{{ s }}", "A"},
		{"variable with a filter on the next line", "{{ 'a'\n  | upcase }}", "A"},
		{"when over several lines", "{% case n %}{% when 1,\n  2 %}low{% endcase %}", "low"},
		{"liquid tag with a blank line", "{%- liquid\n  assign a = 1\n\n  echo a\n-%}", "1"},
		{"blank line inside a block of a liquid tag", "{%- liquid\n  for i in (1..2)\n\n    echo i\n  endfor\n\n  echo 'z'\n-%}", "12z"},
		{"named cycle over several lines", "{% cycle 'g':\n 'a', 'b' %}", "a"},
	}
	for _, tt := range tests {
		for _, mode := range []string{"lax", "strict"} {
			t.Run(tt.name+"/"+mode, func(t *testing.T) {
				assertTemplateResult(t, tt.expected, tt.template, map[string]interface{}{"x": true, "n": 2}, TemplateResultOptions{ErrorMode: mode})
			})
		}
	}
}
