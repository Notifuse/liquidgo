package liquid

import (
	"testing"
)

func TestVariableLookupParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*VariableLookup) bool
	}{
		{"simple", "var", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) == 0
		}},
		{"with dot", "var.method", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) == 1
		}},
		{"with brackets", "var[0]", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) == 1
		}},
		{"nested brackets", "var[method][0]", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) >= 2
		}},
		{"command method", "items.size", func(vl *VariableLookup) bool {
			if vl.name != "items" || len(vl.lookups) != 1 {
				return false
			}
			return vl.LookupCommand(0)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VariableLookupParse(tt.input, nil, nil)
			if result == nil {
				t.Fatal("Expected VariableLookup, got nil")
			}
			if !tt.check(result) {
				t.Errorf("VariableLookupParse(%q) did not pass check", tt.input)
			}
		})
	}
}

func TestVariableLookupCommandMethods(t *testing.T) {
	vl := VariableLookupParse("items.size", nil, nil)
	if !vl.LookupCommand(0) {
		t.Error("Expected size to be a command method")
	}

	vl = VariableLookupParse("items.first", nil, nil)
	if !vl.LookupCommand(0) {
		t.Error("Expected first to be a command method")
	}

	vl = VariableLookupParse("items.last", nil, nil)
	if !vl.LookupCommand(0) {
		t.Error("Expected last to be a command method")
	}

	vl = VariableLookupParse("items.name", nil, nil)
	if vl.LookupCommand(0) {
		t.Error("Expected name NOT to be a command method")
	}
}

// TestVariableLookupEvaluate tests comprehensive variable lookup evaluation
func TestVariableLookupEvaluate(t *testing.T) {
	ctx := NewContext()
	ctx.Set("items", []interface{}{"a", "b", "c"})

	vl := VariableLookupParse("items", nil, nil)
	result := vl.Evaluate(ctx)
	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Test with nested lookup
	ctx.Set("user", map[string]interface{}{
		"name": "John",
	})
	vl2 := VariableLookupParse("user.name", nil, nil)
	result2 := vl2.Evaluate(ctx)
	if result2 != "John" {
		t.Errorf("Expected 'John', got %v", result2)
	}

	// Test with array index
	vl3 := VariableLookupParse("items[0]", nil, nil)
	result3 := vl3.Evaluate(ctx)
	if result3 != "a" {
		t.Errorf("Expected 'a', got %v", result3)
	}
}

// TestVariableLookupName tests Name method
func TestVariableLookupName(t *testing.T) {
	vl := VariableLookupParse("test", nil, nil)
	if vl.Name() != "test" {
		t.Errorf("Expected 'test', got %q", vl.Name())
	}
}

// TestVariableLookupLookups tests Lookups method
func TestVariableLookupLookups(t *testing.T) {
	vl := VariableLookupParse("user.name", nil, nil)
	lookups := vl.Lookups()
	if len(lookups) != 1 {
		t.Errorf("Expected 1 lookup, got %d", len(lookups))
	}
}

// Custom type aliases for testing
type MapOfAny map[string]any
type CustomMap map[string]interface{}
type CustomStringMap map[string]string

// TestVariableLookupCustomMapTypes tests custom map type aliases
func TestVariableLookupCustomMapTypes(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		variable string
		expected interface{}
	}{
		{
			name: "MapOfAny type alias",
			data: MapOfAny{
				"id": "test-123",
			},
			variable: "workspace.id",
			expected: "test-123",
		},
		{
			name: "CustomMap type alias",
			data: CustomMap{
				"name": "John Doe",
			},
			variable: "user.name",
			expected: "John Doe",
		},
		{
			name: "CustomStringMap type alias",
			data: CustomStringMap{
				"title": "Hello",
			},
			variable: "page.title",
			expected: "Hello",
		},
		{
			name: "Nested custom map types",
			data: MapOfAny{
				"nested": MapOfAny{
					"value": 42,
				},
			},
			variable: "obj.nested.value",
			expected: 42,
		},
		{
			name: "Mixed native and custom maps",
			data: map[string]interface{}{
				"custom": MapOfAny{
					"id": "nested-123",
				},
			},
			variable: "root.custom.id",
			expected: "nested-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()

			// Extract the root variable name and set it in context
			vl := VariableLookupParse(tt.variable, nil, nil)
			rootVar := vl.Name().(string)
			ctx.Set(rootVar, tt.data)

			// Evaluate the variable lookup
			result := vl.Evaluate(ctx)

			if result != tt.expected {
				// For struct comparisons, result might be converted to map or something else by Liquid
				// But here we expect direct value access
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper to create a context
func buildTestContext(config ContextConfig) *Context {
	// This function is not exported by liquid package but used in tests
	// We can use NewContext and configure it
	ctx := NewContext()
	// ... configuration ...
	return ctx
}

// ContextConfig is already defined in context.go
// type ContextConfig struct {
// 	Environment *Environment
// }

// TestTryMapAccess tests the tryMapAccess helper function directly
func TestTryMapAccess(t *testing.T) {
	tests := []struct {
		name      string
		obj       interface{}
		key       string
		wantValue interface{}
		wantFound bool
	}{
		{
			name:      "native map[string]interface{}",
			obj:       map[string]interface{}{"key": "value"},
			key:       "key",
			wantValue: "value",
			wantFound: true,
		},
		{
			name:      "custom MapOfAny",
			obj:       MapOfAny{"id": "123"},
			key:       "id",
			wantValue: "123",
			wantFound: true,
		},
		{
			name:      "custom CustomMap",
			obj:       CustomMap{"name": "test"},
			key:       "name",
			wantValue: "test",
			wantFound: true,
		},
		{
			name:      "key not found",
			obj:       MapOfAny{"id": "123"},
			key:       "missing",
			wantValue: nil,
			wantFound: false,
		},
		{
			name:      "not a map",
			obj:       "string",
			key:       "key",
			wantValue: nil,
			wantFound: false,
		},
		{
			name:      "nil object",
			obj:       nil,
			key:       "key",
			wantValue: nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotFound := tryMapAccess(tt.obj, tt.key)

			if gotFound != tt.wantFound {
				t.Errorf("tryMapAccess() gotFound = %v, want %v", gotFound, tt.wantFound)
			}

			if gotValue != tt.wantValue {
				t.Errorf("tryMapAccess() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestVariableLookupTypedSlices tests that typed slices (e.g., []string, []int) work with array index access
func TestVariableLookupTypedSlices(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	tests := []struct {
		name     string
		data     map[string]interface{}
		markup   string
		expected interface{}
	}{
		{
			name:     "[]string index access",
			data:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			markup:   "items[1]",
			expected: "b",
		},
		{
			name:     "[]int index access",
			data:     map[string]interface{}{"nums": []int{10, 20, 30}},
			markup:   "nums[0]",
			expected: 10,
		},
		{
			name:     "[]float64 index access",
			data:     map[string]interface{}{"floats": []float64{1.1, 2.2, 3.3}},
			markup:   "floats[2]",
			expected: 3.3,
		},
		{
			name:     "[]struct index access",
			data:     map[string]interface{}{"structs": []TestStruct{{Name: "first", Value: 1}, {Name: "second", Value: 2}}},
			markup:   "structs[1].Name",
			expected: "second",
		},
		{
			name:     "[]string .size command",
			data:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			markup:   "items.size",
			expected: 3,
		},
		{
			name:     "[]int .first command",
			data:     map[string]interface{}{"nums": []int{10, 20, 30}},
			markup:   "nums.first",
			expected: 10,
		},
		{
			name:     "[]string .last command",
			data:     map[string]interface{}{"items": []string{"a", "b", "c"}},
			markup:   "items.last",
			expected: "c",
		},
		{
			name:     "[]string .empty command - not empty",
			data:     map[string]interface{}{"items": []string{"a", "b"}},
			markup:   "items.empty",
			expected: false,
		},
		{
			name:     "[]string .empty command - empty",
			data:     map[string]interface{}{"items": []string{}},
			markup:   "items.empty",
			expected: true,
		},
		{
			name:     "array [3]int index access",
			data:     map[string]interface{}{"arr": [3]int{5, 10, 15}},
			markup:   "arr[1]",
			expected: 10,
		},
		{
			name:     "array .size command",
			data:     map[string]interface{}{"arr": [4]string{"w", "x", "y", "z"}},
			markup:   "arr.size",
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vl := VariableLookupParse(tt.markup, nil, nil)
			if vl == nil {
				t.Fatal("VariableLookupParse returned nil")
			}

			ctx := buildTestContext(ContextConfig{
				Environment: NewEnvironment(),
			})
			for k, v := range tt.data {
				ctx.Set(k, v)
			}

			result := vl.Evaluate(ctx)
			if result != tt.expected {
				t.Errorf("Expected %v (type %T), got %v (type %T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

// TestVariableLookupStructFieldAccess tests that struct fields can be accessed with snake_case
func TestVariableLookupStructFieldAccess(t *testing.T) {
	type BlogPost struct {
		Title         string
		Author        string
		CommentsCount int
	}

	tests := []struct {
		name     string
		post     BlogPost
		markup   string
		expected interface{}
	}{
		{
			name:     "CamelCase field",
			post:     BlogPost{Title: "Test Post", Author: "Alice", CommentsCount: 5},
			markup:   "post.Title",
			expected: "Test Post",
		},
		{
			name:     "snake_case to CamelCase",
			post:     BlogPost{Title: "Test Post", Author: "Alice", CommentsCount: 5},
			markup:   "post.comments_count",
			expected: 5,
		},
		{
			name:     "lowercase to CamelCase",
			post:     BlogPost{Title: "Test Post", Author: "Alice", CommentsCount: 5},
			markup:   "post.author",
			expected: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vl := VariableLookupParse(tt.markup, nil, nil)
			if vl == nil {
				t.Fatal("VariableLookupParse returned nil")
			}

			ctx := buildTestContext(ContextConfig{
				Environment: NewEnvironment(),
			})
			ctx.Set("post", tt.post)

			result := vl.Evaluate(ctx)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
