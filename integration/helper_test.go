package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TemplateResultOptions contains options for assertTemplateResult.
type TemplateResultOptions struct {
	Message       string
	Partials      map[string]string
	ErrorMode     string
	RenderErrors  bool
	TemplateFactory interface{}
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
	
	// Create environment
	env := liquid.NewEnvironment()
	env.SetFileSystem(fileSystem)
	
	// Set error mode
	if options.ErrorMode != "" {
		env.SetErrorMode(options.ErrorMode)
	} else {
		// Default to strict mode (Ruby uses :strict by default from ENV, but test_helper sets it)
		env.SetErrorMode("strict")
	}
	
	// Register standard tags - need to import tags package
	// Note: This will be done by importing tags.RegisterStandardTags
	tags.RegisterStandardTags(env)
	
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
	if options.TemplateFactory != nil {
		registersMap["template_factory"] = options.TemplateFactory
	}
	
	// Build context with proper rethrowErrors setting
	// When render_errors is false, rethrow_errors is true (default)
	rethrowErrors := !options.RenderErrors
	
	contextConfig := liquid.ContextConfig{
		Environment:       env,
		StaticEnvironments: []map[string]interface{}{assigns},
		Registers:         liquid.NewRegisters(registersMap),
		RethrowErrors:      rethrowErrors,
	}
	ctx := liquid.BuildContext(contextConfig)
	
	// Render template with context
	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	
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
func assertMatchSyntaxError(t *testing.T, match string, template string, errorMode ...string) {
	t.Helper()
	
	mode := "strict"
	if len(errorMode) > 0 && errorMode[0] != "" {
		mode = errorMode[0]
	}
	
	env := liquid.NewEnvironment()
	env.SetErrorMode(mode)
	tags.RegisterStandardTags(env)
	
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

// _assertSyntaxError is a simplified version that just checks for a SyntaxError.
// Prefixed with _ to indicate it's intentionally unused but kept for future use.
//
//nolint:unused
func _assertSyntaxError(t *testing.T, template string, errorMode ...string) {
	t.Helper()
	assertMatchSyntaxError(t, "", template, errorMode...)
}

// _withCustomTag temporarily registers a custom tag, runs the test function, then restores.
// Prefixed with _ to indicate it's intentionally unused but kept for future use.
//
//nolint:unused
func _withCustomTag(t *testing.T, tagName string, tagClass interface{}, fn func()) {
	t.Helper()
	
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	
	// Register custom tag
	env.RegisterTag(tagName, tagClass)
	
	// Store original environment
	originalEnv := liquid.NewEnvironment()
	tags.RegisterStandardTags(originalEnv)
	
	// Run test with custom tag
	fn()
	
	// Restore is implicit - we're using a new environment for each test
	_ = originalEnv
}

// _withErrorModes runs a test function with different error modes.
// Prefixed with _ to indicate it's intentionally unused but kept for future use.
//
//nolint:unused
func _withErrorModes(t *testing.T, modes []string, fn func()) {
	t.Helper()
	
	originalMode := "strict" // Default
	
	for _, mode := range modes {
		env := liquid.NewEnvironment()
		env.SetErrorMode(mode)
		tags.RegisterStandardTags(env)
		
		// Run test
		fn()
	}
	
	_ = originalMode // Restore not needed as we create new env each time
}

// _withGlobalFilter temporarily adds global filters to the environment.
// Prefixed with _ to indicate it's intentionally unused but kept for future use.
//
//nolint:unused
func _withGlobalFilter(t *testing.T, filters []interface{}, fn func()) {
	t.Helper()
	
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	
	// Register filters
	for _, filter := range filters {
		if err := env.RegisterFilter(filter); err != nil {
			t.Fatalf("RegisterFilter() error = %v", err)
		}
	}
	
	// Run test
	fn()
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
	return "", fmt.Errorf("template not found: %s", templatePath)
}

// FileReadCount returns the number of times ReadTemplateFile was called.
func (s *StubFileSystem) FileReadCount() int {
	return s.fileReadCount
}

// StubTemplateFactory is a mock template factory for testing.
type StubTemplateFactory struct {
	count int
}

// NewStubTemplateFactory creates a new StubTemplateFactory.
func NewStubTemplateFactory() *StubTemplateFactory {
	return &StubTemplateFactory{
		count: 0,
	}
}

// For returns a template instance for the given template name.
func (s *StubTemplateFactory) For(templateName string) interface{} {
	s.count++
	template := liquid.NewTemplate(&liquid.TemplateOptions{})
	template.SetName("some/path/" + templateName)
	return template
}

// Count returns the number of times For was called.
func (s *StubTemplateFactory) Count() int {
	return s.count
}

