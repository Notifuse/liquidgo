package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
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
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% assign foo = 42 %}{% assign bar = 23 %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 1
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if result != "Liquid error: Memory limits exceeded" {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !rubyLimits(tmpl).Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 2
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	})
	result = renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
	if rubyLimits(tmpl).AssignScore() == 0 {
		t.Error("Expected assign_score to be set")
	}
}

func TestAssignScoreExceedingLimitFromCompositeObject(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% assign foo = 'aaaa' | reverse %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	limit := 3
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	})
	result := renderRuby(tmpl, nil, &rubyOptions{})
	if result != "Liquid error: Memory limits exceeded" {
		t.Errorf("Expected memory limit error, got %q", result)
	}
	if !rubyLimits(tmpl).Reached() {
		t.Error("Expected resource limits to be reached")
	}

	limit = 5
	setRubyLimits(tmpl, liquid.ResourceLimitsConfig{
		AssignScoreLimit: &limit,
	})
	result = renderRubyBang(tmpl, nil, &rubyOptions{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// ObjectWrapperDrop is a drop that wraps an object for testing assign scores.
type ObjectWrapperDrop struct {
	obj interface{}
}

// NewObjectWrapperDrop creates a new ObjectWrapperDrop.
func NewObjectWrapperDrop(obj interface{}) *ObjectWrapperDrop {
	return &ObjectWrapperDrop{obj: obj}
}

// LiquidGet returns the key a template reads.
func (o *ObjectWrapperDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "value":
		return o.Value(), true, nil
	}
	return nil, false, nil
}

// Value returns the wrapped object.
func (o *ObjectWrapperDrop) Value() interface{} {
	return o.obj
}

// assignScoreOf calculates the assign score for an object by assigning it and checking the score.
func assignScoreOf(obj interface{}) int {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	drop := NewObjectWrapperDrop(obj)
	// Create resource limits for tracking assign score
	rl := liquid.ResourceLimitsConfig{}
	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{{"drop": drop}},
		Limits:             &rl,
	}
	tmpl, err := liquid.ParseTemplate(`{% assign obj = drop.value %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		return 0
	}
	renderRubyBang(tmpl, ctx, &rubyOptions{})
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
