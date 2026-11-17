package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestAConditionIsLexedWholeInStrictMode checks the results of Shopify
// Liquid 5.11.0 in strict mode for a character the lexer refuses in the
// markup of an if, an unless or a for: a syntax error that names the whole
// markup. Lax mode renders them.
func TestAConditionIsLexedWholeInStrictMode(t *testing.T) {
	for _, c := range []struct{ source, strict, lax string }{
		{"{% unless 'a' != nothing['k] %}b{% endunless %}", `Liquid syntax error: Unexpected character ' in "'a' != nothing['k]"`, ""},
		{"{% if nothing['k] %}b{% endif %}", `Liquid syntax error: Unexpected character ' in "nothing['k]"`, ""},
		{"{% if a['k] == 1 %}b{% endif %}", `Liquid syntax error: Unexpected character ' in "a['k] == 1"`, ""},
		{"{% for i in a['k] %}x{% endfor %}", `Liquid syntax error: Unexpected character ' in "i in a['k]"`, ""},
		{"{% if a == 1 and b contains 'x' or c <> 2 %}y{% endif %}", "", ""},
		{"{% for i in (1..2) reversed limit: 1 offset: 1 %}{{ i }}{% endfor %}", "2", "2"},
	} {
		for _, m := range []struct{ mode, want string }{{"strict", c.strict}, {"lax", c.lax}} {
			r := renderWith(newEnv(liquid.ErrorMode(m.mode), limits{}), c.source, nil)
			got := r.output
			if r.parseErr != nil {
				got = r.parseErr.Error()
			}
			if got != m.want {
				t.Errorf("%s: %q gives %q, want %q", m.mode, c.source, got, m.want)
			}
		}
	}
}
