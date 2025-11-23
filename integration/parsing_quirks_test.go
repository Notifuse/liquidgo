package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestParsingQuirks_ParsingCSS tests that CSS without Liquid syntax parses correctly.
// Ported from: test_parsing_css
func TestParsingQuirks_ParsingCSS(t *testing.T) {
	text := " div { font-weight: bold; } "
	assertTemplateResult(t, text, text, map[string]interface{}{})
}

// TestParsingQuirks_RaiseOnSingleCloseBracket tests that single close bracket raises error.
// Ported from: test_raise_on_single_close_bracet
func TestParsingQuirks_RaiseOnSingleCloseBracket(t *testing.T) {
	assertMatchSyntaxError(t, "", "text {{method} oh nos!")
}

// TestParsingQuirks_RaiseOnLabelAndNoCloseBrackets tests that unclosed variable raises error.
// Ported from: test_raise_on_label_and_no_close_bracets
func TestParsingQuirks_RaiseOnLabelAndNoCloseBrackets(t *testing.T) {
	assertMatchSyntaxError(t, "", "TEST {{ ")
}

// TestParsingQuirks_RaiseOnLabelAndNoCloseBracketsPercent tests that unclosed tag raises error.
// Ported from: test_raise_on_label_and_no_close_bracets_percent
func TestParsingQuirks_RaiseOnLabelAndNoCloseBracketsPercent(t *testing.T) {
	assertMatchSyntaxError(t, "", "TEST {% ")
}

// TestParsingQuirks_ErrorOnEmptyFilter tests empty filter handling.
// Ported from: test_error_on_empty_filter
func TestParsingQuirks_ErrorOnEmptyFilter(t *testing.T) {
	// Should parse fine
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate("{{test}}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// In lax mode, empty filter should parse
	env2 := liquid.NewEnvironment()
	env2.SetErrorMode("lax")
	tags.RegisterStandardTags(env2)

	_, err = liquid.ParseTemplate("{{|test}}", &liquid.TemplateOptions{
		Environment: env2,
	})
	if err != nil {
		t.Errorf("Expected no error in lax mode, got %v", err)
	}

	// In strict mode, empty filter should raise error
	env3 := liquid.NewEnvironment()
	env3.SetErrorMode("strict")
	tags.RegisterStandardTags(env3)

	_, err = liquid.ParseTemplate("{{|test}}", &liquid.TemplateOptions{
		Environment: env3,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	_, err = liquid.ParseTemplate("{{test |a|b|}}", &liquid.TemplateOptions{
		Environment: env3,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}
}

// TestParsingQuirks_MeaninglessParensError tests meaningless parentheses error.
// Ported from: test_meaningless_parens_error
func TestParsingQuirks_MeaninglessParensError(t *testing.T) {
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	markup := "a == 'foo' or (b == 'bar' and c == 'baz') or false"
	template := "{% if " + markup + " %} YES {% endif %}"

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}
}

// TestParsingQuirks_UnexpectedCharactersSyntaxError tests unexpected characters error.
// Ported from: test_unexpected_characters_syntax_error
func TestParsingQuirks_UnexpectedCharactersSyntaxError(t *testing.T) {
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	markup1 := "true && false"
	template1 := "{% if " + markup1 + " %} YES {% endif %}"

	_, err := liquid.ParseTemplate(template1, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	markup2 := "false || true"
	template2 := "{% if " + markup2 + " %} YES {% endif %}"

	_, err = liquid.ParseTemplate(template2, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}
}

// TestParsingQuirks_NoErrorOnLaxEmptyFilter tests that lax mode allows empty filters.
// Ported from: test_no_error_on_lax_empty_filter
func TestParsingQuirks_NoErrorOnLaxEmptyFilter(t *testing.T) {
	env := liquid.NewEnvironment()
	env.SetErrorMode("lax")
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate("{{test |a|b|}}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Errorf("Expected no error in lax mode, got %v", err)
	}

	_, err = liquid.ParseTemplate("{{test}}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = liquid.ParseTemplate("{{|test|}}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Errorf("Expected no error in lax mode, got %v", err)
	}
}

// TestParsingQuirks_MeaninglessParensLax tests meaningless parentheses in lax mode.
// Ported from: test_meaningless_parens_lax
func TestParsingQuirks_MeaninglessParensLax(t *testing.T) {
	assigns := map[string]interface{}{"b": "bar", "c": "baz"}
	markup := "a == 'foo' or (b == 'bar' and c == 'baz') or false"
	assertTemplateResult(t, " YES ", "{% if "+markup+" %} YES {% endif %}", assigns, TemplateResultOptions{ErrorMode: "lax"})
}

// TestParsingQuirks_UnexpectedCharactersSilentlyEatLogicLax tests unexpected characters in lax mode.
// Ported from: test_unexpected_characters_silently_eat_logic_lax
func TestParsingQuirks_UnexpectedCharactersSilentlyEatLogicLax(t *testing.T) {
	markup1 := "true && false"
	assertTemplateResult(t, " YES ", "{% if "+markup1+" %} YES {% endif %}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})

	markup2 := "false || true"
	assertTemplateResult(t, "", "{% if "+markup2+" %} YES {% endif %}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
}

// TestParsingQuirks_RaiseOnInvalidTagDelimiter tests invalid tag delimiter error.
// Ported from: test_raise_on_invalid_tag_delimiter
func TestParsingQuirks_RaiseOnInvalidTagDelimiter(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	tmpl := liquid.NewTemplate(&liquid.TemplateOptions{Environment: env})
	err := tmpl.Parse("{% end %}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError")
	}
}

// TestParsingQuirks_UnanchoredFilterArguments tests unanchored filter arguments.
// Ported from: test_unanchored_filter_arguments
func TestParsingQuirks_UnanchoredFilterArguments(t *testing.T) {
	assertTemplateResult(t, "hi", "{{ 'hi there' | split$$$:' ' | first }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "x", "{{ 'X' | downcase) }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})

	// After the messed up quotes a filter without parameters (reverse) should work
	// but one with parameters (remove) shouldn't be detected.
	assertTemplateResult(t, "here", "{{ 'hi there' | split:\"t\"\" | reverse | first}}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "hi ", "{{ 'hi there' | split:\"t\"\" | remove:\"i\" | first}}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
}

// TestParsingQuirks_InvalidVariablesWork tests invalid variables in lax mode.
// Ported from: test_invalid_variables_work
func TestParsingQuirks_InvalidVariablesWork(t *testing.T) {
	assertTemplateResult(t, "bar", "{% assign 123foo = 'bar' %}{{ 123foo }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "123", "{% assign 123 = 'bar' %}{{ 123 }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
}

// TestParsingQuirks_ExtraDotsInRanges tests extra dots in ranges.
// Ported from: test_extra_dots_in_ranges
func TestParsingQuirks_ExtraDotsInRanges(t *testing.T) {
	assertTemplateResult(t, "12345", "{% for i in (1...5) %}{{ i }}{% endfor %}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
}

// TestParsingQuirks_BlankVariableMarkup tests blank variable markup.
// Ported from: test_blank_variable_markup
func TestParsingQuirks_BlankVariableMarkup(t *testing.T) {
	assertTemplateResult(t, "", "{{}}", map[string]interface{}{})
}

// TestParsingQuirks_LookupOnVarWithLiteralName tests lookup on variable with literal name.
// Ported from: test_lookup_on_var_with_literal_name
func TestParsingQuirks_LookupOnVarWithLiteralName(t *testing.T) {
	assigns := map[string]interface{}{"blank": map[string]interface{}{"x": "result"}}
	assertTemplateResult(t, "result", "{{ blank.x }}", assigns)
	assertTemplateResult(t, "result", "{{ blank['x'] }}", assigns)
}

// TestParsingQuirks_ContainsInId tests contains in identifier.
// Ported from: test_contains_in_id
func TestParsingQuirks_ContainsInId(t *testing.T) {
	assertTemplateResult(t, " YES ", "{% if containsallshipments == true %} YES {% endif %}", map[string]interface{}{"containsallshipments": true})
}

// TestParsingQuirks_IncompleteExpression tests incomplete expressions.
// Ported from: test_incomplete_expression
func TestParsingQuirks_IncompleteExpression(t *testing.T) {
	assertTemplateResult(t, "false", "{{ false - }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "false", "{{ false > }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "false", "{{ false < }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "false", "{{ false = }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "false", "{{ false ! }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "false", "{{ false 1 }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
	assertTemplateResult(t, "false", "{{ false a }}", map[string]interface{}{}, TemplateResultOptions{ErrorMode: "lax"})
}
