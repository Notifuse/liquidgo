package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIfchangedTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected IfchangedTag, got nil")
	}
}

func TestIfchangedTagRendersWhenChanged(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}

	// Parse ifchanged block
	tokenizer := pc.NewTokenizer("content {% endifchanged %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// First render should output content
	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestIfchangedTagDoesNotRenderWhenUnchanged(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}

	// Parse ifchanged block
	tokenizer := pc.NewTokenizer("content {% endifchanged %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string

	// First render
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}

	// Second render with same content should not output
	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "" {
		t.Errorf("Expected empty output on second render, got %q", output)
	}
}

func TestIfchangedTagRendersWhenContentChanges(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}

	// Parse ifchanged block
	tokenizer := pc.NewTokenizer("{{item}} {% endifchanged %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string

	// First render with item=1
	ctx.Set("item", 1)
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "1 " {
		t.Errorf("Expected output '1 ', got %q", output)
	}

	// Second render with same item=1 should not output
	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}

	// Third render with item=2 should output
	ctx.Set("item", 2)
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "2 " {
		t.Errorf("Expected output '2 ', got %q", output)
	}
}
