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
