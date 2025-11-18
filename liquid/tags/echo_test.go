package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestEchoTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewEchoTag("echo", "test", pc)
	if tag == nil {
		t.Fatal("Expected EchoTag, got nil")
	}

	ctx := liquid.NewContext()
	ctx.Set("test", "value")
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "value" {
		t.Errorf("Expected 'value', got %q", output)
	}
}
