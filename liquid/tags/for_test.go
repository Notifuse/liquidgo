package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestForTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected ForTag, got nil")
	}

	if tag.VariableName() != "item" {
		t.Errorf("Expected variable name 'item', got %q", tag.VariableName())
	}
}

func TestForTagSimpleLoop(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "1 2 3 " {
		t.Errorf("Expected output '1 2 3 ', got %q", output)
	}
}

func TestForTagEmptyCollection(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block with else
	tokenizer := pc.NewTokenizer("content {% else %}empty{% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "empty" {
		t.Errorf("Expected output 'empty', got %q", output)
	}
}

func TestForTagReversed(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array reversed", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "3 2 1 " {
		t.Errorf("Expected output '3 2 1 ', got %q", output)
	}
}

func TestForTagWithLimit(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array limit:2", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3, 4, 5})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "1 2 " {
		t.Errorf("Expected output '1 2 ', got %q", output)
	}
}

func TestForTagWithOffset(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewForTag("for", "item in array offset:2 limit:2", pc)
	if err != nil {
		t.Fatalf("NewForTag() error = %v", err)
	}

	// Parse for block
	tokenizer := pc.NewTokenizer("{{item}} {% endfor %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	ctx.Set("array", []interface{}{1, 2, 3, 4, 5})
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "3 4 " {
		t.Errorf("Expected output '3 4 ', got %q", output)
	}
}
