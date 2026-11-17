package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestALoneQuoteParsesAsRuby checks the results of Shopify Liquid 5.11.0 for
// a lone quote: an empty string in lax mode, in an output, an assign or a
// filter argument, Ruby's syntax error in a condition or a tag, and a syntax
// error everywhere in strict mode. In strict mode, Ruby ends the message of
// cycle and case with ` in "'"`, which liquidgo leaves out.
func TestALoneQuoteParsesAsRuby(t *testing.T) {
	for _, c := range []struct{ source, lax, strict string }{
		{"{{ ' }}", "", `Liquid syntax error: Unexpected character ' in "{{ ' }}"`},
		{`{{ " }}`, "", `Liquid syntax error: Unexpected character " in "{{ " }}"`},
		{"{% assign v = ' %}[{{ v }}]", "[]", `Liquid syntax error: Unexpected character ' in "{{' }}"`},
		{`{{ x | default: " }}`, "", `Liquid syntax error: Unexpected character " in "{{ x | default: " }}"`},
		{"{% echo ' %}", "", `Liquid syntax error: Unexpected character ' in "{{' }}"`},
		{"{% if ' %}y{% else %}n{% endif %}", "Liquid syntax error: Syntax Error in tag 'if' - Valid syntax: if [expression]", `Liquid syntax error: Unexpected character ' in "'"`},
		{"{% for i in ' %}x{% endfor %}", "Liquid syntax error: Syntax Error in 'for loop' - Valid syntax: for [item] in [collection]", `Liquid syntax error: Unexpected character ' in "i in '"`},
		{"{% cycle ' %}", "Liquid syntax error: Syntax Error in 'cycle' - Valid syntax: cycle [name :] var [, var2, var3 ...]", "Liquid syntax error: Syntax Error in 'cycle' - Valid syntax: cycle [name :] var [, var2, var3 ...]"},
		{"{% case ' %}{% when ' %}w{% endcase %}", "Liquid syntax error: Syntax Error in 'case' - Valid syntax: case [condition]", "Liquid syntax error: Syntax Error in 'case' - Valid syntax: case [condition]"},
	} {
		for _, m := range []struct{ mode, want string }{{"lax", c.lax}, {"strict", c.strict}} {
			r := renderWith(newEnv(liquid.ErrorMode(m.mode), limits{}), c.source, nil)
			got := r.output
			if r.parseErr != nil {
				got = r.parseErr.Error()
			}
			if got != m.want || len(r.errors) > 0 {
				t.Errorf("%s: %q gives %q (%v), want %q", m.mode, c.source, got, r.errors, m.want)
			}
		}
	}
}
