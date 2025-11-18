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
