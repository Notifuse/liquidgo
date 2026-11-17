package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/internal/methodfilters"
	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
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

	// Register filters in order: MoneyFilter, then CanadianMoneyFilter
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), Filters: append(methodfilters.Of(&MoneyFilter{}), methodfilters.Of(&CanadianMoneyFilter{})...)})

	// The last registered filter should take precedence
	tmpl, err := liquid.ParseTemplate("{{1000 | money}}", &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	ctx := &rubyContext{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{},
		RethrowErrors:      false,
	}

	output := renderRuby(tmpl, ctx, &rubyOptions{})
	expected := " 1000$ CAD "

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
