package liquid

import (
	"strings"
	"testing"
)

func TestParserExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"identifier", "hello", "hello"},
		{"number", "42", "42"},
		{"string", "\"hello\"", "\"hello\""},
		{"dot notation", "user.name", "user.name"},
		{"brackets", "items[0]", "items[0]"},
		{"nested", "user.profile.name", "user.profile.name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			result, err := parser.Expression()
			if err != nil {
				t.Fatalf("Expression() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Expression() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParserJump(t *testing.T) {
	parser := NewParser("hello world test")

	// Consume first token
	_, err := parser.Consume(":id")
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	// Jump back to start
	parser.Jump(0)

	// Should be able to consume "hello" again
	result, err := parser.Consume(":id")
	if err != nil {
		t.Fatalf("Consume() after Jump error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello' after Jump(0), got %q", result)
	}

	// Jump to middle
	parser.Jump(1)
	result2, err := parser.Consume(":id")
	if err != nil {
		t.Fatalf("Consume() after Jump(1) error = %v", err)
	}
	if result2 != "world" {
		t.Errorf("Expected 'world' after Jump(1), got %q", result2)
	}
}

func TestParserRange(t *testing.T) {
	parser := NewParser("(1..10)")
	result, err := parser.Expression()
	if err != nil {
		t.Fatalf("Expression() error = %v", err)
	}
	if result != "(1..10)" {
		t.Errorf("Expression() = %v, want (1..10)", result)
	}
}

func TestParserConsume(t *testing.T) {
	parser := NewParser("hello world")

	first, err := parser.Consume(":id")
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if first != "hello" {
		t.Errorf("Consume() = %v, want hello", first)
	}

	// Should fail on wrong type
	_, err = parser.Consume(":number")
	if err == nil {
		t.Error("Expected error when consuming wrong type")
	}
}

func TestParserConsumeOptional(t *testing.T) {
	parser := NewParser("hello")

	val, ok := parser.ConsumeOptional(":id")
	if !ok {
		t.Error("Expected ConsumeOptional to succeed")
	}
	if val != "hello" {
		t.Errorf("ConsumeOptional() = %v, want hello", val)
	}

	// Try consuming something that doesn't match
	_, ok = parser.ConsumeOptional(":number")
	if ok {
		t.Error("Expected ConsumeOptional to fail")
	}
}

func TestParserID(t *testing.T) {
	parser := NewParser("hello")

	val, ok := parser.ID("hello")
	if !ok {
		t.Error("Expected ID() to succeed")
	}
	if val != "hello" {
		t.Errorf("ID() = %v, want hello", val)
	}

	// Try wrong ID
	parser = NewParser("world")
	_, ok = parser.ID("hello")
	if ok {
		t.Error("Expected ID() to fail for wrong ID")
	}
}

func TestParserLook(t *testing.T) {
	parser := NewParser("hello world")

	if !parser.Look(":id", 0) {
		t.Error("Expected Look() to return true")
	}
	if parser.Look(":number", 0) {
		t.Error("Expected Look() to return false for wrong type")
	}
	if !parser.Look(":id", 1) {
		t.Error("Expected Look() to return true for next token")
	}
}

func TestParserVariableLookups(t *testing.T) {
	parser := NewParser("user.name")

	expr, err := parser.Expression()
	if err != nil {
		t.Fatalf("Expression() error = %v", err)
	}
	if expr != "user.name" {
		t.Errorf("Expression() = %v, want user.name", expr)
	}
}

func TestParserArgument(t *testing.T) {
	parser := NewParser("name: value")

	arg, err := parser.Argument()
	if err != nil {
		t.Fatalf("Argument() error = %v", err)
	}
	if arg == "" {
		t.Error("Expected Argument() to return non-empty string")
	}
}

func TestParserExpressionArrayBrackets(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple array access", "[0]", "[0]"},
		{"array access with expression", "[items[0]]", "[items[0]]"},
		{"nested brackets", "[[0]]", "[[0]]"},
		{"array with string index", "[\"key\"]", "[\"key\"]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			result, err := parser.Expression()
			if err != nil {
				t.Fatalf("Expression() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Expression() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParserExpressionRange(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"number range", "(1..10)", "(1..10)"},
		{"string range", "(\"a\"..\"z\")", "(\"a\"..\"z\")"},
		{"variable range", "(start..end)", "(start..end)"},
		{"complex range", "(items[0]..items[1])", "(items[0]..items[1])"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			result, err := parser.Expression()
			if err != nil {
				t.Fatalf("Expression() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Expression() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParserExpressionErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"invalid token", "+++"},
		{"incomplete range", "(1.."},
		{"incomplete brackets", "[0"},
		{"unclosed brackets", "items[0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Expression()
			if err == nil {
				t.Error("Expected error from Expression()")
			}
		})
	}
}

func TestParserExpressionComplex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"nested array access", "items[0][1]", "items[0][1]"},
		{"array with dot notation", "items[0].name", "items[0].name"},
		{"complex nested", "user.profile.items[0].name", "user.profile.items[0].name"},
		{"brackets with expression", "items[user.id]", "items[user.id]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			result, err := parser.Expression()
			if err != nil {
				t.Fatalf("Expression() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Expression() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParserConsumeEndOfStream(t *testing.T) {
	parser := NewParser("hello")
	_, _ = parser.Consume(":id") // Consume "hello"

	// Should error on end of stream
	_, err := parser.Consume(":id")
	if err == nil {
		t.Error("Expected error when consuming at end of stream")
	}
}

func TestParserConsumeOptionalEndOfStream(t *testing.T) {
	parser := NewParser("hello")
	_, _ = parser.ConsumeOptional(":id") // Consume "hello"

	// Should return false at end of stream
	_, ok := parser.ConsumeOptional(":id")
	if ok {
		t.Error("Expected ConsumeOptional to return false at end of stream")
	}
}

func TestParserLookAhead(t *testing.T) {
	parser := NewParser("hello world test")

	// Look ahead beyond available tokens
	if parser.Look(":id", 10) {
		t.Error("Expected Look() to return false when looking too far ahead")
	}
}

func TestParserIDEndOfStream(t *testing.T) {
	parser := NewParser("hello")
	_, _ = parser.Consume(":id") // Consume "hello"

	// Should return false at end of stream
	_, ok := parser.ID("world")
	if ok {
		t.Error("Expected ID() to return false at end of stream")
	}
}

func TestParserNewParserWithStringScanner(t *testing.T) {
	ss := NewStringScanner("test")
	parser := NewParser(ss)
	if parser == nil {
		t.Fatal("Expected parser, got nil")
	}
}

func TestParserNewParserWithOtherType(t *testing.T) {
	// Test with non-string, non-scanner type
	parser := NewParser(42)
	if parser == nil {
		t.Fatal("Expected parser, got nil")
	}
}

func TestParserConsumeWithNilTokenType(t *testing.T) {
	parser := NewParser("hello world")

	// Consume with nil tokenType should accept any token
	result, err := parser.Consume(nil)
	if err != nil {
		t.Fatalf("Consume(nil) error = %v", err)
	}
	if result != "hello" {
		t.Errorf("Consume(nil) = %v, want hello", result)
	}

	// Consume another with nil tokenType
	result2, err := parser.Consume(nil)
	if err != nil {
		t.Fatalf("Consume(nil) error = %v", err)
	}
	if result2 != "world" {
		t.Errorf("Consume(nil) = %v, want world", result2)
	}
}

func TestParserConsumeTokenWithNilValue(t *testing.T) {
	// Create a parser with tokens that have nil values
	// This happens with certain token types like operators
	parser := &Parser{
		tokens: []Token{
			{":dot", nil},
			{":colon", nil},
		},
		p: 0,
	}

	result, err := parser.Consume(":dot")
	if err != nil {
		t.Fatalf("Consume(:dot) error = %v", err)
	}
	if result != "" {
		t.Errorf("Consume(:dot) with nil value = %v, want empty string", result)
	}

	result2, err := parser.Consume(":colon")
	if err != nil {
		t.Fatalf("Consume(:colon) error = %v", err)
	}
	if result2 != "" {
		t.Errorf("Consume(:colon) with nil value = %v, want empty string", result2)
	}
}

func TestParserConsumeOptionalWithNilValue(t *testing.T) {
	// Create a parser with tokens that have nil values
	parser := &Parser{
		tokens: []Token{
			{":dot", nil},
			{":open_square", nil},
		},
		p: 0,
	}

	result, ok := parser.ConsumeOptional(":dot")
	if !ok {
		t.Fatal("ConsumeOptional(:dot) should succeed")
	}
	if result != "" {
		t.Errorf("ConsumeOptional(:dot) with nil value = %v, want empty string", result)
	}

	result2, ok := parser.ConsumeOptional(":open_square")
	if !ok {
		t.Fatal("ConsumeOptional(:open_square) should succeed")
	}
	if result2 != "" {
		t.Errorf("ConsumeOptional(:open_square) with nil value = %v, want empty string", result2)
	}
}

func TestParserIDWithNonIDToken(t *testing.T) {
	// Create parser with non-ID token
	parser := NewParser("42")

	// Try to match ID, but token is :number
	_, ok := parser.ID("42")
	if ok {
		t.Error("Expected ID() to return false when token is not :id type")
	}
}

func TestParserExpressionInvalidToken(t *testing.T) {
	// Create parser with invalid token type for expression
	parser := &Parser{
		tokens: []Token{
			{":colon", ":"},
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for invalid expression token")
	}
}

func TestParserVariableLookupsErrorInBracketExpression(t *testing.T) {
	// Create parser with incomplete bracket expression
	parser := &Parser{
		tokens: []Token{
			{":open_square", "["},
			{":close_square", "]"}, // Missing expression between brackets
		},
		p: 0,
	}

	_, err := parser.VariableLookups()
	if err == nil {
		t.Error("Expected error for incomplete bracket expression in VariableLookups")
	}
}

func TestParserVariableLookupsErrorMissingCloseBracket(t *testing.T) {
	// Create parser with missing close bracket
	parser := &Parser{
		tokens: []Token{
			{":open_square", "["},
			{":number", "0"},
			// Missing :close_square
		},
		p: 0,
	}

	_, err := parser.VariableLookups()
	if err == nil {
		t.Error("Expected error for missing close bracket in VariableLookups")
	}
}

func TestParserVariableLookupsErrorAfterDot(t *testing.T) {
	// Create parser with dot but no following identifier
	parser := &Parser{
		tokens: []Token{
			{":dot", "."},
			{":number", "42"}, // Not an :id after dot
		},
		p: 0,
	}

	_, err := parser.VariableLookups()
	if err == nil {
		t.Error("Expected error for non-identifier after dot in VariableLookups")
	}
}

func TestParserArgumentErrorInExpression(t *testing.T) {
	// Create parser with invalid expression in argument
	parser := &Parser{
		tokens: []Token{
			{":colon", ":"}, // Invalid token for expression
		},
		p: 0,
	}

	_, err := parser.Argument()
	if err == nil {
		t.Error("Expected error for invalid expression in Argument")
	}
}

func TestParserExpressionOpenSquareWithError(t *testing.T) {
	// Test :open_square case with error in nested expression
	parser := &Parser{
		tokens: []Token{
			{":open_square", "["},
			{":colon", ":"}, // Invalid expression token
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for invalid nested expression in open_square case")
	}
}

func TestParserExpressionOpenSquareMissingClose(t *testing.T) {
	// Test :open_square case with missing close bracket
	parser := &Parser{
		tokens: []Token{
			{":open_square", "["},
			{":number", "0"},
			// Missing :close_square
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for missing close bracket in Expression")
	}
}

func TestParserExpressionOpenRoundErrorInFirst(t *testing.T) {
	// Test :open_round case with error in first expression
	parser := &Parser{
		tokens: []Token{
			{":open_round", "("},
			{":colon", ":"}, // Invalid expression token
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for invalid first expression in range")
	}
}

func TestParserExpressionOpenRoundMissingDotDot(t *testing.T) {
	// Test :open_round case with missing dotdot
	parser := &Parser{
		tokens: []Token{
			{":open_round", "("},
			{":number", "1"},
			{":number", "10"}, // Missing :dotdot
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for missing dotdot in range")
	}
}

func TestParserExpressionOpenRoundErrorInLast(t *testing.T) {
	// Test :open_round case with error in last expression
	parser := &Parser{
		tokens: []Token{
			{":open_round", "("},
			{":number", "1"},
			{":dotdot", ".."},
			{":colon", ":"}, // Invalid expression token
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for invalid last expression in range")
	}
}

func TestParserExpressionOpenRoundMissingCloseRound(t *testing.T) {
	// Test :open_round case with missing close round
	parser := &Parser{
		tokens: []Token{
			{":open_round", "("},
			{":number", "1"},
			{":dotdot", ".."},
			{":number", "10"},
			// Missing :close_round
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for missing close round bracket in range")
	}
}

func TestParserExpressionIDWithVariableLookupsError(t *testing.T) {
	// Test :id case with error in variable lookups
	parser := &Parser{
		tokens: []Token{
			{":id", "test"},
			{":dot", "."},
			{":number", "42"}, // Not an :id after dot - will cause error in VariableLookups
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for invalid variable lookup in Expression")
	}
}

func TestParserExpressionOpenSquareWithVariableLookupsError(t *testing.T) {
	// Test :open_square case with error in variable lookups after close bracket
	parser := &Parser{
		tokens: []Token{
			{":open_square", "["},
			{":number", "0"},
			{":close_square", "]"},
			{":dot", "."},
			{":number", "42"}, // Not an :id after dot - will cause error in VariableLookups
		},
		p: 0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error for invalid variable lookup after bracket expression")
	}
}

func TestParserArgumentWithKeyword(t *testing.T) {
	// Test keyword argument: id: expression
	parser := NewParser("name: \"John\"")

	arg, err := parser.Argument()
	if err != nil {
		t.Fatalf("Argument() error = %v", err)
	}
	if arg == "" {
		t.Error("Expected Argument() to return non-empty string for keyword argument")
	}
}

func TestParserConsumeOptionalAtBoundary(t *testing.T) {
	// Test ConsumeOptional at token boundary
	parser := NewParser("hello")

	// First consume should succeed
	val1, ok1 := parser.ConsumeOptional(":id")
	if !ok1 {
		t.Error("First ConsumeOptional should succeed")
	}
	if val1 != "hello" {
		t.Errorf("ConsumeOptional() = %v, want hello", val1)
	}

	// At end of tokens (EOS token), should return false
	val2, ok2 := parser.ConsumeOptional(":id")
	if ok2 {
		t.Error("ConsumeOptional should return false at EOS")
	}
	if val2 != "" {
		t.Errorf("ConsumeOptional at EOS should return empty string, got %v", val2)
	}
}

func TestParserExpressionEmptyAtEOS(t *testing.T) {
	// Create parser already at end of stream
	parser := &Parser{
		tokens: []Token{},
		p:      0,
	}

	_, err := parser.Expression()
	if err == nil {
		t.Error("Expected error when calling Expression() at end of stream")
	}
}

func TestParserConsumeErrorMessage(t *testing.T) {
	parser := NewParser("hello")

	// Try to consume wrong type and check error message
	_, err := parser.Consume(":number")
	if err == nil {
		t.Fatal("Expected error when consuming wrong type")
	}

	// Error message should be user-friendly (without colon prefixes)
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Expected") {
		t.Errorf("Error message should contain 'Expected', got: %v", errMsg)
	}
}

func TestParserExpressionStringAndNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single quoted string", "'hello'", "'hello'"},
		{"double quoted string", "\"world\"", "\"world\""},
		{"integer", "123", "123"},
		{"float", "45.67", "45.67"},
		{"negative number", "-99", "-99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			result, err := parser.Expression()
			if err != nil {
				t.Fatalf("Expression() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Expression() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParserVariableLookupsEmpty(t *testing.T) {
	// Create parser with no lookups following
	parser := &Parser{
		tokens: []Token{
			{":id", "test"},
		},
		p: 1, // Position after "test"
	}

	// VariableLookups with no actual lookups should return empty string
	result, err := parser.VariableLookups()
	if err != nil {
		t.Fatalf("VariableLookups() error = %v", err)
	}
	if result != "" {
		t.Errorf("VariableLookups() with no lookups = %v, want empty string", result)
	}
}

func TestParserNewParserTokenizeError(t *testing.T) {
	// NewParser should handle tokenization errors gracefully
	// Create input that might cause tokenization issues
	parser := NewParser("")
	if parser == nil {
		t.Fatal("NewParser should return a parser even with empty input")
	}
	if len(parser.tokens) == 0 {
		t.Error("NewParser should have at least EOS token")
	}
}

func TestParserConsumeOptionalWithValue(t *testing.T) {
	// Test ConsumeOptional with token that has a non-nil value
	parser := NewParser("hello world")

	// Consume first token which should have value "hello"
	val, ok := parser.ConsumeOptional(":id")
	if !ok {
		t.Fatal("ConsumeOptional should succeed for :id")
	}
	if val != "hello" {
		t.Errorf("ConsumeOptional() = %v, want hello", val)
	}

	// Consume second token which should have value "world"
	val2, ok2 := parser.ConsumeOptional(":id")
	if !ok2 {
		t.Fatal("ConsumeOptional should succeed for second :id")
	}
	if val2 != "world" {
		t.Errorf("ConsumeOptional() = %v, want world", val2)
	}
}

func TestParserIDMatchingValue(t *testing.T) {
	// Test ID() when token is :id AND value matches
	parser := NewParser("test")

	val, ok := parser.ID("test")
	if !ok {
		t.Fatal("ID() should succeed when value matches")
	}
	if val != "test" {
		t.Errorf("ID() = %v, want test", val)
	}

	// Parser position should have advanced
	if parser.p != 1 {
		t.Errorf("Parser position should be 1 after ID(), got %v", parser.p)
	}
}

func TestParserExpressionIDPath(t *testing.T) {
	// Ensure the :id path in Expression is fully covered
	parser := NewParser("myvar")

	result, err := parser.Expression()
	if err != nil {
		t.Fatalf("Expression() error = %v", err)
	}
	if result != "myvar" {
		t.Errorf("Expression() = %v, want myvar", result)
	}
}

func TestParserExpressionStringPath(t *testing.T) {
	// Ensure the :string path in Expression is fully covered
	parser := NewParser("\"test string\"")

	result, err := parser.Expression()
	if err != nil {
		t.Fatalf("Expression() error = %v", err)
	}
	if result != "\"test string\"" {
		t.Errorf("Expression() = %v, want \"test string\"", result)
	}
}

func TestParserExpressionNumberPath(t *testing.T) {
	// Ensure the :number path in Expression is fully covered
	parser := NewParser("42")

	result, err := parser.Expression()
	if err != nil {
		t.Fatalf("Expression() error = %v", err)
	}
	if result != "42" {
		t.Errorf("Expression() = %v, want 42", result)
	}
}

func TestParserConsumeAtEndBoundary(t *testing.T) {
	// Test Consume when exactly at the end
	parser := &Parser{
		tokens: []Token{
			{":id", "test"},
		},
		p: 0,
	}

	// First consume should work
	_, err := parser.Consume(":id")
	if err != nil {
		t.Fatalf("First Consume() should succeed, got error: %v", err)
	}

	// Now at position 1, which is >= len(tokens)
	_, err = parser.Consume(":id")
	if err == nil {
		t.Error("Consume() at end should return error")
	}
}

func TestParserExpressionConsumeInIDCase(t *testing.T) {
	// Test the Consume call within the :id case of Expression
	// This is covered by other tests, but ensure explicit coverage
	parser := NewParser("variable")

	result, err := parser.Expression()
	if err != nil {
		t.Fatalf("Expression() error = %v", err)
	}
	if result != "variable" {
		t.Errorf("Expression() = %v, want variable", result)
	}
}
