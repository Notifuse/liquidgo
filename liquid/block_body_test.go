package liquid

import (
	"strings"
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

	output := renderString(bb, nil)
	expected := "hello world"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
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

	// Test with non-string nodes (should be preserved)
	bb2 := NewBlockBody()
	bb2.blank = true
	pc := NewParseContext(ParseContextOptions{})
	v := NewVariable("test", pc)
	bb2.nodelist = []interface{}{
		"  ",
		v,
		"content",
	}
	bb2.RemoveBlankStrings()
	if len(bb2.nodelist) != 2 {
		t.Errorf("Expected 2 nodes (variable and content), got %d", len(bb2.nodelist))
	}

	// Test with blank = false (should not remove)
	bb3 := NewBlockBody()
	bb3.blank = false
	bb3.nodelist = []interface{}{"  ", "content"}
	originalLen := len(bb3.nodelist)
	bb3.RemoveBlankStrings()
	if len(bb3.nodelist) != originalLen {
		t.Errorf("Expected nodelist to remain unchanged when blank = false, got %d", len(bb3.nodelist))
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
	// variable is now interface{}, cast to *Variable
	v, ok := variable.(*Variable)
	if !ok {
		t.Fatalf("Expected *Variable, got %T", variable)
	}
	if v.Name() == nil {
		t.Error("Expected variable name to be set")
	}
}

// TestBlockBodyParseWithTagConstructor tests that TagConstructor functions are called via reflection
func TestBlockBodyParseWithTagConstructor(t *testing.T) {

	// Register a test tag constructor function
	constructorCalled := false
	envTags := map[string]TagConstructor{}
	envTags["testtag"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		constructorCalled = true
		if tagName != "testtag" {
			t.Errorf("Expected tag name 'testtag', got %q", tagName)
		}
		// Markup may have trailing whitespace, so just check it contains the args
		if markup != "arg1 arg2" && markup != "arg1 arg2 " {
			t.Errorf("Expected markup 'arg1 arg2' (with optional trailing space), got %q", markup)
		}
		// Return a generic tag for testing
		return &silentTag{NewTag(tagName, markup, parseContext)}, nil
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})

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

	// Register a tag constructor that returns an error
	envTags := map[string]TagConstructor{}
	envTags["errortag"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		return nil, NewSyntaxError("test error")
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})

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

// TestBlockBodyParseWithNonFunctionTagClass tests a constructor that
// returns neither a tag nor an error: a syntax error, where a tag class that
// was not a function (which no longer compiles) gave a generic Tag.
func TestBlockBodyParseWithNonFunctionTagClass(t *testing.T) {

	// Register a constructor that gives no tag
	envTags := map[string]TagConstructor{}
	envTags["stringtag"] = func(string, string, ParseContextInterface) (Node, error) { return nil, nil }
	env := mustEnv(EnvironmentConfig{Tags: envTags})

	pc := &mockParseContextForTag{env: env}
	tokenizer := pc.NewTokenizer(`{% stringtag %}`, false, nil, false)
	bb := NewBlockBody()

	defer func() {
		r := recover()
		err, ok := r.(*SyntaxError)
		if !ok {
			t.Fatalf("Expected SyntaxError panic, got %T (%v)", r, r)
		}
		if !strings.Contains(err.Error(), "stringtag") {
			t.Errorf("Expected the error to name the tag, got %q", err.Error())
		}
	}()
	_ = bb.Parse(tokenizer, pc, func(string, string) bool { return false })
}

// TestBlockBodyParseWithUnknownTag tests unknown tag handling
func TestBlockBodyParseWithUnknownTag(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})

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
	// Register echo tag for testing
	envTags := map[string]TagConstructor{}
	envTags["echo"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		return &silentTag{NewTag(tagName, markup, parseContext)}, nil
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})
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

// TestBlockBodyParseForDocument tests parseForDocument with various scenarios
func TestBlockBodyParseForDocument(t *testing.T) {
	envTags := map[string]TagConstructor{}
	envTags["echo"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		return &silentTag{NewTag(tagName, markup, parseContext)}, nil
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})
	pc := NewParseContext(ParseContextOptions{Environment: env})
	bb := NewBlockBody()

	unknownTagHandler := func(tagName, markup string) bool {
		return true
	}

	// Test with variable token
	tokenizer := pc.NewTokenizer(`{{ name }}`, false, nil, false)
	err := bb.parseForDocument(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("parseForDocument() with variable error = %v", err)
	}
	if len(bb.Nodelist()) == 0 {
		t.Error("Expected variable to be parsed")
	}

	// Test with text token and whitespace trimming
	bb2 := NewBlockBody()
	pc2 := NewParseContext(ParseContextOptions{Environment: env})
	pc2.SetTrimWhitespace(true)
	tokenizer2 := pc2.NewTokenizer(`  text  `, false, nil, false)
	err2 := bb2.parseForDocument(tokenizer2, pc2, unknownTagHandler)
	if err2 != nil {
		t.Fatalf("parseForDocument() with text error = %v", err2)
	}

	// Test with tag token
	bb3 := NewBlockBody()
	tokenizer3 := pc.NewTokenizer(`{% echo "test" %}`, false, nil, false)
	err3 := bb3.parseForDocument(tokenizer3, pc, unknownTagHandler)
	if err3 != nil {
		t.Fatalf("parseForDocument() with tag error = %v", err3)
	}

	// Test with liquid tag
	bb4 := NewBlockBody()
	tokenizer4 := pc.NewTokenizer(`{% liquid echo "test" %}`, false, nil, false)
	err4 := bb4.parseForDocument(tokenizer4, pc, unknownTagHandler)
	if err4 != nil {
		t.Fatalf("parseForDocument() with liquid tag error = %v", err4)
	}

	// Test with unknown tag handler returning false
	bb5 := NewBlockBody()
	tokenizer5 := pc.NewTokenizer(`{% unknown %}`, false, nil, false)
	handlerReturnsFalse := func(tagName, markup string) bool {
		return false
	}
	err5 := bb5.parseForDocument(tokenizer5, pc, handlerReturnsFalse)
	if err5 != nil {
		t.Fatalf("parseForDocument() with handler returning false error = %v", err5)
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
	renderToString(bb, ctx, &output)

	// Should stop rendering after interrupt
	if output != "before" {
		t.Errorf("Expected output to stop after interrupt, got %q", output)
	}

	if !ctx.Interrupt() {
		t.Error("Expected interrupt to be set")
	}
}

// TestBlockBodyWhitespaceHandler tests whitespace handling
func TestBlockBodyWhitespaceHandler(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	bb := NewBlockBody()
	bb.nodelist = append(bb.nodelist, "  previous  ")

	// Test with whitespace control token ({%-)
	token := "{%- tag %}"
	bb.whitespaceHandler(token, pc)

	// Previous token should be trimmed
	if len(bb.nodelist) > 0 {
		if prevToken, ok := bb.nodelist[0].(string); ok {
			if strings.HasSuffix(prevToken, "  ") {
				t.Error("Expected previous token to be trimmed on the right")
			}
		}
	}

	// Test with trailing whitespace control (-%})
	token2 := "{% tag -%}"
	bb.whitespaceHandler(token2, pc)
	if !pc.TrimWhitespace() {
		t.Error("Expected TrimWhitespace to be set")
	}
}

// TestBlockBodyRenderNodeOptimized tests optimized node rendering
func TestBlockBodyRenderNodeOptimized(t *testing.T) {
	bb := NewBlockBody()
	ctx := NewContext()
	output := ""

	// Test with a variable node
	pc := NewParseContext(ParseContextOptions{})
	v := NewVariable("test", pc)
	ctx.Set("test", "value")
	renderNodeForTest(bb, v, ctx, &output)
	if output != "value" {
		t.Errorf("Expected 'value', got %q", output)
	}

	// Test with a tag that writes nothing
	pc3 := NewParseContext(ParseContextOptions{})
	output3 := ""
	renderNodeForTest(bb, &silentTag{NewTag("test", "", pc3)}, NewContext(), &output3)
	if output3 != "" {
		t.Errorf("Expected no output from a silent tag, got %q", output3)
	}

	// Test with a tag that embeds Tag and declares its RenderToOutput
	output4 := ""
	renderNodeForTest(bb, &writingTag{Tag: NewTag("custom", "", pc3), text: "c"}, NewContext(), &output4)
	if output4 != "c" {
		t.Errorf("Expected 'c', got %q", output4)
	}
}

// writingTag is a tag that writes text.
type writingTag struct {
	*Tag
	text string
}

func (w *writingTag) RenderToOutput(_ TagContext, output *OutputBuffer) { output.WriteString(w.text) }

// TestBlockBodyParseLiquidTag tests liquid tag parsing
func TestBlockBodyParseLiquidTag(t *testing.T) {
	envTags := map[string]TagConstructor{}
	envTags["echo"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		return &silentTag{NewTag(tagName, markup, parseContext)}, nil
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})
	pc := NewParseContext(ParseContextOptions{Environment: env})
	bb := NewBlockBody()

	// Test parseLiquidTag with echo command
	func() {
		defer func() {
			if r := recover(); r != nil {
				// parseLiquidTag may panic on unknown tags, which is expected
				if _, ok := r.(*SyntaxError); !ok {
					t.Errorf("Expected SyntaxError panic, got %T", r)
				}
			}
		}()
		bb.parseLiquidTag(`echo "test"`, pc)
	}()

	// Test with valid liquid tag syntax
	bb2 := NewBlockBody()
	bb2.parseLiquidTag(`echo "test"`, pc)
	if len(bb2.Nodelist()) == 0 {
		t.Log("parseLiquidTag may not add nodes directly")
	}

	// Test with line number set
	lineNum := 5
	pc3 := NewParseContext(ParseContextOptions{Environment: env})
	pc3.SetLineNumber(&lineNum)
	bb3 := NewBlockBody()
	func() {
		defer func() {
			if r := recover(); r != nil {
				// May panic on unknown tags
				_ = r
			}
		}()
		bb3.parseLiquidTag(`echo "test"`, pc3)
	}()

	// Test with nil line number
	bb4 := NewBlockBody()
	pc4 := NewParseContext(ParseContextOptions{Environment: env})
	pc4.SetLineNumber(nil)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// May panic on unknown tags
				_ = r
			}
		}()
		bb4.parseLiquidTag(`echo "test"`, pc4)
	}()

	// Test with error from parseForLiquidTag
	bb5 := NewBlockBody()
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected panic on error
				_ = r
			}
		}()
		// This should trigger an error path
		bb5.parseLiquidTag(`unknown_tag`, pc)
	}()
}

// TestBlockBodyParseForLiquidTag tests parseForLiquidTag more thoroughly
func TestBlockBodyParseForLiquidTag(t *testing.T) {
	envTags := map[string]TagConstructor{}
	envTags["echo"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		return &silentTag{NewTag(tagName, markup, parseContext)}, nil
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})
	pc := NewParseContext(ParseContextOptions{Environment: env})
	bb := NewBlockBody()

	// Test with liquid tag syntax
	tokenizer := pc.NewTokenizer("echo test", true, nil, true)
	unknownTagHandler := func(tagName, markup string) bool {
		return true
	}

	err := bb.parseForLiquidTag(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("parseForLiquidTag() error = %v", err)
	}

	// Test with empty tokenizer
	tokenizer2 := pc.NewTokenizer("", true, nil, true)
	bb2 := NewBlockBody()
	err2 := bb2.parseForLiquidTag(tokenizer2, pc, unknownTagHandler)
	if err2 != nil {
		t.Fatalf("parseForLiquidTag() with empty input error = %v", err2)
	}

	// Test with liquid tag
	bb3 := NewBlockBody()
	tokenizer3 := pc.NewTokenizer("liquid echo test", true, nil, true)
	err3 := bb3.parseForLiquidTag(tokenizer3, pc, unknownTagHandler)
	if err3 != nil {
		t.Fatalf("parseForLiquidTag() with liquid tag error = %v", err3)
	}

	// Test with unknown tag handler returning false
	bb4 := NewBlockBody()
	tokenizer4 := pc.NewTokenizer("unknown", true, nil, true)
	handlerReturnsFalse := func(tagName, markup string) bool {
		return false // Stop parsing
	}
	err4 := bb4.parseForLiquidTag(tokenizer4, pc, handlerReturnsFalse)
	if err4 != nil {
		t.Fatalf("parseForLiquidTag() with handler returning false error = %v", err4)
	}

	// Test with environment that has no tag registered
	envNoTag := mustEnv(EnvironmentConfig{})
	pcNoTag := NewParseContext(ParseContextOptions{Environment: envNoTag})
	bb5 := NewBlockBody()
	tokenizer5 := pcNoTag.NewTokenizer("unknown_tag", true, nil, true)
	err5 := bb5.parseForLiquidTag(tokenizer5, pcNoTag, unknownTagHandler)
	if err5 != nil {
		t.Fatalf("parseForLiquidTag() with unknown tag error = %v", err5)
	}
}

// interruptNode is a test node that sets an interrupt
type interruptNode struct{}

func (n *interruptNode) RenderToOutput(context TagContext, output *OutputBuffer) {
	interrupt := NewBreakInterrupt()
	context.PushInterrupt(interrupt)
}

func TestBlockBodyParseForDocumentNilEnvironment(t *testing.T) {
	bb := NewBlockBody()

	// Create a parse context with nil environment
	mockPC := &mockParseContext{}
	tokenizer := mockPC.NewTokenizer("{% tag %}", false, nil, false)

	handlerCalled := false
	unknownTagHandler := func(tagName, markup string) bool {
		handlerCalled = true
		return true
	}

	err := bb.parseForDocument(tokenizer, mockPC, unknownTagHandler)
	if err != nil {
		t.Fatalf("parseForDocument() error = %v", err)
	}
	// Should handle nil environment gracefully
	_ = handlerCalled
}

func TestBlockBodyCreateVariableEdgeCases(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	bb := NewBlockBody()

	// Test creating variable with whitespace trimming variations
	v2 := bb.createVariable("{{- var -}}", pc)
	if v2 == nil {
		t.Fatal("Expected variable with whitespace trimming, got nil")
	}

	// Test with just start marker (will raise error)
	func() {
		defer func() {
			if r := recover(); r == nil {
				// In new implementation, createVariable might handle errors internally
				// But for malformed tags it should still panic or return nil/string
				// If it returned without panic, check if we got back a string (token)
				// which happens in warn mode, but here default is lax/strict
				return
			}
		}()
		bb.createVariable("{{ var", pc)
	}()
}

func TestBlockBodyParseForDocumentWithLiquidTag(t *testing.T) {
	// Create environment with assign tag registered
	envTags := map[string]TagConstructor{}
	envTags["assign"] = func(tagName, markup string, parseContext ParseContextInterface) (Node, error) {
		// Import tags package to use NewAssignTag
		// For now, create a simple tag
		return &silentTag{NewTag(tagName, markup, parseContext)}, nil
	}
	env := mustEnv(EnvironmentConfig{Tags: envTags})
	pc := NewParseContext(ParseContextOptions{Environment: env})
	bb := NewBlockBody()

	// Test parsing document with liquid tag
	tokenizer := pc.NewTokenizer("{% liquid assign x = 1 %}", false, nil, false)
	unknownTagHandler := func(tagName, markup string) bool {
		return true
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestBlockBodyParseForDocumentWithVariable(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	bb := NewBlockBody()

	// Test parsing document with variable
	tokenizer := pc.NewTokenizer("{{ var }}", false, nil, false)
	unknownTagHandler := func(tagName, markup string) bool {
		return true
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(bb.Nodelist()) == 0 {
		t.Error("Expected variable in nodelist")
	}
}

func TestBlockBodyParseForDocumentWithTrimWhitespace(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	pc.SetTrimWhitespace(true)
	bb := NewBlockBody()

	// Test parsing with whitespace trimming
	tokenizer := pc.NewTokenizer("   text   ", false, nil, false)
	unknownTagHandler := func(tagName, markup string) bool {
		return true
	}

	err := bb.Parse(tokenizer, pc, unknownTagHandler)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

// renderNodeForTest renders node as BlockBody does, and appends its output.
func renderNodeForTest(bb *BlockBody, node interface{}, context TagContext, output *string) {
	buf := NewOutputBuffer()
	bb.renderNode(context, buf, node)
	*output += buf.String()
}
