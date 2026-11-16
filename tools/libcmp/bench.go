package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// benchTemplate is an email a newsletter sends: a greeting, a loop over
// products with conditions and filters, and a footer. It uses only what every
// compared library renders.
const benchTemplate = `<h1>Hello {{ contact.first_name | capitalize }},</h1>
<p>{{ intro | escape }}</p>
{% for product in products %}
  {% if product.available %}
  <div class="product">
    <h2>{{ product.title | upcase }}</h2>
    <p>{{ product.description | truncate: 80 }}</p>
    <p>{{ product.price | times: 2 }} EUR{% if product.discount > 0 %}, {{ product.discount }}% off{% endif %}</p>
    <p>{{ product.tags | join: ", " }}</p>
  </div>
  {% else %}
  <p>{{ product.title }} is sold out.</p>
  {% endif %}
{% endfor %}
<p>{{ products | size }} products, {{ footer | append: "." }}</p>
`

func benchData() map[string]any {
	products := make([]any, 50)
	for i := range products {
		products[i] = map[string]any{
			"title":       fmt.Sprintf("product %d", i),
			"description": strings.Repeat("A fine product, made to last. ", 5),
			"price":       10 + i,
			"discount":    i % 3 * 5,
			"available":   i%4 != 0,
			"tags":        []any{"new", "summer", "sale"},
		}
	}
	return map[string]any{
		"contact":  map[string]any{"first_name": "ann"},
		"intro":    "Our <new> collection is here.",
		"products": products,
		"footer":   "Thank you",
	}
}

// bench prints the median time of a render of the benchmark template, parsed
// once, over 15 rounds of at least 200 ms each after a warm-up.
func bench(lib library) error {
	fn, err := lib.parse(benchTemplate)
	if err != nil {
		return err
	}
	data := benchData()
	if _, err := fn(data); err != nil {
		return err
	}
	var rounds []time.Duration
	for r := 0; r < 15; r++ {
		n, start := 0, time.Now()
		for time.Since(start) < 200*time.Millisecond {
			if _, err := fn(data); err != nil {
				return err
			}
			n++
		}
		rounds = append(rounds, time.Since(start)/time.Duration(n))
	}
	fmt.Printf("%s\t%v\n", lib.name, median(rounds))
	return nil
}

func median(ds []time.Duration) time.Duration {
	sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
	return ds[len(ds)/2]
}

// export writes the benchmark template, its data and the hostile templates
// to dir, for the Ruby scripts of ruby/, which render them with Shopify
// Liquid itself.
func export(dir string) error {
	hostile := make([]string, len(hostileTemplates))
	for i, h := range hostileTemplates {
		hostile[i] = h.source
	}
	b, err := json.Marshal(map[string]any{
		"bench":   map[string]any{"template": benchTemplate, "data": benchData()},
		"hostile": hostile,
		"limits":  map[string]int{"render_length_limit": limitLength, "render_score_limit": limitScore, "assign_score_limit": limitAssign},
	})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "templates.json"), b, 0o644)
}
