package liquid

import (
	gocontext "context"
	"testing"
)

// A panic out of a render that is a Liquid error ends it with that error;
// any other value is a PanicError with the stack.
func TestRenderErrorKeepsLiquidErrorsAndWrapsTheRest(t *testing.T) {
	base := &Error{Message: "base"}
	if got := renderError(base); got != base {
		t.Errorf("an *Error gives %v, want itself", got)
	}
	limit := NewMemoryError("Memory limits exceeded")
	if got := renderError(limit); got != limit {
		t.Errorf("a MemoryError gives %v, want itself", got)
	}
	p, ok := renderError("boom").(*PanicError)
	if !ok || p.Value != "boom" || len(p.Stack) == 0 {
		t.Errorf("a string gives %#v, want a PanicError with its stack", p)
	}
}

// Template.RenderToOutput renders a template as a partial of the render the
// context belongs to: an unparsed template writes nothing, and the context
// takes the template's name when it has none.
func TestTemplateRenderToOutputAsAPartial(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	buf := NewOutputBuffer()
	NewTemplate(&TemplateOptions{Environment: env}).RenderToOutput(NewContext(), buf)
	if buf.String() != "" {
		t.Errorf("an unparsed template wrote %q", buf.String())
	}
	tmpl, err := ParseTemplate("x", &TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	tmpl.SetName("partial")
	ctx := NewContext()
	tmpl.RenderToOutput(ctx, buf)
	if buf.String() != "x" || ctx.TemplateName() != "partial" {
		t.Errorf("output %q, template name %q, want %q and %q", buf.String(), ctx.TemplateName(), "x", "partial")
	}
	named := NewContext()
	named.SetTemplateName("outer")
	tmpl.RenderToOutput(named, NewOutputBuffer())
	if named.TemplateName() != "outer" {
		t.Errorf("the context's own name became %q", named.TemplateName())
	}
}

// Template.Environment is the environment the template was parsed in.
func TestTemplateEnvironment(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	tmpl, err := ParseTemplate("x", &TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Environment() != env {
		t.Error("Environment is not the environment of the parse")
	}
	if res, _ := tmpl.Render(gocontext.Background(), nil); res.Output != "x" {
		t.Errorf("output %q", res.Output)
	}
}

// In warn mode, a constructor that panics with an error gives a warning,
// and the tag stays as text, as a constructor that returns the error.
func TestATagConstructorThatPanicsInWarnMode(t *testing.T) {
	envTags := map[string]TagConstructor{}
	envTags["boom"] = func(string, string, ParseContextInterface) (Node, error) {
		panic(NewSyntaxError("boom constructor"))
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags, ErrorMode: "warn"})
	tmpl, err := ParseTemplate("a{% boom %}b", &TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("warn mode: %v", err)
	}
	if ws := tmpl.Warnings(); len(ws) != 1 || ws[0].Error() == "" {
		t.Errorf("warnings %v, want the constructor's error", ws)
	}
	if res, _ := tmpl.Render(gocontext.Background(), nil); res.Output != "a{% boom %}b" {
		t.Errorf("output %q, want the tag as text", res.Output)
	}
}
