package liquid

import (
	"testing"
)

func TestTokenizerBasic(t *testing.T) {
	source := "Hello {{ name }} World"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, false)

	tokens := []string{}
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}
		tokens = append(tokens, token)
	}

	if len(tokens) < 3 {
		t.Fatalf("Expected at least 3 tokens, got %d", len(tokens))
	}

	// Should have: "Hello ", "{{ name }}", " World"
	expected := "Hello "
	if tokens[0] != expected {
		t.Errorf("Expected first token '%s', got '%s'", expected, tokens[0])
	}
}

func TestTokenizerTags(t *testing.T) {
	source := "{% if condition %}Hello{% endif %}"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, false)

	tokens := []string{}
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}
		tokens = append(tokens, token)
	}

	if len(tokens) < 3 {
		t.Fatalf("Expected at least 3 tokens, got %d", len(tokens))
	}

	// Should have text, tag, text, tag
	foundTag := false
	for _, token := range tokens {
		if len(token) > 2 && token[0:2] == "{%" {
			foundTag = true
			break
		}
	}
	if !foundTag {
		t.Error("Expected to find a tag token")
	}
}

func TestTokenizerVariables(t *testing.T) {
	source := "Hello {{ name }}"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, false)

	tokens := []string{}
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}
		tokens = append(tokens, token)
	}

	foundVariable := false
	for _, token := range tokens {
		if len(token) > 2 && token[0:2] == "{{" {
			foundVariable = true
			break
		}
	}
	if !foundVariable {
		t.Error("Expected to find a variable token")
	}
}

func TestTokenizerForLiquidTag(t *testing.T) {
	source := "line1\nline2\nline3"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, true)

	tokens := []string{}
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}
		tokens = append(tokens, token)
	}

	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens for liquid tag, got %d", len(tokens))
	}
}

func TestTokenizerLineNumbers(t *testing.T) {
	source := "line1\nline2\nline3"
	ss := NewStringScanner(source)
	lineNum := 1
	tokenizer := NewTokenizer(source, ss, true, &lineNum, false)

	// Shift a few tokens
	tokenizer.Shift()
	tokenizer.Shift()

	lineNumber := tokenizer.LineNumber()
	if lineNumber == nil {
		t.Error("Expected line number to be set")
	} else if *lineNumber < 1 {
		t.Errorf("Expected line number >= 1, got %d", *lineNumber)
	}
}

func TestTokenizerNextTagTokenWithStart(t *testing.T) {
	// Test nextTagTokenWithStart indirectly through tokenization
	// This method is called internally by nextVariableToken when it encounters {% inside {{ }}
	// Testing it directly is complex due to internal state requirements
	// Instead, we test the edge cases that exercise this code path

	// Test with nested tags in variables (which uses nextTagTokenWithStart)
	source := "text {{ var {% tag %} }} more"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, false)

	// Tokenize and verify we can handle nested tags
	tokens := []string{}
	for i := 0; i < 10; i++ {
		token := tokenizer.Shift()
		if token == "" {
			break
		}
		tokens = append(tokens, token)
	}

	// Should have parsed tokens successfully
	if len(tokens) == 0 {
		t.Error("Expected at least one token")
	}
}

func TestTokenizerNextVariableTokenEdgeCases(t *testing.T) {
	// Test edge cases through public tokenization API
	// These exercise nextVariableToken and nextTagTokenWithStart internally

	// Test with unclosed variable
	source := "{{ unclosed"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, false)

	// Tokenize and check behavior with unclosed variable
	token := tokenizer.Shift()
	if token == "" {
		t.Error("Expected at least one token")
	}

	// Test with nested braces
	source2 := "{{ outer {{ inner }} }}"
	ss2 := NewStringScanner(source2)
	tokenizer2 := NewTokenizer(source2, ss2, false, nil, false)
	token2 := tokenizer2.Shift()
	if len(token2) == 0 {
		t.Error("Expected non-empty token for nested braces")
	}
}

func TestTokenizerNextTagTokenEdgeCases(t *testing.T) {
	// Test edge cases through public tokenization API
	// These exercise nextTagToken internally

	// Test with unclosed tag
	source := "{% unclosed"
	ss := NewStringScanner(source)
	tokenizer := NewTokenizer(source, ss, false, nil, false)

	// Tokenize and check behavior with unclosed tag
	token := tokenizer.Shift()
	if token == "" {
		t.Error("Expected at least one token")
	}
	// Should handle unclosed tag gracefully
	if len(token) < 2 {
		t.Errorf("Expected token length >= 2, got %d", len(token))
	}
}
