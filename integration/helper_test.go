package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// TemplateResultOptions contains options for assertTemplateResult.
type TemplateResultOptions struct {
	Message      string
	Partials     map[string]string
	ErrorMode    liquid.ErrorMode
	RenderErrors bool
}

// assertTemplateResult is the main helper for testing template rendering.
// It creates an environment, parses the template, renders it with the given assigns,
// and asserts the output matches the expected value.
func assertTemplateResult(t *testing.T, expected, template string, assigns map[string]interface{}, opts ...TemplateResultOptions) {
	t.Helper()

	var options TemplateResultOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Create file system for partials
	fileSystem := NewStubFileSystem(options.Partials)

	// Strict mode by default (Ruby uses :strict by default from ENV, but
	// test_helper sets it)
	errorMode := options.ErrorMode
	if errorMode == "" {
		errorMode = "strict"
	}
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: errorMode, FileSystem: fileSystem})

	// Parse template
	templateOptions := &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	}

	tmpl, err := liquid.ParseTemplate(template, templateOptions)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// Create registers
	registersMap := make(map[string]interface{})
	registersMap["file_system"] = fileSystem

	// Build context with proper rethrowErrors setting
	// When render_errors is false, rethrow_errors is true (default)
	rethrowErrors := !options.RenderErrors

	contextConfig := rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{assigns},
		Registers:          liquid.NewRegisters(registersMap),
		RethrowErrors:      rethrowErrors,
	}
	ctx := &contextConfig

	// Render template with context
	output := renderRuby(tmpl, ctx, &rubyOptions{})

	// Assert result
	if output != expected {
		message := options.Message
		if message == "" {
			message = fmt.Sprintf("Expected %q, got %q", expected, output)
		}
		t.Error(message)
	}
}

// assertMatchSyntaxError tests that parsing a template raises a SyntaxError
// and that the error message matches the given pattern.
func assertMatchSyntaxError(t *testing.T, match string, template string, errorMode ...liquid.ErrorMode) {
	t.Helper()

	mode := liquid.ErrorModeStrict
	if len(errorMode) > 0 && errorMode[0] != "" {
		mode = errorMode[0]
	}

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: mode})

	templateOptions := &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	}

	// ParseTemplate may panic in strict mode, so we need to recover
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Check if it's a SyntaxError
				if syntaxErr, ok := r.(*liquid.SyntaxError); ok {
					err = syntaxErr
				} else if syntaxErr, ok := r.(error); ok {
					err = syntaxErr
				} else {
					// Re-panic if it's not an error
					panic(r)
				}
			}
		}()
		_, parseErr := liquid.ParseTemplate(template, templateOptions)
		if parseErr != nil {
			err = parseErr
		}
	}()

	if err == nil {
		t.Fatal("Expected SyntaxError, got nil")
	}

	// Check if it's a SyntaxError
	syntaxErr, ok := err.(*liquid.SyntaxError)
	if !ok {
		t.Fatalf("Expected SyntaxError, got %T: %v", err, err)
	}

	// Match pattern if provided
	if match != "" {
		matched, regexErr := regexp.MatchString(match, syntaxErr.Error())
		if regexErr != nil {
			t.Fatalf("Invalid regex pattern %q: %v", match, regexErr)
		}
		if !matched {
			t.Errorf("Error message %q does not match pattern %q", syntaxErr.Error(), match)
		}
	}
}

// StubFileSystem is a mock file system for testing partials/includes.
type StubFileSystem struct {
	fileReadCount int
	values        map[string]string
}

// NewStubFileSystem creates a new StubFileSystem.
func NewStubFileSystem(values map[string]string) *StubFileSystem {
	if values == nil {
		values = make(map[string]string)
	}
	return &StubFileSystem{
		fileReadCount: 0,
		values:        values,
	}
}

// ReadTemplateFile reads a template file from the stub file system.
func (s *StubFileSystem) ReadTemplateFile(templatePath string) (string, error) {
	s.fileReadCount++
	if value, ok := s.values[templatePath]; ok {
		return value, nil
	}
	return "", liquid.NewFileSystemError(fmt.Sprintf("template not found: %s", templatePath))
}

// FileReadCount returns the number of times ReadTemplateFile was called.
func (s *StubFileSystem) FileReadCount() int {
	return s.fileReadCount
}
