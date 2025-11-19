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
