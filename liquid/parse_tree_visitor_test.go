package liquid

import (
	"reflect"
	"testing"
)

func TestParseTreeVisitorBasic(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	node := NewTag("test", "", pc)
	ptv := NewParseTreeVisitor(node, nil)
	if ptv == nil {
		t.Fatal("Expected ParseTreeVisitor, got nil")
	}
}

func TestParseTreeVisitorChildren(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	// Create a tag with nodelist
	tag := NewTag("test", "", pc)
	tag.SetNodelist([]interface{}{"hello", "world"})

	ptv := NewParseTreeVisitor(tag, nil)
	children := ptv.children()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestParseTreeVisitorVisit(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	tag := NewTag("test", "", pc)
	tag.SetNodelist([]interface{}{"hello"})

	ptv := NewParseTreeVisitor(tag, nil)
	result := ptv.Visit(nil)
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}
}

func TestForParseTreeVisitor(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	node := NewTag("test", "", pc)
	ptv := ForParseTreeVisitor(node, nil)
	if ptv == nil {
		t.Fatal("Expected ParseTreeVisitor, got nil")
	}
}

// TestForParseTreeVisitorNodeSpecific tests node-specific ParseTreeVisitor
func TestForParseTreeVisitorNodeSpecific(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}

	// Create a node with a ParseTreeVisitor method
	node := &nodeWithParseTreeVisitor{tag: NewTag("test", "", pc)}
	callbacks := make(map[reflect.Type]ParseTreeVisitorCallback)
	ptv := ForParseTreeVisitor(node, callbacks)
	if ptv == nil {
		t.Fatal("Expected ParseTreeVisitor, got nil")
	}
}

// nodeWithParseTreeVisitor is a test node that implements ParseTreeVisitor method
type nodeWithParseTreeVisitor struct {
	tag *Tag
}

func (n *nodeWithParseTreeVisitor) ParseTreeVisitor(callbacks map[reflect.Type]ParseTreeVisitorCallback) *ParseTreeVisitor {
	return NewParseTreeVisitor(n.tag, callbacks)
}

func TestParseTreeVisitorAddCallbackFor(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	tag := NewTag("test", "", pc)
	tag.SetNodelist([]interface{}{"hello", "world"})
	ptv := NewParseTreeVisitor(tag, nil)

	callback := func(node interface{}, context interface{}) (interface{}, interface{}) {
		return node, context
	}

	// Add callback for string type
	ptv.AddCallbackFor([]interface{}{"test"}, callback)

	// Visit should use the callback
	result := ptv.Visit(nil)
	if len(result) == 0 {
		t.Error("Expected result from visit")
	}

	// Test with multiple types
	ptv2 := NewParseTreeVisitor(tag, nil)
	ptv2.AddCallbackFor([]interface{}{"test", 42}, callback)
	result2 := ptv2.Visit(nil)
	if len(result2) == 0 {
		t.Error("Expected result from visit with multiple types")
	}

	// Test with nil type
	ptv3 := NewParseTreeVisitor(tag, nil)
	ptv3.AddCallbackFor([]interface{}{nil}, callback)
	result3 := ptv3.Visit(nil)
	if len(result3) == 0 {
		t.Error("Expected result from visit with nil type")
	}
}

func TestParseTreeVisitorAddCallbackForWithChaining(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	node := NewTag("test", "", pc)
	ptv := NewParseTreeVisitor(node, nil)

	callback := func(node interface{}, context interface{}) (interface{}, interface{}) {
		return node, context
	}

	// Test chaining
	result := ptv.AddCallbackFor([]interface{}{"test"}, callback)
	if result != ptv {
		t.Error("Expected AddCallbackFor to return self for chaining")
	}
}

func TestForParseTreeVisitorWithNilNode(t *testing.T) {
	ptv := ForParseTreeVisitor(nil, nil)
	if ptv == nil {
		t.Fatal("Expected ParseTreeVisitor, got nil")
	}
}

func TestParseTreeVisitorChildrenWithInvalidNode(t *testing.T) {
	ptv := NewParseTreeVisitor(42, nil) // int doesn't have Nodelist method
	children := ptv.children()
	if len(children) != 0 {
		t.Errorf("Expected empty children for invalid node, got %d", len(children))
	}
}

func TestParseTreeVisitorVisitWithCallback(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum, env: NewEnvironment()}
	tag := NewTag("test", "", pc)
	tag.SetNodelist([]interface{}{"hello", "world"})

	callbacks := make(map[reflect.Type]ParseTreeVisitorCallback)
	callbacks[reflect.TypeOf("")] = func(node interface{}, context interface{}) (interface{}, interface{}) {
		return "processed:" + node.(string), context
	}

	ptv := NewParseTreeVisitor(tag, callbacks)
	result := ptv.Visit(nil)
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
}
