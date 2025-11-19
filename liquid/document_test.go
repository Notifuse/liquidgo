package liquid

import (
	"testing"
)

func TestNewDocument(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)
	if doc == nil {
		t.Fatal("Expected Document, got nil")
	}
	if doc.ParseContext() != pc {
		t.Error("Expected parse context to be set")
	}
	if doc.Body() == nil {
		t.Error("Expected body to be initialized")
	}
}

func TestDocumentParseSimple(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	tokenizer := pc.NewTokenizer("Hello World", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if doc == nil {
		t.Fatal("Expected Document, got nil")
	}

	nodelist := doc.Nodelist()
	if len(nodelist) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodelist))
	}
	if nodelist[0] != "Hello World" {
		t.Errorf("Expected 'Hello World', got %v", nodelist[0])
	}
}

func TestDocumentParseWithVariable(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	tokenizer := pc.NewTokenizer("Hello {{ name }}", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	nodelist := doc.Nodelist()
	if len(nodelist) < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", len(nodelist))
	}
}

func TestDocumentRenderSimple(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	tokenizer := pc.NewTokenizer("Hello World", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	ctx := NewContext()
	result := doc.Render(ctx)
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", result)
	}
}

func TestDocumentRenderToOutputBuffer(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	tokenizer := pc.NewTokenizer("Test", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	ctx := NewContext()
	var output string
	doc.RenderToOutputBuffer(ctx, &output)
	if output != "Test" {
		t.Errorf("Expected 'Test', got %q", output)
	}
}

func TestDocumentUnknownTag(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)

	// Test unknown tag
	err := doc.UnknownTag("unknown", "", nil)
	if err == nil {
		t.Fatal("Expected error for unknown tag")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestDocumentUnknownTagElse(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)

	// Test else tag (should error)
	err := doc.UnknownTag("else", "", nil)
	if err == nil {
		t.Fatal("Expected error for else tag")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestDocumentUnknownTagEnd(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)

	// Test end tag (should error)
	err := doc.UnknownTag("end", "", nil)
	if err == nil {
		t.Fatal("Expected error for end tag")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestDocumentNodelist(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	tokenizer := pc.NewTokenizer("text1 text2", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	nodelist := doc.Nodelist()
	if len(nodelist) == 0 {
		t.Error("Expected nodelist to have nodes")
	}
}

func TestDocumentParseEmpty(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	tokenizer := pc.NewTokenizer("", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if doc == nil {
		t.Fatal("Expected Document, got nil")
	}

	nodelist := doc.Nodelist()
	// Empty document should have empty nodelist or just whitespace
	if len(nodelist) > 1 {
		t.Errorf("Expected empty or minimal nodelist, got %d nodes", len(nodelist))
	}
}
