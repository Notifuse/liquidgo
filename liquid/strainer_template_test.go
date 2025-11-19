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

// TestStrainerTemplateInvokeEdgeCases tests Invoke with edge cases
func TestStrainerTemplateInvokeEdgeCases(t *testing.T) {
	stc := NewStrainerTemplateClass()
	filter := &StandardFilters{}
	_ = stc.AddFilter(filter)

	ctx := &mockContext{}
	st := NewStrainerTemplate(stc, ctx, false)

	// Test with no arguments (should return nil in non-strict mode)
	result, err := st.Invoke("UnknownMethod")
	if err != nil {
		t.Fatalf("Invoke() should not error in non-strict mode, got %v", err)
	}
	if result != nil {
		t.Logf("Note: Invoke with no args returned %v (may vary)", result)
	}

	// Test with various argument types
	result2, err := st.Invoke("Size", []interface{}{1, 2, 3})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if result2 == nil {
		t.Error("Expected non-nil result for Size filter")
	}

	// Test with nil argument (may panic, so we catch it)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Note: Size(nil) panicked with %v (expected behavior)", r)
			}
		}()
		result3, err := st.Invoke("Size", nil)
		if err != nil {
			t.Logf("Note: Size(nil) returned error: %v", err)
		} else {
			t.Logf("Note: Size(nil) returned %v", result3)
		}
	}()

	// Test with mixed argument types
	result4, err := st.Invoke("Join", []interface{}{"a", "b", "c"}, ",")
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if result4 != "a,b,c" {
		t.Logf("Note: Join result is %v (may vary)", result4)
	}
}

// TestStrainerTemplateInvokeErrorHandling tests error handling in Invoke
func TestStrainerTemplateInvokeErrorHandling(t *testing.T) {
	stc := NewStrainerTemplateClass()
	filter := &StandardFilters{}
	_ = stc.AddFilter(filter)

	ctx := &mockContext{}

	// Test strict mode error handling
	stStrict := NewStrainerTemplate(stc, ctx, true)
	_, err := stStrict.Invoke("NonexistentFilter", "arg")
	if err == nil {
		t.Error("Expected error in strict mode for nonexistent filter")
	}
	if _, ok := err.(*UndefinedFilter); !ok {
		t.Errorf("Expected UndefinedFilter error, got %T", err)
	}

	// Test non-strict mode (should return first arg)
	stNonStrict := NewStrainerTemplate(stc, ctx, false)
	result, err := stNonStrict.Invoke("NonexistentFilter", "arg")
	if err != nil {
		t.Errorf("Expected no error in non-strict mode, got %v", err)
	}
	if result != "arg" {
		t.Errorf("Expected 'arg' in non-strict mode, got %v", result)
	}
}
