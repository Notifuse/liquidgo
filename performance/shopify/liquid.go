package shopify

import (
	"github.com/Notifuse/liquidgo/v5/internal/methodfilters"
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// AddTo adds the custom tags and filters of Shopify themes to config.
func AddTo(config *liquid.EnvironmentConfig) {
	if config.Tags == nil {
		config.Tags = map[string]liquid.TagConstructor{}
	}
	config.Tags["paginate"] = func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		tag, err := NewPaginate(tagName, markup, parseContext)
		if err != nil {
			return nil, err
		}
		return tag, nil
	}
	config.Tags["form"] = func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		tag, err := NewCommentForm(tagName, markup, parseContext)
		if err != nil {
			return nil, err
		}
		return tag, nil
	}

	// each filter object exposes several methods
	for _, filters := range []interface{}{&JSONFilter{}, &MoneyFilter{}, &WeightFilter{}, &ShopFilter{}, &TagFilter{}} {
		config.Filters = append(config.Filters, methodfilters.Of(filters)...)
	}
}
