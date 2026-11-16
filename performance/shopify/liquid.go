package shopify

import (
	"github.com/Notifuse/liquidgo/v5/internal/methodfilters"
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// RegisterAll registers all Shopify custom tags and filters with the environment
func RegisterAll(env *liquid.Environment) {
	// Register tags
	env.RegisterTag("paginate", func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		tag, err := NewPaginate(tagName, markup, parseContext)
		if err != nil {
			return nil, err
		}
		return tag, nil
	})
	env.RegisterTag("form", func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		tag, err := NewCommentForm(tagName, markup, parseContext)
		if err != nil {
			return nil, err
		}
		return tag, nil
	})

	// Register filters (each filter object exposes multiple methods)
	_ = env.RegisterFilter(methodfilters.Of(&JSONFilter{})...)
	_ = env.RegisterFilter(methodfilters.Of(&MoneyFilter{})...)
	_ = env.RegisterFilter(methodfilters.Of(&WeightFilter{})...)
	_ = env.RegisterFilter(methodfilters.Of(&ShopFilter{})...)
	_ = env.RegisterFilter(methodfilters.Of(&TagFilter{})...)
}
