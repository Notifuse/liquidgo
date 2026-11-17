package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestAnUnknownTagInABlockIsASyntaxError checks that a tag the environment
// does not know raises Ruby's syntax error in every block and in both
// modes, instead of being dropped with the text around it. The first one is
// a URL-encoded brace in a link, {%22k%22}, which reads as a tag.
func TestAnUnknownTagInABlockIsASyntaxError(t *testing.T) {
	for _, c := range []struct{ source, want string }{
		{`{% for i in (1..1) %}<a href="https://x/?q={%22k%22}">l</a>{% endfor %}`, "Liquid syntax error: Unknown tag '22k'"},
		{"{% for i in (1..1) %}{% else %}{% bogus %}{% endfor %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% capture c %}{% bogus %}{% endcapture %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% case 1 %}{% when 1 %}{% bogus %}{% endcase %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% case 1 %}{% when 2 %}{% else %}{% bogus %}{% endcase %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% tablerow i in (1..1) %}{% bogus %}{% endtablerow %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% ifchanged %}{% bogus %}{% endifchanged %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% if false %}a{% elsif true %}{% bogus %}{% endif %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% unless false %}{% bogus %}{% endunless %}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% liquid\ncapture c\nbogus\nendcapture\n%}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% liquid\ncase 1\nwhen 1\nbogus\nendcase\n%}", "Liquid syntax error: Unknown tag 'bogus'"},
		{"{% case 1 %}{% when %}a{% endcase %}", "Liquid syntax error: Syntax Error in tag 'case' - Valid when condition: {% when [condition] [or condition2...] %}"},
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			_, err := liquid.ParseTemplate(c.source, &liquid.TemplateOptions{Environment: newEnv(mode, limits{})})
			if err == nil || err.Error() != c.want {
				t.Errorf("%s: %q gives the error %v, want %q", mode, c.source, err, c.want)
			}
		}
	}
}
