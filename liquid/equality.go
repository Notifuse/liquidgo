package liquid

import (
	"encoding/json"
	"math"
	"math/big"
	"reflect"
	"strconv"
)

// maxEqualityDepth bounds the recursion of liquidEqual, so that self-referencing data
// returns false instead of recursing forever.
const maxEqualityDepth = 100

// number is a Go numeric value kept without loss: signed and unsigned integers stay
// integers, so that values above 2^53 compare exactly.
type number struct {
	kind numberKind
	i    int64
	u    uint64
	f    float64
}

type numberKind int

const (
	signedNumber numberKind = iota + 1
	unsignedNumber
	floatNumber
)

// asNumber reports whether v is a number: any Go integer or float kind, named numeric
// types included, and json.Number. Strings and booleans are not numbers, as in Ruby.
func asNumber(v interface{}) (number, bool) {
	switch n := v.(type) {
	case nil, string, bool:
		return number{}, false
	case int:
		return number{kind: signedNumber, i: int64(n)}, true
	case int64:
		return number{kind: signedNumber, i: n}, true
	case float64:
		return number{kind: floatNumber, f: n}, true
	case float32:
		return number{kind: floatNumber, f: float32Decimal(float64(n))}, true
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return number{kind: signedNumber, i: i}, true
		}
		if f, err := n.Float64(); err == nil {
			return number{kind: floatNumber, f: f}, true
		}
		return number{}, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return number{kind: signedNumber, i: rv.Int()}, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return number{kind: unsignedNumber, u: rv.Uint()}, true
	case reflect.Float32:
		return number{kind: floatNumber, f: float32Decimal(rv.Float())}, true
	case reflect.Float64:
		return number{kind: floatNumber, f: rv.Float()}, true
	}
	return number{}, false
}

// float32Decimal reads a float32 through its shortest decimal, as arithmetic does, so
// that float32(0.1) is 0.1 and not 0.10000000149011612.
func float32Decimal(f float64) float64 {
	d, err := strconv.ParseFloat(strconv.FormatFloat(f, 'g', -1, 32), 64)
	if err != nil {
		return f
	}
	return d
}

func (n number) float64() float64 {
	switch n.kind {
	case signedNumber:
		return float64(n.i)
	case unsignedNumber:
		return float64(n.u)
	}
	return n.f
}

func (n number) bigFloat() *big.Float {
	switch n.kind {
	case signedNumber:
		return new(big.Float).SetInt64(n.i)
	case unsignedNumber:
		return new(big.Float).SetUint64(n.u)
	}
	return new(big.Float).SetFloat64(n.f)
}

// compareNumbers compares two numbers exactly. ok is false when either is NaN.
func compareNumbers(a, b number) (cmp int, ok bool) {
	switch {
	case a.kind == signedNumber && b.kind == signedNumber:
		return compareOrdered(a.i, b.i), true
	case a.kind == unsignedNumber && b.kind == unsignedNumber:
		return compareOrdered(a.u, b.u), true
	case a.kind == signedNumber && b.kind == unsignedNumber:
		if a.i < 0 {
			return -1, true
		}
		return compareOrdered(uint64(a.i), b.u), true
	case a.kind == unsignedNumber && b.kind == signedNumber:
		if b.i < 0 {
			return 1, true
		}
		return compareOrdered(a.u, uint64(b.i)), true
	}
	x, y := a.float64(), b.float64()
	if math.IsNaN(x) || math.IsNaN(y) {
		return 0, false
	}
	if a.kind == floatNumber && b.kind == floatNumber {
		return compareOrdered(x, y), true
	}
	return a.bigFloat().Cmp(b.bigFloat()), true
}

func compareOrdered[T int64 | uint64 | float64](a, b T) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// liquidEqual implements Liquid's == the way Ruby's == behaves on the values Liquid
// handles: numbers by value whatever their Go type (100 == 100.0), lists and maps
// element by element, named string and bool types by their value. It never panics.
func liquidEqual(left, right interface{}) bool {
	return liquidEqualDepth(left, right, 0)
}

func liquidEqualDepth(left, right interface{}, depth int) (equal bool) {
	if depth > maxEqualityDepth {
		return false
	}
	defer func() {
		if recover() != nil {
			equal = false
		}
	}()
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if ln, ok := asNumber(left); ok {
		rn, ok := asNumber(right)
		if !ok {
			return false
		}
		cmp, ok := compareNumbers(ln, rn)
		return ok && cmp == 0
	}
	if _, ok := asNumber(right); ok {
		return false
	}

	lv, rv := reflect.ValueOf(left), reflect.ValueOf(right)
	lk, rk := lv.Kind(), rv.Kind()
	switch {
	case lk == reflect.String && rk == reflect.String:
		return lv.String() == rv.String()
	case lk == reflect.Bool && rk == reflect.Bool:
		return lv.Bool() == rv.Bool()
	case isList(lk) && isList(rk):
		if lv.Len() != rv.Len() {
			return false
		}
		for i := 0; i < lv.Len(); i++ {
			if !liquidEqualDepth(lv.Index(i).Interface(), rv.Index(i).Interface(), depth+1) {
				return false
			}
		}
		return true
	case lk == reflect.Map && rk == reflect.Map:
		if lv.Len() != rv.Len() {
			return false
		}
		// look the keys of one map up in the other, whichever direction the key types allow,
		// so that the comparison is symmetric
		if !lv.Type().Key().AssignableTo(rv.Type().Key()) {
			if !rv.Type().Key().AssignableTo(lv.Type().Key()) {
				return false
			}
			lv, rv = rv, lv
		}
		iter := lv.MapRange()
		for iter.Next() {
			other := rv.MapIndex(iter.Key())
			if !other.IsValid() || !liquidEqualDepth(iter.Value().Interface(), other.Interface(), depth+1) {
				return false
			}
		}
		return true
	}
	if lv.Type() != rv.Type() || !lv.Comparable() || !rv.Comparable() {
		return false
	}
	return left == right
}

func isList(k reflect.Kind) bool {
	return k == reflect.Slice || k == reflect.Array
}

// toNumber returns v as a float64 when v is a number (see asNumber).
func toNumber(v interface{}) (float64, bool) {
	n, ok := asNumber(v)
	return n.float64(), ok
}
