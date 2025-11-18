package liquid

import (
	"testing"
)

type mockContext struct{}

func (m *mockContext) Context() interface{} {
	return m
}

func TestStrainerTemplateClass(t *testing.T) {
	stc := NewStrainerTemplateClass()
	if stc == nil {
		t.Fatal("Expected StrainerTemplateClass, got nil")
	}

	// Add a filter
	filter := &StandardFilters{}
	err := stc.AddFilter(filter)
	if err != nil {
		t.Fatalf("AddFilter() error = %v", err)
	}

	// Check if methods are invokable
	if !stc.Invokable("Size") {
		t.Error("Expected Size to be invokable")
	}
	if !stc.Invokable("Downcase") {
		t.Error("Expected Downcase to be invokable")
	}
}

func TestStrainerTemplateInvoke(t *testing.T) {
	stc := NewStrainerTemplateClass()
	filter := &StandardFilters{}
	_ = stc.AddFilter(filter)

	ctx := &mockContext{}
	st := NewStrainerTemplate(stc, ctx, false)

	// Test invokable method
	result, err := st.Invoke("Size", "hello")
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// Test non-invokable method (should return first arg in non-strict mode)
	result, err = st.Invoke("UnknownMethod", "test")
	if err != nil {
		t.Fatalf("Invoke() should not error in non-strict mode, got %v", err)
	}
	if result != "test" {
		t.Errorf("Expected 'test', got %v", result)
	}
}

func TestStrainerTemplateStrictMode(t *testing.T) {
	stc := NewStrainerTemplateClass()
	filter := &StandardFilters{}
	_ = stc.AddFilter(filter)

	ctx := &mockContext{}
	st := NewStrainerTemplate(stc, ctx, true)

	// Test non-invokable method in strict mode (should error)
	_, err := st.Invoke("UnknownMethod", "test")
	if err == nil {
		t.Error("Expected error in strict mode for unknown method")
	}
	if _, ok := err.(*UndefinedFilter); !ok {
		t.Errorf("Expected UndefinedFilter error, got %T", err)
	}
}

func TestStrainerTemplateFilterMethodNames(t *testing.T) {
	stc := NewStrainerTemplateClass()
	filter := &StandardFilters{}
	_ = stc.AddFilter(filter)

	names := stc.FilterMethodNames()
	if len(names) == 0 {
		t.Error("Expected filter method names, got empty")
	}

	// Check for some expected methods
	found := false
	for _, name := range names {
		if name == "Size" || name == "Downcase" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Size or Downcase in filter method names")
	}
}

// TestStrainerTemplateReflectionInvocation tests reflection-based method invocation
func TestStrainerTemplateReflectionInvocation(t *testing.T) {
	stc := NewStrainerTemplateClass()
	filter := &StandardFilters{}
	_ = stc.AddFilter(filter)
	
	ctx := &mockContext{}
	st := NewStrainerTemplate(stc, ctx, false)
	
	// Test that reflection-based invocation works
	result, err := st.Invoke("Downcase", "HELLO")
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}
	
	// Test with multiple arguments
	result, err = st.Invoke("Slice", "hello", 1, 3)
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if result != "ell" {
		t.Errorf("Expected 'ell', got %v", result)
	}
}

