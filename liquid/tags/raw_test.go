package tags

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestRawTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRawTag("raw", "", pc)
	if err != nil {
		t.Fatalf("NewRawTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected RawTag, got nil")
	}

	// Create a tokenizer with raw content and endraw
	source := "Hello {{ world }} {% endraw %}"
	tokenizer := pc.NewTokenizer(source, false, nil, false)

	// Parse the tag
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	// Raw tag should output the content as-is (without the endraw tag)
	if !strings.Contains(output, "Hello") {
		t.Errorf("Expected output to contain 'Hello', got %q", output)
	}
}

func TestRawTagInvalidMarkup(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	_, err := NewRawTag("raw", "invalid", pc)
	if err == nil {
		t.Fatal("Expected error for invalid markup")
	}
	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestRawTagNodelist(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRawTag("raw", "", pc)
	if err != nil {
		t.Fatalf("NewRawTag() error = %v", err)
	}

	// Parse raw tag with content
	source := "Hello {{ world }} {% endraw %}"
	tokenizer := pc.NewTokenizer(source, false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	nodelist := tag.Nodelist()
	if len(nodelist) != 1 {
		t.Errorf("Expected nodelist length 1, got %d", len(nodelist))
	}
	body, ok := nodelist[0].(string)
	if !ok {
		t.Errorf("Expected string body, got %T", nodelist[0])
	} else if !strings.Contains(body, "Hello") {
		t.Errorf("Expected body to contain 'Hello', got %q", body)
	}
}

func TestRawTagBlank(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewRawTag("raw", "", pc)
	if err != nil {
		t.Fatalf("NewRawTag() error = %v", err)
	}

	// Test with empty body
	if !tag.Blank() {
		t.Error("Expected Blank to be true for empty raw tag")
	}

	// Parse raw tag with content
	source := "Hello {% endraw %}"
	tokenizer := pc.NewTokenizer(source, false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should not be blank after parsing content
	if tag.Blank() {
		t.Error("Expected Blank to be false for raw tag with content")
	}
}
