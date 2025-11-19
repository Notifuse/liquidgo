package liquid

import (
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
