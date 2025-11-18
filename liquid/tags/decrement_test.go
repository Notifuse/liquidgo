package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestDecrementTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewDecrementTag("decrement", "counter", pc)
	if tag == nil {
		t.Fatal("Expected DecrementTag, got nil")
	}

	if tag.VariableName() != "counter" {
		t.Errorf("Expected variable name 'counter', got %q", tag.VariableName())
	}

	ctx := liquid.NewContext()
	var output string

	// First call should output -1
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "-1" {
		t.Errorf("Expected '-1', got %q", output)
	}

	// Second call should output -2
	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "-2" {
		t.Errorf("Expected '-2', got %q", output)
	}
}
