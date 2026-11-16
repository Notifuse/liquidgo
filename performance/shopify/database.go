package shopify

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var cachedTables map[string]interface{}

// Tables loads and returns the vision database as a map
func Tables() (map[string]interface{}, error) {
	if cachedTables != nil {
		return cachedTables, nil
	}

	// Find the database file
	databasePath := filepath.Join("..", "reference-liquid", "performance", "shopify", "vision.database.yml")
	data, err := os.ReadFile(databasePath)
	if err != nil {
		return nil, err
	}

	var db map[string]interface{}
	if err := yaml.Unmarshal(data, &db); err != nil {
		return nil, err
	}

	// From vision source: Add collections to products
	products, _ := db["products"].([]interface{})
	collections, _ := db["collections"].([]interface{})

	for _, p := range products {
		product, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		var productCollections []interface{}
		productID := product["id"]

		for _, c := range collections {
			collection, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			collectionProducts, _ := collection["products"].([]interface{})
			for _, cp := range collectionProducts {
				if cpMap, ok := cp.(map[string]interface{}); ok {
					if cpMap["id"] == productID {
						productCollections = append(productCollections, collection)
						break
					}
				}
			}
		}

		product["collections"] = productCollections
	}

	// Key the tables by handles. ordered keeps each table in the order of
	// the YAML file, as Ruby's hash keeps insertion order: the current
	// collection, product and blog are the first of their table, as Ruby's
	// values.first, so that every run renders the same data.
	assigns := make(map[string]interface{})
	ordered := make(map[string][]interface{})

	for key, values := range db {
		if valueSlice, ok := values.([]interface{}); ok {
			ordered[key] = valueSlice
			handleMap := make(map[string]interface{})
			for _, v := range valueSlice {
				if vMap, ok := v.(map[string]interface{}); ok {
					if handle, ok := vMap["handle"]; ok {
						handleMap[handle.(string)] = vMap
					}
				}
			}
			assigns[key] = handleMap
		} else {
			assigns[key] = values
		}
	}

	// Some standard direct accessors
	if collections := ordered["collections"]; len(collections) > 0 {
		assigns["collection"] = collections[0]
	}

	if products := ordered["products"]; len(products) > 0 {
		assigns["product"] = products[0]
	}

	if blogs := ordered["blogs"]; len(blogs) > 0 {
		assigns["blog"] = blogs[0]
		if blog, ok := blogs[0].(map[string]interface{}); ok {
			if articles, ok := blog["articles"].([]interface{}); ok && len(articles) > 0 {
				assigns["article"] = articles[0]
			}
		}
	}

	// Add shop object (needed by templates)
	assigns["shop"] = map[string]interface{}{
		"name":     "Test Shop",
		"currency": "USD",
	}

	// Add linklists alias (templates use linklists, YAML has link_lists)
	if linkLists, ok := assigns["link_lists"]; ok {
		assigns["linklists"] = linkLists
	}

	// Build cart
	if lineItems, ok := ordered["line_items"]; ok {
		var totalPrice float64
		var itemCount int
		var items []interface{}

		for _, v := range lineItems {
			items = append(items, v)
			if item, ok := v.(map[string]interface{}); ok {
				linePrice, _ := item["line_price"].(int)
				quantity, _ := item["quantity"].(int)
				totalPrice += float64(linePrice * quantity)
				itemCount += quantity
			}
		}

		assigns["cart"] = map[string]interface{}{
			"total_price": totalPrice,
			"item_count":  itemCount,
			"items":       items,
		}
	}

	cachedTables = assigns
	return assigns, nil
}
