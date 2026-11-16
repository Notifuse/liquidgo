package integration

import (
	gocontext "context"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func strictTemplate(t *testing.T, source string) *liquid.Template {
	t.Helper()
	return strictTemplateIn(t, source, "")
}

// strictTemplateIn parses source in errorMode, or the default mode.
func strictTemplateIn(t *testing.T, source, errorMode string) *liquid.Template {
	t.Helper()
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	if errorMode != "" {
		env.SetErrorMode(errorMode)
	}
	tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tmpl
}

var undefinedVariablesData = map[string]interface{}{
	"x": 33,
	"z": map[string]interface{}{"a": 32, "c": map[string]interface{}{"e": 31}},
}

// Ported from: test_undefined_variables (test/integration/template_test.rb)
func TestTemplate_UndefinedVariables(t *testing.T) {
	tmpl := strictTemplate(t, "{{x}} {{y}} {{z.a}} {{z.b}} {{z.c.d}}")
	result := renderRuby(tmpl, undefinedVariablesData, &rubyOptions{StrictVariables: true})

	if result != "33  32  " {
		t.Errorf("result %q, want %q", result, "33  32  ")
	}
	errs := rubyErrors(tmpl)
	if len(errs) != 3 {
		t.Fatalf("errors %v, want 3", errs)
	}
	for i, want := range []string{
		"Liquid error: undefined variable y",
		"Liquid error: undefined variable b",
		"Liquid error: undefined variable d",
	} {
		if _, ok := errs[i].(*liquid.UndefinedVariable); !ok {
			t.Errorf("error %d is a %T, want an UndefinedVariable", i, errs[i])
		}
		if errs[i].Error() != want {
			t.Errorf("error %d is %q, want %q", i, errs[i].Error(), want)
		}
	}
}

// Ported from: test_nil_value_does_not_raise
func TestTemplate_NilValueDoesNotRaise(t *testing.T) {
	tmpl := strictTemplateIn(t, "some{{x}}thing", "strict")
	result := renderRubyBang(tmpl, map[string]interface{}{"x": nil}, &rubyOptions{StrictVariables: true})

	if n := len(rubyErrors(tmpl)); n != 0 {
		t.Errorf("errors %v, want none", rubyErrors(tmpl))
	}
	if result != "something" {
		t.Errorf("result %q, want %q", result, "something")
	}
}

// Ported from: test_undefined_variables_raise
func TestTemplate_UndefinedVariablesRaise(t *testing.T) {
	tmpl := strictTemplate(t, "{{x}} {{y}} {{z.a}} {{z.b}} {{z.c.d}}")
	defer func() {
		if _, ok := recover().(*liquid.UndefinedVariable); !ok {
			t.Error("render! did not raise an UndefinedVariable")
		}
	}()
	renderRubyBang(tmpl, undefinedVariablesData, &rubyOptions{StrictVariables: true})
}

// Render with strict variables records an undefined name of the first
// level, as Ruby's find_variable: it was recorded only below the first level.
func TestRenderWithStrictVariablesRecordsAnUndefinedName(t *testing.T) {
	tmpl := strictTemplate(t, "a{{ nope }}b")
	res, err := tmpl.Render(gocontext.Background(), nil, liquid.WithStrictVariables())
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "ab" {
		t.Errorf("output %q, want %q", res.Output, "ab")
	}
	if len(res.Errors) != 1 || res.Errors[0].Error() != "Liquid error: undefined variable nope" {
		t.Errorf("errors %v, want the undefined variable nope", res.Errors)
	}
	if res, _ := tmpl.Render(gocontext.Background(), nil); len(res.Errors) != 0 {
		t.Errorf("without strict variables: errors %v, want none", res.Errors)
	}
}
