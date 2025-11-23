package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// FunnyFilter provides test filters for output tests
type FunnyFilter struct{}

func (f *FunnyFilter) MakeFunny(_ interface{}) interface{} {
	return "LOL"
}

func (f *FunnyFilter) CiteFunny(input interface{}) interface{} {
	return "LOL: " + liquid.ToS(input, nil)
}

func (f *FunnyFilter) AddSmiley(input interface{}, smiley ...interface{}) interface{} {
	smileyStr := ":-)"
	if len(smiley) > 0 {
		smileyStr = liquid.ToS(smiley[0], nil)
	}
	return liquid.ToS(input, nil) + " " + smileyStr
}

func (f *FunnyFilter) AddTag(input interface{}, tag ...interface{}) interface{} {
	tagStr := "p"
	idStr := "foo"
	if len(tag) > 0 {
		tagStr = liquid.ToS(tag[0], nil)
	}
	if len(tag) > 1 {
		idStr = liquid.ToS(tag[1], nil)
	}
	return "<" + tagStr + " id=\"" + idStr + "\">" + liquid.ToS(input, nil) + "</" + tagStr + ">"
}

func (f *FunnyFilter) Paragraph(input interface{}) interface{} {
	return "<p>" + liquid.ToS(input, nil) + "</p>"
}

func (f *FunnyFilter) LinkTo(name interface{}, url interface{}) interface{} {
	return "<a href=\"" + liquid.ToS(url, nil) + "\">" + liquid.ToS(name, nil) + "</a>"
}

// TestOutput_Variable tests basic variable output.
// Ported from: test_variable
func TestOutput_Variable(t *testing.T) {
	assertTemplateResult(t, " bmw ", " {{best_cars}} ", map[string]interface{}{"best_cars": "bmw"})
}

// TestOutput_VariableTraversingWithTwoBrackets tests variable traversing with two brackets.
// Ported from: test_variable_traversing_with_two_brackets
func TestOutput_VariableTraversingWithTwoBrackets(t *testing.T) {
	source := "{{ site.data.menu[include.menu][include.locale] }}"
	assertTemplateResult(t, "it works!", source, map[string]interface{}{
		"site": map[string]interface{}{
			"data": map[string]interface{}{
				"menu": map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "it works!",
					},
				},
			},
		},
		"include": map[string]interface{}{
			"menu":   "foo",
			"locale": "bar",
		},
	})
}

// TestOutput_VariableTraversing tests variable traversing.
// Ported from: test_variable_traversing
func TestOutput_VariableTraversing(t *testing.T) {
	source := " {{car.bmw}} {{car.gm}} {{car.bmw}} "
	assertTemplateResult(t, " good bad good ", source, map[string]interface{}{
		"car": map[string]interface{}{
			"bmw": "good",
			"gm":  "bad",
		},
	})
}

// TestOutput_VariablePiping tests variable piping with filters.
// Ported from: test_variable_piping
func TestOutput_VariablePiping(t *testing.T) {
	text := ` {{ car.gm | make_funny }} `
	expected := ` LOL `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_VariablePipingWithInput tests variable piping with input.
// Ported from: test_variable_piping_with_input
func TestOutput_VariablePipingWithInput(t *testing.T) {
	text := ` {{ car.gm | cite_funny }} `
	expected := ` LOL: bad `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_VariablePipingWithArgs tests variable piping with arguments.
// Ported from: test_variable_piping_with_args
func TestOutput_VariablePipingWithArgs(t *testing.T) {
	text := ` {{ car.gm | add_smiley: ':-(' }} `
	expected := ` bad :-( `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_VariablePipingWithNoArgs tests variable piping without arguments.
// Ported from: test_variable_piping_with_no_args
func TestOutput_VariablePipingWithNoArgs(t *testing.T) {
	text := ` {{ car.gm | add_smiley }} `
	expected := ` bad :-) `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_MultipleVariablePipingWithArgs tests multiple variable piping with arguments.
// Ported from: test_multiple_variable_piping_with_args
func TestOutput_MultipleVariablePipingWithArgs(t *testing.T) {
	text := ` {{ car.gm | add_smiley: ':-(' | add_smiley: ':-('}} `
	expected := ` bad :-( :-( `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_VariablePipingWithMultipleArgs tests variable piping with multiple arguments.
// Ported from: test_variable_piping_with_multiple_args
func TestOutput_VariablePipingWithMultipleArgs(t *testing.T) {
	text := ` {{ car.gm | add_tag: 'span', 'bar'}} `
	expected := ` <span id="bar">bad</span> `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_VariablePipingWithVariableArgs tests variable piping with variable arguments.
// Ported from: test_variable_piping_with_variable_args
func TestOutput_VariablePipingWithVariableArgs(t *testing.T) {
	text := ` {{ car.gm | add_tag: 'span', car.bmw}} `
	expected := ` <span id="good">bad</span> `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment: env,
		StaticEnvironments: []map[string]interface{}{{
			"car": map[string]interface{}{
				"bmw": "good",
				"gm":  "bad",
			},
		}},
		RethrowErrors: false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_MultiplePipings tests multiple filter pipings.
// Ported from: test_multiple_pipings
func TestOutput_MultiplePipings(t *testing.T) {
	assigns := map[string]interface{}{"best_cars": "bmw"}
	text := ` {{ best_cars | cite_funny | paragraph }} `
	expected := ` <p>LOL: bmw</p> `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{assigns},
		RethrowErrors:      false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

// TestOutput_LinkTo tests link_to filter.
// Ported from: test_link_to
func TestOutput_LinkTo(t *testing.T) {
	text := ` {{ 'Typo' | link_to: 'http://typo.leetsoft.com' }} `
	expected := ` <a href="http://typo.leetsoft.com">Typo</a> `

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&FunnyFilter{})

	tmpl, err := liquid.ParseTemplate(text, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
	})

	output := tmpl.Render(ctx, &liquid.RenderOptions{})
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
