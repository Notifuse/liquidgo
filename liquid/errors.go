package liquid

import (
	"fmt"
	"strings"
)

// Error is the base error type for all Liquid errors.
type Error struct {
	Message       string
	LineNumber    *int
	TemplateName  string
	MarkupContext string
}

func (e *Error) Error() string {
	return e.String(true)
}

// String returns the error message with optional prefix.
func (e *Error) String(withPrefix bool) string {
	var b strings.Builder

	if withPrefix {
		b.WriteString(e.messagePrefix())
	}

	b.WriteString(e.Message)

	if e.MarkupContext != "" {
		b.WriteString(" ")
		b.WriteString(e.MarkupContext)
	}

	return b.String()
}

func (e *Error) messagePrefix() string {
	var b strings.Builder
	b.WriteString("Liquid error")

	if e.LineNumber != nil {
		b.WriteString(" (")
		if e.TemplateName != "" {
			b.WriteString(e.TemplateName)
			b.WriteString(" ")
		}
		b.WriteString("line ")
		b.WriteString(fmt.Sprintf("%d", *e.LineNumber))
		b.WriteString(")")
	}

	b.WriteString(": ")
	return b.String()
}

// ArgumentError represents an argument error.
type ArgumentError struct {
	Err *Error
}

// NewArgumentError creates a new ArgumentError with the given message.
func NewArgumentError(message string) *ArgumentError {
	return &ArgumentError{
		Err: &Error{Message: message},
	}
}

func (e *ArgumentError) Error() string {
	return e.Err.Error()
}

// ContextError represents a context error.
type ContextError struct {
	Err *Error
}

// NewContextError creates a new ContextError with the given message.
func NewContextError(message string) *ContextError {
	return &ContextError{
		Err: &Error{Message: message},
	}
}

func (e *ContextError) Error() string {
	return e.Err.Error()
}

// FileSystemError represents a file system error.
type FileSystemError struct {
	Err *Error
}

// NewFileSystemError creates a new FileSystemError with the given message.
func NewFileSystemError(message string) *FileSystemError {
	return &FileSystemError{
		Err: &Error{Message: message},
	}
}

func (e *FileSystemError) Error() string {
	return e.Err.Error()
}

// StandardError represents a standard error.
type StandardError struct {
	Err *Error
}

// NewStandardError creates a new StandardError with the given message.
func NewStandardError(message string) *StandardError {
	return &StandardError{
		Err: &Error{Message: message},
	}
}

func (e *StandardError) Error() string {
	return e.Err.Error()
}

// SyntaxError represents a syntax error.
type SyntaxError struct {
	Err *Error
}

// NewSyntaxError creates a new SyntaxError with the given message.
func NewSyntaxError(message string) *SyntaxError {
	return &SyntaxError{
		Err: &Error{Message: message},
	}
}

// Error implements the error interface for SyntaxError with custom prefix.
func (e *SyntaxError) Error() string {
	var b strings.Builder
	b.WriteString("Liquid syntax error")

	if e.Err.LineNumber != nil {
		b.WriteString(" (")
		if e.Err.TemplateName != "" {
			b.WriteString(e.Err.TemplateName)
			b.WriteString(" ")
		}
		b.WriteString("line ")
		b.WriteString(fmt.Sprintf("%d", *e.Err.LineNumber))
		b.WriteString(")")
	}

	b.WriteString(": ")
	b.WriteString(e.Err.Message)

	if e.Err.MarkupContext != "" {
		b.WriteString(" ")
		b.WriteString(e.Err.MarkupContext)
	}

	return b.String()
}

// StackLevelError represents a stack level error.
type StackLevelError struct {
	Err *Error
}

// NewStackLevelError creates a new StackLevelError with the given message.
func NewStackLevelError(message string) *StackLevelError {
	return &StackLevelError{
		Err: &Error{Message: message},
	}
}

func (e *StackLevelError) Error() string {
	return e.Err.Error()
}

// MemoryError represents a memory error.
type MemoryError struct {
	Err *Error
}

// NewMemoryError creates a new MemoryError with the given message.
func NewMemoryError(message string) *MemoryError {
	return &MemoryError{
		Err: &Error{Message: message},
	}
}

func (e *MemoryError) Error() string {
	return e.Err.Error()
}

// ZeroDivisionError represents a zero division error.
type ZeroDivisionError struct {
	Err *Error
}

// NewZeroDivisionError creates a new ZeroDivisionError with the given message.
func NewZeroDivisionError(message string) *ZeroDivisionError {
	return &ZeroDivisionError{
		Err: &Error{Message: message},
	}
}

func (e *ZeroDivisionError) Error() string {
	return e.Err.Error()
}

// FloatDomainError represents a float domain error.
type FloatDomainError struct {
	Err *Error
}

// NewFloatDomainError creates a new FloatDomainError with the given message.
func NewFloatDomainError(message string) *FloatDomainError {
	return &FloatDomainError{
		Err: &Error{Message: message},
	}
}

func (e *FloatDomainError) Error() string {
	return e.Err.Error()
}

// UndefinedVariable represents an undefined variable error.
type UndefinedVariable struct {
	Err *Error
}

// NewUndefinedVariable creates a new UndefinedVariable with the given message.
func NewUndefinedVariable(message string) *UndefinedVariable {
	return &UndefinedVariable{
		Err: &Error{Message: message},
	}
}

func (e *UndefinedVariable) Error() string {
	return e.Err.Error()
}

// UndefinedDropMethod represents an undefined drop method error.
type UndefinedDropMethod struct {
	Err *Error
}

// NewUndefinedDropMethod creates a new UndefinedDropMethod with the given message.
func NewUndefinedDropMethod(message string) *UndefinedDropMethod {
	return &UndefinedDropMethod{
		Err: &Error{Message: message},
	}
}

func (e *UndefinedDropMethod) Error() string {
	return e.Err.Error()
}

// UndefinedFilter represents an undefined filter error.
type UndefinedFilter struct {
	Err *Error
}

// NewUndefinedFilter creates a new UndefinedFilter with the given message.
func NewUndefinedFilter(message string) *UndefinedFilter {
	return &UndefinedFilter{
		Err: &Error{Message: message},
	}
}

func (e *UndefinedFilter) Error() string {
	return e.Err.Error()
}

// MethodOverrideError represents a method override error.
type MethodOverrideError struct {
	Err *Error
}

// NewMethodOverrideError creates a new MethodOverrideError with the given message.
func NewMethodOverrideError(message string) *MethodOverrideError {
	return &MethodOverrideError{
		Err: &Error{Message: message},
	}
}

func (e *MethodOverrideError) Error() string {
	return e.Err.Error()
}

// DisabledError represents a disabled error.
type DisabledError struct {
	Err *Error
}

// NewDisabledError creates a new DisabledError with the given message.
func NewDisabledError(message string) *DisabledError {
	return &DisabledError{
		Err: &Error{Message: message},
	}
}

func (e *DisabledError) Error() string {
	return e.Err.Error()
}

// InternalError represents an internal error.
type InternalError struct {
	Err *Error
}

// NewInternalError creates a new InternalError with the given message.
func NewInternalError(message string) *InternalError {
	return &InternalError{
		Err: &Error{Message: message},
	}
}

func (e *InternalError) Error() string {
	return e.Err.Error()
}

// TemplateEncodingError represents a template encoding error.
type TemplateEncodingError struct {
	Err *Error
}

// NewTemplateEncodingError creates a new TemplateEncodingError with the given message.
func NewTemplateEncodingError(message string) *TemplateEncodingError {
	return &TemplateEncodingError{
		Err: &Error{Message: message},
	}
}

func (e *TemplateEncodingError) Error() string {
	return e.Err.Error()
}
