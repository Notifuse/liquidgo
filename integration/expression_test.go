package integration

import (
	"testing"
)

// TestExpression_KeywordLiterals tests keyword literals.
// Ported from: test_keyword_literals
func TestExpression_KeywordLiterals(t *testing.T) {
	assertTemplateResult(t, "true", "{{ true }}", map[string]interface{}{})
}

// TestExpression_String tests string expressions.
// Ported from: test_string
func TestExpression_String(t *testing.T) {
	assertTemplateResult(t, "single quoted", "{{'single quoted'}}", map[string]interface{}{})
	assertTemplateResult(t, "double quoted", `{{"double quoted"}}`, map[string]interface{}{})
	assertTemplateResult(t, "spaced", "{{ 'spaced' }}", map[string]interface{}{})
	assertTemplateResult(t, "spaced2", "{{ 'spaced2' }}", map[string]interface{}{})
	assertTemplateResult(t, "emoji🔥", "{{ 'emoji🔥' }}", map[string]interface{}{})
}

// TestExpression_Int tests integer expressions.
// Ported from: test_int
func TestExpression_Int(t *testing.T) {
	assertTemplateResult(t, "456", "{{ 456 }}", map[string]interface{}{})
}

// TestExpression_Float tests float expressions.
// Ported from: test_float
func TestExpression_Float(t *testing.T) {
	assertTemplateResult(t, "-17.42", "{{ -17.42 }}", map[string]interface{}{})
	assertTemplateResult(t, "2.5", "{{ 2.5 }}", map[string]interface{}{})
}

// TestExpression_Range tests range expressions.
// Ported from: test_range
func TestExpression_Range(t *testing.T) {
	assertTemplateResult(t, "3..4", "{{ ( 3 .. 4 ) }}", map[string]interface{}{})
}

// TestExpression_Comparison tests comparison operators.
func TestExpression_Comparison(t *testing.T) {
	// Equality
	assertTemplateResult(t, "true", "{% if 5 == 5 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 5 == 3 %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Inequality
	assertTemplateResult(t, "true", "{% if 5 != 3 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 5 != 5 %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Less than
	assertTemplateResult(t, "true", "{% if 3 < 5 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 5 < 3 %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Greater than
	assertTemplateResult(t, "true", "{% if 5 > 3 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 3 > 5 %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Less than or equal
	assertTemplateResult(t, "true", "{% if 5 <= 5 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "true", "{% if 3 <= 5 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 5 <= 3 %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Greater than or equal
	assertTemplateResult(t, "true", "{% if 5 >= 5 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "true", "{% if 5 >= 3 %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 3 >= 5 %}true{% else %}false{% endif %}", map[string]interface{}{})
}

// TestExpression_Logical tests logical operators.
func TestExpression_Logical(t *testing.T) {
	// AND
	assertTemplateResult(t, "true", "{% if true and true %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if true and false %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if false and true %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if false and false %}true{% else %}false{% endif %}", map[string]interface{}{})

	// OR
	assertTemplateResult(t, "true", "{% if true or true %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "true", "{% if true or false %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "true", "{% if false or true %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if false or false %}true{% else %}false{% endif %}", map[string]interface{}{})
}

// TestExpression_OperatorPrecedence tests operator precedence.
func TestExpression_OperatorPrecedence(t *testing.T) {
	// Multiplication before addition
	assertTemplateResult(t, "11", "{% if 5 + 3 * 2 == 11 %}11{% else %}wrong{% endif %}", map[string]interface{}{})

	// AND before OR
	assertTemplateResult(t, "true", "{% if true or false and false %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Comparison before logical
	assertTemplateResult(t, "true", "{% if 5 > 3 and 2 < 4 %}true{% else %}false{% endif %}", map[string]interface{}{})
}

// TestExpression_Contains tests contains operator.
func TestExpression_Contains(t *testing.T) {
	assertTemplateResult(t, "true", "{% if 'hello' contains 'ell' %}true{% else %}false{% endif %}", map[string]interface{}{})
	assertTemplateResult(t, "false", "{% if 'hello' contains 'xyz' %}true{% else %}false{% endif %}", map[string]interface{}{})

	// Array contains
	assertTemplateResult(t, "true", "{% if arr contains 2 %}true{% else %}false{% endif %}", map[string]interface{}{
		"arr": []interface{}{1, 2, 3},
	})
	assertTemplateResult(t, "false", "{% if arr contains 4 %}true{% else %}false{% endif %}", map[string]interface{}{
		"arr": []interface{}{1, 2, 3},
	})
}
