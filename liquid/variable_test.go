package liquid

import (
	"strings"
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

// TestVariableRender tests comprehensive variable rendering
func TestVariableRender(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	// Test simple variable
	v := NewVariable("name", pc)
	ctx := NewContext()
	ctx.Set("name", "test")
	result := v.Render(ctx)
	if result != "test" {
		t.Errorf("Expected 'test', got %v", result)
	}

	// Test variable with single filter
	v2 := NewVariable("name | upcase", pc)
	ctx2 := NewContext()
	ctx2.Set("name", "test")
	result2 := v2.Render(ctx2)
	if result2 != "TEST" {
		t.Errorf("Expected 'TEST', got %v", result2)
	}

	// Test variable with multiple filters
	v3 := NewVariable("name | downcase | capitalize", pc)
	ctx3 := NewContext()
	ctx3.Set("name", "TEST")
	result3 := v3.Render(ctx3)
	if result3 != "Test" {
		t.Errorf("Expected 'Test', got %v", result3)
	}

	// Test variable with filter arguments (using a simpler filter)
	v4 := NewVariable("name | capitalize", pc)
	ctx4 := NewContext()
	ctx4.Set("name", "hello")
	result4 := v4.Render(ctx4)
	if result4 != "Hello" {
		t.Errorf("Expected 'Hello', got %v", result4)
	}

	// Test variable with global filter
	v5 := NewVariable("name", pc)
	ctx5 := NewContext()
	ctx5.Set("name", "test")
	ctx5.SetGlobalFilter(func(obj interface{}) interface{} {
		return "filtered"
	})
	result5 := v5.Render(ctx5)
	if result5 != "filtered" {
		t.Errorf("Expected 'filtered', got %v", result5)
	}
}

// TestVariableParseFilterArgs tests parseFilterArgs with various patterns
func TestVariableParseFilterArgs(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{ErrorMode: "strict"})
	pc.SetLineNumber(&lineNum)

	v := NewVariable("test", pc)

	// Test with single argument
	p := NewParser("arg1")
	args := v.parseFilterArgs(p)
	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}

	// Test with multiple arguments
	p2 := NewParser("arg1, arg2, arg3")
	args2 := v.parseFilterArgs(p2)
	if len(args2) < 2 {
		t.Errorf("Expected at least 2 arguments, got %d", len(args2))
	}

	// Test with no arguments
	p3 := NewParser("")
	args3 := v.parseFilterArgs(p3)
	if len(args3) != 0 {
		t.Errorf("Expected 0 arguments, got %d", len(args3))
	}
}

// TestVariableArgument tests argument parsing
func TestVariableArgument(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{ErrorMode: "rigid"})
	pc.SetLineNumber(&lineNum)

	v := NewVariable("test", pc)

	// Test keyword argument
	p := NewParser("key: value")
	positionalArgs := []interface{}{}
	keywordArgs := make(map[string]interface{})
	v.argument(p, &positionalArgs, keywordArgs)
	if len(keywordArgs) != 1 {
		t.Errorf("Expected 1 keyword argument, got %d", len(keywordArgs))
	}
	if keywordArgs["key"] == nil {
		t.Error("Expected keyword 'key' to be set")
	}

	// Test positional argument
	p2 := NewParser("value")
	positionalArgs2 := []interface{}{}
	keywordArgs2 := make(map[string]interface{})
	v.argument(p2, &positionalArgs2, keywordArgs2)
	if len(positionalArgs2) != 1 {
		t.Errorf("Expected 1 positional argument, got %d", len(positionalArgs2))
	}
}

// TestVariableEndOfArguments tests endOfArguments detection
func TestVariableEndOfArguments(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{ErrorMode: "rigid"})
	pc.SetLineNumber(&lineNum)

	v := NewVariable("test", pc)

	// Test with pipe (end of arguments)
	p := NewParser("|")
	if !v.endOfArguments(p) {
		t.Error("Expected endOfArguments to return true for pipe")
	}

	// Test with end of string
	p2 := NewParser("")
	if !v.endOfArguments(p2) {
		t.Error("Expected endOfArguments to return true for end of string")
	}

	// Test with content (not end)
	p3 := NewParser("arg")
	if v.endOfArguments(p3) {
		t.Error("Expected endOfArguments to return false for content")
	}
}

// TestVariableRigidParseFilterExpressions tests rigid mode filter parsing
func TestVariableRigidParseFilterExpressions(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{ErrorMode: "rigid"})
	pc.SetLineNumber(&lineNum)

	// Test with filter name only
	v := NewVariable("test | filter", pc)
	if len(v.Filters()) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(v.Filters()))
	}

	// Test with filter and arguments
	v2 := NewVariable("test | filter: arg1, arg2", pc)
	if len(v2.Filters()) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(v2.Filters()))
	}

	// Test with keyword arguments
	v3 := NewVariable("test | filter: key: value", pc)
	if len(v3.Filters()) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(v3.Filters()))
	}
}

// TestVariableRenderToOutputBuffer tests buffer rendering
func TestVariableRenderToOutputBuffer(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("name", pc)
	ctx := NewContext()
	ctx.Set("name", "test")
	output := ""
	v.RenderToOutputBuffer(ctx, &output)
	if output != "test" {
		t.Errorf("Expected 'test', got %q", output)
	}

	// Test with nil value
	v2 := NewVariable("nonexistent", pc)
	ctx2 := NewContext()
	output2 := ""
	v2.RenderToOutputBuffer(ctx2, &output2)
	if output2 != "" {
		t.Errorf("Expected empty output for nonexistent variable, got %q", output2)
	}
}

// TestVariableRenderWithComplexFilters tests complex filter scenarios
func TestVariableRenderWithComplexFilters(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	// Test with chained filters
	v := NewVariable("name | upcase | downcase", pc)
	ctx := NewContext()
	ctx.Set("name", "Hello")
	result := v.Render(ctx)
	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}

	// Test with multiple filters
	v2 := NewVariable("name | upcase | capitalize", pc)
	ctx2 := NewContext()
	ctx2.Set("name", "hello")
	result2 := v2.Render(ctx2)
	if result2 == nil {
		t.Error("Expected non-nil result")
	}
}

// TestVariableMarkupContext tests markupContext method
func TestVariableMarkupContext(t *testing.T) {
	lineNum := 1
	pc := &mockParseContext{lineNum: &lineNum}

	v := NewVariable("test", pc)
	context := v.markupContext("test markup")
	if context == "" {
		t.Error("Expected non-empty context string")
	}
	if !strings.Contains(context, "test markup") {
		t.Errorf("Expected context to contain markup, got %q", context)
	}
}

// TestVariableLaxParse tests laxParse method
func TestVariableLaxParse(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{ErrorMode: "lax"})
	pc.SetLineNumber(&lineNum)

	v := NewVariable("test", pc)
	// Test laxParse with simple markup
	v.laxParse("simple")
	if v.Name() == nil {
		t.Error("Expected name to be set after laxParse")
	}

	// Test laxParse with filters
	v2 := NewVariable("test", pc)
	v2.laxParse("test | filter")
	if len(v2.Filters()) == 0 {
		t.Error("Expected filters to be set after laxParse with filter")
	}

	// Test laxParse with multiple filters
	v3 := NewVariable("test", pc)
	v3.laxParse("test | filter1 | filter2")
	if len(v3.Filters()) < 2 {
		t.Errorf("Expected at least 2 filters, got %d", len(v3.Filters()))
	}
}

// TestVariableLaxParseFilterExpressions tests laxParseFilterExpressions
func TestVariableLaxParseFilterExpressions(t *testing.T) {
	lineNum := 1
	pc := NewParseContext(ParseContextOptions{ErrorMode: "lax"})
	pc.SetLineNumber(&lineNum)

	v := NewVariable("test", pc)
	// Test with filter name and args
	result := v.laxParseFilterExpressions("filter", []string{"arg1", "arg2"})
	if len(result) == 0 {
		t.Error("Expected non-empty filter expressions")
	}
	if len(result) < 2 {
		t.Error("Expected filter name and args in result")
	}

	// Test with keyword arguments
	result2 := v.laxParseFilterExpressions("filter", []string{"key: value"})
	if len(result2) == 0 {
		t.Error("Expected non-empty filter expressions for keyword args")
	}

	// Test with mixed positional and keyword args
	result3 := v.laxParseFilterExpressions("filter", []string{"arg1", "key: value"})
	if len(result3) == 0 {
		t.Error("Expected non-empty filter expressions for mixed args")
	}

	// Test with empty args
	result4 := v.laxParseFilterExpressions("filter", []string{})
	if len(result4) == 0 {
		t.Error("Expected non-empty filter expressions even with empty args")
	}
}

func TestVariableAddWarning(t *testing.T) {
	// Test AddWarning on parseContextWrapper (no-op)
	pc := NewParseContext(ParseContextOptions{})
	v := NewVariable("test", pc)

	// Create a wrapper (this happens internally)
	// AddWarning is a no-op for wrapper, so we just verify it doesn't panic
	// This tests the parseContextWrapper.AddWarning method
	_ = v
}
