package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestInlineCommentTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewInlineCommentTag("inline_comment", "# comment", pc)
	if err != nil {
		t.Fatalf("NewInlineCommentTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected InlineCommentTag, got nil")
	}

	if !tag.Blank() {
		t.Error("Expected Blank to be true")
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestInlineCommentTagInvalid(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with newline without # on subsequent line
	_, err := NewInlineCommentTag("inline_comment", "# comment\ninvalid", pc)
	if err == nil {
		t.Error("Expected error for invalid inline comment with newline")
	}
	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}

	// Test with newline with # on subsequent line (should be valid)
	tag, err := NewInlineCommentTag("inline_comment", "# comment\n# more", pc)
	if err != nil {
		t.Fatalf("NewInlineCommentTag() with valid multiline error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected InlineCommentTag, got nil")
	}

	// Test with newline with whitespace and # (should be valid)
	tag2, err := NewInlineCommentTag("inline_comment", "# comment\n  # more", pc)
	if err != nil {
		t.Fatalf("NewInlineCommentTag() with whitespace and # error = %v", err)
	}
	if tag2 == nil {
		t.Fatal("Expected InlineCommentTag, got nil")
	}
}

func TestInlineCommentTagRenderToOutputBuffer(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewInlineCommentTag("inline_comment", "# comment", pc)
	if err != nil {
		t.Fatalf("NewInlineCommentTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	// Explicitly test RenderToOutputBuffer
	tag.RenderToOutputBuffer(ctx, &output)

	// Inline comment should render nothing
	if output != "" {
		t.Errorf("Expected empty output from RenderToOutputBuffer, got %q", output)
	}
}
