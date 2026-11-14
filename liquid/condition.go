package liquid

import (
	"reflect"
	"strings"
	"time"
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
	left           interface{}
	right          interface{}
	attachment     interface{}
	childCondition *Condition
	operator       string
	childRelation  string
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

// ConditionContext provides context for evaluating conditions.
type ConditionContext interface {
	Evaluate(expr interface{}) interface{}
}

// Evaluate evaluates the condition in the given context.
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
		return orderOperator(func(cmp int) bool { return cmp < 0 })
	case ">":
		return orderOperator(func(cmp int) bool { return cmp > 0 })
	case ">=":
		return orderOperator(func(cmp int) bool { return cmp >= 0 })
	case "<=":
		return orderOperator(func(cmp int) bool { return cmp <= 0 })
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

	return liquidEqual(left, right)
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
		// Reflection fallback for typed slices/arrays/maps
		// This matches Ruby's duck-typing behavior: objects respond to .empty?
		if obj != nil {
			v := reflect.ValueOf(obj)
			kind := v.Kind()
			if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
				return v.Len() == 0
			}
		}
		return obj == nil || obj == ""
	default:
		return false
	}
}

// orderOperator returns <, >, <= or >=, holding when holds accepts the
// comparison of the two sides.
func orderOperator(holds func(cmp int) bool) ConditionOperator {
	return func(_ *Condition, left, right interface{}) (bool, error) {
		cmp, ok, err := compareValues(left, right)
		if err != nil || !ok {
			return false, err
		}
		return holds(cmp), nil
	}
}

// orderKind is how a value takes part in <, >, <= and >=.
type orderKind int

const (
	// unordered: the value has no comparison methods in Ruby, as nil, a
	// boolean, an array, a hash or a drop.
	unordered orderKind = iota
	orderedNumber
	orderedString
	orderedTime
)

func orderKindOf(v interface{}) orderKind {
	if _, ok := asNumber(v); ok {
		return orderedNumber
	}
	if _, ok := v.(time.Time); ok {
		return orderedTime
	}
	if v != nil && reflect.ValueOf(v).Kind() == reflect.String {
		return orderedString
	}
	return unordered
}

// compareValues compares two values as Ruby's left.send(op, right) in a
// condition. ok is false when either side has no comparison methods, which
// makes the condition false, and when a float is NaN. Numbers, strings and
// times compare only among their kind; any other pair raises Ruby's
// ArgumentError.
func compareValues(left, right interface{}) (cmp int, ok bool, err error) {
	lk, rk := orderKindOf(left), orderKindOf(right)
	if lk == unordered || rk == unordered {
		return 0, false, nil
	}
	if lk != rk {
		return 0, false, NewArgumentError("comparison of " + rubyClassName(left) + " with " + comparisonOperand(right) + " failed")
	}
	switch lk {
	case orderedNumber:
		ln, _ := asNumber(left)
		rn, _ := asNumber(right)
		cmp, ok = compareNumbers(ln, rn)
		return cmp, ok, nil
	case orderedTime:
		return left.(time.Time).Compare(right.(time.Time)), true, nil
	}
	return strings.Compare(reflect.ValueOf(left).String(), reflect.ValueOf(right).String()), true, nil
}

// rubyClassName returns the name of the Ruby class of an ordered value.
func rubyClassName(v interface{}) string {
	switch orderKindOf(v) {
	case orderedNumber:
		if n, _ := asNumber(v); n.kind == floatNumber {
			return "Float"
		}
		return "Integer"
	case orderedTime:
		return "Time"
	}
	return "String"
}

// comparisonOperand writes the right side of a failed comparison as Ruby's
// rb_cmperr: a number by its value, anything else by its class.
func comparisonOperand(v interface{}) string {
	if orderKindOf(v) == orderedNumber {
		return ToS(v, nil)
	}
	return rubyClassName(v)
}

// containsOperator implements contains as Ruby's include?: a substring for a string
// (the right side converted to text), an element equal to the right side for a list,
// a key for a map.
func containsOperator(left, right interface{}) bool {
	if left == nil || right == nil {
		return false
	}

	rightStr := ToS(right, nil)

	if leftStr, ok := left.(string); ok {
		return strings.Contains(leftStr, rightStr)
	}

	v := reflect.ValueOf(left)
	switch v.Kind() {
	case reflect.String:
		return strings.Contains(v.String(), rightStr)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if liquidEqual(v.Index(i).Interface(), right) {
				return true
			}
		}
		return false
	case reflect.Map:
		if m, ok := left.(map[string]interface{}); ok {
			_, exists := m[rightStr]
			return exists
		}
		keyVal := reflect.ValueOf(rightStr)
		if keyVal.Type().AssignableTo(v.Type().Key()) && v.MapIndex(keyVal).IsValid() {
			return true
		}
		// Also try to find the key by iterating (in case key types don't match exactly)
		iter := v.MapRange()
		for iter.Next() {
			if ToS(iter.Key().Interface(), nil) == rightStr {
				return true
			}
		}
	}

	return false
}

// ParseConditionExpression parses an expression for use in conditions.
func ParseConditionExpression(parseContext ParseContextInterface, markup string, safe bool) interface{} {
	if ml, ok := conditionMethodLiterals[markup]; ok {
		return ml
	}

	// Check if markup contains filter syntax (|)
	// If so, parse as Variable to support filters in conditions
	if strings.Contains(markup, "|") {
		// Use Variable which supports filter syntax
		return NewVariable(markup, parseContext)
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
