package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestVariable_SimpleVariable tests simple variable rendering.
// Ported from: test_simple_variable
func TestVariable_SimpleVariable(t *testing.T) {
	assertTemplateResult(t, "worked", "{{test}}", map[string]interface{}{"test": "worked"})
	assertTemplateResult(t, "worked wonderfully", "{{test}}", map[string]interface{}{"test": "worked wonderfully"})
}

// TestVariable_RenderCallsToLiquid tests that variables call ToLiquid.
// Ported from: test_variable_render_calls_to_liquid
func TestVariable_RenderCallsToLiquid(t *testing.T) {
	assertTemplateResult(t, "foobar", "{{ foo }}", map[string]interface{}{"foo": &ThingWithToLiquid{}})
}

// TestVariable_LookupCallsToLiquidValue tests that variable lookup calls ToLiquidValue.
// Ported from: test_variable_lookup_calls_to_liquid_value
func TestVariable_LookupCallsToLiquidValue(t *testing.T) {
	assertTemplateResult(t, "1", "{{ foo }}", map[string]interface{}{"foo": NewIntegerDrop("1")})
	assertTemplateResult(t, "2", "{{ list[foo] }}", map[string]interface{}{
		"foo":  NewIntegerDrop("1"),
		"list": []interface{}{1, 2, 3},
	})
	assertTemplateResult(t, "one", "{{ list[foo] }}", map[string]interface{}{
		"foo":  NewIntegerDrop("1"),
		"list": map[int]interface{}{1: "one"},
	})
	assertTemplateResult(t, "Yay", "{{ foo }}", map[string]interface{}{"foo": NewBooleanDrop(true)})
	assertTemplateResult(t, "YAY", "{{ foo | upcase }}", map[string]interface{}{"foo": NewBooleanDrop(true)})
}

// TestVariable_IfTagCallsToLiquidValue tests that if tag calls ToLiquidValue.
// Ported from: test_if_tag_calls_to_liquid_value
func TestVariable_IfTagCallsToLiquidValue(t *testing.T) {
	assertTemplateResult(t, "one", "{% if foo == 1 %}one{% endif %}", map[string]interface{}{"foo": NewIntegerDrop("1")})
	assertTemplateResult(t, "one", "{% if foo == eqv %}one{% endif %}", map[string]interface{}{
		"foo": NewIntegerDrop(1),
		"eqv": NewIntegerDrop(1),
	})
	assertTemplateResult(t, "one", "{% if 0 < foo %}one{% endif %}", map[string]interface{}{"foo": NewIntegerDrop("1")})
	assertTemplateResult(t, "one", "{% if foo > 0 %}one{% endif %}", map[string]interface{}{"foo": NewIntegerDrop("1")})
	assertTemplateResult(t, "one", "{% if b > a %}one{% endif %}", map[string]interface{}{
		"b": NewIntegerDrop(1),
		"a": NewIntegerDrop(0),
	})
	assertTemplateResult(t, "true", "{% if foo == true %}true{% endif %}", map[string]interface{}{"foo": NewBooleanDrop(true)})
	assertTemplateResult(t, "true", "{% if foo %}true{% endif %}", map[string]interface{}{"foo": NewBooleanDrop(true)})

	assertTemplateResult(t, "", "{% if foo %}true{% endif %}", map[string]interface{}{"foo": NewBooleanDrop(false)})
	assertTemplateResult(t, "", "{% if foo == true %}True{% endif %}", map[string]interface{}{"foo": NewBooleanDrop(false)})
	assertTemplateResult(t, "", "{% if foo and true %}SHOULD NOT HAPPEN{% endif %}", map[string]interface{}{"foo": NewBooleanDrop(false)})

	assertTemplateResult(t, "one", "{% if a contains x %}one{% endif %}", map[string]interface{}{
		"a": []interface{}{1},
		"x": NewIntegerDrop(1),
	})
}

// TestVariable_UnlessTagCallsToLiquidValue tests that unless tag calls ToLiquidValue.
// Ported from: test_unless_tag_calls_to_liquid_value
func TestVariable_UnlessTagCallsToLiquidValue(t *testing.T) {
	assertTemplateResult(t, "", "{% unless foo %}true{% endunless %}", map[string]interface{}{"foo": NewBooleanDrop(true)})
	assertTemplateResult(t, "true", "{% unless foo %}true{% endunless %}", map[string]interface{}{"foo": NewBooleanDrop(false)})
}

// TestVariable_CaseTagCallsToLiquidValue tests that case tag calls ToLiquidValue.
// Ported from: test_case_tag_calls_to_liquid_value
func TestVariable_CaseTagCallsToLiquidValue(t *testing.T) {
	assertTemplateResult(t, "One", "{% case foo %}{% when 1 %}One{% endcase %}", map[string]interface{}{"foo": NewIntegerDrop("1")})
}

// TestVariable_SimpleWithWhitespaces tests variables with whitespace.
// Ported from: test_simple_with_whitespaces
func TestVariable_SimpleWithWhitespaces(t *testing.T) {
	assertTemplateResult(t, "  worked  ", "  {{ test }}  ", map[string]interface{}{"test": "worked"})
	assertTemplateResult(t, "  worked wonderfully  ", "  {{ test }}  ", map[string]interface{}{"test": "worked wonderfully"})
}

// TestVariable_ExpressionWithWhitespaceInSquareBrackets tests expressions with whitespace in brackets.
// Ported from: test_expression_with_whitespace_in_square_brackets
func TestVariable_ExpressionWithWhitespaceInSquareBrackets(t *testing.T) {
	assertTemplateResult(t, "result", "{{ a[ 'b' ] }}", map[string]interface{}{"a": map[string]interface{}{"b": "result"}})
	assertTemplateResult(t, "result", "{{ a[ [ 'b' ] ] }}", map[string]interface{}{
		"b": "c",
		"a": map[string]interface{}{"c": "result"},
	})
}

// TestVariable_IgnoreUnknown tests that unknown variables are ignored.
// Ported from: test_ignore_unknown
func TestVariable_IgnoreUnknown(t *testing.T) {
	assertTemplateResult(t, "", "{{ test }}", map[string]interface{}{})
}

// TestVariable_UsingBlankAsVariableName tests using blank as a variable name.
// Ported from: test_using_blank_as_variable_name
func TestVariable_UsingBlankAsVariableName(t *testing.T) {
	assertTemplateResult(t, "", "{% assign foo = blank %}{{ foo }}", map[string]interface{}{})
}

// TestVariable_UsingEmptyAsVariableName tests using empty as a variable name.
// Ported from: test_using_empty_as_variable_name
func TestVariable_UsingEmptyAsVariableName(t *testing.T) {
	assertTemplateResult(t, "", "{% assign foo = empty %}{{ foo }}", map[string]interface{}{})
}

// TestVariable_HashScoping tests hash scoping.
// Ported from: test_hash_scoping
func TestVariable_HashScoping(t *testing.T) {
	assertTemplateResult(t, "worked", "{{ test.test }}", map[string]interface{}{"test": map[string]interface{}{"test": "worked"}})
	assertTemplateResult(t, "worked", "{{ test . test }}", map[string]interface{}{"test": map[string]interface{}{"test": "worked"}})
}

// TestVariable_FalseRendersAsFalse tests that false renders as "false".
// Ported from: test_false_renders_as_false
func TestVariable_FalseRendersAsFalse(t *testing.T) {
	assertTemplateResult(t, "false", "{{ foo }}", map[string]interface{}{"foo": false})
	assertTemplateResult(t, "false", "{{ false }}", map[string]interface{}{})
}

// TestVariable_NilRendersAsEmptyString tests that nil renders as empty string.
// Ported from: test_nil_renders_as_empty_string
func TestVariable_NilRendersAsEmptyString(t *testing.T) {
	assertTemplateResult(t, "", "{{ nil }}", map[string]interface{}{})
	assertTemplateResult(t, "cat", "{{ nil | append: 'cat' }}", map[string]interface{}{})
}

// TestVariable_PresetAssigns tests preset assigns on template.
// Ported from: test_preset_assigns
func TestVariable_PresetAssigns(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	tmpl, err := liquid.ParseTemplate("{{ test }}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// In Go, we don't have template.assigns like Ruby
	// Instead, we pass assigns to Render
	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"test": "worked"}},
		RethrowErrors:      false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != "worked" {
		t.Errorf("Expected 'worked', got %q", output)
	}
}

// TestVariable_ReuseParsedTemplate tests reusing a parsed template.
// Ported from: test_reuse_parsed_template
func TestVariable_ReuseParsedTemplate(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	tmpl, err := liquid.ParseTemplate("{{ greeting }} {{ name }}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// First render
	ctx1 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"greeting": "Hello", "name": "Tobi"}},
		RethrowErrors:      false,
	})
	output1 := tmpl.Render(ctx1, &liquid.RenderOptions{})
	if output1 != "Hello Tobi" {
		t.Errorf("Expected 'Hello Tobi', got %q", output1)
	}

	// Second render with missing variable
	ctx2 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"greeting": "Hello", "unknown": "Tobi"}},
		RethrowErrors:      false,
	})
	output2 := tmpl.Render(ctx2, &liquid.RenderOptions{})
	if output2 != "Hello " {
		t.Errorf("Expected 'Hello ', got %q", output2)
	}

	// Third render
	ctx3 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"greeting": "Hello", "name": "Brian"}},
		RethrowErrors:      false,
	})
	output3 := tmpl.Render(ctx3, &liquid.RenderOptions{})
	if output3 != "Hello Brian" {
		t.Errorf("Expected 'Hello Brian', got %q", output3)
	}
}

// TestVariable_AssignsNotPollutedFromTemplate tests that assigns are not polluted from template.
// Ported from: test_assigns_not_polluted_from_template
func TestVariable_AssignsNotPollutedFromTemplate(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	tmpl, err := liquid.ParseTemplate("{{ test }}{% assign test = 'bar' %}{{ test }}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// First render with preset assign
	ctx1 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"test": "baz"}},
		RethrowErrors:      false,
	})
	output1 := tmpl.Render(ctx1, &liquid.RenderOptions{})
	if output1 != "bazbar" {
		t.Errorf("Expected 'bazbar', got %q", output1)
	}

	// Second render - should be same
	ctx2 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"test": "baz"}},
		RethrowErrors:      false,
	})
	output2 := tmpl.Render(ctx2, &liquid.RenderOptions{})
	if output2 != "bazbar" {
		t.Errorf("Expected 'bazbar', got %q", output2)
	}

	// Third render with different assign
	ctx3 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"test": "foo"}},
		RethrowErrors:      false,
	})
	output3 := tmpl.Render(ctx3, &liquid.RenderOptions{})
	if output3 != "foobar" {
		t.Errorf("Expected 'foobar', got %q", output3)
	}

	// Fourth render - should revert to preset
	ctx4 := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"test": "baz"}},
		RethrowErrors:      false,
	})
	output4 := tmpl.Render(ctx4, &liquid.RenderOptions{})
	if output4 != "bazbar" {
		t.Errorf("Expected 'bazbar', got %q", output4)
	}
}

// TestVariable_MultilineVariable tests multiline variables.
// Ported from: test_multiline_variable
func TestVariable_MultilineVariable(t *testing.T) {
	assertTemplateResult(t, "worked", "{{\ntest\n}}", map[string]interface{}{"test": "worked"})
}

// TestVariable_RenderSymbol tests rendering symbols (not applicable in Go, but test for compatibility).
// Ported from: test_render_symbol
func TestVariable_RenderSymbol(t *testing.T) {
	// Go doesn't have symbols like Ruby, but we can test string rendering
	assertTemplateResult(t, "bar", "{{ foo }}", map[string]interface{}{"foo": "bar"})
}

// TestVariable_NestedArray tests nested arrays.
// Ported from: test_nested_array
func TestVariable_NestedArray(t *testing.T) {
	assertTemplateResult(t, "", "{{ foo }}", map[string]interface{}{"foo": [][]interface{}{{nil}}})
}

// TestVariable_DynamicFindVar tests dynamic variable lookup.
// Ported from: test_dynamic_find_var
func TestVariable_DynamicFindVar(t *testing.T) {
	assertTemplateResult(t, "bar", "{{ [key] }}", map[string]interface{}{
		"key": "foo",
		"foo": "bar",
	})
}

// TestVariable_RawValueVariable tests raw value variable lookup.
// Ported from: test_raw_value_variable
func TestVariable_RawValueVariable(t *testing.T) {
	assertTemplateResult(t, "bar", "{{ [key] }}", map[string]interface{}{
		"key": "foo",
		"foo": "bar",
	})
}

// TestVariable_DynamicFindVarWithDrop tests dynamic variable lookup with drops.
// Ported from: test_dynamic_find_var_with_drop
func TestVariable_DynamicFindVarWithDrop(t *testing.T) {
	assertTemplateResult(t, "bar", "{{ [list[settings.zero]] }}", map[string]interface{}{
		"list":     []interface{}{"foo"},
		"settings": NewSettingsDrop(map[string]interface{}{"zero": 0}),
		"foo":      "bar",
	})

	assertTemplateResult(t, "foo", "{{ [list[settings.zero]['foo']] }}", map[string]interface{}{
		"list":     []interface{}{map[string]interface{}{"foo": "bar"}},
		"settings": NewSettingsDrop(map[string]interface{}{"zero": 0}),
		"bar":      "foo",
	})
}

// TestVariable_DoubleNestedVariableLookup tests double nested variable lookup.
// Ported from: test_double_nested_variable_lookup
func TestVariable_DoubleNestedVariableLookup(t *testing.T) {
	assertTemplateResult(t, "bar", "{{ list[list[settings.zero]]['foo'] }}", map[string]interface{}{
		"list":     []interface{}{1, map[string]interface{}{"foo": "bar"}},
		"settings": NewSettingsDrop(map[string]interface{}{"zero": 0}),
		"bar":      "foo",
	})
}

// TestVariable_FilterWithSingleTrailingComma tests filters with trailing commas.
// Ported from: test_filter_with_single_trailing_comma
func TestVariable_FilterWithSingleTrailingComma(t *testing.T) {
	template := `{{ "hello" | append: "world", }}`

	// In strict mode, this should raise an error
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	// In rigid mode, it should work
	env2 := liquid.NewEnvironment()
	env2.SetErrorMode("rigid")
	tags.RegisterStandardTags(env2)

	assertTemplateResult(t, "helloworld", template, map[string]interface{}{}, TemplateResultOptions{ErrorMode: "rigid"})
}

// TestVariable_MultipleFiltersWithTrailingCommas tests multiple filters with trailing commas.
// Ported from: test_multiple_filters_with_trailing_commas
func TestVariable_MultipleFiltersWithTrailingCommas(t *testing.T) {
	template := `{{ "hello" | append: "1", | append: "2", }}`

	// In strict mode, this should raise an error
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	// In rigid mode, it should work
	assertTemplateResult(t, "hello12", template, map[string]interface{}{}, TemplateResultOptions{ErrorMode: "rigid"})
}

// TestVariable_FilterWithColonButNoArguments tests filters with colon but no arguments.
// Ported from: test_filter_with_colon_but_no_arguments
func TestVariable_FilterWithColonButNoArguments(t *testing.T) {
	template := `{{ "test" | upcase: }}`

	// In strict mode, this should raise an error
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	// In rigid mode, it should work
	assertTemplateResult(t, "TEST", template, map[string]interface{}{}, TemplateResultOptions{ErrorMode: "rigid"})
}

// TestVariable_FilterChainWithColonNoArgs tests filter chains with colon but no args.
// Ported from: test_filter_chain_with_colon_no_args
func TestVariable_FilterChainWithColonNoArgs(t *testing.T) {
	template := `{{ "test" | append: "x" | upcase: }}`

	// In strict mode, this should raise an error
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	// In rigid mode, it should work
	assertTemplateResult(t, "TESTX", template, map[string]interface{}{}, TemplateResultOptions{ErrorMode: "rigid"})
}

// TestVariable_CombiningTrailingCommaAndEmptyArgs tests combining trailing comma and empty args.
// Ported from: test_combining_trailing_comma_and_empty_args
func TestVariable_CombiningTrailingCommaAndEmptyArgs(t *testing.T) {
	template := `{{ "test" | append: "x", | upcase: }}`

	// In strict mode, this should raise an error
	env := liquid.NewEnvironment()
	env.SetErrorMode("strict")
	tags.RegisterStandardTags(env)

	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err == nil {
		t.Error("Expected SyntaxError in strict mode")
	}

	// In rigid mode, it should work
	assertTemplateResult(t, "TESTX", template, map[string]interface{}{}, TemplateResultOptions{ErrorMode: "rigid"})
}
