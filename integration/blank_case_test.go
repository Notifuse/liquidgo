package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestACaseKeepsItsWhitespaceUnlessItIsBlank checks the results of Shopify
// Liquid 5.11.0 for a case whose bodies are whitespace: as Ruby, the
// whitespace is removed only when every body of the case is blank, the text
// before the first when included.
func TestACaseKeepsItsWhitespaceUnlessItIsBlank(t *testing.T) {
	for _, c := range []struct{ source, want string }{
		{"{% case 1 %}{% when 2 %}a{% else %} {% endcase %}|", " |"},
		{"{% case 1 %}{% when 2 %} {% else %} {% endcase %}|", "|"},
		{"{% case 1 %}x{% when 2 %} {% else %} {% endcase %}|", " |"},
		{"{% case 1 %}{% when 1 %} {% endcase %}|", "|"},
		{"{% case 1 %}{% when 1 %} {% when 2 %}b{% endcase %}|", " |"},
		{"{% case 1 %}{% when 1 %} {% assign q = 1 %}{% endcase %}|", "|"},
		{`{% case 1 %}{% when 1 %}{% if 1 < "a" %}{% endif %}{% endcase %}|`, "|"},
		{`{% case 1 %}{% when 1 %}{% if 1 < "a" %}{% endif %}z{% endcase %}|`, "z|"},
		{"{% case 1 %}{% when 2 %}a{% else %} {% endcase %}|{% if false %}a{% else %} {% endif %}|", " | |"},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			r := renderWith(newEnv(mode, limits{}), c.source, nil)
			if r.parseErr != nil || r.output != c.want {
				t.Errorf("%s: %q gives %q (%v), want %q", mode, c.source, r.output, r.parseErr, c.want)
			}
		}
	}
}
