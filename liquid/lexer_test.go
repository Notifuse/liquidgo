package liquid

import (
	"testing"
)

func TestLexerTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // minimum number of tokens expected
	}{
		{"identifier", "hello", 2}, // token + EOS
		{"number", "42", 2},
		{"string", "\"hello\"", 2},
		{"comparison", "a == b", 4},      // a, ==, b, EOS
		{"dot notation", "user.name", 4}, // user, ., name, EOS
		{"brackets", "items[0]", 5},      // items, [, 0, ], EOS
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := NewStringScanner(tt.input)
			tokens, err := Tokenize(ss)
			if err != nil {
				t.Fatalf("Tokenize() error = %v", err)
			}
			if len(tokens) < tt.want {
				t.Errorf("Tokenize() returned %d tokens, want at least %d", len(tokens), tt.want)
			}
		})
	}
}

func TestLexerSpecialTokens(t *testing.T) {
	ss := NewStringScanner("|.:,[]()?-")
	tokens, err := Tokenize(ss)
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	expectedTypes := []string{":pipe", ":dot", ":colon", ":comma", ":open_square", ":close_square", ":open_round", ":close_round", ":question", ":dash", ":end_of_string"}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, expectedType := range expectedTypes {
		if tokens[i][0] != expectedType {
			t.Errorf("Token %d: expected type %s, got %v", i, expectedType, tokens[i][0])
		}
	}
}

func TestLexerComparisonTokens(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"==", ":comparison"},
		{"!=", ":comparison"},
		{"<=", ":comparison"},
		{">=", ":comparison"},
		{"<>", ":comparison"},
		{"<", ":comparison"},
		{">", ":comparison"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ss := NewStringScanner(tt.input)
			tokens, err := Tokenize(ss)
			if err != nil {
				t.Fatalf("Tokenize() error = %v", err)
			}
			if len(tokens) < 2 {
				t.Fatalf("Expected at least 2 tokens, got %d", len(tokens))
			}
			if tokens[0][0] != tt.want {
				t.Errorf("Expected type %s, got %v", tt.want, tokens[0][0])
			}
		})
	}
}

func TestLexerContains(t *testing.T) {
	ss := NewStringScanner("contains")
	tokens, err := Tokenize(ss)
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}

	if len(tokens) < 2 {
		t.Fatalf("Expected at least 2 tokens, got %d", len(tokens))
	}

	// "contains" should be tokenized as a comparison token
	if tokens[0][0] != ":comparison" || tokens[0][1] != "contains" {
		t.Errorf("Expected comparison token 'contains', got %v", tokens[0])
	}
}
