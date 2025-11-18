package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIncludeTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected IncludeTag, got nil")
	}
}

func TestIncludeTagWithAttributes(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template' key:value", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	if len(tag.Attributes()) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(tag.Attributes()))
	}
}

func TestIncludeTagWithWith(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template' with var", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	if tag.VariableNameExpr() == nil {
		t.Error("Expected variable name expression, got nil")
	}
}

func TestIncludeTagWithAs(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template' as alias", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	if tag.AliasName() != "alias" {
		t.Errorf("Expected alias name 'alias', got %q", tag.AliasName())
	}
}
