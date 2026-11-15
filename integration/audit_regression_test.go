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

type auditRecipientDrop struct {
	*liquid.Drop
	Email string
}

// Data given as a drop belongs to its render only: a later render of the
// same template, with a map, nil or a Context, does not see it.
func TestAuditDropDoesNotLeakBetweenRenders(t *testing.T) {
	for _, persist := range []bool{false, true} {
		tpl := auditParse(t, auditEnv(), `[{{ email }}{{ __drop__.email }}]`, liquid.TemplateOptions{PersistInstanceAssigns: persist})
		if got := tpl.Render(&auditRecipientDrop{Drop: liquid.NewDrop(), Email: "alice@example.com"}, nil); got != "[alice@example.comalice@example.com]" {
			t.Fatalf("persist %v: render A = %q", persist, got)
		}
		for name, data := range map[string]interface{}{
			"map":     map[string]interface{}{"name": "bob"},
			"nil":     nil,
			"context": liquid.NewContext(),
		} {
			if got := tpl.Render(data, nil); got != "[]" {
				t.Errorf("persist %v: render B with %s = %q, want %q: the drop of A leaks into B", persist, name, got, "[]")
			}
		}
	}
}

// Turning PersistInstanceAssigns off forgets what earlier renders kept.
func TestAuditTurningPersistenceOffForgetsAssigns(t *testing.T) {
	src := `{% if admin %}{% assign secret = token %}{% endif %}[{{ secret }}]`
	tpl := auditParse(t, auditEnv(), src, liquid.TemplateOptions{PersistInstanceAssigns: true})
	tpl.Render(map[string]interface{}{"admin": true, "token": "T"}, nil)
	tpl.SetPersistInstanceAssigns(false)
	if got := tpl.Render(map[string]interface{}{"admin": false}, nil); got != "[]" {
		t.Errorf("render after SetPersistInstanceAssigns(false) = %q, want %q", got, "[]")
	}
	tpl.SetPersistInstanceAssigns(true)
	if got := tpl.Render(map[string]interface{}{"admin": false}, nil); got != "[]" {
		t.Errorf("render after turning it on again = %q, want %q", got, "[]")
	}
}

// Without PersistInstanceAssigns, a render does not start from instance
// assigns, even ones the caller set through InstanceAssigns.
func TestAuditInstanceAssignsUnusedWithoutPersistence(t *testing.T) {
	tpl := auditParse(t, auditEnv(), `[{{ secret }}]`)
	tpl.InstanceAssigns()["secret"] = "T"
	if got := tpl.Render(map[string]interface{}{}, nil); got != "[]" {
		t.Errorf("render = %q, want %q", got, "[]")
	}
}

// Errors can be read while other renders run.
func TestAuditErrorsDuringConcurrentRenders(t *testing.T) {
	tpl := auditParse(t, auditEnv(), `{{ 1 | divided_by: x }}{{ a.b.c }}`)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			tpl.Render(map[string]interface{}{"x": 0}, nil)
		}
	}()
	for i := 0; i < 200; i++ {
		_ = tpl.Errors()
	}
	<-done
}

// The resource limits and the profiler of a template can be read while other
// renders of it run (detected under -race). Each is read alone, so that the
// lock of one does not order the reads of the other.
func TestAuditTemplateStateDuringConcurrentRenders(t *testing.T) {
	reads := map[string]func(tpl *liquid.Template){
		"limits": func(tpl *liquid.Template) {
			if rl := tpl.ResourceLimits(); rl != nil {
				_ = rl.Reached()
				_ = rl.RenderScore()
			}
		},
		"profiler": func(tpl *liquid.Template) { _ = tpl.Profiler() },
	}
	for name, read := range reads {
		t.Run(name, func(t *testing.T) {
			env := auditEnv()
			env.SetDefaultResourceLimits(map[string]interface{}{"render_length_limit": 50})
			tpl := auditParse(t, env, `{% for i in (1..n) %}x{% endfor %}`, liquid.TemplateOptions{Profile: true})
			done := make(chan struct{})
			go func() {
				defer close(done)
				for i := 0; i < 200; i++ {
					tpl.Render(map[string]interface{}{"n": i % 100}, nil)
				}
			}()
			for i := 0; i < 200; i++ {
				read(tpl)
			}
			<-done
		})
	}
}
