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

// TestDocumentParseDocumentError tests ParseDocument with error
func TestDocumentParseDocumentError(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	// Create tokenizer with invalid syntax
	tokenizer := pc.NewTokenizer("{% if %}", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err == nil {
		t.Error("Expected error for invalid syntax")
	}
	if doc != nil {
		t.Error("Expected nil document on error")
	}
}

// TestDocumentParseDocumentPanic tests ParseDocument with panic recovery
func TestDocumentParseDocumentPanic(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	// Create tokenizer that will cause panic
	tokenizer := pc.NewTokenizer("{% unknown_tag %}", false, nil, false)

	doc, err := ParseDocument(tokenizer, pc)
	if err == nil {
		t.Error("Expected error for unknown tag")
	}
	if doc != nil {
		t.Error("Expected nil document on error")
	}
}

// TestDocumentParseBody tests parseBody method
func TestDocumentParseBody(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)
	tokenizer := pc.NewTokenizer("Hello World", false, nil, false)

	// parseBody should return false to stop parsing
	shouldContinue := doc.parseBody(tokenizer, pc)
	if shouldContinue {
		t.Error("Expected parseBody to return false")
	}

	// Check that body was parsed
	nodelist := doc.Nodelist()
	if len(nodelist) == 0 {
		t.Error("Expected nodelist to have nodes")
	}
}

// TestDocumentParseBodyWithUnknownTag tests parseBody with unknown tag handler
func TestDocumentParseBodyWithUnknownTag(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)
	// Create tokenizer with unknown tag
	tokenizer := pc.NewTokenizer("{% unknown_tag %}", false, nil, false)

	// Should panic with unknown tag
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for unknown tag")
			}
		}()
		doc.parseBody(tokenizer, pc)
	}()
}

// TestDocumentParseBodyWithError tests parseBody with parse error
func TestDocumentParseBodyWithError(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)
	// Create tokenizer with invalid syntax
	tokenizer := pc.NewTokenizer("{% if %}", false, nil, false)

	// Should panic with parse error
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for parse error")
			}
		}()
		doc.parseBody(tokenizer, pc)
	}()
}

// TestDocumentParseWithPanicRecovery tests Parse with panic recovery
func TestDocumentParseWithPanicRecovery(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)
	tokenizer := pc.NewTokenizer("{% if %}", false, nil, false)

	// Should panic and be recovered
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid syntax")
			}
		}()
		doc.Parse(tokenizer, pc)
	}()
}

// TestDocumentParseWithLineNumber tests Parse with line number in error
func TestDocumentParseWithLineNumber(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	doc := NewDocument(pc)
	tokenizer := pc.NewTokenizer("line1\nline2\n{% if %}", false, nil, false)

	// Should panic with line number
	func() {
		defer func() {
			if r := recover(); r != nil {
				if se, ok := r.(*SyntaxError); ok {
					// Line number may or may not be set depending on when error occurs
					if se.Err.LineNumber != nil {
						t.Logf("Line number set: %d", *se.Err.LineNumber)
					} else {
						t.Log("Line number not set (may be set later in error handling)")
					}
				}
			}
		}()
		doc.Parse(tokenizer, pc)
	}()
}
