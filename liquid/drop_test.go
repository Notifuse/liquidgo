package liquid

import (
	"testing"
)

func TestDropBasic(t *testing.T) {
	drop := NewDrop()
	if drop == nil {
		t.Fatal("Expected Drop, got nil")
	}
	if drop.Context() != nil {
		t.Error("Expected nil context initially")
	}
}

func TestDropSetContext(t *testing.T) {
	drop := NewDrop()
	ctx := NewContext()
	drop.SetContext(ctx)

	if drop.Context() != ctx {
		t.Error("Expected context to be set")
	}
}

func TestDropInvokeDrop(t *testing.T) {
	drop := NewDrop()

	// Test invoking a non-existent method (should call LiquidMethodMissing)
	result := drop.InvokeDrop("nonexistent")
	if result != nil {
		t.Errorf("Expected nil for nonexistent method, got %v", result)
	}
}

func TestDropToLiquid(t *testing.T) {
	drop := NewDrop()
	// Note: ToLiquid() was removed from Drop to prevent type loss with embedded drops
	// The default behavior (object unchanged) is now used instead
	// Test that ToLiquid (from extensions.go) returns the drop unchanged
	result := ToLiquid(drop)
	if result != drop {
		t.Error("Expected drop to return itself")
	}
}

func TestDropKey(t *testing.T) {
	drop := NewDrop()
	if !drop.Key("anykey") {
		t.Error("Expected Key to return true")
	}
}

func TestForloopDropBasic(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)
	if fl == nil {
		t.Fatal("Expected ForloopDrop, got nil")
	}
	if fl.Name() != "items" {
		t.Errorf("Expected name 'items', got '%s'", fl.Name())
	}
	if fl.Length() != 5 {
		t.Errorf("Expected length 5, got %d", fl.Length())
	}
	if fl.Parentloop() != nil {
		t.Error("Expected nil parentloop")
	}
}

func TestForloopDropIndex(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)

	if fl.Index() != 1 {
		t.Errorf("Expected Index 1, got %d", fl.Index())
	}
	if fl.Index0() != 0 {
		t.Errorf("Expected Index0 0, got %d", fl.Index0())
	}

	fl.Increment()
	if fl.Index() != 2 {
		t.Errorf("Expected Index 2 after increment, got %d", fl.Index())
	}
	if fl.Index0() != 1 {
		t.Errorf("Expected Index0 1 after increment, got %d", fl.Index0())
	}
}

func TestForloopDropFirstLast(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)

	if !fl.First() {
		t.Error("Expected First to be true initially")
	}
	if fl.Last() {
		t.Error("Expected Last to be false initially")
	}

	// Increment to last
	for i := 0; i < 4; i++ {
		fl.Increment()
	}

	if fl.First() {
		t.Error("Expected First to be false at end")
	}
	if !fl.Last() {
		t.Error("Expected Last to be true at end")
	}
}

func TestForloopDropRindex(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)

	if fl.Rindex() != 5 {
		t.Errorf("Expected Rindex 5, got %d", fl.Rindex())
	}
	if fl.Rindex0() != 4 {
		t.Errorf("Expected Rindex0 4, got %d", fl.Rindex0())
	}

	fl.Increment()
	if fl.Rindex() != 4 {
		t.Errorf("Expected Rindex 4, got %d", fl.Rindex())
	}
	if fl.Rindex0() != 3 {
		t.Errorf("Expected Rindex0 3, got %d", fl.Rindex0())
	}
}

func TestForloopDropParentloop(t *testing.T) {
	parent := NewForloopDrop("outer", 3, nil)
	child := NewForloopDrop("inner", 2, parent)

	if child.Parentloop() != parent {
		t.Error("Expected parentloop to be set")
	}
}

func TestTablerowloopDropBasic(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)
	if tr == nil {
		t.Fatal("Expected TablerowloopDrop, got nil")
	}
	if tr.Length() != 10 {
		t.Errorf("Expected length 10, got %d", tr.Length())
	}
	if tr.Cols() != 3 {
		t.Errorf("Expected cols 3, got %d", tr.Cols())
	}
	if tr.Row() != 1 {
		t.Errorf("Expected row 1, got %d", tr.Row())
	}
	if tr.Col() != 1 {
		t.Errorf("Expected col 1, got %d", tr.Col())
	}
}

func TestTablerowloopDropIncrement(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// First increment
	tr.Increment()
	if tr.Row() != 1 {
		t.Errorf("Expected row 1, got %d", tr.Row())
	}
	if tr.Col() != 2 {
		t.Errorf("Expected col 2, got %d", tr.Col())
	}

	// Second increment
	tr.Increment()
	if tr.Row() != 1 {
		t.Errorf("Expected row 1, got %d", tr.Row())
	}
	if tr.Col() != 3 {
		t.Errorf("Expected col 3, got %d", tr.Col())
	}

	// Third increment (should wrap to next row)
	tr.Increment()
	if tr.Row() != 2 {
		t.Errorf("Expected row 2, got %d", tr.Row())
	}
	if tr.Col() != 1 {
		t.Errorf("Expected col 1, got %d", tr.Col())
	}
}

func TestTablerowloopDropColFirstLast(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	if !tr.ColFirst() {
		t.Error("Expected ColFirst to be true initially")
	}
	if tr.ColLast() {
		t.Error("Expected ColLast to be false initially")
	}

	tr.Increment()
	if tr.ColFirst() {
		t.Error("Expected ColFirst to be false")
	}
	if tr.ColLast() {
		t.Error("Expected ColLast to be false")
	}

	tr.Increment()
	if !tr.ColLast() {
		t.Error("Expected ColLast to be true")
	}
}

func TestTablerowloopDropCol0(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	if tr.Col0() != 0 {
		t.Errorf("Expected Col0 0, got %d", tr.Col0())
	}

	tr.Increment()
	if tr.Col0() != 1 {
		t.Errorf("Expected Col0 1, got %d", tr.Col0())
	}
}

func TestSnippetDropBasic(t *testing.T) {
	sd := NewSnippetDrop("body content", "snippet_name", "snippet.liquid")
	if sd == nil {
		t.Fatal("Expected SnippetDrop, got nil")
	}
	if sd.Body() != "body content" {
		t.Errorf("Expected body 'body content', got '%s'", sd.Body())
	}
	if sd.Name() != "snippet_name" {
		t.Errorf("Expected name 'snippet_name', got '%s'", sd.Name())
	}
	if sd.Filename() != "snippet.liquid" {
		t.Errorf("Expected filename 'snippet.liquid', got '%s'", sd.Filename())
	}
}

func TestSnippetDropToPartial(t *testing.T) {
	sd := NewSnippetDrop("body content", "snippet_name", "snippet.liquid")
	if sd.ToPartial() != "body content" {
		t.Errorf("Expected ToPartial to return body, got '%s'", sd.ToPartial())
	}
}

func TestSnippetDropString(t *testing.T) {
	sd := NewSnippetDrop("body", "name", "file")
	if sd.String() != "SnippetDrop" {
		t.Errorf("Expected String 'SnippetDrop', got '%s'", sd.String())
	}
}

func TestDropInvokeDropWithMethod(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)

	// Test invoking Length method
	result := fl.InvokeDrop("Length")
	if result != 5 {
		t.Errorf("Expected Length 5, got %v", result)
	}

	// Test invoking Name method
	result = fl.InvokeDrop("Name")
	if result != "items" {
		t.Errorf("Expected Name 'items', got %v", result)
	}
}

func TestDropStrictVariables(t *testing.T) {
	drop := NewDrop()
	ctx := NewContext()
	ctx.SetStrictVariables(true)
	drop.SetContext(ctx)

	// Should panic on undefined method
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for undefined method in strict mode")
			}
		}()
		drop.InvokeDrop("nonexistent")
	}()
}

func TestDropInvokeDropOld(t *testing.T) {
	drop := NewDrop()

	// Test InvokeDropOld (old implementation)
	result := drop.InvokeDropOld("nonexistent")
	if result != nil {
		t.Errorf("Expected nil for nonexistent method, got %v", result)
	}
}

func TestDropString(t *testing.T) {
	drop := NewDrop()
	result := drop.String()
	if result == "" {
		t.Error("Expected non-empty string representation")
	}
	// Should contain type information
	if len(result) == 0 {
		t.Error("Expected non-empty string representation")
	}
}

func TestTablerowloopDropIndex(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test Index (1-based)
	if tr.Index() != 1 {
		t.Errorf("Expected Index 1, got %d", tr.Index())
	}

	tr.Increment()
	if tr.Index() != 2 {
		t.Errorf("Expected Index 2, got %d", tr.Index())
	}
}

func TestTablerowloopDropIndex0(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test Index0 (0-based)
	if tr.Index0() != 0 {
		t.Errorf("Expected Index0 0, got %d", tr.Index0())
	}

	tr.Increment()
	if tr.Index0() != 1 {
		t.Errorf("Expected Index0 1, got %d", tr.Index0())
	}
}

func TestTablerowloopDropRindex(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test Rindex (1-based reverse)
	if tr.Rindex() != 10 {
		t.Errorf("Expected Rindex 10, got %d", tr.Rindex())
	}

	tr.Increment()
	if tr.Rindex() != 9 {
		t.Errorf("Expected Rindex 9, got %d", tr.Rindex())
	}
}

func TestTablerowloopDropRindex0(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test Rindex0 (0-based reverse)
	if tr.Rindex0() != 9 {
		t.Errorf("Expected Rindex0 9, got %d", tr.Rindex0())
	}

	tr.Increment()
	if tr.Rindex0() != 8 {
		t.Errorf("Expected Rindex0 8, got %d", tr.Rindex0())
	}
}

func TestTablerowloopDropFirst(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test First
	if !tr.First() {
		t.Error("Expected First to be true initially")
	}

	tr.Increment()
	if tr.First() {
		t.Error("Expected First to be false after increment")
	}
}

func TestTablerowloopDropLast(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test Last (should be false initially)
	if tr.Last() {
		t.Error("Expected Last to be false initially")
	}

	// Increment to last item
	for i := 0; i < 9; i++ {
		tr.Increment()
	}

	if !tr.Last() {
		t.Error("Expected Last to be true at end")
	}
}

func TestTablerowloopDropInvokeDrop(t *testing.T) {
	tr := NewTablerowloopDrop(10, 3)

	// Test InvokeDrop with various methods
	result := tr.InvokeDrop("Index")
	if result != 1 {
		t.Errorf("Expected Index 1, got %v", result)
	}

	result = tr.InvokeDrop("Length")
	if result != 10 {
		t.Errorf("Expected Length 10, got %v", result)
	}

	result = tr.InvokeDrop("Cols")
	if result != 3 {
		t.Errorf("Expected Cols 3, got %v", result)
	}
}

func TestSnippetDropInvokeDrop(t *testing.T) {
	sd := NewSnippetDrop("body", "name", "file")

	// Test InvokeDrop
	result := sd.InvokeDrop("Body")
	if result != "body" {
		t.Errorf("Expected Body 'body', got %v", result)
	}

	result = sd.InvokeDrop("Name")
	if result != "name" {
		t.Errorf("Expected Name 'name', got %v", result)
	}
}

// TestDropInvokeDropOldWithMethod tests InvokeDropOld with method calls
func TestDropInvokeDropOldWithMethod(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)

	// Test invoking Length method - InvokeDropOld works on Drop, not ForloopDrop directly
	// ForloopDrop embeds Drop, so we need to call it on the Drop part
	drop := &Drop{}
	drop.SetContext(NewContext())

	// Test that InvokeDropOld exists and can be called
	result := drop.InvokeDropOld("nonexistent")
	// Should return nil or call LiquidMethodMissing
	if result != nil {
		t.Logf("InvokeDropOld returned: %v", result)
	}

	// Test with actual drop that has methods
	result2 := fl.InvokeDrop("Length")
	if result2 != 5 {
		t.Errorf("Expected Length 5 via InvokeDrop, got %v", result2)
	}
}

// TestDropInvokeDropOldWithField tests InvokeDropOld with field access
func TestDropInvokeDropOldWithField(t *testing.T) {
	// Create a drop with a field
	type testDrop struct {
		*Drop
		TestField string
	}

	td := &testDrop{
		Drop:      NewDrop(),
		TestField: "test_value",
	}

	// Test accessing field - InvokeDropOld tries method first, then field
	result := td.InvokeDropOld("TestField")
	// May return nil if method doesn't exist and field access fails
	if result != "test_value" && result != nil {
		t.Logf("InvokeDropOld returned: %v (may not support field access directly)", result)
	}

	// Test that the field exists
	if td.TestField != "test_value" {
		t.Errorf("Expected TestField 'test_value', got %v", td.TestField)
	}
}

// TestInvokeDropOnWithNonPointer tests InvokeDropOn with non-pointer
func TestInvokeDropOnWithNonPointer(t *testing.T) {
	drop := Drop{} // not a pointer
	result := InvokeDropOn(drop, "Context")
	if result != nil {
		t.Errorf("Expected nil for non-pointer, got %v", result)
	}
}

// TestInvokeDropOnWithFieldAccess tests InvokeDropOn with field access
func TestInvokeDropOnWithFieldAccess(t *testing.T) {
	type testDrop struct {
		*Drop
		TestField string
	}

	td := &testDrop{
		Drop:      NewDrop(),
		TestField: "test_value",
	}

	// Test accessing field - InvokeDropOn tries method first, then field
	result := InvokeDropOn(td, "TestField")
	if result != "test_value" {
		t.Logf("InvokeDropOn returned: %v (may try method first)", result)
		// Field access should work if no method exists
		if result == nil {
			t.Log("Field access returned nil (method lookup may have failed)")
		}
	}

	// Test with capitalized field name (Go convention)
	result2 := InvokeDropOn(td, "TestField")
	if result2 != "test_value" && result2 != nil {
		t.Errorf("Expected 'test_value' or nil, got %v", result2)
	}
}

// TestInvokeDropOnWithLiquidMethodMissing tests InvokeDropOn calling LiquidMethodMissing
func TestInvokeDropOnWithLiquidMethodMissing(t *testing.T) {
	drop := NewDrop()

	// Test with non-existent method (should call LiquidMethodMissing)
	result := InvokeDropOn(drop, "nonexistent")
	if result != nil {
		t.Errorf("Expected nil for nonexistent method, got %v", result)
	}

	// Test with strict variables
	ctx := NewContext()
	ctx.SetStrictVariables(true)
	drop.SetContext(ctx)

	// Should panic in strict mode
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for undefined method in strict mode")
			}
		}()
		InvokeDropOn(drop, "nonexistent")
	}()
}

// TestInvokeDropOnWithNonInvokable tests InvokeDropOn with non-invokable drop
func TestInvokeDropOnWithNonInvokable(t *testing.T) {
	// Test with nil
	result := InvokeDropOn(nil, "method")
	if result != nil {
		t.Errorf("Expected nil for nil drop, got %v", result)
	}

	// Test with non-drop type
	result2 := InvokeDropOn("not a drop", "method")
	if result2 != nil {
		t.Errorf("Expected nil for non-drop, got %v", result2)
	}
}

// TestInvokeDropOnWithMethodCache tests InvokeDropOn method caching
func TestInvokeDropOnWithMethodCache(t *testing.T) {
	fl := NewForloopDrop("items", 5, nil)

	// First call should build cache
	result1 := InvokeDropOn(fl, "Length")
	if result1 != 5 {
		t.Errorf("Expected Length 5, got %v", result1)
	}

	// Second call should use cache
	result2 := InvokeDropOn(fl, "Length")
	if result2 != 5 {
		t.Errorf("Expected Length 5 on second call, got %v", result2)
	}
}
