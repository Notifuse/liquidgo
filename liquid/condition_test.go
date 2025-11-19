package liquid

import (
	"testing"
)

// MockConditionContext for testing
type mockConditionContext struct{}

func (m *mockConditionContext) Evaluate(expr interface{}) interface{} {
	// Simple mock - return the expression itself for testing
	return expr
}

func TestConditionBasic(t *testing.T) {
	c := NewCondition(1, "==", 1)
	if c == nil {
		t.Fatal("Expected Condition, got nil")
	}
	if c.Left() != 1 {
		t.Error("Left mismatch")
	}
	if c.Operator() != "==" {
		t.Error("Operator mismatch")
	}
	if c.Right() != 1 {
		t.Error("Right mismatch")
	}
}

func TestConditionEvaluate(t *testing.T) {
	ctx := &mockConditionContext{}

	tests := []struct {
		name     string
		left     interface{}
		operator string
		right    interface{}
		want     bool
	}{
		{"equal", 1, "==", 1, true},
		{"not equal", 1, "!=", 2, true},
		{"less than", 1, "<", 2, true},
		{"greater than", 2, ">", 1, true},
		{"less or equal", 1, "<=", 1, true},
		{"greater or equal", 2, ">=", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCondition(tt.left, tt.operator, tt.right)
			result, err := c.Evaluate(ctx)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if result != tt.want {
				t.Errorf("Evaluate() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestConditionOr(t *testing.T) {
	ctx := &mockConditionContext{}

	c1 := NewCondition(false, "", nil)
	c2 := NewCondition(true, "", nil)
	c1.Or(c2)

	result, err := c1.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result {
		t.Error("Expected OR condition to be true")
	}
}

func TestConditionAnd(t *testing.T) {
	ctx := &mockConditionContext{}

	c1 := NewCondition(true, "", nil)
	c2 := NewCondition(true, "", nil)
	c1.And(c2)

	result, err := c1.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result {
		t.Error("Expected AND condition to be true")
	}
}

func TestConditionContains(t *testing.T) {
	ctx := &mockConditionContext{}

	c := NewCondition("hello world", "contains", "world")
	result, err := c.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result {
		t.Error("Expected contains to return true")
	}
}

func TestElseCondition(t *testing.T) {
	ctx := &mockConditionContext{}

	ec := NewElseCondition()
	if !ec.Else() {
		t.Error("Expected Else() to return true")
	}

	result, err := ec.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result {
		t.Error("Expected ElseCondition.Evaluate() to return true")
	}
}

// TestConditionToNumber tests number conversion in conditions
func TestConditionToNumber(t *testing.T) {
	// Test with string number
	result, ok := toNumber("42")
	if !ok || result != 42 {
		t.Errorf("Expected 42, got %v, ok=%v", result, ok)
	}

	// Test with float string
	result2, ok2 := toNumber("3.14")
	if !ok2 || result2 != 3.14 {
		t.Errorf("Expected 3.14, got %v, ok=%v", result2, ok2)
	}

	// Test with actual number
	result3, ok3 := toNumber(42)
	if !ok3 || result3 != 42 {
		t.Errorf("Expected 42, got %v, ok=%v", result3, ok3)
	}
}

// TestConditionContainsOperator tests contains operator evaluation
func TestConditionContainsOperator(t *testing.T) {
	// Test string contains
	result := containsOperator("hello world", "world")
	if !result {
		t.Error("Expected contains to return true")
	}

	// Test array contains
	result2 := containsOperator([]interface{}{1, 2, 3}, 2)
	if !result2 {
		t.Error("Expected array contains to return true")
	}

	// Test map contains key
	result3 := containsOperator(map[string]interface{}{"key": "value"}, "key")
	if !result3 {
		t.Error("Expected map contains key to return true")
	}
}

// TestConditionCompareValues tests value comparison logic
func TestConditionCompareValues(t *testing.T) {
	// Test equal (compareValues returns int: -1, 0, 1)
	result := compareValues(1, 1)
	if result != 0 {
		t.Error("Expected 1 == 1 to return 0")
	}

	// Test less than
	result2 := compareValues(1, 2)
	if result2 >= 0 {
		t.Error("Expected 1 < 2 to return negative")
	}

	// Test greater than
	result3 := compareValues(2, 1)
	if result3 <= 0 {
		t.Error("Expected 2 > 1 to return positive")
	}
}

// TestConditionChildCondition tests ChildCondition method
func TestConditionChildCondition(t *testing.T) {
	c1 := NewCondition(1, "==", 1)
	c2 := NewCondition(2, "==", 2)

	c1.Or(c2)
	child := c1.ChildCondition()
	if child != c2 {
		t.Error("Expected ChildCondition to return the child condition")
	}
}

// TestConditionAttachment tests Attachment and Attach methods
func TestConditionAttachment(t *testing.T) {
	c := NewCondition(1, "==", 1)
	attachment := "test attachment"

	c.Attach(attachment)
	attached := c.Attachment()
	if attached != attachment {
		t.Errorf("Expected attachment %v, got %v", attachment, attached)
	}
}

// TestConditionElse tests Else method
func TestConditionElse(t *testing.T) {
	c := NewCondition(1, "==", 1)
	if c.Else() {
		t.Error("Expected Else() to return false for regular condition")
	}

	ec := NewElseCondition()
	if !ec.Else() {
		t.Error("Expected Else() to return true for ElseCondition")
	}
}

// TestParseConditionExpression tests ParseConditionExpression
func TestParseConditionExpression(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})

	// Test with method literal
	result := ParseConditionExpression(pc, "blank", false)
	if result == nil {
		t.Error("Expected non-nil result for method literal")
	}

	// Test with regular expression
	result2 := ParseConditionExpression(pc, "var", false)
	if result2 == nil {
		t.Error("Expected non-nil result for variable expression")
	}

	// Test with safe parsing
	result3 := ParseConditionExpression(pc, "var", true)
	if result3 == nil {
		t.Error("Expected non-nil result for safe parsing")
	}
}

// TestConditionMethodLiteral tests checkMethodLiteral indirectly through condition evaluation
func TestConditionMethodLiteral(t *testing.T) {
	ctx := NewContext()

	// Test blank method literal - blank == ""
	pc := NewParseContext(ParseContextOptions{})
	blankML := ParseConditionExpression(pc, "blank", false)
	c1 := NewCondition(blankML, "==", "")
	result1, err := c1.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result1 {
		t.Error("Expected blank == \"\" to evaluate to true")
	}

	// Test empty method literal with array - empty == []
	emptyML := ParseConditionExpression(pc, "empty", false)
	c2 := NewCondition(emptyML, "==", []interface{}{})
	result2, err := c2.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result2 {
		t.Error("Expected empty == [] to evaluate to true")
	}

	// Test empty method literal with map - empty == {}
	c3 := NewCondition(emptyML, "==", map[string]interface{}{})
	result3, err := c3.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !result3 {
		t.Error("Expected empty == {} to evaluate to true")
	}
}
