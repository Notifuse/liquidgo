package liquid

import (
	"errors"
	"testing"
)

// Mock parse context for testing
type mockParseContextForSwitching struct {
	errorMode string
	warnings  []error
}

func (m *mockParseContextForSwitching) ErrorMode() string {
	return m.errorMode
}

func (m *mockParseContextForSwitching) AddWarning(err error) {
	m.warnings = append(m.warnings, err)
}

func TestParserSwitchingRigidMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "rigid"}
	ps := &ParserSwitching{
		parseContext: pc,
		lineNumber:   intPtr(1),
		markupContext: func(m string) string {
			return "in \"" + m + "\""
		},
	}

	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestParserSwitchingStrictMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict"}
	ps := &ParserSwitching{
		parseContext: pc,
		lineNumber:   intPtr(1),
		markupContext: func(m string) string {
			return "in \"" + m + "\""
		},
	}

	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestParserSwitchingLaxMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "lax"}
	ps := &ParserSwitching{
		parseContext: pc,
	}

	strictParse := func(m string) error { return errors.New("strict error") }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return errors.New("rigid error") }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error in lax mode, got %v", err)
	}
}

func TestParserSwitchingWarnMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "warn"}
	ps := &ParserSwitching{
		parseContext: pc,
	}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return syntaxErr }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error in warn mode (falls back to lax), got %v", err)
	}
	if len(pc.warnings) == 0 {
		t.Error("Expected warning to be added")
	}
}

func TestParserSwitchingRigidModeCheck(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "rigid"}
	ps := &ParserSwitching{parseContext: pc}

	if !ps.RigidMode() {
		t.Error("Expected RigidMode() to return true")
	}

	pc.errorMode = "strict"
	if ps.RigidMode() {
		t.Error("Expected RigidMode() to return false for strict mode")
	}
}

func TestParserSwitchingErrorContext(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict"}
	lineNum := 42
	ps := &ParserSwitching{
		parseContext: pc,
		lineNumber:   &lineNum,
		markupContext: func(m string) string {
			return "in \"" + m + "\""
		},
	}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.ParseWithSelectedParser("test markup", strictParse, laxParse, rigidParse)
	if err == nil {
		t.Fatal("Expected error in strict mode")
	}

	if se, ok := err.(*SyntaxError); ok {
		if se.Err.LineNumber == nil || *se.Err.LineNumber != 42 {
			t.Errorf("Expected line number 42, got %v", se.Err.LineNumber)
		}
		if se.Err.MarkupContext != "in \"test markup\"" {
			t.Errorf("Expected markup context 'in \"test markup\"', got '%s'", se.Err.MarkupContext)
		}
	} else {
		t.Errorf("Expected SyntaxError, got %T", err)
	}
}

func TestStrictParseWithErrorModeFallback(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "lax"}
	ps := &ParserSwitching{parseContext: pc}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.StrictParseWithErrorModeFallback("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error (falls back to lax), got %v", err)
	}
}

func TestMarkupContext(t *testing.T) {
	result := MarkupContext("test markup")
	expected := "in \"test markup\""
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with whitespace
	result2 := MarkupContext("  test  ")
	expected2 := "in \"test\""
	if result2 != expected2 {
		t.Errorf("Expected %q, got %q", expected2, result2)
	}

	// Test with empty string
	result3 := MarkupContext("")
	expected3 := "in \"\""
	if result3 != expected3 {
		t.Errorf("Expected %q, got %q", expected3, result3)
	}
}

func TestParserSwitchingDefaultMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "unknown"}
	ps := &ParserSwitching{parseContext: pc}

	strictParse := func(m string) error { return NewSyntaxError("strict error") }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return NewSyntaxError("rigid error") }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error in default mode (falls back to lax), got %v", err)
	}
}

func TestParserSwitchingWarnModeNonSyntaxError(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "warn"}
	ps := &ParserSwitching{parseContext: pc}

	strictParse := func(m string) error { return NewInternalError("internal error") }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return NewInternalError("internal error") }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err == nil {
		t.Error("Expected error for non-SyntaxError in warn mode")
	}
}

func TestStrictParseWithErrorModeFallbackRigidMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "rigid"}
	ps := &ParserSwitching{parseContext: pc}

	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.StrictParseWithErrorModeFallback("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestStrictParseWithErrorModeFallbackStrictMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict"}
	ps := &ParserSwitching{parseContext: pc}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.StrictParseWithErrorModeFallback("test", strictParse, laxParse, rigidParse)
	if err == nil {
		t.Fatal("Expected error in strict mode")
	}
	if se, ok := err.(*SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T", err)
	} else if se.Err.Message != "test error" {
		t.Errorf("Expected 'test error', got %q", se.Err.Message)
	}
}

func TestStrictParseWithErrorModeFallbackWarnMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "warn"}
	ps := &ParserSwitching{parseContext: pc}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.StrictParseWithErrorModeFallback("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error (falls back to lax), got %v", err)
	}
	if len(pc.warnings) == 0 {
		t.Error("Expected warning to be added")
	}
}

func TestStrictParseWithErrorModeFallbackNonSyntaxError(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "lax"}
	ps := &ParserSwitching{parseContext: pc}

	strictParse := func(m string) error { return NewInternalError("internal error") }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := ps.StrictParseWithErrorModeFallback("test", strictParse, laxParse, rigidParse)
	if err == nil {
		t.Error("Expected error for non-SyntaxError")
	}
}

// Tests for strict2 mode (alias for rigid) - Shopify Liquid v5.11.0
func TestStrict2ModeIsAliasForRigid(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict2"}
	ps := &ParserSwitching{parseContext: pc}

	if !ps.RigidMode() {
		t.Error("Expected strict2 to be treated as rigid mode (RigidMode() should return true)")
	}
}

func TestParseWithSelectedParserStrict2(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict2"}
	ps := &ParserSwitching{
		parseContext: pc,
		lineNumber:   intPtr(1),
		markupContext: func(m string) string {
			return "in \"" + m + "\""
		},
	}

	rigidCalled := false
	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { rigidCalled = true; return nil }

	err := ps.ParseWithSelectedParser("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !rigidCalled {
		t.Error("Expected strict2 to route to rigid parser")
	}
}

func TestStrictParseWithErrorModeFallbackStrict2Mode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict2"}
	ps := &ParserSwitching{parseContext: pc}

	rigidCalled := false
	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { rigidCalled = true; return nil }

	err := ps.StrictParseWithErrorModeFallback("test", strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !rigidCalled {
		t.Error("Expected strict2 to route to rigid parser in deprecated method")
	}
}
