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

	// Test ParseContext (returns nil currently)
	parseCtx := ctx.ParseContext()
	if parseCtx != nil {
		t.Logf("ParseContext returned %v (may be nil as per TODO)", parseCtx)
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

// TestContextHandleErrorComprehensive tests comprehensive error handling
func TestContextHandleErrorComprehensive(t *testing.T) {
	ctx := NewContext()

	// Test with SyntaxError
	syntaxErr := NewSyntaxError("syntax error")
	result := ctx.HandleError(syntaxErr, nil)
	if result == "" {
		t.Error("Expected error message, got empty string")
	}
	if len(ctx.Errors()) != 1 {
		t.Errorf("Expected 1 error, got %d", len(ctx.Errors()))
	}

	// Test with ContextError
	ctx2 := NewContext()
	contextErr := NewContextError("context error")
	result2 := ctx2.HandleError(contextErr, nil)
	if result2 == "" {
		t.Error("Expected error message, got empty string")
	}

	// Test with UndefinedVariable
	ctx3 := NewContext()
	undefinedErr := NewUndefinedVariable("undefined variable")
	result3 := ctx3.HandleError(undefinedErr, nil)
	if result3 == "" {
		t.Error("Expected error message, got empty string")
	}

	// Test with ExceptionRenderer
	ctx4 := NewContext()
	ctx4.SetExceptionRenderer(func(err error) interface{} {
		return "custom error"
	})
	result4 := ctx4.HandleError(NewSyntaxError("test"), nil)
	if result4 != "custom error" {
		t.Errorf("Expected 'custom error', got %q", result4)
	}

	// Test with line number
	lineNum := 42
	ctx5 := NewContext()
	ctx5.SetTemplateName("test.liquid")
	result5 := ctx5.HandleError(NewSyntaxError("test"), &lineNum)
	if result5 == "" {
		t.Error("Expected error message with line number")
	}
}

// TestContextLookupAndEvaluate tests LookupAndEvaluate method
func TestContextLookupAndEvaluate(t *testing.T) {
	ctx := NewContext()
	obj := map[string]interface{}{
		"key": "value",
	}

	result := ctx.LookupAndEvaluate(obj, "key", false)
	if result != "value" {
		t.Errorf("Expected 'value', got %v", result)
	}

	// Test with nonexistent key
	result2 := ctx.LookupAndEvaluate(obj, "nonexistent", false)
	if result2 != nil {
		t.Errorf("Expected nil, got %v", result2)
	}

	// Test with function value
	obj2 := map[string]interface{}{
		"func": func() interface{} {
			return "function result"
		},
	}
	result3 := ctx.LookupAndEvaluate(obj2, "func", false)
	if result3 != "function result" {
		t.Errorf("Expected 'function result', got %v", result3)
	}
}

// TestContextEvaluateComprehensive tests comprehensive evaluation
func TestContextEvaluateComprehensive(t *testing.T) {
	ctx := NewContext()

	// Test with nil
	result := ctx.Evaluate(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}

	// Test with simple value
	result = ctx.Evaluate("test")
	if result != "test" {
		t.Errorf("Expected 'test', got %v", result)
	}

	// Test with VariableLookup
	ctx.Set("name", "value")
	vl := VariableLookupParse("name", nil, nil)
	result = ctx.Evaluate(vl)
	if result != "value" {
		t.Errorf("Expected 'value', got %v", result)
	}

	// Test with RangeLookup
	rl := &RangeLookup{
		startObj: 1,
		endObj:   5,
	}
	result = ctx.Evaluate(rl)
	if result == nil {
		t.Error("Expected non-nil Range result")
	}

	// Test with evaluable object
	evaluable := &testEvaluable{value: "evaluated"}
	result = ctx.Evaluate(evaluable)
	if result != "evaluated" {
		t.Errorf("Expected 'evaluated', got %v", result)
	}
}

// testEvaluable is a test type that implements Evaluate
type testEvaluable struct {
	value string
}

func (t *testEvaluable) Evaluate(ctx *Context) interface{} {
	return t.value
}

// TestContextGettersSetters tests all getters and setters
func TestContextGettersSetters(t *testing.T) {
	ctx := NewContext()

	// Test Scopes
	scopes := ctx.Scopes()
	if scopes == nil {
		t.Error("Expected scopes, got nil")
	}

	// Test Warnings
	warnings := ctx.Warnings()
	if warnings == nil {
		t.Error("Expected warnings slice, got nil")
	}

	// Test AddWarning
	ctx.AddWarning(NewSyntaxError("warning"))
	if len(ctx.Warnings()) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(ctx.Warnings()))
	}

	// Test Partial
	if ctx.Partial() {
		t.Error("Expected Partial to be false initially")
	}
	ctx.SetPartial(true)
	if !ctx.Partial() {
		t.Error("Expected Partial to be true after SetPartial")
	}

	// Test StrictFilters
	if ctx.StrictFilters() {
		t.Error("Expected StrictFilters to be false initially")
	}
	ctx.SetStrictFilters(true)
	if !ctx.StrictFilters() {
		t.Error("Expected StrictFilters to be true after SetStrictFilters")
	}

	// Test GlobalFilter
	if ctx.GlobalFilter() != nil {
		t.Error("Expected GlobalFilter to be nil initially")
	}
	filter := func(obj interface{}) interface{} {
		return "filtered"
	}
	ctx.SetGlobalFilter(filter)
	if ctx.GlobalFilter() == nil {
		t.Error("Expected GlobalFilter to be set")
	}

	// Test ExceptionRenderer
	if ctx.ExceptionRenderer() == nil {
		t.Error("Expected ExceptionRenderer to be set")
	}
	renderer := func(err error) interface{} {
		return "rendered"
	}
	ctx.SetExceptionRenderer(renderer)
	if ctx.ExceptionRenderer() == nil {
		t.Error("Expected ExceptionRenderer to be set")
	}

	// Test AddFilters
	ctx.AddFilters([]interface{}{&StandardFilters{}})
	if len(ctx.filters) == 0 {
		t.Error("Expected filters to be added")
	}

	// Test Key
	ctx.Set("testkey", "testvalue")
	if !ctx.Key("testkey") {
		t.Error("Expected Key to return true for existing key")
	}
	if ctx.Key("nonexistent") {
		t.Error("Expected Key to return false for nonexistent key")
	}

	// Test SetLast
	ctx.Push(map[string]interface{}{"inner": "inner_value"})
	ctx.SetLast("lastkey", "lastvalue")
	if ctx.Get("lastkey") != "lastvalue" {
		t.Error("Expected SetLast to set value in last scope")
	}
}

// TestContextLookupAndEvaluateMethod tests public LookupAndEvaluate method
func TestContextLookupAndEvaluateMethod(t *testing.T) {
	ctx := NewContext()
	obj := map[string]interface{}{
		"key": "value",
	}

	result := ctx.LookupAndEvaluate(obj, "key", false)
	if result != "value" {
		t.Errorf("Expected 'value', got %v", result)
	}
}

// TestContextEvaluateComplexExpressions tests complex expression evaluation
func TestContextEvaluateComplexExpressions(t *testing.T) {
	ctx := NewContext()
	ctx.Set("user", map[string]interface{}{
		"name": "John",
		"age":  30,
	})

	// Test nested variable lookup
	vl := VariableLookupParse("user.name", nil, nil)
	result := ctx.Evaluate(vl)
	if result != "John" {
		t.Errorf("Expected 'John', got %v", result)
	}

	// Test with array
	ctx.Set("items", []interface{}{"a", "b", "c"})
	vl2 := VariableLookupParse("items", nil, nil)
	result2 := ctx.Evaluate(vl2)
	if result2 == nil {
		t.Error("Expected non-nil result for array")
	}

	// Test with map
	ctx.Set("data", map[string]interface{}{
		"nested": map[string]interface{}{
			"value": "deep",
		},
	})
	vl3 := VariableLookupParse("data.nested.value", nil, nil)
	result3 := ctx.Evaluate(vl3)
	if result3 != "deep" {
		t.Errorf("Expected 'deep', got %v", result3)
	}
}
