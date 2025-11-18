package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestCaseTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected CaseTag, got nil")
	}
}

func TestCaseTagWithWhen(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with when
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 1 {
		t.Errorf("Expected 1 block, got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 1)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "one" {
		t.Errorf("Expected output 'one', got %q", output)
	}
}

func TestCaseTagWithWhenAndElse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with when and else
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% else %}other{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks (when, else), got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 2)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Should render else block
	if output != "other" {
		t.Errorf("Expected output 'other', got %q", output)
	}
}

func TestCaseTagWithMultipleWhen(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaseTag("case", "var", pc)
	if err != nil {
		t.Fatalf("NewCaseTag() error = %v", err)
	}

	// Parse case block with multiple when
	tokenizer := pc.NewTokenizer("{% when 1 %}one{% when 2 %}two{% endcase %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(tag.Blocks()) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(tag.Blocks()))
	}

	ctx := liquid.NewContext()
	ctx.Set("var", 2)
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "two" {
		t.Errorf("Expected output 'two', got %q", output)
	}
}
