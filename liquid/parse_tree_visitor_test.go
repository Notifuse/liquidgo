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

