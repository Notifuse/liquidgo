package liquid

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

type testStatus int
type testRole string
type testFlag bool

type equalityCase struct {
	name  string
	left  interface{}
	right interface{}
	want  bool
}

// evalEquality evaluates left == right and left != right through Condition, in both
// argument orders, and checks that != is the exact negation of ==.
func evalEquality(t *testing.T, tt equalityCase) {
	t.Helper()
	ctx := &mockConditionContext{}
	for _, order := range [][2]interface{}{{tt.left, tt.right}, {tt.right, tt.left}} {
		eq, err := NewCondition(order[0], "==", order[1]).Evaluate(ctx)
		if err != nil {
			t.Fatalf("%s: == error %v", tt.name, err)
		}
		ne, err := NewCondition(order[0], "!=", order[1]).Evaluate(ctx)
		if err != nil {
			t.Fatalf("%s: != error %v", tt.name, err)
		}
		if eq != tt.want {
			t.Errorf("%s: %#v == %#v is %v, want %v", tt.name, order[0], order[1], eq, tt.want)
		}
		if ne == eq {
			t.Errorf("%s: != is not the negation of == (%v, %v)", tt.name, eq, ne)
		}
	}
}

func TestEqualVariablesNumericTypes(t *testing.T) {
	for _, tt := range []equalityCase{
		{"int and float64", 100, 100.0, true},
		{"int32 and int", int32(100), 100, true},
		{"uint8 and float64", uint8(1), 1.0, true},
		{"float32 and int", float32(1.5), 1.5, true},
		{"different values", 100, 100.5, false},
		{"int64 exact above 2^53", int64(1<<53 + 1), int64(1 << 53), false},
		{"uint64 exact above 2^53", uint64(1<<53 + 1), int64(1<<53 + 1), true},
		{"int64 above 2^53 against float64", int64(1<<53 + 1), float64(1 << 53), false},
		{"negative int and max uint64", int64(-1), uint64(math.MaxUint64), false},
		{"string is not a number", "100", 100, false},
		{"bool is not a number", true, 1, false},
		{"nil and nil", nil, nil, true},
		{"nil and zero", nil, 0, false},
		{"NaN", math.NaN(), math.NaN(), false},
		{"infinity", math.Inf(1), math.Inf(1), true},
		{"named int", testStatus(1), 1, true},
		{"named string", testRole("admin"), "admin", true},
		{"named bool", testFlag(true), true, true},
		{"json.Number", json.Number("100"), 100, true},
		{"json.Number float", json.Number("1.5"), 1.5, true},
		{"strings", "a", "a", true},
		{"different strings", "a", "b", false},
		{"time values", time.Unix(0, 0).UTC(), time.Unix(0, 0).UTC(), true},
	} {
		evalEquality(t, tt)
	}
}

func TestEqualVariablesCollections(t *testing.T) {
	type holder struct{ V interface{} }
	for _, tt := range []equalityCase{
		{"lists of numbers", []interface{}{1}, []interface{}{1.0}, true},
		{"order matters", []interface{}{1, 2}, []interface{}{2, 1}, false},
		{"empty lists", []interface{}{}, []interface{}{}, true},
		{"lengths differ", []interface{}{1}, []interface{}{1, 1}, false},
		{"nested lists", []interface{}{[]interface{}{"a"}}, []interface{}{[]interface{}{"a"}}, true},
		{"typed and untyped slices", []string{"a"}, []interface{}{"a"}, true},
		{"array and slice", [2]int{1, 2}, []interface{}{1.0, 2.0}, true},
		{"maps", map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1.0}, true},
		{"map with extra key", map[string]interface{}{"a": 1}, map[string]interface{}{"a": 1, "b": 2}, false},
		{"same size, other keys", map[string]interface{}{"a": nil}, map[string]interface{}{"b": nil}, false},
		{"typed map", map[string]int{"a": 1}, map[string]interface{}{"a": 1.0}, true},
		{"map key types differ", map[int]int{1: 1}, map[string]int{"1": 1}, false},
		{"list and map", []interface{}{}, map[string]interface{}{}, false},
		{"list and string", []interface{}{"a"}, "a", false},
		{"struct holding a slice", holder{V: []int{1}}, holder{V: []int{1}}, false},
	} {
		evalEquality(t, tt)
	}
}

func TestEqualVariablesSelfReference(t *testing.T) {
	a := map[string]interface{}{}
	a["self"] = a
	b := map[string]interface{}{}
	b["self"] = b
	done := make(chan bool, 1)
	go func() { done <- liquidEqual(a, b) }()
	select {
	case got := <-done:
		if got {
			t.Errorf("self-referencing maps: got true, want false at the depth limit")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("liquidEqual did not return on self-referencing maps")
	}
}

func TestCompareValuesNumericTypes(t *testing.T) {
	ctx := &mockConditionContext{}
	for _, tt := range []struct {
		left  interface{}
		op    string
		right interface{}
		want  bool
	}{
		{int32(5), ">", 3, true},
		{uint(5), ">=", 5.0, true},
		{float32(1.5), "<", 2, true},
		{int64(1<<53 + 1), ">", int64(1 << 53), true},
		{int64(-1), "<", uint64(math.MaxUint64), true},
		{json.Number("10"), ">", 9, true},
		{testStatus(2), "<=", 1, false},
	} {
		got, err := NewCondition(tt.left, tt.op, tt.right).Evaluate(ctx)
		if err != nil {
			t.Errorf("%#v %s %#v: error %v", tt.left, tt.op, tt.right, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%#v %s %#v = %v, want %v", tt.left, tt.op, tt.right, got, tt.want)
		}
	}
	if _, err := NewCondition("a", ">", 1).Evaluate(ctx); err == nil {
		t.Error(`"a" > 1: want a comparison error, as in Ruby`)
	}
}

func TestContainsOperatorElementEquality(t *testing.T) {
	for _, tt := range []struct {
		name  string
		left  interface{}
		right interface{}
		want  bool
	}{
		{"element equal", []interface{}{"vip"}, "vip", true},
		{"no substring of an element", []interface{}{"vip_gold"}, "vip", false},
		{"number against float", []interface{}{1, 2}, 1.0, true},
		{"number against its text", []interface{}{1, 2}, "1", false},
		{"typed slice", []string{"a"}, "a", true},
		{"typed slice no substring", []string{"apple"}, "app", false},
		{"list element", []interface{}{[]interface{}{"a"}}, []interface{}{"a"}, true},
		{"string substring", "hello world", "world", true},
		{"string with number", "a1", 1, true},
		{"map key", map[string]interface{}{"a": 1}, "a", true},
		{"map missing key", map[string]interface{}{"a": 1}, "b", false},
		{"nil left", nil, "a", false},
	} {
		if got := containsOperator(tt.left, tt.right); got != tt.want {
			t.Errorf("%s: %#v contains %#v = %v, want %v", tt.name, tt.left, tt.right, got, tt.want)
		}
	}
}

func TestValuesEqualCollections(t *testing.T) {
	if !valuesEqual([]interface{}{"a"}, []interface{}{"a"}) {
		t.Error("equal lists should match")
	}
	if valuesEqual([]interface{}{"a"}, []interface{}{"b"}) {
		t.Error("different lists should not match")
	}
	if !valuesEqual(10.0, 10) {
		t.Error("10.0 should match 10")
	}
	if !valuesEqual("10", 10) {
		t.Error(`"10" should still match 10 (string fallback)`)
	}
}
