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

	// Key the tables by handles
	assigns := make(map[string]interface{})

	for key, values := range db {
		if valueSlice, ok := values.([]interface{}); ok {
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
	if collectionsMap, ok := assigns["collections"].(map[string]interface{}); ok {
		for _, v := range collectionsMap {
			assigns["collection"] = v
			break
		}
	}

	if productsMap, ok := assigns["products"].(map[string]interface{}); ok {
		for _, v := range productsMap {
			assigns["product"] = v
			break
		}
	}

	if blogsMap, ok := assigns["blogs"].(map[string]interface{}); ok {
		for _, v := range blogsMap {
			assigns["blog"] = v
			if blog, ok := v.(map[string]interface{}); ok {
				if articles, ok := blog["articles"].([]interface{}); ok && len(articles) > 0 {
					assigns["article"] = articles[0]
				}
			}
			break
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
	if lineItemsMap, ok := assigns["line_items"].(map[string]interface{}); ok {
		var totalPrice float64
		var itemCount int
		var items []interface{}

		for _, v := range lineItemsMap {
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
