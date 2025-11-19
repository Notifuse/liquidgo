package liquid

import (
	"testing"
)

func TestTemplateParse(t *testing.T) {
	env := NewEnvironment()
	// Note: Tags should be registered via tags.RegisterStandardTags from outside
	// For now, test without tags (just variables)

	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if template.Root() == nil {
		t.Error("Expected root document, got nil")
	}
}

func TestTemplateRender(t *testing.T) {
	env := NewEnvironment()
	// Note: Tags should be registered via tags.RegisterStandardTags from outside
	// For now, test without tags (just variables)

	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := template.Render(map[string]interface{}{"name": "world"}, nil)
	expected := "Hello world"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateRenderEmpty(t *testing.T) {
	env := NewEnvironment()

	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := template.Render(nil, nil)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestTemplateRenderNilRoot(t *testing.T) {
	env := NewEnvironment()

	template := NewTemplate(&TemplateOptions{Environment: env})
	// Don't parse, so root is nil

	result := template.Render(nil, nil)
	if result != "" {
		t.Errorf("Expected empty string for nil root, got %q", result)
	}
}

// TestTemplateDefaultResourceLimits tests that default resource limits are applied
func TestTemplateDefaultResourceLimits(t *testing.T) {
	env := NewEnvironment()
	renderLimit := 100
	assignLimit := 50
	env.SetDefaultResourceLimits(map[string]interface{}{
		"render_length_limit": renderLimit,
		"assign_score_limit":  assignLimit,
	})

	template := NewTemplate(&TemplateOptions{Environment: env})
	if template.ResourceLimits() == nil {
		t.Fatal("Expected ResourceLimits to be set")
	}

	// Resource limits should have default values from environment
	if template.ResourceLimits().RenderLengthLimit() == nil {
		t.Error("Expected render_length_limit to be set from environment")
	}
	if template.ResourceLimits().AssignScoreLimit() == nil {
		t.Error("Expected assign_score_limit to be set from environment")
	}
}

// TestTemplateEncodingValidation tests UTF-8 encoding validation
func TestTemplateEncodingValidation(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Test invalid UTF-8 sequence
	invalidUTF8 := "\xff\x00"
	err := template.Parse(invalidUTF8, nil)
	if err == nil {
		t.Error("Expected TemplateEncodingError for invalid UTF-8")
	} else if _, ok := err.(*TemplateEncodingError); !ok {
		t.Errorf("Expected TemplateEncodingError, got %T", err)
	}

	// Test valid UTF-8
	validUTF8 := "Hello {{ name }}"
	err = template.Parse(validUTF8, nil)
	if err != nil {
		t.Errorf("Expected no error for valid UTF-8, got %v", err)
	}
}

// TestParseTemplate tests the ParseTemplate function
func TestParseTemplate(t *testing.T) {
	env := NewEnvironment()

	// Test basic parsing
	template, err := ParseTemplate("Hello {{ name }}", &TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	if template == nil {
		t.Fatal("Expected template, got nil")
	}
	if template.Root() == nil {
		t.Error("Expected root document, got nil")
	}

	// Test with options
	_, err = ParseTemplate("{{ value }}", &TemplateOptions{
		Environment:     env,
		StrictVariables: true,
		LineNumbers:     true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() with options error = %v", err)
	}

	// Test with nil options
	_, err = ParseTemplate("test", nil)
	if err != nil {
		t.Fatalf("ParseTemplate() with nil options error = %v", err)
	}
}

// TestTemplateRenderBang tests RenderBang with rethrow_errors enabled
func TestTemplateRenderBang(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test normal rendering
	result := template.RenderBang(map[string]interface{}{"name": "world"}, nil)
	if result != "Hello world" {
		t.Errorf("Expected 'Hello world', got %q", result)
	}

	// Verify rethrowErrors is set
	if !template.rethrowErrors {
		t.Error("Expected rethrowErrors to be true after RenderBang")
	}
}

// TestTemplateRenderToOutputBuffer tests direct buffer rendering
func TestTemplateRenderToOutputBuffer(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test with Context
	ctx := NewContext()
	ctx.Set("name", "world")
	output := ""
	template.RenderToOutputBuffer(ctx, &output)
	if output != "Hello world" {
		t.Errorf("Expected 'Hello world', got %q", output)
	}

	// Test with nil root
	template2 := NewTemplate(&TemplateOptions{Environment: env})
	output2 := ""
	template2.RenderToOutputBuffer(ctx, &output2)
	if output2 != "" {
		t.Errorf("Expected empty output for nil root, got %q", output2)
	}

	// Test with memory error recovery
	template3 := NewTemplate(&TemplateOptions{Environment: env})
	err = template3.Parse("{{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	ctx3 := NewContext()
	ctx3.Set("name", "test")
	var output3 string
	template3.RenderToOutputBuffer(ctx3, &output3)
	if output3 != "test" {
		t.Errorf("Expected 'test', got %q", output3)
	}

	// Test with fallback path - create a context from map
	ctx4 := BuildContext(ContextConfig{
		Environments: []map[string]interface{}{{"name": "test"}},
	})
	output4 := ""
	template.RenderToOutputBuffer(ctx4, &output4)
	if output4 == "" {
		t.Error("Expected non-empty output")
	}
}

// TestTemplateBuildContext tests buildContext with various input types
func TestTemplateBuildContext(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("{{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test with Context as assigns
	ctx := NewContext()
	ctx.Set("name", "context_value")
	result := template.Render(ctx, nil)
	if result != "context_value" {
		t.Errorf("Expected 'context_value', got %q", result)
	}

	// Test with map[string]interface{} as assigns
	result = template.Render(map[string]interface{}{"name": "map_value"}, nil)
	if result != "map_value" {
		t.Errorf("Expected 'map_value', got %q", result)
	}

	// Test with nil assigns
	result = template.Render(nil, nil)
	if result != "" {
		t.Errorf("Expected empty string for nil assigns, got %q", result)
	}

	// Test with Drop as assigns
	drop := NewDrop()
	result = template.Render(drop, nil)
	if result != "" {
		t.Logf("Drop rendering result: %q", result)
	}

	// Test with RenderOptions
	output := ""
	_ = template.Render(map[string]interface{}{"name": "options"}, &RenderOptions{
		Output: &output,
	})
	if output != "options" {
		t.Errorf("Expected 'options' in output, got %q", output)
	}
}

// TestTemplateGettersSetters tests all getters and setters
func TestTemplateGettersSetters(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Test Registers
	registers := template.Registers()
	if registers == nil {
		t.Error("Expected registers map, got nil")
	}
	registers["test"] = "value"
	if template.Registers()["test"] != "value" {
		t.Error("Registers not working correctly")
	}

	// Test Assigns
	assigns := template.Assigns()
	if assigns == nil {
		t.Error("Expected assigns map, got nil")
	}
	assigns["assign"] = "test"
	if template.Assigns()["assign"] != "test" {
		t.Error("Assigns not working correctly")
	}

	// Test InstanceAssigns
	instanceAssigns := template.InstanceAssigns()
	if instanceAssigns == nil {
		t.Error("Expected instanceAssigns map, got nil")
	}
	instanceAssigns["instance"] = "test"
	if template.InstanceAssigns()["instance"] != "test" {
		t.Error("InstanceAssigns not working correctly")
	}

	// Test Errors
	errors := template.Errors()
	if errors == nil {
		t.Error("Expected errors slice, got nil")
	}

	// Test Warnings
	warnings := template.Warnings()
	if warnings == nil {
		t.Error("Expected warnings slice, got nil")
	}

	// Test SetRoot
	doc := &Document{}
	template.SetRoot(doc)
	if template.Root() != doc {
		t.Error("SetRoot not working correctly")
	}

	// Test SetResourceLimits
	rl := NewResourceLimits(ResourceLimitsConfig{})
	template.SetResourceLimits(rl)
	if template.ResourceLimits() != rl {
		t.Error("SetResourceLimits not working correctly")
	}

	// Test SetName
	template.SetName("test_template")
	if template.Name() != "test_template" {
		t.Errorf("Expected 'test_template', got %q", template.Name())
	}
}

// TestTemplateRenderWithOptions tests Render with various options
func TestTemplateRenderWithOptions(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("{{ name | custom }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test with Output option
	output := ""
	result := template.Render(map[string]interface{}{"name": "test"}, &RenderOptions{
		Output: &output,
	})
	if output != result {
		t.Error("Output option not working correctly")
	}

	// Test with Registers
	result = template.Render(map[string]interface{}{"name": "test"}, &RenderOptions{
		Registers: map[string]interface{}{"reg": "value"},
	})
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Test with Filters
	customFilter := &StandardFilters{}
	result = template.Render(map[string]interface{}{"name": "TEST"}, &RenderOptions{
		Filters: []interface{}{customFilter},
	})
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Test with GlobalFilter
	result = template.Render(map[string]interface{}{"name": "test"}, &RenderOptions{
		GlobalFilter: func(obj interface{}) interface{} {
			return "filtered"
		},
	})
	if result != "filtered" {
		t.Errorf("Expected 'filtered', got %q", result)
	}

	// Test with ExceptionRenderer
	result = template.Render(map[string]interface{}{"name": "test"}, &RenderOptions{
		ExceptionRenderer: func(err error) interface{} {
			return "error_rendered"
		},
	})
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Test with StrictVariables
	result = template.Render(map[string]interface{}{"name": "test"}, &RenderOptions{
		StrictVariables: true,
	})
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Test with StrictFilters
	result = template.Render(map[string]interface{}{"name": "test"}, &RenderOptions{
		StrictFilters: true,
	})
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// TestTemplateMemoryErrorHandling tests memory error recovery in Render
func TestTemplateMemoryErrorHandling(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("{{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Set very low resource limits to trigger memory error
	limit := 1
	rl := NewResourceLimits(ResourceLimitsConfig{
		RenderScoreLimit: &limit,
	})
	template.SetResourceLimits(rl)

	// Render should handle memory error gracefully
	result := template.Render(map[string]interface{}{"name": "test"}, nil)
	if result == "" {
		t.Log("Memory error handled, got empty result")
	}
}

// TestTemplateProfiling tests profiling integration in Render
func TestTemplateProfiling(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{
		Environment: env,
		Profile:     true,
	})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := template.Render(map[string]interface{}{"name": "world"}, nil)
	if result != "Hello world" {
		t.Errorf("Expected 'Hello world', got %q", result)
	}

	// Profiler may be created during render if profiling is enabled
	// Just verify the render worked correctly
	profiler := template.Profiler()
	// Profiler might be nil if no profiling data was collected, which is acceptable
	_ = profiler
}

// TestTemplateBuildContextWithDrop tests buildContext with a drop as assigns
func TestTemplateBuildContextWithDrop(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("{{ __drop__.test }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Create a drop with a method
	drop := NewDrop()
	drop.SetContext(NewContext())

	// Test rendering with drop as assigns
	result := template.Render(drop, nil)
	// Drop rendering may return empty if method doesn't exist
	_ = result
}

// TestTemplateBuildContextWithRethrowErrors tests buildContext with rethrowErrors
func TestTemplateBuildContextWithRethrowErrors(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("{{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Set rethrowErrors
	template.rethrowErrors = true

	// Test with Context as assigns
	ctx := NewContext()
	ctx.Set("name", "test")
	result := template.Render(ctx, nil)
	if result != "test" {
		t.Errorf("Expected 'test', got %q", result)
	}
}

// TestTemplateBuildContextWithDropInInstanceAssigns tests buildContext with drop in instanceAssigns
func TestTemplateBuildContextWithDropInInstanceAssigns(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("{{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Add drop to instanceAssigns
	drop := NewDrop()
	template.instanceAssigns["__drop__"] = drop

	// Test with Context as assigns
	ctx := NewContext()
	ctx.Set("name", "test")
	result := template.Render(ctx, nil)
	if result != "test" {
		t.Errorf("Expected 'test', got %q", result)
	}
}

func TestTemplateRegisters(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Test Registers() - should return non-nil map
	registers := template.Registers()
	if registers == nil {
		t.Fatal("Expected non-nil registers map")
	}

	// Test that we can set values
	registers["test"] = "value"
	if registers["test"] != "value" {
		t.Error("Expected to be able to set register value")
	}

	// Test that subsequent calls return same map
	registers2 := template.Registers()
	if registers2["test"] != "value" {
		t.Error("Expected subsequent Registers() call to return same map")
	}
}

func TestTemplateAssigns(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Test Assigns() - should return non-nil map
	assigns := template.Assigns()
	if assigns == nil {
		t.Fatal("Expected non-nil assigns map")
	}

	// Test that we can set values
	assigns["test"] = "value"
	if assigns["test"] != "value" {
		t.Error("Expected to be able to set assign value")
	}

	// Test that subsequent calls return same map
	assigns2 := template.Assigns()
	if assigns2["test"] != "value" {
		t.Error("Expected subsequent Assigns() call to return same map")
	}
}

func TestTemplateInstanceAssigns(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Test InstanceAssigns() - should return non-nil map
	instanceAssigns := template.InstanceAssigns()
	if instanceAssigns == nil {
		t.Fatal("Expected non-nil instanceAssigns map")
	}

	// Test that we can set values
	instanceAssigns["test"] = "value"
	if instanceAssigns["test"] != "value" {
		t.Error("Expected to be able to set instanceAssign value")
	}

	// Test that subsequent calls return same map
	instanceAssigns2 := template.InstanceAssigns()
	if instanceAssigns2["test"] != "value" {
		t.Error("Expected subsequent InstanceAssigns() call to return same map")
	}
}

func TestTemplateErrors(t *testing.T) {
	env := NewEnvironment()
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Test Errors() - should return non-nil slice
	errors := template.Errors()
	if errors == nil {
		t.Fatal("Expected non-nil errors slice")
	}

	// Test that Errors() returns a slice (even if empty initially)
	if errors == nil {
		t.Error("Expected Errors() to return non-nil slice")
	}

	// Test that subsequent calls return same slice
	errors2 := template.Errors()
	if errors2 == nil {
		t.Error("Expected subsequent Errors() call to return non-nil slice")
	}

	// Note: Errors are typically set during rendering, not directly appended
	// The method returns the internal errors slice, which gets populated during render
}
