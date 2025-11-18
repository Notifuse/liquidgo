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

	// Mock context
	type mockTagContext struct{}

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
