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
