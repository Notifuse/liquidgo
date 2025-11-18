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
