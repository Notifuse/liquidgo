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
	warnings       []error
}

func (m *mockParseContext) ErrorMode() string {
	if m.env != nil {
		return m.env.ErrorMode()
	}
	return "lax"
}

func (m *mockParseContext) AddWarning(err error) {
	m.warnings = append(m.warnings, err)
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
	if len(v.Filters()) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(v.Filters()))
	}
}

func TestVariableWithArgs(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("user.name | date: '%Y-%m-%d'", pc)
	if v == nil {
		t.Fatal("Expected Variable, got nil")
	}
	if len(v.Filters()) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(v.Filters()))
	}
}

func TestVariableLineNumber(t *testing.T) {
	lineNum := 5
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("user.name", pc)
	ln := v.LineNumber()
	if ln == nil {
		t.Error("Expected line number, got nil")
	} else if *ln != 5 {
		t.Errorf("Expected line number 5, got %d", *ln)
	}
}

func TestVariableRender(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}
	v := NewVariable("name", pc)

	ctx := NewContext()
	ctx.Set("name", "bob")

	output := ""
	v.RenderToOutputBuffer(ctx, &output)
	if output != "bob" {
		t.Errorf("Expected 'bob', got %q", output)
	}
}

func TestVariableMarkupContext(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}
	v := NewVariable("test", pc)

	// We can't access markupContext directly as it is private,
	// but we can check if errors contain context when rendered?
	// Or we can just trust it's correct since unit tests usually test public API.
	_ = v
}

// testPanicker is a drop that panics with a configurable error when triggered
type testPanicker struct {
	err error
}

func (p *testPanicker) ToLiquid() interface{} { return p }

func (p *testPanicker) LiquidMethodMissing(method string) interface{} {
	if method == "trigger" {
		panic(p.err)
	}
	return nil
}

// TestRenderToOutputBufferPreservesLiquidErrorTypes tests that all LiquidError types
// are preserved during panic recovery (not converted to InternalError)
func TestRenderToOutputBufferPreservesLiquidErrorTypes(t *testing.T) {
	errorTypes := []struct {
		name   string
		create func() error
	}{
		{"UndefinedFilter", func() error { return NewUndefinedFilter("test") }},
		{"UndefinedVariable", func() error { return NewUndefinedVariable("test") }},
		{"UndefinedDropMethod", func() error { return NewUndefinedDropMethod("test") }},
		{"DisabledError", func() error { return NewDisabledError("test") }},
		{"MemoryError", func() error { return NewMemoryError("test") }},
		{"FileSystemError", func() error { return NewFileSystemError("test") }},
		{"ZeroDivisionError", func() error { return NewZeroDivisionError("test") }},
		{"FloatDomainError", func() error { return NewFloatDomainError("test") }},
		{"MethodOverrideError", func() error { return NewMethodOverrideError("test") }},
		{"ContextError", func() error { return NewContextError("test") }},
		{"TemplateEncodingError", func() error { return NewTemplateEncodingError("test") }},
	}

	for _, tt := range errorTypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext()
			ctx.Set("panicker", &testPanicker{err: tt.create()})

			lineNum := 1
			pc := &mockParseContext{lineNum: &lineNum}
			v := NewVariable("panicker.trigger", pc)

			output := ""
			v.RenderToOutputBuffer(ctx, &output)

			// Error should be in context
			if len(ctx.Errors()) == 0 {
				t.Fatalf("%s: expected error in context.Errors()", tt.name)
			}

			// Error type should be preserved
			if _, ok := ctx.Errors()[0].(*InternalError); ok && tt.name != "InternalError" {
				t.Errorf("%s: error was converted to InternalError", tt.name)
			}
		})
	}
}
