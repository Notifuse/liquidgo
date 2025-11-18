package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestRenderTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected RenderTag, got nil")
	}
}

func TestRenderTagWithWith(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' with var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	if tag.VariableNameExpr() == nil {
		t.Error("Expected variable name expression, got nil")
	}

	if tag.IsForLoop() {
		t.Error("Expected IsForLoop to be false for 'with', got true")
	}
}

func TestRenderTagWithFor(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' for items", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	if tag.VariableNameExpr() == nil {
		t.Error("Expected variable name expression, got nil")
	}

	if !tag.IsForLoop() {
		t.Error("Expected IsForLoop to be true for 'for', got false")
	}
}

func TestRenderTagWithAs(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' as alias", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	if tag.AliasName() != "alias" {
		t.Errorf("Expected alias name 'alias', got %q", tag.AliasName())
	}
}
