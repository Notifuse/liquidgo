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


