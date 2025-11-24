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
	// Test with string number (should fail as conditions don't implicitly convert strings)
	result, ok := toNumber("42")
	if ok {
		t.Errorf("Expected ok=false for string number, got ok=true, result=%v", result)
	}

	// Test with float string (should fail)
	result2, ok2 := toNumber("3.14")
	if ok2 {
		t.Errorf("Expected ok=false for float string, got ok=true, result=%v", result2)
	}

	// Test with actual number
	result3, ok3 := toNumber(42)
	if !ok3 || result3 != 42.0 {
		t.Errorf("Expected 42.0, got %v, ok=%v", result3, ok3)
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
	result, err := compareValues(1, 1)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result != 0 {
		t.Error("Expected 1 == 1 to return 0")
	}

	// Test less than
	result2, err := compareValues(1, 2)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result2 >= 0 {
		t.Error("Expected 1 < 2 to return negative")
	}

	// Test greater than
	result3, err := compareValues(2, 1)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
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

// TestConditionCheckMethodLiteral tests checkMethodLiteral with various inputs
func TestConditionCheckMethodLiteral(t *testing.T) {
	pc := NewParseContext(ParseContextOptions{})

	// Test blank method literal
	blankML := ParseConditionExpression(pc, "blank", false)
	if ml, ok := blankML.(*MethodLiteral); ok {
		// Test blank with empty string
		if !checkMethodLiteral(ml, "") {
			t.Error("Expected blank == \"\" to be true")
		}

		// Test blank with whitespace string
		if !checkMethodLiteral(ml, "   ") {
			t.Error("Expected blank == \"   \" to be true")
		}

		// Test blank with non-empty string
		if checkMethodLiteral(ml, "hello") {
			t.Error("Expected blank == \"hello\" to be false")
		}

		// Test blank with nil
		if !checkMethodLiteral(ml, nil) {
			t.Error("Expected blank == nil to be true")
		}

		// Test blank with non-string, non-nil
		if checkMethodLiteral(ml, 42) {
			t.Error("Expected blank == 42 to be false")
		}
	}

	// Test empty method literal
	emptyML := ParseConditionExpression(pc, "empty", false)
	if ml, ok := emptyML.(*MethodLiteral); ok {
		// Test empty with empty string
		if !checkMethodLiteral(ml, "") {
			t.Error("Expected empty == \"\" to be true")
		}

		// Test empty with non-empty string
		if checkMethodLiteral(ml, "hello") {
			t.Error("Expected empty == \"hello\" to be false")
		}

		// Test empty with empty array
		if !checkMethodLiteral(ml, []interface{}{}) {
			t.Error("Expected empty == [] to be true")
		}

		// Test empty with non-empty array
		if checkMethodLiteral(ml, []interface{}{1, 2}) {
			t.Error("Expected empty == [1,2] to be false")
		}

		// Test empty with empty map
		if !checkMethodLiteral(ml, map[string]interface{}{}) {
			t.Error("Expected empty == {} to be true")
		}

		// Test empty with non-empty map
		if checkMethodLiteral(ml, map[string]interface{}{"key": "value"}) {
			t.Error("Expected empty == {key:value} to be false")
		}

		// Test empty with nil
		if !checkMethodLiteral(ml, nil) {
			t.Error("Expected empty == nil to be true")
		}
	}

	// Test unknown method literal
	unknownML := &MethodLiteral{MethodName: "unknown"}
	if checkMethodLiteral(unknownML, "anything") {
		t.Error("Expected unknown method literal to return false")
	}
}

// TestConditionCompareValuesEdgeCases tests compareValues with various edge case inputs
func TestConditionCompareValuesEdgeCases(t *testing.T) {
	// Test numeric comparison
	result, err := compareValues(1, 2)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result >= 0 {
		t.Error("Expected 1 < 2 to return negative")
	}

	result2, err := compareValues(2, 1)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result2 <= 0 {
		t.Error("Expected 2 > 1 to return positive")
	}

	result3, err := compareValues(1, 1)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result3 != 0 {
		t.Error("Expected 1 == 1 to return 0")
	}

	// Test float comparison
	result4, err := compareValues(1.5, 2.5)
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result4 >= 0 {
		t.Error("Expected 1.5 < 2.5 to return negative")
	}

	// Test string comparison
	result5, err := compareValues("a", "b")
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result5 >= 0 {
		t.Error("Expected \"a\" < \"b\" to return negative")
	}

	result6, err := compareValues("b", "a")
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result6 <= 0 {
		t.Error("Expected \"b\" > \"a\" to return positive")
	}

	result7, err := compareValues("a", "a")
	if err != nil {
		t.Errorf("compareValues returned error: %v", err)
	}
	if result7 != 0 {
		t.Error("Expected \"a\" == \"a\" to return 0")
	}

	// Test mixed types (should return error now)
	_, err = compareValues(42, "hello")
	// Should return error for mixed type comparison
	if err == nil {
		t.Error("Expected mixed type comparison to return error")
	}
}

// TestConditionCompareValuesNilComparisons tests nil comparisons (matching Shopify Liquid behavior)
func TestConditionCompareValuesNilComparisons(t *testing.T) {
	// Test nil < number (nil is less than any number)
	result, err := compareValues(nil, 10)
	if err != nil {
		t.Errorf("compareValues(nil, 10) returned error: %v", err)
	}
	if result >= 0 {
		t.Errorf("Expected compareValues(nil, 10) to return negative, got %d", result)
	}

	// Test number > nil (should return negative to make condition false)
	result2, err := compareValues(10, nil)
	if err != nil {
		t.Errorf("compareValues(10, nil) returned error: %v", err)
	}
	if result2 >= 0 {
		t.Errorf("Expected compareValues(10, nil) to return negative, got %d", result2)
	}

	// Test nil == nil (both are equal)
	result3, err := compareValues(nil, nil)
	if err != nil {
		t.Errorf("compareValues(nil, nil) returned error: %v", err)
	}
	if result3 != 0 {
		t.Errorf("Expected compareValues(nil, nil) to return 0, got %d", result3)
	}

	// Test nil with float
	result4, err := compareValues(nil, 3.14)
	if err != nil {
		t.Errorf("compareValues(nil, 3.14) returned error: %v", err)
	}
	if result4 >= 0 {
		t.Errorf("Expected compareValues(nil, 3.14) to return negative, got %d", result4)
	}

	// Test float with nil
	result5, err := compareValues(3.14, nil)
	if err != nil {
		t.Errorf("compareValues(3.14, nil) returned error: %v", err)
	}
	if result5 >= 0 {
		t.Errorf("Expected compareValues(3.14, nil) to return negative, got %d", result5)
	}
}

// TestConditionToNumberEdgeCases tests toNumber with various edge case inputs
func TestConditionToNumberEdgeCases(t *testing.T) {
	// Test with int
	result, ok := toNumber(42)
	if !ok || result != 42.0 {
		t.Errorf("Expected 42.0, got %v, ok=%v", result, ok)
	}

	// Test with int64
	result2, ok2 := toNumber(int64(42))
	if !ok2 || result2 != 42.0 {
		t.Errorf("Expected 42.0, got %v, ok=%v", result2, ok2)
	}

	// Test with float64
	result3, ok3 := toNumber(3.14)
	if !ok3 || result3 != 3.14 {
		t.Errorf("Expected 3.14, got %v, ok=%v", result3, ok3)
	}

	// Test with string number (should fail)
	result4, ok4 := toNumber("42")
	if ok4 {
		t.Errorf("Expected ok=false for string number, got ok=true, result=%v", result4)
	}

	// Test with string float (should fail)
	result5, ok5 := toNumber("3.14")
	if ok5 {
		t.Errorf("Expected ok=false for string float, got ok=true, result=%v", result5)
	}

	// Test with invalid string
	result6, ok6 := toNumber("abc")
	if ok6 {
		t.Errorf("Expected ok=false for invalid string, got ok=true, result=%v", result6)
	}

	// Test with nil
	result7, ok7 := toNumber(nil)
	if ok7 {
		t.Errorf("Expected ok=false for nil, got ok=true, result=%v", result7)
	}

	// Test with bool
	result8, ok8 := toNumber(true)
	if ok8 {
		t.Errorf("Expected ok=false for bool, got ok=true, result=%v", result8)
	}
}

// TestContainsOperatorTypedSlices tests the contains operator with typed slices
func TestContainsOperatorTypedSlices(t *testing.T) {
	tests := []struct {
		name     string
		left     interface{}
		right    interface{}
		expected bool
	}{
		// Typed slices
		{
			name:     "[]string contains element",
			left:     []string{"apple", "banana", "cherry"},
			right:    "banana",
			expected: true,
		},
		{
			name:     "[]string does not contain element",
			left:     []string{"apple", "banana", "cherry"},
			right:    "grape",
			expected: false,
		},
		{
			name:     "[]int contains element",
			left:     []int{10, 20, 30},
			right:    "20",
			expected: true,
		},
		{
			name:     "[]int does not contain element",
			left:     []int{10, 20, 30},
			right:    "40",
			expected: false,
		},
		// Arrays
		{
			name:     "[3]string contains element",
			left:     [3]string{"red", "green", "blue"},
			right:    "green",
			expected: true,
		},
		// Substring matching for structs
		{
			name: "[]struct contains substring",
			left: []struct {
				Name string
			}{
				{Name: "First Item"},
				{Name: "Second Item"},
			},
			right:    "First",
			expected: true,
		},
		// String contains (existing functionality)
		{
			name:     "string contains substring",
			left:     "hello world",
			right:    "world",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsOperator(tt.left, tt.right)
			if result != tt.expected {
				t.Errorf("containsOperator(%v, %v) = %v, want %v", tt.left, tt.right, result, tt.expected)
			}
		})
	}
}

// TestCheckMethodLiteralTypedSlices tests the empty method literal with typed slices
func TestCheckMethodLiteralTypedSlices(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		obj      interface{}
		expected bool
	}{
		// empty on typed slices
		{
			name:     "[]string empty - true",
			method:   "empty",
			obj:      []string{},
			expected: true,
		},
		{
			name:     "[]string empty - false",
			method:   "empty",
			obj:      []string{"a", "b"},
			expected: false,
		},
		{
			name:     "[]int empty - true",
			method:   "empty",
			obj:      []int{},
			expected: true,
		},
		{
			name:     "[]int empty - false",
			method:   "empty",
			obj:      []int{1, 2, 3},
			expected: false,
		},
		// empty on arrays
		{
			name:     "[3]string not empty",
			method:   "empty",
			obj:      [3]string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "[0]int empty",
			method:   "empty",
			obj:      [0]int{},
			expected: true,
		},
		// empty on typed maps
		{
			name:     "map[string]string empty - true",
			method:   "empty",
			obj:      map[string]string{},
			expected: true,
		},
		{
			name:     "map[string]string empty - false",
			method:   "empty",
			obj:      map[string]string{"key": "value"},
			expected: false,
		},
		{
			name:     "map[string]int empty - true",
			method:   "empty",
			obj:      map[string]int{},
			expected: true,
		},
		// blank (existing functionality should still work)
		{
			name:     "blank - empty string",
			method:   "blank",
			obj:      "",
			expected: true,
		},
		{
			name:     "blank - whitespace",
			method:   "blank",
			obj:      "   ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &MethodLiteral{MethodName: tt.method}
			result := checkMethodLiteral(ml, tt.obj)
			if result != tt.expected {
				t.Errorf("checkMethodLiteral(%s, %v) = %v, want %v", tt.method, tt.obj, result, tt.expected)
			}
		})
	}
}
