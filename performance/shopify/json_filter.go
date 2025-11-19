package shopify

import (
	"encoding/json"
)

// JsonFilter provides JSON encoding functionality
type JsonFilter struct{}

// Json converts an object to JSON, excluding the "collections" key
func (f *JsonFilter) Json(input interface{}) (string, error) {
	// If input is a map, remove collections key
	if m, ok := input.(map[string]interface{}); ok {
		filtered := make(map[string]interface{})
		for k, v := range m {
			if k != "collections" {
				filtered[k] = v
			}
		}
		input = filtered
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
