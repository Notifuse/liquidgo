package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

func TestCapturesBlockContentInVariable(t *testing.T) {
	assertTemplateResult(t, "test string", `{% capture 'var' %}test string{% endcapture %}{{var}}`, nil)
}

func TestCaptureWithHyphenInVariableName(t *testing.T) {
	templateSource := `{% capture this-thing %}Print this-thing{% endcapture -%}
{{ this-thing -}}`
	assertTemplateResult(t, "Print this-thing", templateSource, nil)
}

func TestCaptureToVariableFromOuterScopeIfExisting(t *testing.T) {
	templateSource := `{% assign var = '' -%}
{% if true -%}
  {% capture var %}first-block-string{% endcapture -%}
{% endif -%}
{% if true -%}
  {% capture var %}test-string{% endcapture -%}
{% endif -%}
{{var-}}`
	assertTemplateResult(t, "test-string", templateSource, nil)
}

func TestAssigningFromCapture(t *testing.T) {
	templateSource := `{% assign first = '' -%}
{% assign second = '' -%}
{% for number in (1..3) -%}
  {% capture first %}{{number}}{% endcapture -%}
  {% assign second = first -%}
{% endfor -%}
{{ first }}-{{ second -}}`
	assertTemplateResult(t, "3-3", templateSource, nil)
}

func TestIncrementAssignScoreByBytesNotCharacters(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env) // Register tags for parsing
	tmpl, err := liquid.ParseTemplate(`{% capture foo %}すごい{% endcapture %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	tmpl.RenderBang(nil, &liquid.RenderOptions{})
	if tmpl.ResourceLimits().AssignScore() != 9 {
		t.Errorf("Expected assign_score 9, got %d", tmpl.ResourceLimits().AssignScore())
	}
}

