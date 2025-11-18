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

