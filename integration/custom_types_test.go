package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// Custom type alias (like domain.MapOfAny in the bug report)
type MapOfAny map[string]any
type CustomMap map[string]interface{}

// TestCustomMapTypeAlias tests the exact reproduction case from the bug report.
// It verifies that custom type aliases based on map[string]any are rendered correctly.
func TestCustomMapTypeAlias(t *testing.T) {
	// Test 1: Direct map[string]interface{} - WORKS ✅
	data1 := map[string]interface{}{
		"workspace": map[string]interface{}{
			"id": "test-123",
		},
	}

	// Test 2: Custom type MapOfAny - Should now WORK ✅
	workspaceData := MapOfAny{
		"id": "test-456",
	}
	data2 := map[string]interface{}{
		"workspace": workspaceData,
	}

	// Test 3: After conversion to map[string]interface{} - WORKS ✅
	workspaceConverted := make(map[string]interface{})
	for k, v := range workspaceData {
		workspaceConverted[k] = v
	}
	data3 := map[string]interface{}{
		"workspace": workspaceConverted,
	}

	template := "{{ workspace.id }}"

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result1 := tmpl.Render(data1, nil)
	result2 := tmpl.Render(data2, nil)
	result3 := tmpl.Render(data3, nil)

	// All three should produce the expected results
	if result1 != "test-123" {
		t.Errorf("map[string]interface{}: Expected 'test-123', got %q", result1)
	}

	if result2 != "test-456" {
		t.Errorf("Custom type MapOfAny: Expected 'test-456', got %q", result2)
	}

	if result3 != "test-456" {
		t.Errorf("MapOfAny after conversion: Expected 'test-456', got %q", result3)
	}
}

// TestCustomMapTypeAliasWithFilters tests custom map types with Liquid filters
func TestCustomMapTypeAliasWithFilters(t *testing.T) {
	workspaceData := MapOfAny{
		"name": "example workspace",
		"id":   "ws-123",
	}

	data := map[string]interface{}{
		"workspace": workspaceData,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "upcase filter",
			template: "{{ workspace.name | upcase }}",
			expected: "EXAMPLE WORKSPACE",
		},
		{
			name:     "downcase filter",
			template: "{{ workspace.name | downcase }}",
			expected: "example workspace",
		},
		{
			name:     "capitalize filter",
			template: "{{ workspace.name | capitalize }}",
			expected: "Example workspace",
		},
		{
			name:     "append filter",
			template: "{{ workspace.id | append: '-suffix' }}",
			expected: "ws-123-suffix",
		},
	}

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result := tmpl.Render(data, nil)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestNestedCustomMapTypes tests deeply nested custom map type aliases
func TestNestedCustomMapTypes(t *testing.T) {
	data := map[string]interface{}{
		"company": MapOfAny{
			"workspace": MapOfAny{
				"project": MapOfAny{
					"task": MapOfAny{
						"id": "deep-nested-id",
					},
				},
			},
		},
	}

	template := "{{ company.workspace.project.task.id }}"

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result := tmpl.Render(data, nil)
	expected := "deep-nested-id"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestCustomMapInLoops tests custom map types in for loops
func TestCustomMapInLoops(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{
			MapOfAny{"name": "Item 1", "value": 10},
			MapOfAny{"name": "Item 2", "value": 20},
			MapOfAny{"name": "Item 3", "value": 30},
		},
	}

	template := "{% for item in items %}{{ item.name }}: {{ item.value }}\n{% endfor %}"

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result := tmpl.Render(data, nil)
	expected := "Item 1: 10\nItem 2: 20\nItem 3: 30\n"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestCustomMapInConditionals tests custom map types in if statements
func TestCustomMapInConditionals(t *testing.T) {
	data := map[string]interface{}{
		"user": MapOfAny{
			"active": true,
			"name":   "John",
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "if statement with custom map",
			template: "{% if user.active %}{{ user.name }} is active{% endif %}",
			expected: "John is active",
		},
		{
			name:     "unless statement with custom map",
			template: "{% unless user.active %}Inactive{% else %}Active{% endunless %}",
			expected: "Active",
		},
	}

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result := tmpl.Render(data, nil)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestMixedCustomAndNativeMaps tests mixing custom and native map types
func TestMixedCustomAndNativeMaps(t *testing.T) {
	data := map[string]interface{}{
		"native": map[string]interface{}{
			"id": "native-123",
			"custom": MapOfAny{
				"value": "custom-456",
			},
		},
		"custom": MapOfAny{
			"id": "custom-789",
			"native": map[string]interface{}{
				"value": "native-012",
			},
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "native with nested custom",
			template: "{{ native.custom.value }}",
			expected: "custom-456",
		},
		{
			name:     "custom with nested native",
			template: "{{ custom.native.value }}",
			expected: "native-012",
		},
		{
			name:     "top level native",
			template: "{{ native.id }}",
			expected: "native-123",
		},
		{
			name:     "top level custom",
			template: "{{ custom.id }}",
			expected: "custom-789",
		},
	}

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result := tmpl.Render(data, nil)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestCustomMapWithDifferentTypes tests custom map type aliases with different value types
func TestCustomMapWithDifferentTypes(t *testing.T) {
	data := map[string]interface{}{
		"data": MapOfAny{
			"string": "text",
			"int":    42,
			"float":  3.14,
			"bool":   true,
			"nil":    nil,
			"array":  []interface{}{"a", "b", "c"},
			"nested": MapOfAny{"key": "value"},
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "string value",
			template: "{{ data.string }}",
			expected: "text",
		},
		{
			name:     "int value",
			template: "{{ data.int }}",
			expected: "42",
		},
		{
			name:     "float value",
			template: "{{ data.float }}",
			expected: "3.14",
		},
		{
			name:     "bool value",
			template: "{{ data.bool }}",
			expected: "true",
		},
		{
			name:     "nil value",
			template: "{{ data.nil }}",
			expected: "",
		},
		{
			name:     "array value",
			template: "{{ data.array[0] }}",
			expected: "a",
		},
		{
			name:     "nested map",
			template: "{{ data.nested.key }}",
			expected: "value",
		},
	}

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := liquid.ParseTemplate(tt.template, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result := tmpl.Render(data, nil)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestCustomMapTypeAssignTag tests custom map types with assign tag
func TestCustomMapTypeAssignTag(t *testing.T) {
	data := map[string]interface{}{
		"workspace": MapOfAny{
			"id":   "ws-123",
			"name": "My Workspace",
		},
	}

	template := "{% assign ws_id = workspace.id %}{% assign ws_name = workspace.name %}ID: {{ ws_id }}, Name: {{ ws_name }}"

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result := tmpl.Render(data, nil)
	expected := "ID: ws-123, Name: My Workspace"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
