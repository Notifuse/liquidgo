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
