package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIncrementTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewIncrementTag("increment", "counter", pc)
	if tag == nil {
		t.Fatal("Expected IncrementTag, got nil")
	}

	if tag.VariableName() != "counter" {
		t.Errorf("Expected variable name 'counter', got %q", tag.VariableName())
	}

	ctx := liquid.NewContext()
	var output string

	// First call should output 0
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "0" {
		t.Errorf("Expected '0', got %q", output)
	}

	// Second call should output 1
	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "1" {
		t.Errorf("Expected '1', got %q", output)
	}
}
