package integration

import "testing"

// raw, comment and doc end at the first tag naming their delimiter, even with text
// before it in the same token or whitespace control, as Ruby's FullTokenPossiblyInvalid.
func TestRawCommentDocDelimiters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"raw with whitespace control", "> {%- raw -%} inner {%- endraw -%} <", "> inner <"},
		{"raw body is never trimmed", "> {%- raw -%} inner {%- endraw %} <", "> inner  <"},
		{"raw around a variable", "> {%- raw -%}{{ test }}{%- endraw -%} <", ">{{ test }}<"},
		{"raw with an unclosed tag inside", "{% raw %} Foobar {% invalid {% endraw %}", " Foobar {% invalid "},
		{"raw with an unclosed variable inside", "{% raw %} Foobar {{ invalid {% endraw %}{{ 1 }}", " Foobar {{ invalid 1"},
		{"raw across lines", "{% raw %}a\n{{ b }}\n{%\nendraw\n%}", "a\n{{ b }}\n"},
		{"comment with raw inside", "{% comment %}{% raw %} {{%%%%}}  }} { {% endcomment %} {% comment {% endraw %} {% endcomment %}", ""},
		{"comment closing tag over lines", "a{% comment %}x{%\nendcomment\n%}b", "ab"},
		{"comment with whitespace control", "a {% comment %}x{% endcomment -%} b", "a b"},
		{"comment with a blank line in a liquid tag", "{%- liquid\n  comment\n\n  x\n  endcomment\n  echo 'y'\n-%}", "y"},
		{"doc with whitespace control", "a {% doc %}x{% enddoc -%} b", "a b"},
	}
	for _, tt := range tests {
		for _, mode := range []string{"lax", "strict"} {
			t.Run(tt.name+"/"+mode, func(t *testing.T) {
				assertTemplateResult(t, tt.expected, tt.template, nil, TemplateResultOptions{ErrorMode: mode})
			})
		}
	}
}
