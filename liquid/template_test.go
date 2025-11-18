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
