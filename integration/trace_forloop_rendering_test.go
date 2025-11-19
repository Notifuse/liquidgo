package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

func TestTraceForloopRendering(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Test direct variable lookup in context
	ctx := liquid.BuildContext(liquid.ContextConfig{Environment: env})

	// Create a forloop drop and add it to context
	drop := liquid.NewForloopDrop("test", 3, nil)
	ctx.Set("forloop", drop)

	// Test 1: Direct lookup
	forloopVar := ctx.FindVariable("forloop", false)
	t.Logf("FindVariable('forloop') = %#v (type: %T)", forloopVar, forloopVar)

	// Test 2: Evaluate a VariableLookup
	vl := liquid.VariableLookupParse("forloop.last", liquid.NewStringScanner(""), nil)
	result := vl.Evaluate(ctx)
	t.Logf("VariableLookup.Evaluate('forloop.last') = %#v (type: %T)", result, result)

	// Test 3: ToS on the result
	resultStr := liquid.ToS(result, nil)
	t.Logf("ToS(result) = %q", resultStr)

	// Test 4: Full template rendering
	template := `{{ forloop.last }}`
	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	err := tmpl.Parse(template, nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	vars := map[string]interface{}{"forloop": drop}
	got := tmpl.Render(vars, nil)
	t.Logf("Template render result = %q", got)

	if got == "" {
		t.Error("Template rendered empty string instead of 'false'")
	}
}
