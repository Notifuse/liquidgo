package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestContinueTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewContinueTag("continue", "", pc)
	if tag == nil {
		t.Fatal("Expected ContinueTag, got nil")
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if !ctx.Interrupt() {
		t.Error("Expected interrupt to be set")
	}
}
