package liquid

import (
	"fmt"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	err := &Error{
		Message: "test error",
	}

	msg := err.Error()
	if !strings.Contains(msg, "Liquid error") {
		t.Errorf("Expected error message to contain 'Liquid error', got: %s", msg)
	}
	if !strings.Contains(msg, "test error") {
		t.Errorf("Expected error message to contain 'test error', got: %s", msg)
	}
}

func TestErrorWithLineNumber(t *testing.T) {
	lineNum := 42
	err := &Error{
		Message:    "test error",
		LineNumber: &lineNum,
	}

	msg := err.Error()
	if !strings.Contains(msg, "line 42") {
		t.Errorf("Expected error message to contain 'line 42', got: %s", msg)
	}
}

func TestErrorWithTemplateName(t *testing.T) {
	lineNum := 10
	err := &Error{
		Message:      "test error",
		TemplateName: "template.liquid",
		LineNumber:   &lineNum,
	}

	msg := err.Error()
	if !strings.Contains(msg, "template.liquid") {
		t.Errorf("Expected error message to contain 'template.liquid', got: %s", msg)
	}
	if !strings.Contains(msg, "line 10") {
		t.Errorf("Expected error message to contain 'line 10', got: %s", msg)
	}
}

func TestErrorWithMarkupContext(t *testing.T) {
	err := &Error{
		Message:       "test error",
		MarkupContext: "in \"{{name}}\"",
	}

	msg := err.Error()
	if !strings.Contains(msg, "in \"{{name}}\"") {
		t.Errorf("Expected error message to contain markup context, got: %s", msg)
	}
}

func TestSyntaxError(t *testing.T) {
	err := NewSyntaxError("syntax error")
	msg := err.Error()

	if !strings.Contains(msg, "Liquid syntax error") {
		t.Errorf("Expected SyntaxError to contain 'Liquid syntax error', got: %s", msg)
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name    string
		create  func(string) error
		wantErr string
	}{
		{"ArgumentError", func(msg string) error { return NewArgumentError(msg) }, "test"},
		{"ContextError", func(msg string) error { return NewContextError(msg) }, "test"},
		{"FileSystemError", func(msg string) error { return NewFileSystemError(msg) }, "test"},
		{"StandardError", func(msg string) error { return NewStandardError(msg) }, "test"},
		{"SyntaxError", func(msg string) error { return NewSyntaxError(msg) }, "test"},
		{"StackLevelError", func(msg string) error { return NewStackLevelError(msg) }, "test"},
		{"MemoryError", func(msg string) error { return NewMemoryError(msg) }, "test"},
		{"ZeroDivisionError", func(msg string) error { return NewZeroDivisionError(msg) }, "test"},
		{"FloatDomainError", func(msg string) error { return NewFloatDomainError(msg) }, "test"},
		{"UndefinedVariable", func(msg string) error { return NewUndefinedVariable(msg) }, "test"},
		{"UndefinedDropMethod", func(msg string) error { return NewUndefinedDropMethod(msg) }, "test"},
		{"UndefinedFilter", func(msg string) error { return NewUndefinedFilter(msg) }, "test"},
		{"MethodOverrideError", func(msg string) error { return NewMethodOverrideError(msg) }, "test"},
		{"DisabledError", func(msg string) error { return NewDisabledError(msg) }, "test"},
		{"InternalError", func(msg string) error { return NewInternalError(msg) }, "test"},
		{"TemplateEncodingError", func(msg string) error { return NewTemplateEncodingError(msg) }, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create(tt.wantErr)
			if err == nil {
				t.Errorf("Expected error to be created, got nil")
				return
			}

			msg := err.Error()
			if !strings.Contains(msg, tt.wantErr) {
				t.Errorf("Expected error message to contain '%s', got: %s", tt.wantErr, msg)
			}
		})
	}
}

func TestErrorStringWithoutPrefix(t *testing.T) {
	err := &Error{
		Message: "test error",
	}

	msg := err.String(false)
	if strings.Contains(msg, "Liquid error") {
		t.Errorf("Expected error message without prefix, got: %s", msg)
	}
	if !strings.Contains(msg, "test error") {
		t.Errorf("Expected error message to contain 'test error', got: %s", msg)
	}
}

// TestLiquidErrorInterface tests that all error types implement LiquidError
func TestLiquidErrorInterface(t *testing.T) {
	errorTypes := []LiquidError{
		NewArgumentError("test"),
		NewContextError("test"),
		NewFileSystemError("test"),
		NewStandardError("test"),
		NewSyntaxError("test"),
		NewStackLevelError("test"),
		NewMemoryError("test"),
		NewZeroDivisionError("test"),
		NewFloatDomainError("test"),
		NewUndefinedVariable("test"),
		NewUndefinedDropMethod("test"),
		NewUndefinedFilter("test"),
		NewMethodOverrideError("test"),
		NewDisabledError("test"),
		NewInternalError("test"),
		NewTemplateEncodingError("test"),
	}

	for _, le := range errorTypes {
		t.Run(fmt.Sprintf("%T", le), func(t *testing.T) {
			baseErr := le.GetError()
			if baseErr == nil {
				t.Errorf("%T: GetError() returned nil", le)
			}
			if baseErr.Message != "test" {
				t.Errorf("%T: expected message 'test', got '%s'", le, baseErr.Message)
			}
		})
	}
}

// TestGetErrorReturnsMutableReference tests that GetError returns a mutable reference
func TestGetErrorReturnsMutableReference(t *testing.T) {
	err := NewArgumentError("test")
	baseErr := err.GetError()

	lineNum := 42
	baseErr.LineNumber = &lineNum
	baseErr.TemplateName = "test.liquid"

	// Verify modifications are reflected in original error
	if err.Err.LineNumber == nil || *err.Err.LineNumber != 42 {
		t.Error("LineNumber modification not reflected")
	}
	if err.Err.TemplateName != "test.liquid" {
		t.Error("TemplateName modification not reflected")
	}
}
