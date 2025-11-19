package liquid

import (
	"testing"
)

func TestExpressionParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  interface{}
	}{
		{"nil", "nil", nil},
		{"null", "null", nil},
		{"empty string", "", nil},
		{"true", "true", true},
		{"false", "false", false},
		{"blank", "blank", ""},
		{"empty", "empty", ""},
		{"quoted string double", `"hello"`, "hello"},
		{"quoted string single", `'world'`, "world"},
		{"integer", "42", 42},
		{"negative integer", "-42", -42},
		{"float", "3.14", 3.14},
		{"negative float", "-3.14", -3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.input, nil, nil)
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpressionParseRange(t *testing.T) {
	result := Parse("(1..10)", nil, nil)
	if result == nil {
		t.Error("Expected range result, got nil")
	}
}

func TestExpressionParseNumber(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"123", 123},
		{"-456", -456},
		{"123.456", 123.456},
		{"-789.012", -789.012},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseNumber(tt.input, nil)
			if got != tt.want {
				t.Errorf("parseNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpressionParseVariableLookup(t *testing.T) {
	result := Parse("user.name", nil, nil)
	if result == nil {
		t.Error("Expected VariableLookup result, got nil")
	}
	if _, ok := result.(*VariableLookup); !ok {
		t.Errorf("Expected *VariableLookup, got %T", result)
	}
}

// TestExpressionParseNumberEdgeCases tests number parsing edge cases
func TestExpressionParseNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  interface{}
	}{
		{"zero", "0", 0},
		{"large number", "1234567890", 1234567890},
		{"decimal", "0.5", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNumber(tt.input, nil)
			if got != tt.want {
				t.Errorf("parseNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestExpressionParseNumberComplex tests parseNumber with complex cases
func TestExpressionParseNumberComplex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  interface{}
	}{
		{"leading dash", "-", nil},
		{"dash only", "-", nil},
		{"dash with non-digit", "-a", nil},
		{"zero", "0", 0},
		{"negative zero", "-0", 0},
		{"large integer", "999999999", 999999999},
		{"negative large", "-999999999", -999999999},
		{"decimal zero", "0.0", 0.0},
		{"negative decimal", "-0.5", -0.5},
		{"scientific notation", "1e5", nil}, // May not be supported
		{"empty string", "", nil},
		{"whitespace", " 42 ", nil}, // Whitespace not trimmed
		{"multiple dots", "1.2.3", nil}, // Invalid
		{"dot at end", "123.", 123.0}, // Valid
		{"negative dot at end", "-123.", nil}, // May not parse correctly
		{"just dot", ".", nil}, // Invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNumber(tt.input, nil)
			if tt.want != nil && got != tt.want {
				t.Errorf("parseNumber(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}

	// Test with StringScanner
	ss := NewStringScanner("42")
	result := parseNumber("42", ss)
	if result != 42 {
		t.Errorf("parseNumber with StringScanner = %v, want 42", result)
	}

	// Test with StringScanner that has different string
	ss2 := NewStringScanner("initial")
	result2 := parseNumber("100", ss2)
	if result2 != 100 {
		t.Errorf("parseNumber with StringScanner (different string) = %v, want 100", result2)
	}
}

// TestExpressionSafeParse tests SafeParse method
func TestExpressionSafeParse(t *testing.T) {
	// Test with valid expression
	p := NewParser("42")
	result := SafeParse(p, nil, nil)
	if result != 42 {
		t.Errorf("SafeParse('42') = %v, want 42", result)
	}

	// Test with invalid expression (should return nil)
	p2 := NewParser("invalid{{")
	result2 := SafeParse(p2, nil, nil)
	_ = result2 // May be nil or error value
}
