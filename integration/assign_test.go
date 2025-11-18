package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

func TestAssignWithHyphenInVariableName(t *testing.T) {
	templateSource := `{% assign this-thing = 'Print this-thing' -%}
{{ this-thing -}}`
	assertTemplateResult(t, "Print this-thing", templateSource, nil)
}

func TestAssignedVariable(t *testing.T) {
	assertTemplateResult(t, ".foo.", `{% assign foo = values %}.{{ foo[0] }}.`, map[string]interface{}{
		"values": []interface{}{"foo", "bar", "baz"},
	})

	assertTemplateResult(t, ".bar.", `{% assign foo = values %}.{{ foo[1] }}.`, map[string]interface{}{
		"values": []interface{}{"foo", "bar", "baz"},
	})
}

func TestAssignWithFilter(t *testing.T) {
	assertTemplateResult(t, ".bar.", `{% assign foo = values | split: "," %}.{{ foo[1] }}.`, map[string]interface{}{
		"values": "foo,bar,baz",
	})
}

func TestAssignSyntaxError(t *testing.T) {
	assertMatchSyntaxError(t, "assign", `{% assign foo not values %}.`)
}

func TestAssignUsesErrorMode(t *testing.T) {
	assertMatchSyntaxError(t, "Expected dotdot but found pipe in ", `{% assign foo = ('X' | downcase) %}`, "strict")
	assertTemplateResult(t, "", `{% assign foo = ('X' | downcase) %}`, nil, TemplateResultOptions{ErrorMode: "lax"})
}

func TestExpressionWithWhitespaceInSquareBrackets(t *testing.T) {
	source := `{% assign r = a[ 'b' ] %}{{ r }}`
	assertTemplateResult(t, "result", source, map[string]interface{}{
		"a": map[string]interface{}{"b": "result"},
	})
}

func TestAssignScoreExceedingResourceLimit(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env) // Register tags for parsing
	tmpl, err := liquid.ParseTemplate(`{% assign foo = 42 %}{% assign bar = 23 %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 1
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "Liquid error: Memory limits exceeded" {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !tmpl.ResourceLimits().Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 2
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}))
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
	if tmpl.ResourceLimits().AssignScore() == 0 {
		t.Error("Expected assign_score to be set")
	}
}

func TestAssignScoreExceedingLimitFromCompositeObject(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env) // Register tags for parsing
	tmpl, err := liquid.ParseTemplate(`{% assign foo = 'aaaa' | reverse %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 3
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}))
	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "Liquid error: Memory limits exceeded" {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !tmpl.ResourceLimits().Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 5
	tmpl.SetResourceLimits(liquid.NewResourceLimits(liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	}))
	result = tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// ObjectWrapperDrop is a drop that wraps an object for testing assign scores.
type ObjectWrapperDrop struct {
	*liquid.Drop
	obj interface{}
}

// NewObjectWrapperDrop creates a new ObjectWrapperDrop.
func NewObjectWrapperDrop(obj interface{}) *ObjectWrapperDrop {
	return &ObjectWrapperDrop{
		Drop: liquid.NewDrop(),
		obj:  obj,
	}
}

// Value returns the wrapped object.
func (o *ObjectWrapperDrop) Value() interface{} {
	return o.obj
}

// assignScoreOf calculates the assign score for an object by assigning it and checking the score.
func assignScoreOf(obj interface{}) int {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env) // Register tags for parsing
	drop := NewObjectWrapperDrop(obj)
	// Create resource limits for tracking assign score
	rl := liquid.NewResourceLimits(liquid.ResourceLimitsConfig{})
	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"drop": drop}},
		ResourceLimits:     rl,
	})
	tmpl, err := liquid.ParseTemplate(`{% assign obj = drop.value %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		return 0
	}
	tmpl.RenderBang(ctx, &liquid.RenderOptions{})
	return ctx.ResourceLimits().AssignScore()
}

func TestAssignScoreOfInt(t *testing.T) {
	score := assignScoreOf(123)
	if score != 1 {
		t.Errorf("Expected assign score 1, got %d", score)
	}
}

func TestAssignScoreOfStringCountsBytes(t *testing.T) {
	score := assignScoreOf("123")
	if score != 3 {
		t.Errorf("Expected assign score 3, got %d", score)
	}

	score = assignScoreOf("12345")
	if score != 5 {
		t.Errorf("Expected assign score 5, got %d", score)
	}

	score = assignScoreOf("すごい")
	if score != 9 {
		t.Errorf("Expected assign score 9, got %d", score)
	}
}

func TestAssignScoreOfArray(t *testing.T) {
	score := assignScoreOf([]interface{}{})
	if score != 1 {
		t.Errorf("Expected assign score 1, got %d", score)
	}

	score = assignScoreOf([]interface{}{123})
	if score != 2 {
		t.Errorf("Expected assign score 2, got %d", score)
	}

	score = assignScoreOf([]interface{}{123, "abcd"})
	if score != 6 {
		t.Errorf("Expected assign score 6, got %d", score)
	}
}

func TestAssignScoreOfHash(t *testing.T) {
	score := assignScoreOf(map[string]interface{}{})
	if score != 1 {
		t.Errorf("Expected assign score 1, got %d", score)
	}

	score = assignScoreOf(map[string]interface{}{"int": 123})
	if score != 5 {
		t.Errorf("Expected assign score 5, got %d", score)
	}

	score = assignScoreOf(map[string]interface{}{"int": 123, "str": "abcd"})
	if score != 12 {
		t.Errorf("Expected assign score 12, got %d", score)
	}
}
