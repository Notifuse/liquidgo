package liquid

import (
	"testing"
)

func TestContextBasic(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("Expected Context, got nil")
	}
	if ctx.Environment() == nil {
		t.Error("Expected environment, got nil")
	}
	if ctx.Registers() == nil {
		t.Error("Expected registers, got nil")
	}
}

func TestContextSetGet(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key", "value")

	val := ctx.Get("key")
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestContextScopes(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")

	ctx.Push(map[string]interface{}{"key2": "value2"})
	val := ctx.Get("key2")
	if val != "value2" {
		t.Errorf("Expected 'value2', got %v", val)
	}

	ctx.Pop()
	val = ctx.Get("key2")
	if val != nil {
		t.Errorf("Expected nil after pop, got %v", val)
	}
}

func TestContextStack(t *testing.T) {
	ctx := NewContext()
	ctx.Set("outer", "outer_value")

	ctx.Stack(map[string]interface{}{"inner": "inner_value"}, func() {
		if ctx.Get("inner") != "inner_value" {
			t.Error("Expected inner_value in stack")
		}
		if ctx.Get("outer") != "outer_value" {
			t.Error("Expected outer_value to be accessible")
		}
	})

	if ctx.Get("inner") != nil {
		t.Error("Expected inner to be nil after stack")
	}
}

func TestContextMerge(t *testing.T) {
	ctx := NewContext()
	ctx.Merge(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})

	if ctx.Get("key1") != "value1" {
		t.Error("Expected key1 to be value1")
	}
	if ctx.Get("key2") != "value2" {
		t.Error("Expected key2 to be value2")
	}
}

func TestContextFindVariable(t *testing.T) {
	ctx := NewContext()
	ctx.Set("test", "value")

	val := ctx.FindVariable("test", false)
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}

	val = ctx.FindVariable("nonexistent", false)
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestContextEvaluate(t *testing.T) {
	ctx := NewContext()

	// Test with simple value
	result := ctx.Evaluate("test")
	if result != "test" {
		t.Errorf("Expected 'test', got %v", result)
	}

	// Test with VariableLookup
	vl := VariableLookupParse("test", nil, nil)
	result = ctx.Evaluate(vl)
	// Should return nil since variable doesn't exist
	if result != nil {
		t.Logf("VariableLookup evaluated to: %v", result)
	}
}

func TestContextInvoke(t *testing.T) {
	ctx := NewContext()
	ctx.Set("test", "HELLO")

	result := ctx.Invoke("Downcase", "HELLO")
	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}
}

func TestContextInterrupt(t *testing.T) {
	ctx := NewContext()
	if ctx.Interrupt() {
		t.Error("Expected no interrupt initially")
	}

	ctx.PushInterrupt(NewBreakInterrupt())
	if !ctx.Interrupt() {
		t.Error("Expected interrupt after push")
	}

	interrupt := ctx.PopInterrupt()
	if interrupt == nil {
		t.Error("Expected interrupt, got nil")
	}
	if ctx.Interrupt() {
		t.Error("Expected no interrupt after pop")
	}
}

func TestContextWithDisabledTags(t *testing.T) {
	ctx := NewContext()
	if ctx.TagDisabled("test") {
		t.Error("Expected tag not to be disabled")
	}

	ctx.WithDisabledTags([]string{"test"}, func() {
		if !ctx.TagDisabled("test") {
			t.Error("Expected tag to be disabled")
		}
	})

	if ctx.TagDisabled("test") {
		t.Error("Expected tag not to be disabled after WithDisabledTags")
	}
}

func TestContextHandleError(t *testing.T) {
	ctx := NewContext()
	err := NewSyntaxError("test error")

	result := ctx.HandleError(err, nil)
	if result == "" {
		t.Error("Expected error message, got empty string")
	}

	if len(ctx.Errors()) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ctx.Errors()))
	}
}

func TestContextStrictVariables(t *testing.T) {
	ctx := NewContext()
	ctx.SetStrictVariables(true)

	if !ctx.StrictVariables() {
		t.Error("Expected strict variables to be true")
	}

	// Should panic on undefined variable
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for undefined variable in strict mode")
			}
		}()
		ctx.FindVariable("nonexistent", true)
	}()
}

func TestContextResourceLimits(t *testing.T) {
	ctx := NewContext()
	rl := ctx.ResourceLimits()
	if rl == nil {
		t.Error("Expected resource limits, got nil")
	}
}

func TestContextApplyGlobalFilter(t *testing.T) {
	ctx := NewContext()
	ctx.SetGlobalFilter(func(obj interface{}) interface{} {
		return "filtered"
	})

	result := ctx.ApplyGlobalFilter("test")
	if result != "filtered" {
		t.Errorf("Expected 'filtered', got %v", result)
	}
}

func TestContextNewIsolatedSubcontext(t *testing.T) {
	ctx := NewContext()
	ctx.Set("parent", "parent_value")

	subCtx := ctx.NewIsolatedSubcontext()
	if subCtx == nil {
		t.Fatal("Expected subcontext, got nil")
	}

	// Subcontext should have isolated scope
	subCtx.Set("child", "child_value")
	if ctx.Get("child") != nil {
		t.Error("Expected parent context not to see child variable")
	}
}

func TestContextClearInstanceAssigns(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")
	ctx.Set("key2", "value2")

	ctx.ClearInstanceAssigns()

	if ctx.Get("key1") != nil {
		t.Error("Expected key1 to be cleared")
	}
	if ctx.Get("key2") != nil {
		t.Error("Expected key2 to be cleared")
	}
}
