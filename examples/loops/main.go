package main

import (
	"context"
	"fmt"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func main() {
	env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard()})
	if err != nil {
		panic(err)
	}

	template := `
<h1>Product List</h1>
<ul>
{% for product in products %}
  <li>
    <strong>{{ product.name }}</strong>
    - ${{ product.price }}
    {% if product.on_sale %}
      <span class="sale">ON SALE!</span>
    {% endif %}
  </li>
{% endfor %}
</ul>

<p>Total products: {{ products | size }}</p>
`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		panic(err)
	}

	res, err := tmpl.Render(context.Background(), map[string]interface{}{
		"products": []map[string]interface{}{
			{"name": "Widget", "price": 19.99, "on_sale": false},
			{"name": "Gadget", "price": 29.99, "on_sale": true},
			{"name": "Doohickey", "price": 9.99, "on_sale": false},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Output)
}
