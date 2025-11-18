package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestBreakTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewBreakTag("break", "", pc)
	if tag == nil {
		t.Fatal("Expected BreakTag, got nil")
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if !ctx.Interrupt() {
		t.Error("Expected interrupt to be set")
	}
}
