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

func TestCycleTagParseMarkup(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with named syntax
	tag, err := NewCycleTag("cycle", "colors: red, blue", pc)
	if err != nil {
		t.Fatalf("NewCycleTag() with named syntax error = %v", err)
	}
	if !tag.Named() {
		t.Error("Expected Named to be true for named syntax")
	}

	// Test with simple syntax
	tag2, err := NewCycleTag("cycle", `"one", "two"`, pc)
	if err != nil {
		t.Fatalf("NewCycleTag() with simple syntax error = %v", err)
	}
	if len(tag2.Variables()) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(tag2.Variables()))
	}

	// Test with invalid syntax (empty or non-matching)
	_, err3 := NewCycleTag("cycle", "", pc)
	if err3 == nil {
		t.Error("Expected error for empty cycle syntax")
	}
}

func TestCycleTagVariablesFromString(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with comma-separated values
	tag, err := NewCycleTag("cycle", `"a", "b", "c"`, pc)
	if err != nil {
		t.Fatalf("NewCycleTag() error = %v", err)
	}
	if len(tag.Variables()) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(tag.Variables()))
	}

	// Test with empty values (should be skipped)
	tag2, err := NewCycleTag("cycle", `"a", , "b"`, pc)
	if err != nil {
		t.Fatalf("NewCycleTag() error = %v", err)
	}
	// Empty values should be skipped
	_ = tag2
}

func TestCycleTagRenderToOutputBufferWithArray(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCycleTag("cycle", `"one", "two"`, pc)
	if err != nil {
		t.Fatalf("NewCycleTag() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "one" {
		t.Errorf("Expected 'one', got %q", output)
	}

	// Test cycle wraps around
	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "two" {
		t.Errorf("Expected 'two', got %q", output)
	}

	output = ""
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "one" {
		t.Errorf("Expected 'one' (wrapped), got %q", output)
	}
}
