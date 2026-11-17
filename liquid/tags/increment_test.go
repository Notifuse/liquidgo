package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

func TestIncrementTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncrementTag("increment", "counter", pc)
	if err != nil {
		t.Fatal(err)
	}
	if tag == nil {
		t.Fatal("Expected IncrementTag, got nil")
	}

	if tag.VariableName() != "counter" {
		t.Errorf("Expected variable name 'counter', got %q", tag.VariableName())
	}

	ctx := liquid.NewContext()
	var output string

	// First call should output 0
	renderToString(tag, ctx, &output)
	if output != "0" {
		t.Errorf("Expected '0', got %q", output)
	}

	// Second call should output 1
	output = ""
	renderToString(tag, ctx, &output)
	if output != "1" {
		t.Errorf("Expected '1', got %q", output)
	}
}

// TestIncrementTagRenderToOutputBufferEdgeCases tests RenderToOutputBuffer edge cases
func TestIncrementTagRenderToOutputBufferEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIncrementTag("increment", "counter", pc)
	if err != nil {
		t.Fatal(err)
	}

	ctx := liquid.NewContext()

	// Test with empty scopes (should initialize)
	ctx.Scopes() // Ensure scopes exist
	var output string
	renderToString(tag, ctx, &output)

	// Should output 0 on first call
	if output != "0" {
		t.Errorf("Expected '0' on first call, got %q", output)
	}

	// Test multiple increments
	for i := 1; i <= 5; i++ {
		output = ""
		renderToString(tag, ctx, &output)
		expected := liquid.ToS(i, nil)
		if output != expected {
			t.Errorf("Expected %q on call %d, got %q", expected, i+1, output)
		}
	}
}
