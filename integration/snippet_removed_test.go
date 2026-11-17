package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// TestSnippetTagReturnsUnknownTagError verifies that the snippet tag
// was removed in Shopify Liquid v5.11.0 and now returns an unknown tag error.
func TestSnippetTagReturnsUnknownTagError(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})

	template := `{% snippet foo %}content{% endsnippet %}`
	_, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})

	if err == nil {
		t.Fatal("Expected error for unknown 'snippet' tag, but parsing succeeded")
	}
	if err.Error() != "Liquid syntax error: Unknown tag 'snippet'" {
		t.Errorf("Expected Ruby's error, Liquid syntax error: Unknown tag 'snippet', got: %v", err)
	}
}
