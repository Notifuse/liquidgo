package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestUnlessTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "false", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected UnlessTag, got nil")
	}

	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}
}

func TestUnlessTagFalseCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "false", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	// Parse unless block with false condition (should render)
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
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

func TestUnlessTagTrueCondition(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "true", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	// Parse unless block with true condition (should not render)
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
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

func TestUnlessTagWithElse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "true", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	// Parse unless-else block
	tokenizer := pc.NewTokenizer("unless content {% else %} else content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks (unless, else), got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Since unless condition is true, it won't render, so else should render
	if output != " else content " {
		t.Errorf("Expected output ' else content ', got %q", output)
	}
}

func TestUnlessTagWithNilValue(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "nil_var", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	// Don't set nil_var, so it will be nil
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// nil value should render (unless renders when false/nil)
	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestUnlessTagWithEmptyString(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", `""`, pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Empty string should render (unless renders when false/empty)
	if output != "content " {
		t.Errorf("Expected output 'content ', got %q", output)
	}
}

func TestUnlessTagWithErrorInEvaluation(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewUnlessTag("unless", "invalid_var", pc)
	if err != nil {
		t.Fatalf("NewUnlessTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	tokenizer := pc.NewTokenizer("content {% endunless %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Create a context that will cause an error during evaluation
	// Use a variable that causes an error when evaluated
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should handle error gracefully (output may contain error message or be empty)
	// The exact behavior depends on error handling implementation
}
