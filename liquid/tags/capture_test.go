package tags

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestCaptureTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaptureTag("capture", "var", pc)
	if err != nil {
		t.Fatalf("NewCaptureTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected CaptureTag, got nil")
	}

	if tag.To() != "var" {
		t.Errorf("Expected To 'var', got %q", tag.To())
	}

	if !tag.Blank() {
		t.Error("Expected Blank to be true")
	}
}

func TestCaptureTagSyntaxError(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	_, err := NewCaptureTag("capture", "", pc)
	if err == nil {
		t.Fatal("Expected error for invalid syntax")
	}
	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestCaptureTagRender(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	// Use NewCaptureTag to properly create the tag
	tag, err := NewCaptureTag("capture", "var", pc)
	if err != nil {
		t.Fatalf("NewCaptureTag() error = %v", err)
	}

	// Parse the block body
	tokenizer := pc.NewTokenizer("test content {% endcapture %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Check that variable was captured
	val := ctx.Get("var")
	if val == nil {
		t.Error("Expected variable to be captured")
	} else if valStr, ok := val.(string); !ok {
		t.Errorf("Expected string value, got %T", val)
	} else if valStr != "test content " {
		t.Errorf("Expected variable value 'test content ', got %q", valStr)
	}

	// Capture tag should not output anything
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestCaptureTagWithHyphenInVariableName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewCaptureTag("capture", "this-thing", pc)
	if err != nil {
		t.Fatalf("NewCaptureTag() error = %v", err)
	}

	if tag.To() != "this-thing" {
		t.Errorf("Expected To 'this-thing', got %q", tag.To())
	}
}

func TestCaptureTagWithQuotedVariableName(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})

	// Test with single quotes
	tag, err := NewCaptureTag("capture", "'var'", pc)
	if err != nil {
		t.Fatalf("NewCaptureTag() error = %v", err)
	}
	if tag.To() != "var" {
		t.Errorf("Expected To 'var', got %q", tag.To())
	}

	// Test with double quotes
	tag, err = NewCaptureTag("capture", `"var"`, pc)
	if err != nil {
		t.Fatalf("NewCaptureTag() error = %v", err)
	}
	if tag.To() != "var" {
		t.Errorf("Expected To 'var', got %q", tag.To())
	}

	// Test with quoted string containing hyphen
	tag, err = NewCaptureTag("capture", "'this-thing'", pc)
	if err != nil {
		t.Fatalf("NewCaptureTag() error = %v", err)
	}
	if tag.To() != "this-thing" {
		t.Errorf("Expected To 'this-thing', got %q", tag.To())
	}
}
