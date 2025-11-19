package liquid

import (
	"strconv"
	"strings"
)

// MethodLiteral represents a method literal (blank, empty).
type MethodLiteral struct {
	MethodName string
	ToString   string
}

var (
	conditionMethodLiterals = map[string]*MethodLiteral{
		"blank": {MethodName: "blank", ToString: ""},
		"empty": {MethodName: "empty", ToString: ""},
	}
)

// ConditionOperator is a function type for condition operators.
type ConditionOperator func(cond *Condition, left, right interface{}) (bool, error)

// Condition represents a Liquid condition.
type Condition struct {
	left          interface{}
	operator      string
	right         interface{}
	childRelation string // "or" or "and"
	childCondition *Condition
	attachment    interface{}
}

// NewCondition creates a new Condition.
func NewCondition(left interface{}, operator string, right interface{}) *Condition {
	return &Condition{
		left:     left,
		operator: operator,
		right:    right,
	}
}

// Left returns the left side of the condition.
func (c *Condition) Left() interface{} {
	return c.left
}

// Operator returns the operator.
func (c *Condition) Operator() string {
	return c.operator
}

// Right returns the right side of the condition.
func (c *Condition) Right() interface{} {
	return c.right
}

// ChildCondition returns the child condition.
func (c *Condition) ChildCondition() *Condition {
	return c.childCondition
}

// Attachment returns the attachment.
func (c *Condition) Attachment() interface{} {
	return c.attachment
}

// Or chains this condition with another using OR.
func (c *Condition) Or(condition *Condition) {
	c.childRelation = "or"
	c.childCondition = condition
}

// And chains this condition with another using AND.
func (c *Condition) And(condition *Condition) {
	c.childRelation = "and"
	c.childCondition = condition
}

// Attach attaches an attachment to this condition.
func (c *Condition) Attach(attachment interface{}) {
	c.attachment = attachment
}

// Else returns true if this is an else condition.
func (c *Condition) Else() bool {
	return false
}

// Evaluate evaluates the condition in the given context.
type ConditionContext interface {
	Evaluate(expr interface{}) interface{}
}

func (c *Condition) Evaluate(context ConditionContext) (bool, error) {
	condition := c
	var result bool
	var err error

	for {
		result, err = c.interpretCondition(condition.left, condition.right, condition.operator, context)
		if err != nil {
			return false, err
		}

		resultVal := ToLiquidValue(result)
		shouldContinue := false
		switch condition.childRelation {
		case "or":
			if resultVal == nil || resultVal == false || resultVal == "" {
				shouldContinue = true
			}
		case "and":
			if resultVal != nil && resultVal != false && resultVal != "" {
				shouldContinue = true
			}
		}

		if !shouldContinue || condition.childCondition == nil {
			break
		}
		condition = condition.childCondition
	}

	return result, nil
}

func (c *Condition) interpretCondition(left, right interface{}, op string, context ConditionContext) (bool, error) {
	// If operator is empty, just evaluate the left side
	if op == "" {
		result := context.Evaluate(left)
		resultVal := ToLiquidValue(result)
		// Convert to bool
		if resultVal == nil || resultVal == false || resultVal == "" {
			return false, nil
		}
		return true, nil
	}

	leftVal := ToLiquidValue(context.Evaluate(left))
	rightVal := ToLiquidValue(context.Evaluate(right))

	operator := getConditionOperator(op)
	if operator == nil {
		return false, NewArgumentError("Unknown operator " + op)
	}

	return operator(c, leftVal, rightVal)
}

func getConditionOperator(op string) ConditionOperator {
	switch op {
	case "==":
		return func(cond *Condition, left, right interface{}) (bool, error) {
			return cond.equalVariables(left, right), nil
		}
	case "!=", "<>":
		return func(cond *Condition, left, right interface{}) (bool, error) {
			return !cond.equalVariables(left, right), nil
		}
	case "<":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			return compareValues(left, right) < 0, nil
		}
	case ">":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			return compareValues(left, right) > 0, nil
		}
	case ">=":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			return compareValues(left, right) >= 0, nil
		}
	case "<=":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			return compareValues(left, right) <= 0, nil
		}
	case "contains":
		return func(_ *Condition, left, right interface{}) (bool, error) {
			return containsOperator(left, right), nil
		}
	default:
		return nil
	}
}

func (c *Condition) equalVariables(left, right interface{}) bool {
	if ml, ok := left.(*MethodLiteral); ok {
		return checkMethodLiteral(ml, right)
	}

	if ml, ok := right.(*MethodLiteral); ok {
		return checkMethodLiteral(ml, left)
	}

	return left == right
}

func checkMethodLiteral(ml *MethodLiteral, obj interface{}) bool {
	switch ml.MethodName {
	case "blank":
		if str, ok := obj.(string); ok {
			return strings.TrimSpace(str) == ""
		}
		return obj == nil || obj == ""
	case "empty":
		if str, ok := obj.(string); ok {
			return str == ""
		}
		if arr, ok := obj.([]interface{}); ok {
			return len(arr) == 0
		}
		if m, ok := obj.(map[string]interface{}); ok {
			return len(m) == 0
		}
		return obj == nil || obj == ""
	default:
		return false
	}
}

func compareValues(left, right interface{}) int {
	// Simple numeric comparison
	leftNum, leftOk := toNumber(left)
	rightNum, rightOk := toNumber(right)
	if leftOk && rightOk {
		if leftNum < rightNum {
			return -1
		} else if leftNum > rightNum {
			return 1
		}
		return 0
	}

	// String comparison
	leftStr := ToS(left, nil)
	rightStr := ToS(right, nil)
	if leftStr < rightStr {
		return -1
	} else if leftStr > rightStr {
		return 1
	}
	return 0
}

func toNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case string:
		if num, err := strconv.ParseFloat(n, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

func containsOperator(left, right interface{}) bool {
	if left == nil || right == nil {
		return false
	}

	rightStr := ToS(right, nil)

	// Check if left is a string
	if leftStr, ok := left.(string); ok {
		return strings.Contains(leftStr, rightStr)
	}

	// Check if left is a slice/array
	if arr, ok := left.([]interface{}); ok {
		for _, item := range arr {
			if ToS(item, nil) == rightStr {
				return true
			}
		}
	}

	// Check if left is a map
	if m, ok := left.(map[string]interface{}); ok {
		_, exists := m[rightStr]
		return exists
	}

	return false
}

// ParseConditionExpression parses an expression for use in conditions.
func ParseConditionExpression(parseContext ParseContextInterface, markup string, safe bool) interface{} {
	if ml, ok := conditionMethodLiterals[markup]; ok {
		return ml
	}
	if safe {
		// For safe parsing, we'd use SafeParseExpression if available
		return parseContext.ParseExpression(markup)
	}
	return parseContext.ParseExpression(markup)
}

// ElseCondition represents an else condition.
type ElseCondition struct {
	*Condition
}

// NewElseCondition creates a new ElseCondition.
func NewElseCondition() *ElseCondition {
	return &ElseCondition{
		Condition: NewCondition(nil, "", nil),
	}
}

// Else returns true for else conditions.
func (e *ElseCondition) Else() bool {
	return true
}

// Evaluate always returns true for else conditions.
func (e *ElseCondition) Evaluate(context ConditionContext) (bool, error) {
	return true, nil
}

