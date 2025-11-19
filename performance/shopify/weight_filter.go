package shopify

import (
	"fmt"
)

// WeightFilter provides weight conversion functionality
type WeightFilter struct{}

// Weight converts grams to kilograms
func (f *WeightFilter) Weight(grams interface{}) string {
	var g float64
	switch v := grams.(type) {
	case int:
		g = float64(v)
	case int64:
		g = float64(v)
	case float64:
		g = v
	default:
		return "0.00"
	}

	return fmt.Sprintf("%.2f", g/1000.0)
}

// WeightWithUnit converts grams to kilograms with unit
func (f *WeightFilter) WeightWithUnit(grams interface{}) string {
	return f.Weight(grams) + " kg"
}
