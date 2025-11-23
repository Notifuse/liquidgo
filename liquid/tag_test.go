package liquid

import (
	"testing"
)

// MockParseContext for testing
type mockParseContextForTag struct {
	lineNum        *int
	env            *Environment
	trimWhitespace bool
	depth          int
	warnings       []error
}

func (m *mockParseContextForTag) ErrorMode() string {
	if m.env != nil {
		return m.env.ErrorMode()
	}
	return "lax"
}

func (m *mockParseContextForTag) AddWarning(err error) {
	m.warnings = append(m.warnings, err)
}

func (m *mockParseContextForTag) ParseExpression(markup string) interface{} {
	return VariableLookupParse(markup, nil, nil)
}

func (m *mockParseContextForTag) SafeParseExpression(parser *Parser) interface{} {
	expr, err := parser.Expression()
	if err != nil {
		return nil
	}
	return VariableLookupParse(expr, nil, nil)
}

func (m *mockParseContextForTag) NewParser(markup string) *Parser {
	return NewParser(markup)
}

func (m *mockParseContextForTag) LineNumber() *int {
	return m.lineNum
}

func (m *mockParseContextForTag) SetLineNumber(ln *int) {
	m.lineNum = ln
}

func (m *mockParseContextForTag) Environment() *Environment {
	return m.env
}

func (m *mockParseContextForTag) TrimWhitespace() bool {
	return m.trimWhitespace
}

func (m *mockParseContextForTag) SetTrimWhitespace(tw bool) {
	m.trimWhitespace = tw
}

func (m *mockParseContextForTag) Depth() int {
	return m.depth
}

func (m *mockParseContextForTag) IncrementDepth() {
	m.depth++
}

func (m *mockParseContextForTag) DecrementDepth() {
	m.depth--
}

func (m *mockParseContextForTag) NewBlockBody() *BlockBody {
	return NewBlockBody()
}

func (m *mockParseContextForTag) NewTokenizer(source string, lineNumbers bool, startLineNumber *int, forLiquidTag bool) *Tokenizer {
	return NewTokenizer(source, nil, lineNumbers, startLineNumber, forLiquidTag)
}

func TestTagBasic(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("test", "arg1 arg2", pc)
	if tag == nil {
		t.Fatal("Expected Tag, got nil")
	}
	if tag.TagName() != "test" {
		t.Errorf("Expected tag name 'test', got '%s'", tag.TagName())
	}
	if tag.Markup() != "arg1 arg2" {
		t.Errorf("Expected markup 'arg1 arg2', got '%s'", tag.Markup())
	}
}

func TestTagRaw(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("test", "arg1", pc)
	raw := tag.Raw()
	if raw != "test arg1" {
		t.Errorf("Expected 'test arg1', got '%s'", raw)
	}
}

func TestTagName(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("TestTag", "arg", pc)
	name := tag.Name()
	if name != "testtag" {
		t.Errorf("Expected 'testtag', got '%s'", name)
	}
}

func TestTagRender(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("test", "", pc)

	result := tag.Render(nil)
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestTagBlank(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("test", "", pc)
	if tag.Blank() {
		t.Error("Expected tag NOT to be blank by default (returns false)")
	}
}

func TestTagRenderToOutputBuffer(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("test", "", pc)
	ctx := NewContext()
	output := ""

	// Test with empty render result
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}

	// Test with non-empty render result (Tag.Render returns empty, so this tests the path)
	// Since Tag.Render always returns empty, output should remain empty
	tag.RenderToOutputBuffer(ctx, &output)
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

func TestParseTag(t *testing.T) {
	lineNum := 1
	pc := &mockParseContextForTag{lineNum: &lineNum}

	// Test ParseTag function
	tokenizer := pc.NewTokenizer("content", false, nil, false)
	tag, err := ParseTag("test", "arg", tokenizer, pc)
	if err != nil {
		t.Fatalf("ParseTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected Tag, got nil")
	}
	if tag.TagName() != "test" {
		t.Errorf("Expected tag name 'test', got '%s'", tag.TagName())
	}
}

func TestTagLineNumber(t *testing.T) {
	lineNum := 5
	pc := &mockParseContextForTag{lineNum: &lineNum}

	tag := NewTag("test", "", pc)
	ln := tag.LineNumber()
	if ln == nil {
		t.Error("Expected line number, got nil")
	} else if *ln != 5 {
		t.Errorf("Expected line number 5, got %d", *ln)
	}

	// Test with nil line number
	pc2 := &mockParseContextForTag{lineNum: nil}
	tag2 := NewTag("test", "", pc2)
	ln2 := tag2.LineNumber()
	if ln2 != nil {
		t.Errorf("Expected nil line number, got %v", ln2)
	}
}

// Remove tests that use methods that don't exist on Tag (like SafeParseExpression directly)
// or update them to use context if available
