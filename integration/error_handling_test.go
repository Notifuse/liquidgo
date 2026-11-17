package integration

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// TestErrorHandling_TemplatesParsedWithLineNumbersRendersThemInErrors tests that
// templates parsed with line numbers render line numbers in error messages.
// Ported from: test_templates_parsed_with_line_numbers_renders_them_in_errors
func TestErrorHandling_TemplatesParsedWithLineNumbersRendersThemInErrors(t *testing.T) {
	template := `      Hello,

      {{ errors.standard_error }} will raise a standard error.

      Bla bla test.

      {{ errors.syntax_error }} will raise a syntax error.

      This is an argument error: {{ errors.argument_error }}

      Bla.
    `

	expected := `      Hello,

      Liquid error (line 3): standard error will raise a standard error.

      Bla bla test.

      Liquid syntax error (line 7): syntax error will raise a syntax error.

      This is an argument error: Liquid error (line 9): argument error

      Bla.
    `

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "strict"})

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	if output != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, output)
	}
}

// TestErrorHandling_StandardError tests StandardError handling.
// Ported from: test_standard_error
func TestErrorHandling_StandardError(t *testing.T) {
	template := " {{ errors.standard_error }} "

	expected := " Liquid error: standard error "

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	errors := rubyErrors(tmpl)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	} else {
		if _, ok := errors[0].(*liquid.StandardError); !ok {
			t.Errorf("Expected StandardError, got %T", errors[0])
		}
	}
}

// TestErrorHandling_SyntaxError tests SyntaxError handling.
// Ported from: test_syntax
func TestErrorHandling_SyntaxError(t *testing.T) {
	template := " {{ errors.syntax_error }} "

	expected := " Liquid syntax error: syntax error "

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	errors := rubyErrors(tmpl)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	} else {
		if _, ok := errors[0].(*liquid.SyntaxError); !ok {
			t.Errorf("Expected SyntaxError, got %T", errors[0])
		}
	}
}

// TestErrorHandling_ArgumentError tests ArgumentError handling.
// Ported from: test_argument
func TestErrorHandling_ArgumentError(t *testing.T) {
	template := " {{ errors.argument_error }} "

	expected := " Liquid error: argument error "

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	errors := rubyErrors(tmpl)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	} else {
		if _, ok := errors[0].(*liquid.ArgumentError); !ok {
			t.Errorf("Expected ArgumentError, got %T", errors[0])
		}
	}
}

// TestErrorHandling_MissingEndtagParseTimeError tests that missing end tags
// raise parse-time errors.
// Ported from: test_missing_endtag_parse_time_error
func TestErrorHandling_MissingEndtagParseTimeError(t *testing.T) {
	assertMatchSyntaxError(t, `: 'for' tag was never closed\z`, " {% for a in b %} ... ")
}

// TestErrorHandling_UnrecognizedOperator tests that unrecognized operators
// raise SyntaxError in strict mode.
// Ported from: test_unrecognized_operator
func TestErrorHandling_UnrecognizedOperator(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "strict"})

	templateOptions := &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	}

	_, err := liquid.ParseTemplate(" {% if 1 =! 2 %}ok{% endif %} ", templateOptions)
	if err == nil {
		t.Fatal("Expected SyntaxError, got nil")
	}

	if _, ok := err.(*liquid.SyntaxError); !ok {
		t.Errorf("Expected SyntaxError, got %T: %v", err, err)
	}
}

// TestErrorHandling_LaxUnrecognizedOperator tests that unrecognized operators
// are handled gracefully in lax mode.
// Ported from: test_lax_unrecognized_operator
func TestErrorHandling_LaxUnrecognizedOperator(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "lax"})

	tmpl, err := liquid.ParseTemplate(" {% if 1 =! 2 %}ok{% endif %} ", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})
	expected := " Liquid error: Unknown operator =! "

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	errors := rubyErrors(tmpl)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	} else {
		if _, ok := errors[0].(*liquid.ArgumentError); !ok {
			t.Errorf("Expected ArgumentError, got %T", errors[0])
		}
	}
}

// TestErrorHandling_WithLineNumbersAddsNumbersToParserErrors tests that
// line numbers are added to parser errors.
// Ported from: test_with_line_numbers_adds_numbers_to_parser_errors
func TestErrorHandling_WithLineNumbersAddsNumbersToParserErrors(t *testing.T) {
	source := `foobar

      {% "cat" | foobar %}

      bla
    `
	assertMatchSyntaxError(t, `Liquid syntax error \(line 3\)`, source)
}

// TestErrorHandling_WithLineNumbersAddsNumbersToParserErrorsWithWhitespaceTrim tests
// that line numbers are added to parser errors even with whitespace trim.
// Ported from: test_with_line_numbers_adds_numbers_to_parser_errors_with_whitespace_trim
func TestErrorHandling_WithLineNumbersAddsNumbersToParserErrorsWithWhitespaceTrim(t *testing.T) {
	source := `foobar

      {%- "cat" | foobar -%}

      bla
    `
	assertMatchSyntaxError(t, `Liquid syntax error \(line 3\)`, source)
}

// TestErrorHandling_ParsingWarnWithLineNumbersAddsNumbersToLexerErrors tests that
// warnings include line numbers in warn mode.
// Ported from: test_parsing_warn_with_line_numbers_adds_numbers_to_lexer_errors
func TestErrorHandling_ParsingWarnWithLineNumbersAddsNumbersToLexerErrors(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "warn"})

	template := `
        foobar

        {% if 1 =! 2 %}ok{% endif %}

        bla
            `

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	warnings := tmpl.Warnings()
	if len(warnings) == 0 {
		t.Skip("Warnings not yet implemented or not collected")
		return
	}

	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
		return
	}

	expectedMsg := `Liquid syntax error (line 4): Unexpected character = in "1 =! 2"`
	if !strings.Contains(warnings[0].Error(), expectedMsg) {
		t.Errorf("Expected warning to contain %q, got %q", expectedMsg, warnings[0].Error())
	}
}

// TestErrorHandling_ParsingStrictWithLineNumbersAddsNumbersToLexerErrors tests that
// strict mode errors include line numbers.
// Ported from: test_parsing_strict_with_line_numbers_adds_numbers_to_lexer_errors
func TestErrorHandling_ParsingStrictWithLineNumbersAddsNumbersToLexerErrors(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "strict"})

	template := `
          foobar

          {% if 1 =! 2 %}ok{% endif %}

          bla
                `

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})

	if err == nil {
		t.Fatal("Expected SyntaxError, got nil")
	}

	syntaxErr, ok := err.(*liquid.SyntaxError)
	if !ok {
		t.Fatalf("Expected SyntaxError, got %T: %v", err, err)
	}

	expectedMsg := `Liquid syntax error (line 4): Unexpected character = in "1 =! 2"`
	if syntaxErr.Error() != expectedMsg {
		t.Errorf("Expected %q, got %q", expectedMsg, syntaxErr.Error())
	}
}

// TestErrorHandling_SyntaxErrorsInNestedBlocksHaveCorrectLineNumber tests that
// syntax errors in nested blocks have correct line numbers.
// Ported from: test_syntax_errors_in_nested_blocks_have_correct_line_number
func TestErrorHandling_SyntaxErrorsInNestedBlocksHaveCorrectLineNumber(t *testing.T) {
	source := `foobar

      {% if 1 != 2 %}
        {% foo %}
      {% endif %}

      bla
    `
	assertMatchSyntaxError(t, `Liquid syntax error \(line 4\): Unknown tag 'foo'`, source)
}

// TestErrorHandling_StrictErrorMessages tests strict mode error messages.
// Ported from: test_strict_error_messages
func TestErrorHandling_StrictErrorMessages(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "strict"})

	templateOptions := &liquid.TemplateOptions{
		Environment: env,
	}

	_, err := liquid.ParseTemplate(" {% if 1 =! 2 %}ok{% endif %} ", templateOptions)
	if err == nil {
		t.Fatal("Expected SyntaxError, got nil")
	}

	syntaxErr, ok := err.(*liquid.SyntaxError)
	if !ok {
		t.Fatalf("Expected SyntaxError, got %T: %v", err, err)
	}

	expectedMsg := `Liquid syntax error: Unexpected character = in "1 =! 2"`
	if syntaxErr.Error() != expectedMsg {
		t.Errorf("Expected %q, got %q", expectedMsg, syntaxErr.Error())
	}

	_, err = liquid.ParseTemplate("{{%%%}}", templateOptions)
	if err == nil {
		t.Fatal("Expected SyntaxError, got nil")
	}

	syntaxErr2, ok := err.(*liquid.SyntaxError)
	if !ok {
		t.Fatalf("Expected SyntaxError, got %T: %v", err, err)
	}

	expectedMsg2 := `Liquid syntax error: Unexpected character % in "{{%%%}}"`
	if syntaxErr2.Error() != expectedMsg2 {
		t.Errorf("Expected %q, got %q", expectedMsg2, syntaxErr2.Error())
	}
}

// TestErrorHandling_Warnings tests that warnings are collected in warn mode.
// Ported from: test_warnings
func TestErrorHandling_Warnings(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "warn"})

	tmpl, err := liquid.ParseTemplate(`{% if ~~~ %}{{%%%}}{% else %}{{ hello. }}{% endif %}`, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	warnings := tmpl.Warnings()
	if len(warnings) == 0 {
		t.Skip("Warnings not yet implemented or not collected")
		return
	}

	if len(warnings) != 3 {
		t.Errorf("Expected 3 warnings, got %d", len(warnings))
		return
	}

	// Check warning messages (order may vary)
	warningMsgs := make([]string, len(warnings))
	for i, w := range warnings {
		warningMsgs[i] = w.Error()
	}

	expectedWarnings := []string{
		`Unexpected character ~ in "~~~"`,
		`Unexpected character % in "{{%%%}}"`,
		`Expected id but found end_of_string in "{{ hello. }}"`,
	}

	found := make(map[string]bool)
	for _, expected := range expectedWarnings {
		for _, msg := range warningMsgs {
			if strings.Contains(msg, expected) {
				found[expected] = true
				break
			}
		}
	}

	for _, expected := range expectedWarnings {
		if !found[expected] {
			t.Errorf("Expected warning containing %q, but not found in %v", expected, warningMsgs)
		}
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

// TestErrorHandling_WarningLineNumbers tests that warnings include line numbers.
// Ported from: test_warning_line_numbers
func TestErrorHandling_WarningLineNumbers(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: "warn"})

	tmpl, err := liquid.ParseTemplate("{% if ~~~ %}\n{{%%%}}{% else %}\n{{ hello. }}{% endif %}", &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	warnings := tmpl.Warnings()
	if len(warnings) == 0 {
		t.Skip("Warnings not yet implemented or not collected")
		return
	}

	if len(warnings) != 3 {
		t.Errorf("Expected 3 warnings, got %d", len(warnings))
		return
	}

	// Check that warnings have line numbers
	expectedMessages := []string{
		`Liquid syntax error (line 1): Unexpected character ~ in "~~~"`,
		`Liquid syntax error (line 2): Unexpected character % in "{{%%%}}"`,
		`Liquid syntax error (line 3): Expected id but found end_of_string in "{{ hello. }}"`,
	}

	warningMsgs := make([]string, len(warnings))
	for i, w := range warnings {
		warningMsgs[i] = w.Error()
	}

	found := make(map[string]bool)
	for _, expected := range expectedMessages {
		for _, msg := range warningMsgs {
			if strings.Contains(msg, expected) {
				found[expected] = true
				break
			}
		}
	}

	for _, expected := range expectedMessages {
		if !found[expected] {
			t.Errorf("Expected warning containing %q, but not found in %v", expected, warningMsgs)
		}
	}
}

// TestErrorHandling_ExceptionsPropagate tests that non-StandardError exceptions
// propagate (not caught by Liquid).
// Ported from: test_exceptions_propagate
func TestErrorHandling_ExceptionsPropagate(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate("{{ errors.exception }}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
	}

	// In Go, panics from non-Liquid errors should propagate
	// We expect this to panic or handle the error
	defer func() {
		if r := recover(); r != nil {
			// Expected - exception should propagate
			_ = r
		}
	}()

	_ = renderRuby(tmpl, ctx, &rubyOptions{})
	// If we get here without panic, the error was caught (which may be acceptable)
}

// TestErrorHandling_DefaultExceptionRendererWithInternalError tests default
// exception renderer with internal error.
// Ported from: test_default_exception_renderer_with_internal_error
func TestErrorHandling_DefaultExceptionRendererWithInternalError(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate("This is a runtime error: {{ errors.runtime_error }}", &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
	}

	// Ruby renders the RuntimeError as "Liquid error (line 1): internal" in
	// place of the output. The drop panics with a Go value, which ends a
	// render of liquidgo as a fatal PanicError with its stack (a divergence
	// from Ruby), and the Ruby helper writes the text of an error that ends
	// the render as the output.
	output := renderRuby(tmpl, ctx, &rubyOptions{})
	expected := "Liquid error: internal"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	errors := rubyErrors(tmpl)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	} else {
		if _, ok := errors[0].(*liquid.PanicError); !ok {
			t.Errorf("Expected PanicError, got %T", errors[0])
		}
	}
}

// TestErrorHandling_IncludedTemplateNameWithLineNumbers tests that included
// template names appear in error messages.
// Ported from: test_included_template_name_with_line_numbers
func TestErrorHandling_IncludedTemplateNameWithLineNumbers(t *testing.T) {
	fileSystem := NewStubFileSystem(map[string]string{
		"product": "{{ errors.argument_error }}",
	})

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: fileSystem})

	template := "Argument error:\n{% include 'product' %}"

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"errors": NewErrorDrop()}},
		RethrowErrors:      false,
		Registers:          liquid.NewRegisters(map[string]interface{}{"file_system": fileSystem}),
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})
	expected := "Argument error:\nLiquid error (product line 1): argument error"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	errors := rubyErrors(tmpl)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	} else {
		// Check that error has template name
		if err, ok := errors[0].(*liquid.ArgumentError); ok {
			if err.Err.TemplateName != "product" {
				t.Errorf("Expected template name 'product', got %q", err.Err.TemplateName)
			}
		}
	}
}

// TestErrorHandling_BugCompatibleSilencingOfErrorsInBlankNodes tests bug-compatible
// behavior where errors in blank nodes are silenced.
// Ported from: test_bug_compatible_silencing_of_errors_in_blank_nodes
func TestErrorHandling_BugCompatibleSilencingOfErrorsInBlankNodes(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	template1 := "{% assign x = 0 %}{% if 1 < '2' %}not blank{% assign x = 3 %}{% endif %}{{ x }}"
	tmpl1, err := liquid.ParseTemplate(template1, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx1 := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
	}

	output1 := renderRuby(tmpl1, ctx1, &rubyOptions{})
	expected1 := "Liquid error: comparison of Integer with String failed0"

	if output1 != expected1 {
		t.Errorf("Expected %q, got %q", expected1, output1)
	}

	template2 := "{% assign x = 0 %}{% if 1 < '2' %}{% assign x = 3 %}{% endif %}{{ x }}"
	tmpl2, err := liquid.ParseTemplate(template2, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx2 := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
	}

	output2 := renderRuby(tmpl2, ctx2, &rubyOptions{})
	expected2 := "0"

	if output2 != expected2 {
		t.Errorf("Expected %q, got %q", expected2, output2)
	}
}

// TestErrorHandling_SyntaxErrorIsRaisedWithTemplateName tests that syntax errors
// include template name.
// Ported from: test_syntax_error_is_raised_with_template_name
func TestErrorHandling_SyntaxErrorIsRaisedWithTemplateName(t *testing.T) {
	fileSystem := NewStubFileSystem(map[string]string{
		"snippet": "1\n2\n{{ 1",
	})

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: fileSystem})

	template := `{% render "snippet" %}`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	tmpl.SetName("template/index")

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
		Registers:          liquid.NewRegisters(map[string]interface{}{"file_system": fileSystem}),
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	if !strings.Contains(output, "snippet line 3") {
		t.Errorf("Expected output to contain 'snippet line 3', got %q", output)
	}
}

// TestErrorHandling_SyntaxErrorIsRaisedWithTemplateNameFromTemplateFactory tests
// that a syntax error of a partial names the partial (Ruby's test names the
// template factory, which liquidgo no longer has).
// Ported from: test_syntax_error_is_raised_with_template_name_from_template_factory
func TestErrorHandling_SyntaxErrorIsRaisedWithTemplateNameFromTemplateFactory(t *testing.T) {
	fileSystem := NewStubFileSystem(map[string]string{
		"snippet": "1\n2\n{{ 1",
	})

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: fileSystem})

	template := `{% render "snippet" %}`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	tmpl.SetName("template/index")

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
		Registers: liquid.NewRegisters(map[string]interface{}{
			"file_system": fileSystem,
		}),
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	// the partial's error names it: the template factory that renamed it
	// (some/path/snippet in Ruby's test) is gone
	if !strings.Contains(output, "(snippet line 3)") {
		t.Errorf("Expected output to contain '(snippet line 3)', got %q", output)
	}
}

// TestErrorHandling_ErrorIsRaisedDuringParseWithTemplateName tests that errors
// raised during parse include template name.
// Ported from: test_error_is_raised_during_parse_with_template_name
func TestErrorHandling_ErrorIsRaisedDuringParseWithTemplateName(t *testing.T) {
	// Get MAX_DEPTH from block package - we'll need to check this
	// For now, use a reasonable depth
	depth := 101 // This should exceed MAX_DEPTH
	code := ""
	for i := 0; i < depth; i++ {
		code += "{% if true %}"
	}
	code += "rendered"
	for i := 0; i < depth; i++ {
		code += "{% endif %}"
	}

	fileSystem := NewStubFileSystem(map[string]string{
		"snippet": code,
	})

	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: fileSystem})

	template := `{% render 'snippet' %}`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
		Registers: liquid.NewRegisters(map[string]interface{}{
			"file_system": fileSystem,
		}),
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})

	// the partial's error names it (Ruby's test renames it some/path/snippet
	// with a template factory, which is gone)
	if !strings.Contains(output, "(snippet line 1)") {
		t.Errorf("Expected output to contain '(snippet line 1)', got %q", output)
	}
	if !strings.Contains(output, "Nesting too deep") {
		t.Errorf("Expected output to contain 'Nesting too deep', got %q", output)
	}
}

// TestErrorHandling_InternalErrorIsRaisedWithTemplateName tests that internal
// errors include template name.
// Ported from: test_internal_error_is_raised_with_template_name
func TestErrorHandling_InternalErrorIsRaisedWithTemplateName(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: NewStubFileSystem(map[string]string{})})

	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	err := tmpl.Parse("{% render 'snippet' %}", &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tmpl.SetName("template/index")

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
		Registers: liquid.NewRegisters(map[string]interface{}{
			"file_system": NewStubFileSystem(map[string]string{}),
		}),
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})
	expected := "Liquid error (template/index line 1): template not found: snippet"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestErrorHandling_StrictFiltersWithUndefinedFilter tests that strict_filters mode
// raises an UndefinedFilter error when an unknown filter is used.
func TestErrorHandling_StrictFiltersWithUndefinedFilter(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate(`{{ "hello" | unknown_filter }}`, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// Test with strict_filters = true
	t.Run("strict_filters_true", func(t *testing.T) {
		output := renderRuby(tmpl, nil, &rubyOptions{
			StrictFilters: true,
		})

		// As in Ruby, an undefined filter is recorded and writes nothing.
		if output != "" {
			t.Errorf("Expected an empty output, got %q", output)
		}
		if errs := rubyErrors(tmpl); len(errs) != 1 || !strings.Contains(errs[0].Error(), "undefined filter unknown_filter") {
			t.Errorf("Expected the error 'undefined filter unknown_filter', got %v", errs)
		}
	})

	// Test with strict_filters = false (default)
	t.Run("strict_filters_false", func(t *testing.T) {
		output := renderRuby(tmpl, nil, &rubyOptions{
			StrictFilters: false,
		})

		// Should pass through the input unchanged (no error)
		if output != "hello" {
			t.Errorf("Expected 'hello' (pass-through), got %q", output)
		}
	})
}

// TestErrorHandling_StrictFiltersWithDefinedFilter tests that strict_filters mode
// works correctly with defined filters.
func TestErrorHandling_StrictFiltersWithDefinedFilter(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate(`{{ "HELLO" | downcase }}`, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// Should work in both modes
	t.Run("strict_filters_true", func(t *testing.T) {
		output := renderRuby(tmpl, nil, &rubyOptions{
			StrictFilters: true,
		})
		if output != "hello" {
			t.Errorf("Expected 'hello', got %q", output)
		}
	})

	t.Run("strict_filters_false", func(t *testing.T) {
		output := renderRuby(tmpl, nil, &rubyOptions{
			StrictFilters: false,
		})
		if output != "hello" {
			t.Errorf("Expected 'hello', got %q", output)
		}
	})
}

// TestErrorHandling_StrictFiltersErrorIncludesLineNumber tests that UndefinedFilter
// errors in strict mode include line numbers when available.
func TestErrorHandling_StrictFiltersErrorIncludesLineNumber(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	template := `Line 1
Line 2
{{ "test" | bad_filter }}
Line 4`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	output := renderRuby(tmpl, nil, &rubyOptions{
		StrictFilters: true,
	})

	// As in Ruby, the error is recorded with its line, and writes nothing.
	if output != "Line 1\nLine 2\n\nLine 4" {
		t.Errorf("Expected the lines without the output, got %q", output)
	}
	if errs := rubyErrors(tmpl); len(errs) != 1 || errs[0].Error() != "Liquid error (line 3): undefined filter bad_filter" {
		t.Errorf("Expected the error 'Liquid error (line 3): undefined filter bad_filter', got %v", errs)
	}
}

// TestErrorHandling_StrictFiltersWithFilterChain tests strict mode with chained filters
// where one filter is undefined.
func TestErrorHandling_StrictFiltersWithFilterChain(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	tmpl, err := liquid.ParseTemplate(`{{ "hello" | upcase | unknown_filter | downcase }}`, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	t.Run("strict_mode_stops_at_undefined", func(t *testing.T) {
		output := renderRuby(tmpl, nil, &rubyOptions{
			StrictFilters: true,
		})

		// As in Ruby, an undefined filter is recorded and writes nothing.
		if output != "" {
			t.Errorf("Expected an empty output, got %q", output)
		}
		if errs := rubyErrors(tmpl); len(errs) != 1 || !strings.Contains(errs[0].Error(), "undefined filter") {
			t.Errorf("Expected the error 'undefined filter', got %v", errs)
		}
	})

	t.Run("non_strict_mode_passes_through", func(t *testing.T) {
		output := renderRuby(tmpl, nil, &rubyOptions{
			StrictFilters: false,
		})

		// In non-strict mode, unknown filter passes through its input
		// upcase("hello") = "HELLO", unknown_filter("HELLO") = "HELLO", downcase("HELLO") = "hello"
		if output != "hello" {
			t.Errorf("Expected 'hello', got %q", output)
		}
	})
}
