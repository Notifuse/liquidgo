package integration

import "testing"

// Amounts are computed exactly, with JSON data (float64) as applications pass it.
func TestDecimalArithmeticRendering(t *testing.T) {
	data := `{"price": 19.99, "qty": 3, "total": 1999, "rate": 0.0725, "items": [{"p": 0.1}, {"p": 0.2}]}`
	tests := []struct {
		template string
		expected string
	}{
		{`{{ 0.1 | plus: 0.2 }}`, "0.3"},
		{`{{ 0.3 | minus: 0.1 }}`, "0.2"},
		{`{{ 4.35 | times: 100 }}`, "435"},
		{`{{ price | times: qty }}`, "59.97"},
		{`{{ rate | times: 100 }}`, "7.25"},
		{`{{ total | divided_by: 100 }}`, "19.99"},
		{`{{ 1.005 | round: 2 }}`, "1.01"},
		{`{{ items | sum: "p" }}`, "0.3"},
		{`{{ 14 | divided_by: 3 }}`, "4"},
		{`{{ -7 | modulo: 3 }}`, "2"},
		{`{{ 100 | plus: 1 }}`, "101"},
		{`{{ qty | plus: 1 }}`, "4"},
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, jsonAssigns(t, data))
		})
	}
}
