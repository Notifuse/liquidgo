package liquid

import (
	"testing"
)

func TestNewParseContext(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	if pc == nil {
		t.Fatal("Expected ParseContext, got nil")
	}
	if pc.Environment() == nil {
		t.Error("Expected environment, got nil")
	}
	if pc.Locale() == nil {
		t.Error("Expected locale, got nil")
	}
}

func TestParseContextParseExpressionSafe(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{ErrorMode: "lax"})

	// Test with safe = true
	result := pc.ParseExpressionSafe("42", true)
	if result == nil {
		t.Error("Expected non-nil result from ParseExpressionSafe")
	}

	// Test with safe = false in lax mode (should work)
	result2 := pc.ParseExpressionSafe("100", false)
	if result2 == nil {
		t.Error("Expected non-nil result from ParseExpressionSafe with safe=false in lax mode")
	}

	// Test with safe = false in rigid mode (should panic)
	pc2 := NewParseContext(ParseContextOptions{ErrorMode: "rigid"})
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for unsafe ParseExpressionSafe in rigid mode")
			}
		}()
		pc2.ParseExpressionSafe("42", false)
	}()
}

func TestParseContextWithOptions(t *testing.T) {
	env := NewEnvironment()
	env.SetErrorMode("strict")
	locale := NewI18n("en")

	pc := NewParseContext(ParseContextOptions{
		Environment: env,
		Locale:      locale,
		ErrorMode:   "lax",
	})

	if pc.Environment() != env {
		t.Error("Environment mismatch")
	}
	if pc.Locale() != locale {
		t.Error("Locale mismatch")
	}
	if pc.ErrorMode() != "lax" {
		t.Errorf("Expected error mode 'lax', got '%s'", pc.ErrorMode())
	}
}

func TestParseContextLineNumber(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	lineNum := 42
	pc.SetLineNumber(&lineNum)

	if pc.LineNumber() == nil || *pc.LineNumber() != 42 {
		t.Errorf("Expected line number 42, got %v", pc.LineNumber())
	}
}

func TestParseContextTrimWhitespace(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	pc.SetTrimWhitespace(true)

	if !pc.TrimWhitespace() {
		t.Error("Expected trim whitespace to be true")
	}
}

func TestParseContextDepth(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	if pc.Depth() != 0 {
		t.Errorf("Expected depth 0, got %d", pc.Depth())
	}

	pc.IncrementDepth()
	if pc.Depth() != 1 {
		t.Errorf("Expected depth 1, got %d", pc.Depth())
	}

	pc.DecrementDepth()
	if pc.Depth() != 0 {
		t.Errorf("Expected depth 0 after decrement, got %d", pc.Depth())
	}
}

func TestParseContextPartial(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	if pc.Partial() {
		t.Error("Expected partial to be false initially")
	}

	pc.SetPartial(true)
	if !pc.Partial() {
		t.Error("Expected partial to be true")
	}
}

func TestParseContextWarnings(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	if len(pc.Warnings()) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(pc.Warnings()))
	}

	warning := NewSyntaxError("test warning")
	pc.AddWarning(warning)

	if len(pc.Warnings()) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(pc.Warnings()))
	}
}

func TestParseContextNewBlockBody(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	bb := pc.NewBlockBody()
	if bb == nil {
		t.Fatal("Expected BlockBody, got nil")
	}
}

func TestParseContextNewParser(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	parser := pc.NewParser("1 + 2")
	if parser == nil {
		t.Fatal("Expected Parser, got nil")
	}
}

func TestParseContextNewTokenizer(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	lineNum := 1
	tokenizer := pc.NewTokenizer("hello {{ world }}", true, &lineNum, false)
	if tokenizer == nil {
		t.Fatal("Expected Tokenizer, got nil")
	}
}

func TestParseContextParseExpression(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	result := pc.ParseExpression("42")
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestParseContextSafeParseExpression(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})
	parser := pc.NewParser("42")
	result := pc.SafeParseExpression(parser)
	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestParseContextGetOption(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{
		TemplateOptions: map[string]interface{}{
			"test_option": "test_value",
		},
	})

	val := pc.GetOption("test_option")
	if val != "test_value" {
		t.Errorf("Expected 'test_value', got %v", val)
	}
}

func TestParseContextPartialOptions(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{
		TemplateOptions: map[string]interface{}{
			"test_option":               "test_value",
			"include_options_blacklist": true,
		},
	})

	pc.SetPartial(true)
	val := pc.GetOption("locale")
	if val == nil {
		t.Error("Expected locale in partial options")
	}

	val = pc.GetOption("test_option")
	if val != nil {
		t.Error("Expected test_option to be excluded from partial options")
	}
}

func TestParseContextPartialOptionsWithBlacklistArray(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{
		TemplateOptions: map[string]interface{}{
			"test_option":               "test_value",
			"another_option":            "another_value",
			"include_options_blacklist": []string{"test_option"},
		},
	})

	pc.SetPartial(true)

	// test_option should be excluded
	val := pc.GetOption("test_option")
	if val != nil {
		t.Error("Expected test_option to be excluded from partial options")
	}

	// another_option should be included
	val = pc.GetOption("another_option")
	if val != "another_value" {
		t.Errorf("Expected 'another_value', got %v", val)
	}
}

func TestParseContextPartialOptionsNoBlacklist(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{
		TemplateOptions: map[string]interface{}{
			"test_option": "test_value",
		},
	})

	pc.SetPartial(true)

	// All options should be included when no blacklist
	val := pc.GetOption("test_option")
	if val != "test_value" {
		t.Errorf("Expected 'test_value', got %v", val)
	}
}

func TestParseContextPartialOptionsSetToFalse(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{
		TemplateOptions: map[string]interface{}{
			"test_option": "test_value",
		},
	})

	pc.SetPartial(true)
	pc.SetPartial(false)

	// After setting partial to false, should use template options directly
	val := pc.GetOption("test_option")
	if val != "test_value" {
		t.Errorf("Expected 'test_value', got %v", val)
	}

	// Error mode should reset to environment default
	if pc.ErrorMode() != pc.Environment().ErrorMode() {
		t.Error("Expected error mode to reset to environment default")
	}
}

func TestParseContextSafeParseExpressionStrictMode(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{ErrorMode: "strict"})
	parser := pc.NewParser("invalid+++")

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic in strict mode for invalid expression")
			}
		}()
		pc.SafeParseCompleteExpression(parser)
	}()
}

func TestParseContextSafeParseExpressionRigidMode(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{ErrorMode: "rigid"})
	parser := pc.NewParser("invalid+++")

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic in rigid mode for invalid expression")
			}
		}()
		pc.SafeParseCompleteExpression(parser)
	}()
}

func TestParseContextSafeParseExpressionLaxMode(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{ErrorMode: "lax"})
	parser := pc.NewParser("invalid+++")

	// Should not panic in lax mode
	result := pc.SafeParseCompleteExpression(parser)
	if result != nil {
		t.Errorf("Expected nil result in lax mode for invalid expression, got %v", result)
	}
}
