package shopify

import (
	"fmt"
)

// MoneyFilter provides money formatting functionality
type MoneyFilter struct{}

// MoneyWithCurrency formats a price in cents as dollars with currency symbol
func (f *MoneyFilter) MoneyWithCurrency(money interface{}) string {
	if money == nil {
		return ""
	}

	var cents float64
	switch v := money.(type) {
	case int:
		cents = float64(v)
	case int64:
		cents = float64(v)
	case float64:
		cents = v
	default:
		return ""
	}

	return fmt.Sprintf("$ %.2f USD", cents/100.0)
}

// Money formats a price in cents as dollars
func (f *MoneyFilter) Money(money interface{}) string {
	if money == nil {
		return ""
	}

	var cents float64
	switch v := money.(type) {
	case int:
		cents = float64(v)
	case int64:
		cents = float64(v)
	case float64:
		cents = v
	default:
		return ""
	}

	return fmt.Sprintf("$ %.2f", cents/100.0)
}

