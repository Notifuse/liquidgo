package liquid

import "testing"

// Mock parse context for testing
type mockParseContextForSwitching struct {
	errorMode string
	warnings  []error
}

func (m *mockParseContextForSwitching) ErrorMode() ErrorMode {
	return ErrorMode(m.errorMode)
}

func (m *mockParseContextForSwitching) AddWarning(err error) {
	m.warnings = append(m.warnings, err)
}

func TestParserSwitchingRigidModeCheck(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "rigid"}

	if !rigidMode(pc.ErrorMode()) {
		t.Error("Expected RigidMode() to return true")
	}

	pc.errorMode = "strict"
	if rigidMode(pc.ErrorMode()) {
		t.Error("Expected RigidMode() to return false for strict mode")
	}
}

func TestStrictParseWithErrorModeFallback(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "lax"}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := parseWithErrorModeFallback(pc, "test", nil, MarkupContext, strictParse, laxParse, rigidParse)
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

func TestStrictParseWithErrorModeFallbackRigidMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "rigid"}

	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := parseWithErrorModeFallback(pc, "test", nil, MarkupContext, strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestStrictParseWithErrorModeFallbackStrictMode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict"}

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := parseWithErrorModeFallback(pc, "test", nil, MarkupContext, strictParse, laxParse, rigidParse)
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

	syntaxErr := NewSyntaxError("test error")
	strictParse := func(m string) error { return syntaxErr }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := parseWithErrorModeFallback(pc, "test", nil, MarkupContext, strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error (falls back to lax), got %v", err)
	}
	if len(pc.warnings) == 0 {
		t.Error("Expected warning to be added")
	}
}

func TestStrictParseWithErrorModeFallbackNonSyntaxError(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "lax"}

	strictParse := func(m string) error { return NewInternalError("internal error") }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { return nil }

	err := parseWithErrorModeFallback(pc, "test", nil, MarkupContext, strictParse, laxParse, rigidParse)
	if err == nil {
		t.Error("Expected error for non-SyntaxError")
	}
}

// Tests for strict2 mode (alias for rigid) - Shopify Liquid v5.11.0
func TestStrict2ModeIsAliasForRigid(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict2"}

	if !rigidMode(pc.ErrorMode()) {
		t.Error("Expected strict2 to be treated as rigid mode (RigidMode() should return true)")
	}
}

func TestStrictParseWithErrorModeFallbackStrict2Mode(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict2"}

	rigidCalled := false
	strictParse := func(m string) error { return nil }
	laxParse := func(m string) error { return nil }
	rigidParse := func(m string) error { rigidCalled = true; return nil }

	err := parseWithErrorModeFallback(pc, "test", nil, MarkupContext, strictParse, laxParse, rigidParse)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !rigidCalled {
		t.Error("Expected strict2 to route to rigid parser in deprecated method")
	}
}

// A syntax error of the strict parse gets the line number and the markup
// context.
func TestParseWithErrorModeFallbackErrorContext(t *testing.T) {
	pc := &mockParseContextForSwitching{errorMode: "strict"}
	line := 3
	strictParse := func(string) error { return NewSyntaxError("bad") }
	err := parseWithErrorModeFallback(pc, "test markup", &line, MarkupContext, strictParse, nil, nil)
	syntaxErr, ok := err.(*SyntaxError)
	if !ok || syntaxErr.Err.LineNumber == nil || *syntaxErr.Err.LineNumber != 3 || syntaxErr.Err.MarkupContext != `in "test markup"` {
		t.Errorf("error %#v, want a syntax error at line 3 in \"test markup\"", err)
	}
}
