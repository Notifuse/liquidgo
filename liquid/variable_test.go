package liquid

import (
	"testing"
)

// MockParseContext for testing
type mockParseContext struct {
	lineNum        *int
	env            *Environment
	trimWhitespace bool
	depth          int
}

func (m *mockParseContext) ParseExpression(markup string) interface{} {
	return VariableLookupParse(markup, nil, nil)
}

func (m *mockParseContext) SafeParseExpression(parser *Parser) interface{} {
	expr, err := parser.Expression()
	if err != nil {
		return nil
	}
	return VariableLookupParse(expr, nil, nil)
}

func (m *mockParseContext) NewParser(markup string) *Parser {
	return NewParser(markup)
}

func (m *mockParseContext) LineNumber() *int {
	return m.lineNum
}

func (m *mockParseContext) SetLineNumber(ln *int) {
	m.lineNum = ln
}

func (m *mockParseContext) Environment() *Environment {
	if m.env == nil {
		m.env = NewEnvironment()
	}
	return m.env
}

func (m *mockParseContext) TrimWhitespace() bool {
	return m.trimWhitespace
}

func (m *mockParseContext) SetTrimWhitespace(tw bool) {
	m.trimWhitespace = tw
}

func (m *mockParseContext) Depth() int {
	return m.depth
}

func (m *mockParseContext) IncrementDepth() {
	m.depth++
}

func (m *mockParseContext) DecrementDepth() {
	m.depth--
}

func (m *mockParseContext) NewBlockBody() *BlockBody {
	return NewBlockBody()
}

func (m *mockParseContext) NewTokenizer(source string, lineNumbers bool, startLineNumber *int, forLiquidTag bool) *Tokenizer {
	return NewTokenizer(source, nil, lineNumbers, startLineNumber, forLiquidTag)
}

func TestVariableBasic(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("user.name", pc)
	if v == nil {
		t.Fatal("Expected Variable, got nil")
	}
	if v.Raw() != "user.name" {
		t.Errorf("Expected raw 'user.name', got '%s'", v.Raw())
	}
	if v.Name() == nil {
		t.Error("Expected name to be set")
	}
}

func TestVariableWithFilters(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("user.name | upcase", pc)
	if v == nil {
		t.Fatal("Expected Variable, got nil")
	}
	if len(v.Filters()) == 0 {
		t.Error("Expected filters to be set")
	}
}

func TestVariableLineNumber(t *testing.T) {
	lineNum := 42
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("test", pc)
	if v.LineNumber() == nil || *v.LineNumber() != 42 {
		t.Errorf("Expected line number 42, got %v", v.LineNumber())
	}
}

// TestVariableParserSwitching tests parser switching based on error mode
func TestVariableParserSwitching(t *testing.T) {
	lineNum := 1

	// Test strict mode
	pcStrict := NewParseContext(ParseContextOptions{ErrorMode: "strict"})
	pcStrict.SetLineNumber(&lineNum)
	v := NewVariable("test", pcStrict)
	if v == nil {
		t.Fatal("Expected Variable, got nil")
	}

	// Test lax mode
	pcLax := NewParseContext(ParseContextOptions{ErrorMode: "lax"})
	pcLax.SetLineNumber(&lineNum)
	v = NewVariable("test", pcLax)
	if v == nil {
		t.Fatal("Expected Variable, got nil")
	}

	// Test rigid mode
	pcRigid := NewParseContext(ParseContextOptions{ErrorMode: "rigid"})
	pcRigid.SetLineNumber(&lineNum)
	v = NewVariable("test", pcRigid)
	if v == nil {
		t.Fatal("Expected Variable, got nil")
	}
}
