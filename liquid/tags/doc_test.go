package tags

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestDocTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewDocTag("doc", "", pc)
	if err != nil {
		t.Fatalf("NewDocTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected DocTag, got nil")
	}

	if !tag.Blank() {
		t.Error("Expected Blank to be true initially")
	}

	ctx := liquid.NewContext()
	var output string
	tag.RenderToOutputBuffer(ctx, &output)

	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestDocTagParse(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tokenizer := pc.NewTokenizer("Documentation content {% enddoc %}", false, nil, false)

	tag, err := NewDocTag("doc", "", pc)
	if err != nil {
		t.Fatalf("NewDocTag() error = %v", err)
	}

	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check body content
	nodelist := tag.Nodelist()
	if len(nodelist) != 1 {
		t.Errorf("Expected nodelist length 1, got %d", len(nodelist))
	}

	body, ok := nodelist[0].(string)
	if !ok {
		t.Errorf("Expected string body, got %T", nodelist[0])
	} else if !strings.Contains(body, "Documentation content") {
		t.Errorf("Expected body to contain 'Documentation content', got %q", body)
	}
}

func TestDocTagBlank(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	// Create tokenizer with just the end tag
	tokenizer := pc.NewTokenizer("{% enddoc %}", false, nil, false)

	tag, err := NewDocTag("doc", "", pc)
	if err != nil {
		t.Fatalf("NewDocTag() error = %v", err)
	}

	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Body should be empty (or just whitespace from end tag)
	if !tag.Blank() {
		// If not blank, check if it's just whitespace
		body := tag.Nodelist()[0].(string)
		if strings.TrimSpace(body) != "" {
			t.Errorf("Expected Blank to be true for empty doc, but body is %q", body)
		}
	}
}

func TestDocTagWithContent(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	docContent := "Some documentation"
	tokenizer := pc.NewTokenizer(docContent+" {% enddoc %}", false, nil, false)

	tag, err := NewDocTag("doc", "", pc)
	if err != nil {
		t.Fatalf("NewDocTag() error = %v", err)
	}

	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if tag.Blank() {
		t.Error("Expected Blank to be false for doc with content")
	}

	nodelist := tag.Nodelist()
	if len(nodelist) != 1 {
		t.Errorf("Expected nodelist length 1, got %d", len(nodelist))
	}

	body := nodelist[0].(string)
	// Body should contain the content (may have trailing space before end tag)
	if !strings.Contains(body, docContent) {
		t.Errorf("Expected body to contain %q, got %q", docContent, body)
	}
}

func TestDocTagInvalidMarkup(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	_, err := NewDocTag("doc", "extra", pc)
	if err == nil {
		t.Fatal("Expected error for invalid markup")
	}
	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestDocTagNestedDocError(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tokenizer := pc.NewTokenizer("outer {% doc %} inner {% enddoc %} more {% enddoc %}", false, nil, false)

	tag, err := NewDocTag("doc", "", pc)
	if err != nil {
		t.Fatalf("NewDocTag() error = %v", err)
	}

	err = tag.Parse(tokenizer)
	if err == nil {
		t.Fatal("Expected error for nested doc tag")
	}
	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}
