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
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
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
		if got := renderRuby(tpl, map[string]interface{}{"admin": true, "token": "TOKEN-OF-ALICE"}, nil); got != "[TOKEN-OF-ALICE]" {
			t.Fatalf("%s: render A = %q", src, got)
		}
		if got := renderRuby(tpl, map[string]interface{}{"admin": false}, nil); got != "[]" {
			t.Errorf("%s: render B = %q, want %q: the value assigned for A leaks into B", src, got, "[]")
		}
	}
}

type auditRecipientDrop struct {
	Email string
}

// LiquidGet returns the key a template reads.
func (d *auditRecipientDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "email":
		return d.Email, true, nil
	}
	return nil, false, nil
}

// A drop given to one render belongs to that render only: a later render of
// the same template, with a map, nil or a Context, does not see it.
func TestAuditDropDoesNotLeakBetweenRenders(t *testing.T) {
	tpl := auditParse(t, auditEnv(), `[{{ r.email }}]`)
	if got := renderRuby(tpl, map[string]interface{}{"r": &auditRecipientDrop{Email: "alice@example.com"}}, nil); got != "[alice@example.com]" {
		t.Fatalf("render A = %q", got)
	}
	for name, data := range map[string]interface{}{
		"map":     map[string]interface{}{"name": "bob"},
		"nil":     nil,
		"context": newRubyContext(nil),
	} {
		if got := renderRuby(tpl, data, nil); got != "[]" {
			t.Errorf("render B with %s = %q, want %q: the drop of A leaks into B", name, got, "[]")
		}
	}
}

// RenderBang raises the errors of its own render only: a later Render of the
// same template records its errors, and nothing raises out of Render.
func TestAuditRenderBangDoesNotStick(t *testing.T) {
	tpl := auditParse(t, auditEnv(), `[{{ errors.standard_error }}]`)
	data := map[string]interface{}{"errors": NewErrorDrop()}
	raised := func(render func()) (r interface{}) {
		defer func() { r = recover() }()
		render()
		return nil
	}
	if r := raised(func() { renderRubyBang(tpl, data, nil) }); r == nil {
		t.Fatal("RenderBang of a drop that raises did not raise")
	}
	var got string
	if r := raised(func() { got = renderRuby(tpl, data, nil) }); r != nil {
		t.Fatalf("Render after RenderBang raised %v", r)
	}
	if want := "[Liquid error: standard error]"; got != want {
		t.Errorf("Render after RenderBang = %q, want %q", got, want)
	}
}
