package liquid

import (
	"math"
	"math/big"
)

// Abs returns the absolute value of a number.
// Mirrors Ruby's abs from standardfilters.rb:792
func (sf *StandardFilters) Abs(input interface{}) interface{} {
	n := toLiquidNumber(input)
	switch n.kind {
	case integerNumber:
		if n.i < 0 {
			return arithmetic(integerOf(0), n, '-').value()
		}
		return n.value()
	case bigIntegerNumber:
		return decimalOfInteger(new(big.Int).Abs(n.bigInt())).value()
	case decimalNumber:
		return decimalOf(new(big.Rat).Abs(n.r)).value()
	}
	return math.Abs(n.f)
}

// Plus adds two numbers.
// Mirrors Ruby's plus from standardfilters.rb:804
func (sf *StandardFilters) Plus(input interface{}, operand interface{}) interface{} {
	return arithmetic(toLiquidNumber(input), toLiquidNumber(operand), '+').value()
}

// Minus subtracts two numbers.
// Mirrors Ruby's minus from standardfilters.rb:815
func (sf *StandardFilters) Minus(input interface{}, operand interface{}) interface{} {
	return arithmetic(toLiquidNumber(input), toLiquidNumber(operand), '-').value()
}

// Times multiplies two numbers.
// Mirrors Ruby's times from standardfilters.rb:826
func (sf *StandardFilters) Times(input interface{}, operand interface{}) interface{} {
	return arithmetic(toLiquidNumber(input), toLiquidNumber(operand), '*').value()
}

// DividedBy divides two numbers: two integers give the floored integer quotient.
// Mirrors Ruby's divided_by from standardfilters.rb:837
func (sf *StandardFilters) DividedBy(input interface{}, operand interface{}) (interface{}, error) {
	divisor := toLiquidNumber(operand)
	dividend := toLiquidNumber(input)
	if divisor.isZero() {
		// as Ruby: only an integer by an integer raises; a decimal side
		// gives Infinity, -Infinity, or NaN for 0
		if dividend.isInteger() && divisor.isInteger() {
			return nil, NewZeroDivisionError("divided by 0")
		}
		switch x := dividend.float64(); {
		case x > 0:
			return math.Inf(1), nil
		case x < 0:
			return math.Inf(-1), nil
		default:
			return math.NaN(), nil
		}
	}
	return arithmetic(dividend, divisor, '/').value(), nil
}

// Modulo returns the remainder of division, with the sign of the divisor (Ruby %).
// Mirrors Ruby's modulo from standardfilters.rb:850
func (sf *StandardFilters) Modulo(input interface{}, operand interface{}) (interface{}, error) {
	divisor := toLiquidNumber(operand)
	if divisor.isZero() {
		return nil, NewZeroDivisionError("divided by 0")
	}
	return arithmetic(toLiquidNumber(input), divisor, '%').value(), nil
}

// Round rounds a number to the nearest integer or to n decimal places, half away from
// zero on the exact decimal value (1.005 | round: 2 is 1.01).
// Mirrors Ruby's round from standardfilters.rb:863
func (sf *StandardFilters) Round(input interface{}, n interface{}) (interface{}, error) {
	digits := 0
	if n != nil {
		d := toLiquidNumber(n)
		if d.kind == specialNumber {
			return nil, NewFloatDomainError(ToS(d.f, nil))
		}
		digits = clampRoundDigits(floorRat(d.rat()))
	}

	num := toLiquidNumber(input)
	if num.kind == specialNumber {
		return nil, NewFloatDomainError(ToS(num.f, nil))
	}
	if num.isInteger() && digits >= 0 {
		return num.value(), nil
	}
	// beyond a float64's range of digits the result is known without computing 10^digits
	if digits > maxRoundDigits {
		return num.value(), nil
	}
	if digits < -maxRoundDigits {
		if num.isInteger() {
			return 0, nil
		}
		return 0.0, nil
	}
	rounded := decimalOf(roundRat(num.rat(), digits))
	if digits == 0 || num.isInteger() {
		return decimalOfInteger(floorRat(rounded.r)).value(), nil
	}
	return rounded.value(), nil
}

// maxRoundDigits bounds the precision of round: a float64 has no digit beyond 10^±340,
// so a larger precision keeps the number and a smaller one gives 0.
const maxRoundDigits = 400

func clampRoundDigits(d *big.Int) int {
	switch {
	case d.Cmp(big.NewInt(maxRoundDigits)) > 0:
		return maxRoundDigits + 1
	case d.Cmp(big.NewInt(-maxRoundDigits)) < 0:
		return -maxRoundDigits - 1
	}
	return int(d.Int64())
}

// Ceil rounds a number up to the nearest integer.
// Mirrors Ruby's ceil from standardfilters.rb:879
func (sf *StandardFilters) Ceil(input interface{}) (interface{}, error) {
	num := toLiquidNumber(input)
	if num.kind == specialNumber {
		return 0, NewFloatDomainError(ToS(num.f, nil))
	}
	return decimalOfInteger(ceilRat(num.rat())).value(), nil
}

// Floor rounds a number down to the nearest integer.
// Mirrors Ruby's floor from standardfilters.rb:892
func (sf *StandardFilters) Floor(input interface{}) (interface{}, error) {
	num := toLiquidNumber(input)
	if num.kind == specialNumber {
		return 0, NewFloatDomainError(ToS(num.f, nil))
	}
	return decimalOfInteger(floorRat(num.rat())).value(), nil
}

// AtLeast limits a number to a minimum value.
// Mirrors Ruby's at_least from standardfilters.rb:905
func (sf *StandardFilters) AtLeast(input interface{}, n interface{}) interface{} {
	minValue, result := toLiquidNumber(n), toLiquidNumber(input)
	if compareLiquidNumbers(minValue, result) > 0 {
		return minValue.value()
	}
	return result.value()
}

// AtMost limits a number to a maximum value.
// Mirrors Ruby's at_most from standardfilters.rb:920
func (sf *StandardFilters) AtMost(input interface{}, n interface{}) interface{} {
	maxValue, result := toLiquidNumber(n), toLiquidNumber(input)
	if compareLiquidNumbers(maxValue, result) < 0 {
		return maxValue.value()
	}
	return result.value()
}
