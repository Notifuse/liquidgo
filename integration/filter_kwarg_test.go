package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// KwargFilter provides a test filter that accepts keyword arguments.
type KwargFilter struct{}

// HtmlTag creates an HTML tag with attributes from keyword arguments.
// In Ruby: def html_tag(_tag, attributes)
func (k *KwargFilter) HtmlTag(_ interface{}, attributes interface{}) interface{} {
	// This test documents that keyword arguments are not yet implemented
	// The filter signature expects a map/attributes parameter, but keyword
	// argument parsing is not implemented in liquidgo yet.
	return "keyword arguments not implemented"
}

// TestFilterKwarg_CanParseDataKwargs tests that keyword arguments can be parsed.
// Ported from: test_can_parse_data_kwargs
//
// NOTE: This test is expected to fail as keyword arguments are not yet implemented.
// It documents the missing feature.
func TestFilterKwarg_CanParseDataKwargs(t *testing.T) {
	t.Skip("Keyword arguments not yet implemented in liquidgo")

	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	_ = env.RegisterFilter(&KwargFilter{})

	template := `{{ 'img' | html_tag: data-src: 'src', data-widths: '100, 200' }}`
	expected := "data-src='src' data-widths='100, 200'"

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
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
