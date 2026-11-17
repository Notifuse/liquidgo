package liquid

import (
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// StandardFilters provides standard filter implementations.
// This is a struct with methods that can be used as filters.
type StandardFilters struct {
	context *Context
}

// Size returns the size of a string or array.
func (sf *StandardFilters) Size(input interface{}) interface{} {
	if input == nil {
		return 0
	}
	switch v := input.(type) {
	case string:
		return utf8.RuneCountInString(v)
	case []interface{}:
		return len(v)
	case map[string]interface{}:
		return len(v)
	case *Range:
		return v.Len()
	default:
		// Reflection fallback for typed slices/arrays/maps
		// This matches Ruby's duck-typing behavior: objects respond to .size
		if input != nil {
			val := reflect.ValueOf(input)
			kind := val.Kind()
			if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
				return val.Len()
			}
		}
		return 0
	}
}

// First returns the first element of an array, or the start of a range, as
// Range#first, even when the range is empty.
func (sf *StandardFilters) First(input interface{}) interface{} {
	if s, ok := input.(string); ok {
		return firstChar(s)
	}
	if r, ok := input.(*Range); ok {
		return r.Start
	}
	if arr, ok := input.([]interface{}); ok && len(arr) > 0 {
		return arr[0]
	}
	// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
	// This matches Ruby's duck-typing behavior: arrays respond to .first
	if input != nil {
		v := reflect.ValueOf(input)
		if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && v.Len() > 0 {
			return itemAt(v, 0)
		}
	}
	return nil
}

// Last returns the last element of an array, or the end of a range, as
// Range#last, even when the range is empty.
func (sf *StandardFilters) Last(input interface{}) interface{} {
	if s, ok := input.(string); ok {
		return lastChar(s)
	}
	if r, ok := input.(*Range); ok {
		return r.End
	}
	if arr, ok := input.([]interface{}); ok && len(arr) > 0 {
		return arr[len(arr)-1]
	}
	// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
	// This matches Ruby's duck-typing behavior: arrays respond to .last
	if input != nil {
		v := reflect.ValueOf(input)
		if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && v.Len() > 0 {
			return itemAt(v, v.Len()-1)
		}
	}
	return nil
}

// Date formats a date using strftime-style format codes.
func (sf *StandardFilters) Date(input interface{}, format interface{}) interface{} {
	formatStr := ToS(format, nil)
	if formatStr == "" {
		return input
	}

	date := ToDate(input)
	if date == nil {
		return input
	}

	return strftime(date, formatStr)
}

// Default returns a default value if input is nil, false, or empty.
// Mirrors Ruby's default from standardfilters.rb:940
func (sf *StandardFilters) Default(input interface{}, defaultValue interface{}, options interface{}) interface{} {
	// Parse options
	allowFalse := false
	if opts, ok := options.(map[string]interface{}); ok {
		if val, exists := opts["allow_false"]; exists {
			if b, ok := val.(bool); ok {
				allowFalse = b
			}
		}
	}

	// Check if input should be replaced with default
	if input == nil {
		return defaultValue
	}

	// Check for false
	if !allowFalse {
		if b, ok := input.(bool); ok && !b {
			return defaultValue
		}
	}

	// Check for empty
	switch v := input.(type) {
	case string:
		if v == "" {
			return defaultValue
		}
	case []interface{}:
		if len(v) == 0 {
			return defaultValue
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return defaultValue
		}
	}

	// Reflection fallback for typed slices/arrays/maps
	// This matches Ruby's duck-typing behavior: objects respond to .empty?
	if input != nil {
		val := reflect.ValueOf(input)
		kind := val.Kind()
		if (kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map) && val.Len() == 0 {
			return defaultValue
		}
	}

	return input
}

// InputIterator wraps input for iteration, as Ruby's InputIterator.
type InputIterator struct {
	input []interface{}
}

// NewInputIterator creates a new InputIterator.
func NewInputIterator(input interface{}, context *Context) *InputIterator {
	var items []interface{}

	if input == nil {
		items = []interface{}{}
	} else if arr, ok := input.([]interface{}); ok {
		// Flatten arrays
		items = flattenArray(arr)
	} else if m, ok := input.(map[string]interface{}); ok {
		items = []interface{}{m}
	} else if r, ok := input.(*Range); ok {
		// A range is enumerable: its integers, within the limits
		n := r.Len()
		(&StandardFilters{context: context}).checkRangeItems(n)
		items = make([]interface{}, n)
		for i := range items {
			items[i] = r.Start + i
		}
	} else if input != nil {
		// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
		// This matches Ruby's duck-typing behavior: arrays respond to iteration
		v := reflect.ValueOf(input)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			items = make([]interface{}, v.Len())
			for i := 0; i < v.Len(); i++ {
				items[i] = itemAt(v, i)
			}
		} else {
			// Single item
			items = []interface{}{input}
		}
	} else {
		// Single item
		items = []interface{}{input}
	}

	return &InputIterator{
		input: items,
	}
}

// flattenArray recursively flattens nested arrays. An array that holds
// itself is left out where it recurs, instead of recursing without end.
func flattenArray(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr))
	if hasNestedList(arr) {
		return appendFlattened(result, arr, map[flattenKey]bool{})
	}
	// no nested list: a copy, without the map that stops a list holding itself
	return append(result, arr...)
}

// flattenKey names an array being flattened by its first item and length.
type flattenKey struct {
	first *interface{}
	n     int
}

func appendFlattened(result, arr []interface{}, seen map[flattenKey]bool) []interface{} {
	if len(arr) > 0 {
		key := flattenKey{&arr[0], len(arr)}
		if seen[key] {
			return result
		}
		seen[key] = true
		defer delete(seen, key)
	}
	for _, item := range arr {
		if nested, ok := item.([]interface{}); ok {
			result = appendFlattened(result, nested, seen)
		} else {
			result = append(result, item)
		}
	}
	return result
}

// Each iterates over items, converting them to liquid values.
func (it *InputIterator) Each(fn func(interface{})) {
	for _, item := range it.input {
		// Convert to liquid, as Ruby's InputIterator yields each item's to_liquid
		liquidItem := ToLiquidValue(ToLiquid(item))

		fn(liquidItem)
	}
}

// ToArray converts the iterator to an array.
func (it *InputIterator) ToArray() []interface{} {
	result := make([]interface{}, 0, len(it.input))
	it.Each(func(item interface{}) {
		result = append(result, item)
	})
	return result
}

// Empty checks if the iterator is empty.
func (it *InputIterator) Empty() bool {
	return len(it.input) == 0
}

// Join joins items with a glue string.
func (it *InputIterator) Join(glue string) string {
	var builder strings.Builder
	first := true
	it.Each(func(item interface{}) {
		if first {
			first = false
		} else {
			builder.WriteString(glue)
		}
		builder.WriteString(ToS(item, nil))
	})
	return builder.String()
}

// Concat concatenates with another array.
func (it *InputIterator) Concat(args interface{}) []interface{} {
	result := it.ToArray()
	if arr, ok := args.([]interface{}); ok {
		result = append(result, arr...)
	} else if args != nil {
		// Use reflection to handle typed slices
		v := reflect.ValueOf(args)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			for i := 0; i < v.Len(); i++ {
				result = append(result, itemAt(v, i))
			}
		}
	}
	return result
}

// Reverse returns items in reverse order.
func (it *InputIterator) Reverse() []interface{} {
	arr := it.ToArray()
	result := make([]interface{}, len(arr))
	for i, item := range arr {
		result[len(arr)-1-i] = item
	}
	return result
}

// Uniq returns unique items.
func (it *InputIterator) Uniq(keyFunc func(interface{}) interface{}) []interface{} {
	seen := make(map[interface{}]bool)
	result := make([]interface{}, 0)

	it.Each(func(item interface{}) {
		var key interface{}
		if keyFunc != nil {
			key = keyFunc(ToLiquidValue(item))
		} else {
			key = ToLiquidValue(item)
		}

		// Use string representation for complex types
		keyStr := ""
		if key != nil {
			keyStr = ToS(key, nil)
		}

		if !seen[keyStr] {
			seen[keyStr] = true
			result = append(result, item)
		}
	})

	return result
}

// Compact removes nil items.
func (it *InputIterator) Compact() []interface{} {
	result := make([]interface{}, 0)
	it.Each(func(item interface{}) {
		if item != nil {
			result = append(result, item)
		}
	})
	return result
}

// Helper functions for filters

// nilSafeCompare compares two values, handling nil gracefully.
// Mirrors Ruby's nil_safe_compare from standardfilters.rb:1008-1020
func nilSafeCompare(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1 // nil goes last
	}
	if b == nil {
		return -1 // nil goes last
	}

	// Numbers compare as numbers; strings, numeric strings included, as strings (Ruby <=>)
	aNum, aOk := asNumber(a)
	bNum, bOk := asNumber(b)
	if aOk && bOk {
		cmp, _ := compareNumbers(aNum, bNum)
		return cmp
	}
	if (aOk || isBigInt(a)) && (bOk || isBigInt(b)) {
		return compareBigInts(a, b)
	}

	// String comparison
	aStr := ToS(a, nil)
	bStr := ToS(b, nil)
	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// ToNumberValue converts to float64 for comparison.
func ToNumberValue(obj interface{}) (float64, bool) {
	switch v := obj.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case string:
		// Try to parse as float first
		trimmed := strings.TrimSpace(v)
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f, true
		}
		// Try to parse as int
		if i, err := strconv.Atoi(trimmed); err == nil {
			return float64(i), true
		}
		// Not a number
		return 0, false
	}
	return 0, false
}

// nilSafeCasecmp compares strings case-insensitively, handling nil.
// Mirrors Ruby's nil_safe_casecmp from standardfilters.rb:1022-1030
func nilSafeCasecmp(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}

	aStr := strings.ToLower(ToS(a, nil))
	bStr := strings.ToLower(ToS(b, nil))

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// raisePropertyError creates an ArgumentError for invalid properties.
func raisePropertyError(property interface{}) error {
	return NewArgumentError("cannot select the property '" + ToS(property, nil) + "'")
}

// Array Filters

// firstChar is Ruby's s[0] || "": the first character of s, or "" for an
// empty string (Shopify Liquid 5.14, as ActiveSupport's String#first).
func firstChar(s string) string {
	_, size := utf8.DecodeRuneInString(s)
	return s[:size]
}

// lastChar is Ruby's s[-1] || "": the last character of s, or "".
func lastChar(s string) string {
	_, size := utf8.DecodeLastRuneInString(s)
	return s[len(s)-size:]
}
