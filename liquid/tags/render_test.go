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

func TestRenderTagParse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	// Parse is a no-op for render tags
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestRenderTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'nonexistent'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
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

func TestRenderTagRenderToOutputBufferComprehensive(t *testing.T) {
	env := liquid.NewEnvironment()
	pc := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})

	// Test with non-string template name
	tag2, _ := NewRenderTag("render", "123", pc)
	ctx2 := liquid.NewContext()
	var output2 string
	tag2.RenderToOutputBuffer(ctx2, &output2)
	// Should handle error gracefully
	_ = output2

	// Test with with clause
	tag3, err := NewRenderTag("render", "'template' with person", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with 'with' error = %v", err)
	}
	ctx3 := liquid.NewContext()
	ctx3.Set("person", map[string]interface{}{"name": "Alice"})
	var output3 string
	tag3.RenderToOutputBuffer(ctx3, &output3)
	_ = output3

	// Test with for clause
	tag4, err := NewRenderTag("render", "'template' for items", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with 'for' error = %v", err)
	}
	ctx4 := liquid.NewContext()
	ctx4.Set("items", []interface{}{map[string]interface{}{"name": "Item1"}, map[string]interface{}{"name": "Item2"}})
	var output4 string
	tag4.RenderToOutputBuffer(ctx4, &output4)
	_ = output4

	// Test with as clause
	tag5, err := NewRenderTag("render", "'template' as alias_var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with 'as' error = %v", err)
	}
	ctx5 := liquid.NewContext()
	var output5 string
	tag5.RenderToOutputBuffer(ctx5, &output5)
	_ = output5

	// Test with attributes
	tag6, err := NewRenderTag("render", "'template' key:value", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with attributes error = %v", err)
	}
	ctx6 := liquid.NewContext()
	var output6 string
	tag6.RenderToOutputBuffer(ctx6, &output6)
	_ = output6

	// Test with variable template name
	tag7, err := NewRenderTag("render", "template_var", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() with variable name error = %v", err)
	}
	ctx7 := liquid.NewContext()
	ctx7.Set("template_var", "template_name")
	var output7 string
	tag7.RenderToOutputBuffer(ctx7, &output7)
	_ = output7
}

func TestRenderTagTemplateNameExpr(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template'", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	templateNameExpr := tag.TemplateNameExpr()
	if templateNameExpr == nil {
		t.Error("Expected TemplateNameExpr() to return non-nil expression")
	}
}

func TestRenderTagAttributes(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRenderTag("render", "'template' key:value", pc)
	if err != nil {
		t.Fatalf("NewRenderTag() error = %v", err)
	}

	attributes := tag.Attributes()
	if attributes == nil {
		t.Error("Expected Attributes() to return non-nil map")
	}
	if len(attributes) == 0 {
		t.Error("Expected Attributes() to contain attributes")
	}
}
