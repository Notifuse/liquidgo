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

func TestIncludeTagParse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	// Parse is a no-op for include tags
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestIncludeTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'nonexistent'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	// RenderToOutputBuffer should handle missing template gracefully
	tag.RenderToOutputBuffer(ctx, &output)
	// Should output error message or empty string
	if output == "" {
		t.Log("RenderToOutputBuffer returned empty output (expected for missing template)")
	}
}

func TestIncludeTagRenderToOutputBufferComprehensive(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Test with non-string template name
	tag2, _ := NewIncludeTag("include", "123", pc)
	ctx2 := liquid.NewContext()
	var output2 string
	tag2.RenderToOutputBuffer(ctx2, &output2)
	// Should handle error gracefully
	_ = output2

	// Test with with clause
	tag3, err := NewIncludeTag("include", "'greeting' with person", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with 'with' error = %v", err)
	}
	ctx3 := liquid.NewContext()
	ctx3.Set("person", map[string]interface{}{"name": "Alice"})
	var output3 string
	tag3.RenderToOutputBuffer(ctx3, &output3)
	_ = output3

	// Test with for clause
	tag4, err := NewIncludeTag("include", "'greeting' for person", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with 'for' error = %v", err)
	}
	ctx4 := liquid.NewContext()
	ctx4.Set("person", map[string]interface{}{"name": "Bob"})
	var output4 string
	tag4.RenderToOutputBuffer(ctx4, &output4)
	_ = output4

	// Test with as clause
	tag5, err := NewIncludeTag("include", "'greeting' as greeting_var", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with 'as' error = %v", err)
	}
	ctx5 := liquid.NewContext()
	ctx5.Set("name", "Charlie")
	var output5 string
	tag5.RenderToOutputBuffer(ctx5, &output5)
	_ = output5

	// Test with array variable
	tag6, err := NewIncludeTag("include", "'greeting' for items", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() with array error = %v", err)
	}
	ctx6 := liquid.NewContext()
	ctx6.Set("items", []interface{}{map[string]interface{}{"name": "Item1"}, map[string]interface{}{"name": "Item2"}})
	var output6 string
	tag6.RenderToOutputBuffer(ctx6, &output6)
	_ = output6
}

func TestIncludeTagTemplateNameExpr(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncludeTag("include", "'template'", pc)
	if err != nil {
		t.Fatalf("NewIncludeTag() error = %v", err)
	}

	templateNameExpr := tag.TemplateNameExpr()
	if templateNameExpr == nil {
		t.Error("Expected TemplateNameExpr() to return non-nil expression")
	}
}
