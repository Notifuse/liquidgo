package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestCycleTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	// Use quoted strings (string literals), not unquoted variable names
	tag, err := NewCycleTag("cycle", `"one", "two", "three"`, pc)
	if err != nil {
		t.Fatalf("NewCycleTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected CycleTag, got nil")
	}

	if len(tag.Variables()) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(tag.Variables()))
	}

	ctx := liquid.NewContext()
	var output string

	// First call should output first value
	tag.RenderToOutputBuffer(ctx, &output)
	// Output should be "one" (the first literal string value)
	if output != "one" {
		t.Errorf("Expected \"one\", got %q", output)
	}

	// Second call should output second value
	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "two" {
		t.Errorf("Expected \"two\", got %q", output)
	}
}

func TestCycleTagNamed(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCycleTag("cycle", "colors: red, blue, green", pc)
	if err != nil {
		t.Fatalf("NewCycleTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected CycleTag, got nil")
	}

	if !tag.Named() {
		t.Error("Expected Named to be true")
	}

	if len(tag.Variables()) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(tag.Variables()))
	}
}
