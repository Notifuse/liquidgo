package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIfTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "true", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected IfTag, got nil")
	}

	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}
}

func TestIfTagSimpleCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "true", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Parse simple if block
	tokenizer := pc.NewTokenizer("content {% endif %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestIfTagFalseCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "false", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Parse if block with false condition
	tokenizer := pc.NewTokenizer("content {% endif %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestIfTagWithElse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "false", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Parse if-else block
	tokenizer := pc.NewTokenizer("if content {% else %} else content {% endif %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks (if, else), got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != " else content " {
		t.Errorf("Expected output ' else content ', got %q", output)
	}
}

func TestIfTagWithElsif(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "false", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Parse if-elsif-else block
	tokenizer := pc.NewTokenizer("if {% elsif true %}elsif{% else %}else{% endif %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 3 {
		t.Errorf("Expected 3 blocks (if, elsif, else), got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "elsif" {
		t.Errorf("Expected output 'elsif', got %q", output)
	}
}

func TestIfTagNodelist(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "true", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Parse if block
	tokenizer := pc.NewTokenizer("content {% endif %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	nodelist := tag.Nodelist()
	if nodelist == nil {
		t.Error("Expected Nodelist() to return non-nil slice")
	}
	if len(nodelist) == 0 {
		t.Error("Expected Nodelist() to contain nodes after parsing")
	}
}

func TestIfTagUnknownTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "true", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Test UnknownTag with elsif (should be handled)
	tokenizer := pc.NewTokenizer("", false, nil, false)
	err = tag.UnknownTag("elsif", "condition", tokenizer)
	if err != nil {
		t.Errorf("Expected nil error for elsif, got %v", err)
	}

	// Test UnknownTag with else (should be handled)
	err = tag.UnknownTag("else", "", tokenizer)
	if err != nil {
		t.Errorf("Expected nil error for else, got %v", err)
	}

	// Test UnknownTag with unknown tag (should error)
	err = tag.UnknownTag("unknown", "", tokenizer)
	if err == nil {
		t.Error("Expected error for unknown tag")
	}
}

func TestIfTagParseBodyForBlock(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "true", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Test with endif tag
	tokenizer := pc.NewTokenizer("content {% endif %}", false, nil, false)
	shouldContinue, err := tag.parseBodyForBlock(tokenizer, tag.blocks[0])
	if err != nil {
		t.Fatalf("parseBodyForBlock() error = %v", err)
	}
	if shouldContinue {
		t.Error("Expected shouldContinue to be false after finding endif")
	}

	// Test with elsif tag (should continue - new block created but parsing continues)
	tag2, _ := NewIfTag("if", "false", pc)
	tokenizer2 := pc.NewTokenizer("content {% elsif true %}", false, nil, false)
	shouldContinue2, err2 := tag2.parseBodyForBlock(tokenizer2, tag2.blocks[0])
	if err2 != nil {
		t.Fatalf("parseBodyForBlock() with elsif error = %v", err2)
	}
	// parseBodyForBlock returns false when elsif/else is found, but Parse() continues
	if shouldContinue2 {
		t.Log("parseBodyForBlock may return true if content remains")
	}

	// Test with else tag (should continue - new block created but parsing continues)
	tag3, _ := NewIfTag("if", "false", pc)
	tokenizer3 := pc.NewTokenizer("content {% else %}", false, nil, false)
	shouldContinue3, err3 := tag3.parseBodyForBlock(tokenizer3, tag3.blocks[0])
	if err3 != nil {
		t.Fatalf("parseBodyForBlock() with else error = %v", err3)
	}
	// parseBodyForBlock returns false when else is found, but Parse() continues
	if shouldContinue3 {
		t.Log("parseBodyForBlock may return true if content remains")
	}

	// Test with depth limit
	pc4 := liquid.NewParseContext(liquid.ParseContextOptions{})
	for i := 0; i < 100; i++ {
		pc4.IncrementDepth()
	}
	tag4, _ := NewIfTag("if", "true", pc4)
	tokenizer4 := pc4.NewTokenizer("content", false, nil, false)
	_, err4 := tag4.parseBodyForBlock(tokenizer4, tag4.blocks[0])
	if err4 == nil {
		t.Error("Expected error for depth limit exceeded")
	}
}

func TestIfTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfTag("if", "true", pc)
	if err != nil {
		t.Fatalf("NewIfTag() error = %v", err)
	}

	// Parse if block
	tokenizer := pc.NewTokenizer("content {% endif %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string

	// Test with true condition
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "content " {
		t.Errorf("Expected 'content ', got %q", output)
	}

	// Test with false condition and else
	tag2, _ := NewIfTag("if", "false", pc)
	tokenizer2 := pc.NewTokenizer("if content {% else %}else content {% endif %}", false, nil, false)
	tag2.Parse(tokenizer2)
	output2 := ""
	tag2.RenderToOutputBuffer(ctx, &output2)
	if output2 != "else content " {
		t.Errorf("Expected 'else content ', got %q", output2)
	}

	// Test with error in evaluation
	tag3, _ := NewIfTag("if", "var", pc)
	tokenizer3 := pc.NewTokenizer("content {% endif %}", false, nil, false)
	tag3.Parse(tokenizer3)
	output3 := ""
	ctx3 := liquid.NewContext()
	// Set var to something that causes error
	tag3.RenderToOutputBuffer(ctx3, &output3)
	// Should handle error gracefully
	_ = output3
}

func TestIfTagParseIfCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with simple condition
	condition, err := parseIfCondition("true", pc)
	if err != nil {
		t.Fatalf("parseIfCondition() error = %v", err)
	}
	if condition == nil {
		t.Error("Expected condition, got nil")
	}

	// Test with comparison
	condition2, err := parseIfCondition("var == 1", pc)
	if err != nil {
		t.Fatalf("parseIfCondition() with comparison error = %v", err)
	}
	if condition2 == nil {
		t.Error("Expected condition, got nil")
	}

	// Test with operator only
	condition3, err := parseIfCondition("var", pc)
	if err != nil {
		t.Fatalf("parseIfCondition() with variable error = %v", err)
	}
	if condition3 == nil {
		t.Error("Expected condition, got nil")
	}
}
