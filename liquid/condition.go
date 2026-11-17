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
	// operate is the operator, found once when the condition is made (nil
	// for an unknown one).
	operate       ConditionOperator
	childRelation string
}

// NewCondition creates a new Condition.
func NewCondition(left interface{}, operator string, right interface{}) *Condition {
	return &Condition{
		left:     left,
		operator: operator,
		operate:  getConditionOperator(operator),
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
		result, err = c.interpretCondition(condition.left, condition.right, condition.operator, condition.operate, context)
		if err != nil {
			return false, err
		}

		shouldContinue := false
		switch condition.childRelation {
		case "or":
			shouldContinue = !result
		case "and":
			shouldContinue = result
		}

		if !shouldContinue || condition.childCondition == nil {
			break
		}
		condition = condition.childCondition
	}

	return result, nil
}

// truthy reports whether a condition holds for value: nil, false and an
// empty string do not (an empty string is a deliberate divergence).
func truthy(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case string:
		return v != ""
	}
	return true
}

func (c *Condition) interpretCondition(left, right interface{}, op string, operator ConditionOperator, context ConditionContext) (bool, error) {
	// If operator is empty, just evaluate the left side
	if op == "" {
		return truthy(ToLiquidValue(context.Evaluate(left))), nil
	}

	leftVal := ToLiquidValue(context.Evaluate(left))
	rightVal := ToLiquidValue(context.Evaluate(right))

	if operator == nil {
		return false, NewArgumentError("Unknown operator " + op)
	}

	return operator(c, leftVal, rightVal)
}

// conditionOperators are the operators of a condition, made once.
var conditionOperators = map[string]ConditionOperator{
	"==": func(cond *Condition, left, right interface{}) (bool, error) {
		return cond.equalVariables(left, right), nil
	},
	"!=": notEqualOperator,
	"<>": notEqualOperator,
	"<":  orderOperator(func(cmp int) bool { return cmp < 0 }),
	">":  orderOperator(func(cmp int) bool { return cmp > 0 }),
	">=": orderOperator(func(cmp int) bool { return cmp >= 0 }),
	"<=": orderOperator(func(cmp int) bool { return cmp <= 0 }),
	"contains": func(_ *Condition, left, right interface{}) (bool, error) {
		return containsOperator(left, right), nil
	},
}

func notEqualOperator(cond *Condition, left, right interface{}) (bool, error) {
	return !cond.equalVariables(left, right), nil
}

func getConditionOperator(op string) ConditionOperator {
	return conditionOperators[op]
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
		return liquidBlank(obj)
	case "empty":
		return liquidEmpty(obj)
	default:
		return false
	}
}

// liquidBlank is Ruby's liquid_blank? (Shopify Liquid 5.14), ActiveSupport's
// blank?: nil, a string of whitespace only, and an empty list or hash are
// blank; true and numbers are not. Two differences are deliberate, for the
// display conditions of Notifuse: false is a value, not blank, and the
// whitespace of a string is Unicode's (a no-break space is blank), where
// Ruby makes false blank and counts ASCII whitespace only.
func liquidBlank(obj interface{}) bool {
	switch v := obj.(type) {
	case nil:
		return true
	case bool:
		return false
	case string:
		return strings.TrimSpace(v) == ""
	}
	if _, ok := asNumber(obj); ok {
		return false
	}
	return liquidEmpty(obj)
}

// liquidEmpty is Ruby's liquid_empty?: a string, a list or a hash with nothing
// in it; nil is not empty.
func liquidEmpty(obj interface{}) bool {
	switch v := obj.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	case nil:
		return false
	}
	// typed lists and hashes, which Ruby has as Array and Hash
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	}
	return false
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
	if isBigInt(v) {
		return orderedNumber
	}
	if _, ok := timeOf(v); ok {
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
		ln, lok := asNumber(left)
		rn, rok := asNumber(right)
		if !lok || !rok { // a *big.Int
			return compareBigInts(left, right), true, nil
		}
		cmp, ok = compareNumbers(ln, rn)
		return cmp, ok, nil
	case orderedTime:
		lt, _ := timeOf(left)
		rt, _ := timeOf(right)
		return lt.Compare(rt), true, nil
	}
	return strings.Compare(reflect.ValueOf(left).String(), reflect.ValueOf(right).String()), true, nil
}

// timeOf returns the time of a time.Time or a non-nil *time.Time.
func timeOf(v interface{}) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case *time.Time:
		if t != nil {
			return *t, true
		}
	}
	return time.Time{}, false
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

// MethodLiteralOf returns the method literal markup names (blank, empty), or
// nil: a condition compares a value with it by its method, as Ruby's
// Condition.parse_expression.
func MethodLiteralOf(markup string) *MethodLiteral {
	return conditionMethodLiterals[markup]
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
