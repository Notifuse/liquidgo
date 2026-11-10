package liquid

import (
	"math"
	"reflect"
	"testing"
)

// Arithmetic follows Ruby: integers stay integers (floor division and modulo), and any
// decimal makes the computation exact (BigDecimal in Ruby), returned as a float64.

type arithCase struct {
	name string
	got  interface{}
	want interface{}
}

func checkArith(t *testing.T, cases []arithCase) {
	t.Helper()
	for _, c := range cases {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s = %#v (%T), want %#v (%T)", c.name, c.got, c.got, c.want, c.want)
		}
	}
}

func TestArithmeticExactDecimals(t *testing.T) {
	sf := &StandardFilters{}
	div := func(a, b interface{}) interface{} { r, _ := sf.DividedBy(a, b); return r }
	mod := func(a, b interface{}) interface{} { r, _ := sf.Modulo(a, b); return r }
	checkArith(t, []arithCase{
		{"0.1 + 0.2", sf.Plus(0.1, 0.2), 0.3},
		{"0.3 - 0.1", sf.Minus(0.3, 0.1), 0.2},
		{"0.1 * 3", sf.Times(0.1, 3), 0.3},
		{"4.35 * 100", sf.Times(4.35, 100), 435.0},
		{"0.0725 * 100", sf.Times(0.0725, 100), 7.25},
		{"19.99 * 3", sf.Times(19.99, 3), 59.97},
		{"0.3 / 0.1", div(0.3, 0.1), 3.0},
		{"1 / 3.0", div(1, 3.0), 1.0 / 3.0},
		{"5.5 % 2", mod(5.5, 2), 1.5},
		{"-5.5 % 2", mod(-5.5, 2), 0.5},
		{`"0.1" + "0.2"`, sf.Plus("0.1", "0.2"), 0.3},
		{`"-0.0725" * -100`, sf.Times("-0.0725", -100), 7.25},
		{"float32", sf.Plus(float32(0.1), 0.2), 0.3},
	})
}

func TestArithmeticIntegers(t *testing.T) {
	sf := &StandardFilters{}
	div := func(a, b interface{}) interface{} { r, _ := sf.DividedBy(a, b); return r }
	mod := func(a, b interface{}) interface{} { r, _ := sf.Modulo(a, b); return r }
	checkArith(t, []arithCase{
		{"1 + 1", sf.Plus(1, 1), 2},
		{"5 - 1", sf.Minus(5, 1), 4},
		{"3 * 4", sf.Times(3, 4), 12},
		{"14 / 3", div(14, 3), 4},
		{"-7 / 2", div(-7, 2), -4},
		{"7 / -2", div(7, -2), -4},
		{"-7 % 3", mod(-7, 3), 2},
		{"7 % -3", mod(7, -3), -2},
		{"3 % 2", mod(3, 2), 1},
		{"int32 + int", sf.Plus(int32(2), 3), 5},
		{`"6-3" + 0`, sf.Plus("6-3", 0), 6},
		{`"abc" + 1`, sf.Plus("abc", 1), 1},
		{`" 12 " + 1`, sf.Plus(" 12 ", 1), 13},
		{"nil + 1", sf.Plus(nil, 1), 1},
		{"JSON integer + int", sf.Plus(100.0, 1), 101.0},
		{"JSON integer / int", div(1999.0, 100), 19.99},
		{"overflow", sf.Plus(math.MaxInt64, 1), float64(math.MaxInt64) + 1},
	})
}

func TestArithmeticDivisionByZero(t *testing.T) {
	sf := &StandardFilters{}
	for _, operand := range []interface{}{0, 0.0, "0", nil} {
		if _, err := sf.DividedBy(1, operand); err == nil {
			t.Errorf("DividedBy(1, %#v): want a division by zero error", operand)
		}
		if _, err := sf.Modulo(1, operand); err == nil {
			t.Errorf("Modulo(1, %#v): want a division by zero error", operand)
		}
	}
}

func TestRoundCeilFloorAbsExact(t *testing.T) {
	sf := &StandardFilters{}
	round := func(a, n interface{}) interface{} { r, _ := sf.Round(a, n); return r }
	ceil := func(a interface{}) interface{} { r, _ := sf.Ceil(a); return r }
	floor := func(a interface{}) interface{} { r, _ := sf.Floor(a); return r }
	checkArith(t, []arithCase{
		{"1.005 round 2", round(1.005, 2), 1.01},
		{"2.675 round 2", round(2.675, 2), 2.68},
		{"4.5 round", round(4.5, nil), 5},
		{"-4.5 round", round(-4.5, nil), -5},
		{"4.6 round", round(4.6, nil), 5},
		{`"4.3" round`, round("4.3", nil), 4},
		{"4.5612 round 2", round(4.5612, 2), 4.56},
		{"int round 2", round(5, 2), 5},
		{"int round -1", round(15, -1), 20},
		{"1234.5 round -2", round(1234.5, -2), 1200.0},
		{"ceil 1.2", ceil(1.2), 1 + 1},
		{"ceil -1.2", ceil(-1.2), -1},
		{"floor 1.8", floor(1.8), 1},
		{"floor -1.2", floor(-1.2), -2},
		{"ceil string", ceil("3.1"), 4},
		{"abs -1.5", sf.Abs(-1.5), 1.5},
		{"abs -3", sf.Abs(-3), 3},
		{"abs string", sf.Abs("-0.1"), 0.1},
		{"at_least", sf.AtLeast(4, 5), 5},
		{"at_least decimal", sf.AtLeast(4.5, 4), 4.5},
		{"at_most", sf.AtMost(4, "3.5"), 3.5},
	})
	for _, v := range []interface{}{math.Inf(1), math.NaN()} {
		if _, err := sf.Round(v, nil); err == nil {
			t.Errorf("Round(%v): want a float domain error", v)
		}
	}
}

func TestSumExact(t *testing.T) {
	sf := &StandardFilters{}
	checkArith(t, []arithCase{
		{"ints", sf.Sum([]interface{}{1, 2, 3}, nil), 6},
		{"decimals", sf.Sum([]interface{}{0.1, 0.2}, nil), 0.3},
		{"strings", sf.Sum([]interface{}{"0.1", "0.2", 1}, nil), 1.3},
		{"property", sf.Sum([]interface{}{map[string]interface{}{"p": 19.99}, map[string]interface{}{"p": 0.01}}, "p"), 20.0},
		{"empty", sf.Sum([]interface{}{}, nil), 0},
	})
}
