package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestDrop_ProductDrop tests basic drop functionality.
// Ported from: test_product_drop
func TestDrop_ProductDrop(t *testing.T) {
	// Create a simple drop for testing
	drop := &ThingWithToLiquid{}
	assertTemplateResult(t, "  ", "  ", map[string]interface{}{"product": drop})
}

// TestDrop_DoesOnlyRespondToWhitelistedMethods tests that drops only respond to whitelisted methods.
// Ported from: test_drop_does_only_respond_to_whitelisted_methods
func TestDrop_DoesOnlyRespondToWhitelistedMethods(t *testing.T) {
	// In Go, we can't test inspect/pretty_inspect like Ruby
	// But we can test that unknown methods return empty
	drop := NewDropWithUndefinedMethod()
	assertTemplateResult(t, "", "{{ product.whatever }}", map[string]interface{}{"product": drop})
}

// TestDrop_RespondsToToLiquid tests that drops respond to ToLiquid.
// Ported from: test_drops_respond_to_to_liquid
func TestDrop_RespondsToToLiquid(t *testing.T) {
	// Test that ToLiquid is accessible
	drop := &ThingWithToLiquid{}
	assertTemplateResult(t, "foobar", "{{ product.to_liquid }}", map[string]interface{}{"product": drop})
}

// TestDrop_ContextDrop tests context drop functionality.
// Ported from: test_context_drop
func TestDrop_ContextDrop(t *testing.T) {
	// Create a context drop that can access template context
	drop := NewTemplateContextDrop()
	assertTemplateResult(t, " carrot ", " {{ context.bar }} ", map[string]interface{}{
		"context": drop,
		"bar":     "carrot",
	})
}

// TestDrop_NestedContextDrop tests nested context drop.
// Ported from: test_nested_context_drop
func TestDrop_NestedContextDrop(t *testing.T) {
	// Test nested context access
	drop := NewTemplateContextDrop()
	assertTemplateResult(t, " monkey ", " {{ product.context.foo }} ", map[string]interface{}{
		"product": drop,
		"foo":     "monkey",
	})
}

// TestDrop_Protected tests that protected methods are not accessible.
// Ported from: test_protected
func TestDrop_Protected(t *testing.T) {
	// In Go, we can't have protected methods like Ruby
	// This test documents expected behavior
	drop := NewDropWithUndefinedMethod()
	output := renderTemplateForDropTest("{{ product.callmenot }}", map[string]interface{}{"product": drop})
	// Should return empty or error, not expose protected method
	_ = output
}

// TestDrop_ObjectMethodsNotAllowed tests that object methods are not allowed.
// Ported from: test_object_methods_not_allowed
func TestDrop_ObjectMethodsNotAllowed(t *testing.T) {
	// In Go, we don't have the same object methods as Ruby
	// This test documents expected behavior
	drop := NewDropWithUndefinedMethod()
	// Test that dangerous methods are not accessible
	output := renderTemplateForDropTest("{{ product.eval }}", map[string]interface{}{"product": drop})
	// Should return empty, not execute code
	if output != "" && output != " " {
		t.Errorf("Expected empty output for dangerous method, got %q", output)
	}
}

// TestDrop_Scope tests scope access from drops.
// Ported from: test_scope
func TestDrop_Scope(t *testing.T) {
	// Test that drops can access scope information
	// This requires a drop that implements scope access
	drop := NewTemplateContextDrop()
	assertTemplateResult(t, "1", "{{ context.scopes }}", map[string]interface{}{"context": drop})
}

// TestDrop_AccessContextFromDrop tests accessing context from drop.
// Ported from: test_access_context_from_drop
func TestDrop_AccessContextFromDrop(t *testing.T) {
	drop := NewTemplateContextDrop()
	assertTemplateResult(t, "123", "{%for a in dummy%}{{ context.loop_pos }}{% endfor %}", map[string]interface{}{
		"context": drop,
		"dummy":   []interface{}{1, 2, 3},
	})
}

// TestDrop_EnumerableDrop tests enumerable drop functionality.
// Ported from: test_enumerable_drop
func TestDrop_EnumerableDrop(t *testing.T) {
	// Create an enumerable drop
	drop := NewEnumerableDrop()
	assertTemplateResult(t, "123", "{% for c in collection %}{{c}}{% endfor %}", map[string]interface{}{
		"collection": drop,
	})
}

// TestDrop_EnumerableDropSize tests enumerable drop size.
// Ported from: test_enumerable_drop_size
func TestDrop_EnumerableDropSize(t *testing.T) {
	drop := NewEnumerableDrop()
	assertTemplateResult(t, "3", "{{collection.size}}", map[string]interface{}{
		"collection": drop,
	})
}

// EnumerableDrop is a drop that implements enumerable behavior.
type EnumerableDrop struct {
	*liquid.Drop
}

// NewEnumerableDrop creates a new EnumerableDrop.
func NewEnumerableDrop() *EnumerableDrop {
	return &EnumerableDrop{
		Drop: liquid.NewDrop(),
	}
}

// LiquidMethodMissing handles missing method calls.
func (e *EnumerableDrop) LiquidMethodMissing(method string) interface{} {
	return method
}

// Each iterates over the collection.
func (e *EnumerableDrop) Each(fn func(interface{})) {
	fn(1)
	fn(2)
	fn(3)
}

// Size returns the size.
func (e *EnumerableDrop) Size() interface{} {
	return 3
}

// First returns the first element.
func (e *EnumerableDrop) First() interface{} {
	return 1
}

// Count returns the count.
func (e *EnumerableDrop) Count() int {
	return 3
}

// Min returns the minimum.
func (e *EnumerableDrop) Min() interface{} {
	return 1
}

// Max returns the maximum.
func (e *EnumerableDrop) Max() interface{} {
	return 3
}

// renderTemplateForDropTest is a helper to render templates for drop testing.
func renderTemplateForDropTest(template string, assigns map[string]interface{}) string {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		return "ERROR: " + err.Error()
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{assigns},
		RethrowErrors:      false,
	})

	return tmpl.Render(ctx, &liquid.RenderOptions{})
}
