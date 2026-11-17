package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

func TestDecrementTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewDecrementTag("decrement", "counter", pc)
	if err != nil {
		t.Fatal(err)
	}
	if tag == nil {
		t.Fatal("Expected DecrementTag, got nil")
	}

	if tag.VariableName() != "counter" {
		t.Errorf("Expected variable name 'counter', got %q", tag.VariableName())
	}

	ctx := liquid.NewContext()
	var output string

	// First call should output -1
	renderToString(tag, ctx, &output)
	if output != "-1" {
		t.Errorf("Expected '-1', got %q", output)
	}

	// Second call should output -2
	output = ""
	renderToString(tag, ctx, &output)
	if output != "-2" {
		t.Errorf("Expected '-2', got %q", output)
	}
}

// TestDecrementTagRenderToOutputBufferEdgeCases tests RenderToOutputBuffer edge cases
func TestDecrementTagRenderToOutputBufferEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewDecrementTag("decrement", "counter", pc)
	if err != nil {
		t.Fatal(err)
	}

	ctx := liquid.NewContext()

	// Test with empty scopes (should initialize)
	ctx.Scopes() // Ensure scopes exist
	var output string
	renderToString(tag, ctx, &output)

	// Should output -1 on first call
	if output != "-1" {
		t.Errorf("Expected '-1' on first call, got %q", output)
	}

	// Test multiple decrements
	for i := 2; i <= 5; i++ {
		output = ""
		renderToString(tag, ctx, &output)
		expected := liquid.ToS(-i, nil)
		if output != expected {
			t.Errorf("Expected %q on call %d, got %q", expected, i, output)
		}
	}
}
