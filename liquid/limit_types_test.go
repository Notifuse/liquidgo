package liquid

import (
	"strings"
	"testing"
)

// TestTheEnvironmentLimitsApply checks that the limits of the environment
// apply to the renders of its templates (the template of the former test of
// limits given as any integer type in a map, which the typed config replaces).
func TestTheEnvironmentLimitsApply(t *testing.T) {
	five := 5
	env := mustEnv(EnvironmentConfig{Limits: ResourceLimitsConfig{RenderLengthLimit: &five}})
	tmpl, err := ParseTemplate("0123456789", &TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	if out := renderRuby(tmpl, nil, nil); !strings.Contains(out, "Memory limits exceeded") {
		t.Errorf("render_length_limit 5: got %q, want the limit error", out)
	}
}

type dropWithArguments struct{}

func (d *dropWithArguments) AddDate(years, months, days int) string { return "called" }

func (d *dropWithArguments) Name() string { return "drop" }

func (d *dropWithArguments) LiquidGet(_ *Context, key string) (interface{}, bool, error) {
	if key == "name" {
		return d.Name(), true, nil
	}
	return nil, false, nil
}

// TestDropMethodsThatNeedArgumentsAreNotCalled checks that a lookup does not
// call a method that needs arguments, which would panic, and still calls one
// that takes none.
func TestDropMethodsThatNeedArgumentsAreNotCalled(t *testing.T) {
	tmpl, err := ParseTemplate("{{ d.add_date }}|{{ d.AddDate }}|{{ d.name }}", nil)
	if err != nil {
		t.Fatal(err)
	}
	out := renderRuby(tmpl, map[string]interface{}{"d": &dropWithArguments{}}, nil)
	if out != "||drop" {
		t.Errorf("got %q, want %q (errors %v)", out, "||drop", rubyErrors(tmpl))
	}
}

// TestAHashIndexedByNilFindsNothing checks that a key that evaluates to nil
// finds nothing in a hash, as in Ruby, instead of panicking.
func TestAHashIndexedByNilFindsNothing(t *testing.T) {
	tmpl, err := ParseTemplate("a{{ h[nothing] }}b{{ h[] }}c", nil)
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{"h": map[string]interface{}{"k": "v"}}
	if out := renderRuby(tmpl, data, nil); out != "abc" || len(rubyErrors(tmpl)) > 0 {
		t.Errorf("got %q (errors %v), want %q", out, rubyErrors(tmpl), "abc")
	}
}
