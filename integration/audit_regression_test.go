package integration

// Regression tests from the architecture audit of 2026-11-15, re-verified on
// v5.11.4. Each one reproduced a defect before its fix.
//
// Points of the audit that are not here:
//   - quadratic output and ranges built in memory: fixed in v5.11.1, covered
//     by the hostile-template tests;
//   - filter errors rendered empty (divided_by: 0): a deliberate divergence
//     for Notifuse, see integration/testdata/differential/differential_divergences.json
//     (render-values-11).

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func auditEnv() *liquid.Environment {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	return env
}

func auditParse(t *testing.T, env *liquid.Environment, src string, options ...liquid.TemplateOptions) *liquid.Template {
	t.Helper()
	opts := liquid.TemplateOptions{Environment: env}
	if len(options) > 0 {
		opts = options[0]
		opts.Environment = env
	}
	tpl, err := liquid.ParseTemplate(src, &opts)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return tpl
}

// A value assigned while rendering for one recipient never reaches the next
// render of the same template, by assign or by capture.
func TestAuditAssignDoesNotLeakBetweenRenders(t *testing.T) {
	for _, src := range []string{
		`{% if admin %}{% assign secret = token %}{% endif %}[{{ secret }}]`,
		`{% if admin %}{% capture secret %}{{ token }}{% endcapture %}{% endif %}[{{ secret }}]`,
	} {
		tpl := auditParse(t, auditEnv(), src)
		if got := tpl.Render(map[string]interface{}{"admin": true, "token": "TOKEN-OF-ALICE"}, nil); got != "[TOKEN-OF-ALICE]" {
			t.Fatalf("%s: render A = %q", src, got)
		}
		if got := tpl.Render(map[string]interface{}{"admin": false}, nil); got != "[]" {
			t.Errorf("%s: render B = %q, want %q: the value assigned for A leaks into B", src, got, "[]")
		}
	}
}

// With PersistInstanceAssigns, the template keeps Ruby's behaviour.
func TestAuditPersistInstanceAssignsKeepsRubyBehaviour(t *testing.T) {
	src := `{% if admin %}{% assign secret = token %}{% endif %}[{{ secret }}]`
	tpl := auditParse(t, auditEnv(), src, liquid.TemplateOptions{PersistInstanceAssigns: true})
	tpl.Render(map[string]interface{}{"admin": true, "token": "T"}, nil)
	if got := tpl.Render(map[string]interface{}{"admin": false}, nil); got != "[T]" {
		t.Errorf("render B = %q, want %q with PersistInstanceAssigns", got, "[T]")
	}
}
