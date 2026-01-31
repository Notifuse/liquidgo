package integration

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestSnippetTagReturnsUnknownTagError verifies that the snippet tag
// was removed in Shopify Liquid v5.11.0 and now returns an unknown tag error.
func TestSnippetTagReturnsUnknownTagError(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	template := `{% snippet foo %}content{% endsnippet %}`
	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})

	if err == nil {
		t.Fatal("Expected error for unknown 'snippet' tag, but parsing succeeded")
	}
	if !strings.Contains(err.Error(), "unknown_tag") && !strings.Contains(err.Error(), "unknown tag") {
		t.Errorf("Expected 'unknown tag' or 'unknown_tag' error, got: %v", err)
	}
}
