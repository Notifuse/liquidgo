package liquid

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Arithmetic filters follow Ruby Liquid's Utils.to_number and apply_operation:
// integers stay integers, with floor division and floor modulo; a float or a decimal
// string is converted to an exact decimal (BigDecimal in Ruby, big.Rat here), and a
// result that involved a decimal is returned as a float64. So 0.1 | plus: 0.2 is 0.3,
// not 0.30000000000000004.

// liquidNumber is an integer, an exact decimal, or a float that has no exact value
// (NaN, ±Inf).
type liquidNumber struct {
	kind liquidNumberKind
	i    int64
	r    *big.Rat
	f    float64
}

type liquidNumberKind int

const (
	integerNumber liquidNumberKind = iota
	decimalNumber
	specialNumber
)

var rubyToIPrefix = regexp.MustCompile(`^[+-]?\d+(?:_\d+)*`)

func integerOf(i int64) liquidNumber { return liquidNumber{kind: integerNumber, i: i} }

func decimalOf(r *big.Rat) liquidNumber { return liquidNumber{kind: decimalNumber, r: r} }

// decimalOfInteger keeps an integer too large for int64 exact.
func decimalOfInteger(i *big.Int) liquidNumber {
	if i.IsInt64() {
		return integerOf(i.Int64())
	}
	return decimalOf(new(big.Rat).SetInt(i))
}

// decimalOfFloat converts a float as Ruby's BigDecimal(float.to_s): through its shortest
// decimal representation, so 0.1 is exactly 0.1.
func decimalOfFloat(f float64, bitSize int) liquidNumber {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return liquidNumber{kind: specialNumber, f: f}
	}
	r, ok := new(big.Rat).SetString(strconv.FormatFloat(f, 'g', -1, bitSize))
	if !ok {
		return liquidNumber{kind: specialNumber, f: f}
	}
	return decimalOf(r)
}

// toLiquidNumber is Ruby's Utils.to_number.
func toLiquidNumber(v interface{}) liquidNumber {
	switch x := v.(type) {
	case nil, bool:
		return integerOf(0)
	case float32:
		return decimalOfFloat(float64(x), 32)
	case float64:
		return decimalOfFloat(x, 64)
	case string:
		return stringNumber(x)
	case interface{ ToNumber() interface{} }:
		return toLiquidNumber(x.ToNumber())
	}
	if n, ok := asNumber(v); ok {
		switch n.kind {
		case signedNumber:
			return integerOf(n.i)
		case unsignedNumber:
			return decimalOfInteger(new(big.Int).SetUint64(n.u))
		default:
			return decimalOfFloat(n.f, 64)
		}
	}
	return integerOf(0)
}

// stringNumber parses a string as Ruby does: a decimal ("-1.5") exactly, anything else
// with to_i (leading integer, 0 when there is none: "6-3" is 6).
func stringNumber(s string) liquidNumber {
	s = strings.Trim(s, " \t\n\v\f\r\x00")
	if DecimalRegex.MatchString(s) {
		if r, ok := new(big.Rat).SetString(s); ok {
			return decimalOf(r)
		}
	}
	digits := strings.ReplaceAll(rubyToIPrefix.FindString(s), "_", "")
	if digits == "" {
		return integerOf(0)
	}
	i, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return integerOf(0)
	}
	return decimalOfInteger(i)
}

func (n liquidNumber) rat() *big.Rat {
	if n.kind == integerNumber {
		return new(big.Rat).SetInt64(n.i)
	}
	return n.r
}

func (n liquidNumber) float64() float64 {
	switch n.kind {
	case integerNumber:
		return float64(n.i)
	case decimalNumber:
		f, _ := n.r.Float64()
		return f
	}
	return n.f
}

func (n liquidNumber) isZero() bool {
	switch n.kind {
	case integerNumber:
		return n.i == 0
	case decimalNumber:
		return n.r.Sign() == 0
	}
	return n.f == 0
}

// value returns the filter result: an int for an integer, a float64 otherwise (Ruby's
// BigDecimal#to_f). A decimal that came from an integer too large for int64 is a float.
func (n liquidNumber) value() interface{} {
	if n.kind == integerNumber {
		return int(n.i)
	}
	return n.float64()
}

// arithmetic applies +, -, *, / or % (Ruby's apply_operation). Division and modulo by
// zero must be checked by the caller.
func arithmetic(a, b liquidNumber, op byte) liquidNumber {
	if a.kind == specialNumber || b.kind == specialNumber {
		x, y := a.float64(), b.float64()
		switch op {
		case '+':
			return decimalOfFloat(x+y, 64)
		case '-':
			return decimalOfFloat(x-y, 64)
		case '*':
			return decimalOfFloat(x*y, 64)
		case '/':
			return decimalOfFloat(x/y, 64)
		default:
			return decimalOfFloat(x-y*math.Floor(x/y), 64)
		}
	}
	if a.kind == integerNumber && b.kind == integerNumber {
		x, y := big.NewInt(a.i), big.NewInt(b.i)
		switch op {
		case '+':
			return decimalOfInteger(x.Add(x, y))
		case '-':
			return decimalOfInteger(x.Sub(x, y))
		case '*':
			return decimalOfInteger(x.Mul(x, y))
		case '/':
			q, m := new(big.Int).DivMod(x, y, new(big.Int))
			// DivMod is Euclidean (m >= 0); Ruby floors toward negative infinity, which
			// differs only for a negative divisor and a remainder
			if m.Sign() != 0 && y.Sign() < 0 {
				q.Sub(q, big.NewInt(1))
			}
			return decimalOfInteger(q)
		default:
			m := new(big.Int).Mod(x, y) // 0 <= m < |y|
			if m.Sign() != 0 && y.Sign() < 0 {
				m.Add(m, y)
			}
			return decimalOfInteger(m)
		}
	}
	x, y := a.rat(), b.rat()
	switch op {
	case '+':
		return decimalOf(new(big.Rat).Add(x, y))
	case '-':
		return decimalOf(new(big.Rat).Sub(x, y))
	case '*':
		return decimalOf(new(big.Rat).Mul(x, y))
	case '/':
		return decimalOf(new(big.Rat).Quo(x, y))
	default:
		// floor modulo: x - y * floor(x / y)
		q := floorRat(new(big.Rat).Quo(x, y))
		return decimalOf(new(big.Rat).Sub(x, new(big.Rat).Mul(y, new(big.Rat).SetInt(q))))
	}
}

func floorRat(r *big.Rat) *big.Int {
	// Denom is positive, so Euclidean division is floor division
	return new(big.Int).Div(r.Num(), r.Denom())
}

func ceilRat(r *big.Rat) *big.Int {
	q := floorRat(r)
	if !r.IsInt() {
		q.Add(q, big.NewInt(1))
	}
	return q
}

// roundRat rounds half away from zero to digits decimal places (BigDecimal ROUND_HALF_UP).
func roundRat(r *big.Rat, digits int) *big.Rat {
	scale := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt(digits))), nil))
	scaled := new(big.Rat).Set(r)
	if digits >= 0 {
		scaled.Mul(scaled, scale)
	} else {
		scaled.Quo(scaled, scale)
	}
	half := big.NewRat(1, 2)
	var rounded *big.Int
	if scaled.Sign() >= 0 {
		rounded = floorRat(scaled.Add(scaled, half))
	} else {
		rounded = new(big.Int).Neg(floorRat(scaled.Neg(scaled).Add(scaled, half)))
	}
	result := new(big.Rat).SetInt(rounded)
	if digits >= 0 {
		return result.Quo(result, scale)
	}
	return result.Mul(result, scale)
}

func absInt(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func compareLiquidNumbers(a, b liquidNumber) int {
	if a.kind == specialNumber || b.kind == specialNumber {
		x, y := a.float64(), b.float64()
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	}
	return a.rat().Cmp(b.rat())
}
