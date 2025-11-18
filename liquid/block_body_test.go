package liquid

import (
	"testing"
)

func TestBlockBodyBasic(t *testing.T) {
	bb := NewBlockBody()
	if bb == nil {
		t.Fatal("Expected BlockBody, got nil")
	}
	if bb.Nodelist() == nil {
		t.Error("Expected nodelist, got nil")
	}
	if len(bb.Nodelist()) != 0 {
		t.Errorf("Expected empty nodelist, got %d items", len(bb.Nodelist()))
	}
	if !bb.Blank() {
		t.Error("Expected block body to be blank initially")
	}
}

func TestBlockBodyBlank(t *testing.T) {
	bb := NewBlockBody()
	if !bb.Blank() {
		t.Error("Expected block body to be blank")
	}
}

func TestBlockBodyRender(t *testing.T) {
	bb := NewBlockBody()

	// Add some text nodes
	bb.nodelist = append(bb.nodelist, "hello")
	bb.nodelist = append(bb.nodelist, " ")
	bb.nodelist = append(bb.nodelist, "world")

	output := bb.Render(nil)
	if output == "" {
		t.Error("Expected rendered output, got empty string")
	}
}

func TestBlockBodyRemoveBlankStrings(t *testing.T) {
	bb := NewBlockBody()
	bb.blank = true
	bb.nodelist = []interface{}{
		"   ",
		"text",
		"\t\n",
	}

	bb.RemoveBlankStrings()

	if len(bb.nodelist) != 1 {
		t.Errorf("Expected 1 node after removing blanks, got %d", len(bb.nodelist))
	}
	if bb.nodelist[0] != "text" {
		t.Errorf("Expected 'text', got %v", bb.nodelist[0])
	}
}

func TestBlockBodyCreateVariable(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)
	bb := NewBlockBody()

	variable := bb.createVariable("{{ name }}", pc)
	if variable == nil {
		t.Fatal("Expected Variable, got nil")
	}
	if variable.Name() == nil {
		t.Error("Expected variable name to be set")
	}
}

// TestBlockBodyParseWithTagConstructor tests that TagConstructor functions are called via reflection
func TestBlockBodyParseWithTagConstructor(t *testing.T) {
	env := NewEnvironment()

	// Register a test tag constructor function
	constructorCalled := false
	env.RegisterTag("testtag", func(tagName, markup string, parseContext ParseContextInterface) (interface{}, error) {
		constructorCalled = true
		if tagName != "testtag" {
			t.Errorf("Expected tag name 'testtag', got %q", tagName)
		}
		// Markup may have trailing whitespace, so just check it contains the args
		if markup != "arg1 arg2" && markup != "arg1 arg2 " {
			t.Errorf("Expected markup 'arg1 arg2' (with optional trailing space), got %q", markup)
		}
		// Return a generic tag for testing
		return NewTag(tagName, markup, parseContext), nil
	})

	pc := &mockParseContextForTag{env: env}
	tokenizer := pc.NewTokenizer(`{% testtag arg1 arg2 %}`, false, nil, false)

	bb := NewBlockBody()
	unknownTagHandler := func(tagName, markup string) bool {
		// Handler may be called with empty tagName at end of parsing (this is expected)
		if tagName != "" {
			// Should not be called for registered tags
			t.Errorf("Unknown tag handler called for tag: %s", tagName)
		}
		return false
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !constructorCalled {
		t.Error("Expected tag constructor to be called via reflection")
	}

	// Should have parsed the tag
	if len(bb.Nodelist()) == 0 {
		t.Error("Expected tag to be parsed, got empty nodelist")
	}

	// Verify it's a tag (not nil)
	tag := bb.Nodelist()[0]
	if tag == nil {
		t.Error("Expected tag, got nil")
	}
}

// TestBlockBodyParseWithTagConstructorError tests that errors from TagConstructor are handled
func TestBlockBodyParseWithTagConstructorError(t *testing.T) {
	env := NewEnvironment()

	// Register a tag constructor that returns an error
	env.RegisterTag("errortag", func(tagName, markup string, parseContext ParseContextInterface) (interface{}, error) {
		return nil, NewSyntaxError("test error")
	})

	pc := &mockParseContextForTag{env: env}
	tokenizer := pc.NewTokenizer(`{% errortag %}`, false, nil, false)
	bb := NewBlockBody()

	// Should panic with SyntaxError
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with SyntaxError")
			} else if _, ok := r.(*SyntaxError); !ok {
				t.Errorf("Expected SyntaxError panic, got %T", r)
			}
		}()
		unknownTagHandler := func(tagName, markup string) bool {
			return false
		}
		_ = bb.Parse(tokenizer, pc, unknownTagHandler)
	}()
}

// TestBlockBodyParseWithNonFunctionTagClass tests fallback to NewTag for non-function tag classes
func TestBlockBodyParseWithNonFunctionTagClass(t *testing.T) {
	env := NewEnvironment()

	// Register a non-function tag class (just a string)
	env.RegisterTag("stringtag", "not a function")

	pc := &mockParseContextForTag{env: env}
	tokenizer := pc.NewTokenizer(`{% stringtag %}`, false, nil, false)
	bb := NewBlockBody()

	unknownTagHandler := func(tagName, markup string) bool {
		return false
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should fallback to generic Tag
	if len(bb.Nodelist()) == 0 {
		t.Error("Expected tag to be parsed, got empty nodelist")
	}

	tag := bb.Nodelist()[0]
	if genericTag, ok := tag.(*Tag); ok {
		if genericTag.TagName() != "stringtag" {
			t.Errorf("Expected tag name 'stringtag', got %q", genericTag.TagName())
		}
	} else {
		t.Errorf("Expected generic Tag, got %T", tag)
	}
}

// TestBlockBodyParseWithUnknownTag tests unknown tag handling
func TestBlockBodyParseWithUnknownTag(t *testing.T) {
	env := NewEnvironment()

	pc := &mockParseContextForTag{env: env}
	tokenizer := pc.NewTokenizer(`{% unknowntag %}`, false, nil, false)
	bb := NewBlockBody()

	handlerCalled := false
	unknownTagHandler := func(tagName, markup string) bool {
		// Handler may be called with empty tagName at end of parsing
		if tagName != "" {
			handlerCalled = true
			if tagName != "unknowntag" {
				t.Errorf("Expected tag name 'unknowntag', got %q", tagName)
			}
		}
		return true // Continue parsing
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !handlerCalled {
		t.Error("Expected unknown tag handler to be called")
	}
}

// TestBlockBodyLiquidTagParsing tests liquid tag parsing
func TestBlockBodyLiquidTagParsing(t *testing.T) {
	env := NewEnvironment()
	// Register echo tag for testing
	env.RegisterTag("echo", func(tagName, markup string, parseContext ParseContextInterface) (interface{}, error) {
		return NewTag(tagName, markup, parseContext), nil
	})
	pc := &mockParseContextForTag{env: env}

	// Test liquid tag in document context
	tokenizer := pc.NewTokenizer(`{% liquid echo "test" %}`, false, nil, false)
	bb := NewBlockBody()

	unknownTagHandler := func(tagName, markup string) bool {
		return false
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have parsed the liquid tag content
	if len(bb.Nodelist()) == 0 {
		t.Error("Expected liquid tag content to be parsed")
	}
}

// TestBlockBodyMissingVariableTerminator tests missing variable terminator error
func TestBlockBodyMissingVariableTerminator(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{})
	pc.SetLineNumber(&lineNum)
	bb := NewBlockBody()

	// Should panic with SyntaxError for missing terminator
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with SyntaxError for missing variable terminator")
			} else if _, ok := r.(*SyntaxError); !ok {
				t.Errorf("Expected SyntaxError panic, got %T", r)
			}
		}()
		_ = bb.createVariable("{{ name", pc)
	}()
}

// TestBlockBodyInterruptChecking tests interrupt checking during rendering
func TestBlockBodyInterruptChecking(t *testing.T) {
	bb := NewBlockBody()
	ctx := NewContext()

	// Create a node that sets an interrupt
	interruptNode := &interruptNode{}
	bb.nodelist = append(bb.nodelist, "before")
	bb.nodelist = append(bb.nodelist, interruptNode)
	bb.nodelist = append(bb.nodelist, "after")

	var output string
	bb.RenderToOutputBuffer(ctx, &output)

	// Should stop rendering after interrupt
	if output != "before" {
		t.Errorf("Expected output to stop after interrupt, got %q", output)
	}

	if !ctx.Interrupt() {
		t.Error("Expected interrupt to be set")
	}
}

// interruptNode is a test node that sets an interrupt
type interruptNode struct{}

func (n *interruptNode) RenderToOutputBuffer(context TagContext, output *string) {
	interrupt := NewBreakInterrupt()
	context.PushInterrupt(interrupt)
}
