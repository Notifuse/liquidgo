package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// MoneyFilter provides a test filter for money formatting.
type MoneyFilter struct{}

func (m *MoneyFilter) Money(input interface{}) interface{} {
	return " " + liquid.ToS(input, nil) + "$ "
}

// CanadianMoneyFilter provides a Canadian money filter.
type CanadianMoneyFilter struct{}

func (c *CanadianMoneyFilter) Money(input interface{}) interface{} {
	return " " + liquid.ToS(input, nil) + "$ CAD "
}

// TestHashOrdering_GlobalRegisterOrder tests that filter registration order matters.
// Ported from: test_global_register_order
func TestHashOrdering_GlobalRegisterOrder(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Register filters in order: MoneyFilter, then CanadianMoneyFilter
	_ = env.RegisterFilter(&MoneyFilter{})
	_ = env.RegisterFilter(&CanadianMoneyFilter{})

	// The last registered filter should take precedence
	tmpl, err := liquid.ParseTemplate("{{1000 | money}}", &liquid.TemplateOptions{
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
	expected := " 1000$ CAD "

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
