package liquid

import (
	gocontext "context"
	"fmt"
	"sync"
	"testing"
)

func TestTemplateParse(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	// Note: the standard tags come from tags.Standard(), outside this package
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
	env := mustEnv(EnvironmentConfig{})
	// Note: the standard tags come from tags.Standard(), outside this package
	// For now, test without tags (just variables)

	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := renderRuby(template, map[string]interface{}{"name": "world"}, nil)
	expected := "Hello world"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateRenderEmpty(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})

	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	result := renderRuby(template, nil, nil)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestTemplateRenderNilRoot(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})

	template := NewTemplate(&TemplateOptions{Environment: env})
	// Don't parse, so root is nil

	result := renderRuby(template, nil, nil)
	if result != "" {
		t.Errorf("Expected empty string for nil root, got %q", result)
	}
}

// TestTemplateDefaultResourceLimits tests that default resource limits are applied
func TestTemplateDefaultResourceLimits(t *testing.T) {
	renderLimit := 100
	assignLimit := 50
	env := mustEnv(EnvironmentConfig{Limits: ResourceLimitsConfig{RenderLengthLimit: &renderLimit, AssignScoreLimit: &assignLimit}})

	template := NewTemplate(&TemplateOptions{Environment: env})

	// Resource limits should have default values from environment
	if template.limits.RenderLengthLimit == nil || *template.limits.RenderLengthLimit != renderLimit {
		t.Error("Expected render_length_limit to be set from environment")
	}
	if template.limits.AssignScoreLimit == nil || *template.limits.AssignScoreLimit != assignLimit {
		t.Error("Expected assign_score_limit to be set from environment")
	}
}

// TestTemplateEncodingValidation tests UTF-8 encoding validation
func TestTemplateEncodingValidation(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
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
	env := mustEnv(EnvironmentConfig{})

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
		Environment: env,
		LineNumbers: true,
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

// TestTemplateGettersSetters tests all getters and setters
func TestTemplateGettersSetters(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
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

	// Test SetName
	template.SetName("test_template")
	if template.Name() != "test_template" {
		t.Errorf("Expected 'test_template', got %q", template.Name())
	}
}

func TestTemplateRegisters(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
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
	env := mustEnv(EnvironmentConfig{})
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

// TestTemplateLazyInitialization tests lazy initialization of nil fields
func TestTemplateLazyInitialization(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	template := NewTemplate(&TemplateOptions{Environment: env})

	// Manually set fields to nil to test lazy initialization
	template.registers = nil
	template.assigns = nil

	// Test Registers() lazy initialization
	registers := template.Registers()
	if registers == nil {
		t.Fatal("Expected Registers() to initialize nil field")
	}
	if len(registers) != 0 {
		t.Errorf("Expected empty registers map, got %d items", len(registers))
	}

	// Test Assigns() lazy initialization
	assigns := template.Assigns()
	if assigns == nil {
		t.Fatal("Expected Assigns() to initialize nil field")
	}
	if len(assigns) != 0 {
		t.Errorf("Expected empty assigns map, got %d items", len(assigns))
	}

}

// TestTemplateRenderToOutputBufferNilRoot tests a render of a template with
// nil root, through Render, where RenderToOutputBuffer is gone
func TestTemplateRenderToOutputBufferNilRoot(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	template := NewTemplate(&TemplateOptions{Environment: env})
	// Don't parse, so root is nil

	res, err := template.Render(gocontext.Background(), nil)

	// Should not panic and output should be empty
	if err != nil || res.Output != "" {
		t.Errorf("Expected empty output for nil root, got %q (%v)", res.Output, err)
	}
}

// TestTemplateConcurrentRender tests that the same template can be rendered
// concurrently without race conditions.
// This test reproduces the issue from https://github.com/Notifuse/liquidgo/issues/2
func TestTemplateConcurrentRender(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	template := NewTemplate(&TemplateOptions{Environment: env})
	err := template.Parse("Hello {{ name }}", nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			assigns := map[string]interface{}{"name": fmt.Sprintf("user%d", id)}
			res, err := template.Render(gocontext.Background(), assigns)
			expected := fmt.Sprintf("Hello user%d", id)
			if err != nil || res.Output != expected {
				errors <- fmt.Errorf("expected %q, got %q (%v)", expected, res.Output, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}
