package integration

import (
	gocontext "context"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// leakSecret is a value of the data that no error message may show.
const leakSecret = "S3CR3T-mail@example.com"

// knownLeaks are the templates whose error shows a value of the data, as in
// Ruby: the value is the name the template looks up (a key, a partial), and
// the template could write it anyway. A listed template that no longer
// shows it fails the test until its line is removed.
var knownLeaks = map[string]string{
	"{{ h[s] }}":      "strict variables name the undefined key, as Ruby's undefined variable",
	"{{ a[s] }}":      "strict variables name the key that is no index, as Ruby's undefined variable",
	"{% include s %}": "the partial is named in its error, as Ruby's No such template",
}

// No error of a parse or a render, with or without strict variables and
// filters, shows a value of the data, but for knownLeaks.
func TestErrorsShowNoValueOfTheData(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: liquid.NewLocalFileSystem(t.TempDir(), "")})
	data := map[string]interface{}{
		"s": leakSecret, "n": "12" + leakSecret,
		"h": map[string]interface{}{"k": leakSecret}, "a": []interface{}{leakSecret, map[string]interface{}{"k": leakSecret}},
	}
	leaked := map[string]bool{}
	for _, src := range []string{
		"{{ s | nofilter }}", "{{ s | divided_by: 0 }}", "{{ s | divided_by: s }}", "{{ s | modulo: s }}",
		"{{ s | date: s }}", "{{ s | slice: s }}", "{{ s | truncate: s }}", "{{ s | truncatewords: s }}",
		"{{ s | round: s }}", "{{ s | plus: s }}", "{{ s | times: s }}", "{{ s | at_least: s }}", "{{ s | abs }}",
		"{{ n | ceil }}", "{{ s | sum }}", "{{ s | base64_decode }}", "{{ s | url_decode }}",
		"{{ a | map: s }}", "{{ a | where: s, s }}", "{{ a | sort: s }}", "{{ a | sort_natural: s }}",
		"{{ a | group_by: s }}", "{{ h | sort: s }}", "{{ s | split: s | join: s }}",
		"{% include s %}", "{% render s %}", "{% for i in s %}{% endfor %}", "{% cycle s, s %}",
		"{% tablerow i in s cols: s %}{% endtablerow %}", "{% if s > 1 %}{% endif %}", "{% if s contains 1 %}{% endif %}",
		"{{ s.nope }}", "{{ h[s] }}", "{{ a[s] }}", "{% assign x = s | nofilter %}{{ x }}",
	} {
		for _, strict := range []bool{false, true} {
			var msgs []string
			tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env, LineNumbers: true})
			if err != nil {
				msgs = append(msgs, err.Error())
			} else {
				var opts []liquid.RenderOption
				if strict {
					opts = append(opts, liquid.WithStrictVariables(), liquid.WithStrictFilters())
				}
				res, err := tmpl.Render(gocontext.Background(), data, opts...)
				for _, e := range res.Errors {
					msgs = append(msgs, e.Error())
				}
				if err != nil {
					msgs = append(msgs, err.Error())
				}
			}
			for _, m := range msgs {
				if !strings.Contains(m, leakSecret) {
					continue
				}
				leaked[src] = true
				if _, known := knownLeaks[src]; !known {
					t.Errorf("%s (strict %v): the error %q shows a value of the data", src, strict, m)
				}
			}
		}
	}
	for src, reason := range knownLeaks {
		if !leaked[src] {
			t.Errorf("knownLeaks lists %s (%s), which no longer shows the value: remove it", src, reason)
		}
	}
}
