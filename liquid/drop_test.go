package liquid

import (
	gocontext "context"
	"errors"
	"testing"
)

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

// Note: SnippetDrop tests removed - snippet tag was removed in Shopify Liquid v5.11.0

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

// keysDrop declares the keys a and err; a missing key is not found.
type keysDrop struct{ calls int }

func (d *keysDrop) LiquidGet(ctx *Context, key string) (interface{}, bool, error) {
	d.calls++
	switch key {
	case "a":
		return "A", true, nil
	case "err":
		return nil, false, NewArgumentError("bad key")
	case "ctx":
		return ctx.Get("x"), true, nil
	}
	return nil, false, nil
}

func renderDrop(t *testing.T, src string, data map[string]interface{}, opts ...RenderOption) Result {
	t.Helper()
	tmpl, err := ParseTemplate(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := tmpl.Render(gocontext.Background(), data, opts...)
	return res
}

// A template reads the keys a drop declares, with the context of the render;
// a missing key is nil, or an UndefinedDropMethod with strict variables; an
// error the drop returns is recorded.
func TestADropGivesTheKeysItDeclares(t *testing.T) {
	data := map[string]interface{}{"d": &keysDrop{}, "x": "X"}
	res := renderDrop(t, "{{ d.a }}|{{ d.ctx }}|{{ d.nope }}|{{ d.A }}", data)
	if res.Output != "A|X||" || len(res.Errors) != 0 {
		t.Errorf("lax: %q %v", res.Output, res.Errors)
	}
	res = renderDrop(t, "{{ d.nope }}", data, WithStrictVariables())
	var undefined *UndefinedDropMethod
	if len(res.Errors) != 1 || !errors.As(res.Errors[0], &undefined) {
		t.Errorf("strict: errors %v, want an UndefinedDropMethod", res.Errors)
	}
	res = renderDrop(t, "[{{ d.err }}]", data)
	var argument *ArgumentError
	if len(res.Errors) != 1 || !errors.As(res.Errors[0], &argument) {
		t.Errorf("an error of the drop: errors %v, want the ArgumentError", res.Errors)
	}
}

// The array filters read the keys of drops, a missing one being undefined with
// strict variables; a []T whose *T is a drop is read too.
func TestArrayFiltersReadTheKeysOfDrops(t *testing.T) {
	data := map[string]interface{}{"ds": []*keysDrop{{}, {}}, "vs": []keysDrop{{}}}
	res := renderDrop(t, `{{ ds | map: "a" | join: "" }}|{{ vs | map: "a" | join }}|{{ ds | where: "a", "A" | size }}`, data)
	if res.Output != "AA|A|2" || len(res.Errors) != 0 {
		t.Errorf("got %q %v", res.Output, res.Errors)
	}
	res = renderDrop(t, `{{ ds | map: "nope" | join }}`, data, WithStrictVariables())
	var undefined *UndefinedDropMethod
	if len(res.Errors) != 1 || !errors.As(res.Errors[0], &undefined) {
		t.Errorf("strict: errors %v, want an UndefinedDropMethod", res.Errors)
	}
}

var errDBDown = errors.New("db down")

// failingDrop returns a Go error that is not a Liquid error.
type failingDrop struct{}

func (failingDrop) LiquidGet(*Context, string) (interface{}, bool, error) {
	return nil, true, errDBDown
}

// A Go error a drop returns is recorded as an internal error, which keeps it
// as its cause, and the render goes on, as Ruby rescues a StandardError.
func TestAGoErrorOfADropIsAnInternalError(t *testing.T) {
	tmpl, err := ParseTemplate(`{{ d.orders }}|{{ ds | map: "orders" | size }}|after`, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"d": failingDrop{}, "ds": []failingDrop{{}}})
	if err != nil {
		t.Fatalf("the render ended: %v", err)
	}
	var internal *InternalError
	if len(res.Errors) != 2 || !errors.As(res.Errors[0], &internal) || !errors.Is(res.Errors[0], errDBDown) {
		t.Errorf("errors %v, want two internal errors caused by %v", res.Errors, errDBDown)
	}
	if res.Output != "Liquid error: internal|Liquid error: internal|after" {
		t.Errorf("output %q", res.Output)
	}
}

// A drop value that has no address (in a map, in an array) is read through a
// copy of it, when its pointer is the drop.
func TestADropValueWithoutAnAddressIsRead(t *testing.T) {
	data := map[string]interface{}{"v": keysDrop{}, "a": [1]keysDrop{}}
	res := renderDrop(t, `{{ v.a }}|{{ a[0].a }}|{{ a | map: "a" | join }}`, data)
	if res.Output != "A|A|A" || len(res.Errors) != 0 {
		t.Errorf("got %q %v", res.Output, res.Errors)
	}
}
