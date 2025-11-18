package shopify

import (
	"github.com/Notifuse/liquidgo/liquid"
)

// TagConstructor is a function type that creates a tag instance.
type TagConstructor func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error)

// RegisterAll registers all Shopify custom tags and filters with the environment
func RegisterAll(env *liquid.Environment) {
	// Register tags
	env.RegisterTag("paginate", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewPaginate(tagName, markup, parseContext)
	}))
	env.RegisterTag("form", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewCommentForm(tagName, markup, parseContext)
	}))

	// Register filters (each filter object exposes multiple methods)
	env.RegisterFilter(&JsonFilter{})
	env.RegisterFilter(&MoneyFilter{})
	env.RegisterFilter(&WeightFilter{})
	env.RegisterFilter(&ShopFilter{})
	env.RegisterFilter(&TagFilter{})
}

