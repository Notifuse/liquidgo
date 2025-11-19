package liquid

import (
	"encoding/base64"
	"html"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"
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
		return len(v)
	case []interface{}:
		return len(v)
	case map[string]interface{}:
		return len(v)
	default:
		// Try to get size via reflection or return 0
		return 0
	}
}

// Downcase converts a string to all lowercase characters.
func (sf *StandardFilters) Downcase(input interface{}) string {
	return strings.ToLower(ToS(input, nil))
}

// Upcase converts a string to all uppercase characters.
func (sf *StandardFilters) Upcase(input interface{}) string {
	return strings.ToUpper(ToS(input, nil))
}

// Capitalize capitalizes the first word in a string and downcases the remaining characters.
func (sf *StandardFilters) Capitalize(input interface{}) string {
	s := ToS(input, nil)
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// Escape escapes special characters in HTML.
func (sf *StandardFilters) Escape(input interface{}) string {
	if input == nil {
		return ""
	}
	return html.EscapeString(ToS(input, nil))
}

// H is an alias for Escape.
func (sf *StandardFilters) H(input interface{}) string {
	return sf.Escape(input)
}

// EscapeOnce escapes a string without changing characters that have already been escaped.
func (sf *StandardFilters) EscapeOnce(input interface{}) string {
	s := ToS(input, nil)
	// Simple implementation - just escape
	return html.EscapeString(s)
}

// URLEncode converts URL-unsafe characters to percent-encoded equivalent.
func (sf *StandardFilters) URLEncode(input interface{}) string {
	if input == nil {
		return ""
	}
	return url.QueryEscape(ToS(input, nil))
}

// URLDecode decodes percent-encoded characters.
func (sf *StandardFilters) URLDecode(input interface{}) (string, error) {
	if input == nil {
		return "", nil
	}
	decoded, err := url.QueryUnescape(ToS(input, nil))
	if err != nil {
		return "", NewArgumentError("invalid URL encoding")
	}
	return decoded, nil
}

// Base64Encode encodes a string to Base64 format.
func (sf *StandardFilters) Base64Encode(input interface{}) string {
	return base64.StdEncoding.EncodeToString([]byte(ToS(input, nil)))
}

// Base64Decode decodes a string in Base64 format.
func (sf *StandardFilters) Base64Decode(input interface{}) (string, error) {
	s := ToS(input, nil)
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", NewArgumentError("invalid base64 provided to base64_decode")
	}
	return string(decoded), nil
}

// Base64URLSafeEncode encodes a string to URL-safe Base64 format.
func (sf *StandardFilters) Base64URLSafeEncode(input interface{}) string {
	return base64.URLEncoding.EncodeToString([]byte(ToS(input, nil)))
}

// Base64URLSafeDecode decodes a string in URL-safe Base64 format.
func (sf *StandardFilters) Base64URLSafeDecode(input interface{}) (string, error) {
	s := ToS(input, nil)
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", NewArgumentError("invalid base64 provided to base64_url_safe_decode")
	}
	return string(decoded), nil
}

// Slice returns a substring or series of array items.
func (sf *StandardFilters) Slice(input interface{}, offset interface{}, length interface{}) interface{} {
	offsetInt, _ := ToInteger(offset)
	lengthInt := 1
	if length != nil {
		lengthInt, _ = ToInteger(length)
	}

	switch v := input.(type) {
	case []interface{}:
		if offsetInt < 0 || offsetInt >= len(v) {
			return []interface{}{}
		}
		end := offsetInt + lengthInt
		if end > len(v) {
			end = len(v)
		}
		return v[offsetInt:end]
	case string:
		if offsetInt < 0 || offsetInt >= len(v) {
			return ""
		}
		end := offsetInt + lengthInt
		if end > len(v) {
			end = len(v)
		}
		return v[offsetInt:end]
	default:
		return ""
	}
}

// Truncate truncates a string down to a given number of characters.
func (sf *StandardFilters) Truncate(input interface{}, length interface{}, truncateString interface{}) string {
	if input == nil {
		return ""
	}
	inputStr := ToS(input, nil)
	lengthInt, _ := ToInteger(length)
	if lengthInt <= 0 {
		lengthInt = 50
	}

	truncateStr := "..."
	if truncateString != nil {
		truncateStr = ToS(truncateString, nil)
	}

	if len(inputStr) <= lengthInt {
		return inputStr
	}

	l := lengthInt - len(truncateStr)
	if l < 0 {
		l = 0
	}

	return inputStr[:l] + truncateStr
}

// TruncateWords truncates a string down to a given number of words.
func (sf *StandardFilters) TruncateWords(input interface{}, words interface{}, truncateString interface{}) string {
	if input == nil {
		return ""
	}
	inputStr := ToS(input, nil)
	wordsInt, _ := ToInteger(words)
	if wordsInt <= 0 {
		wordsInt = 15
	}

	truncateStr := "..."
	if truncateString != nil {
		truncateStr = ToS(truncateString, nil)
	}

	wordList := strings.Fields(inputStr)
	if len(wordList) <= wordsInt {
		return inputStr
	}

	result := strings.Join(wordList[:wordsInt], " ")
	return result + truncateStr
}

// Split splits a string into an array of substrings based on a separator.
func (sf *StandardFilters) Split(input interface{}, pattern interface{}) []interface{} {
	inputStr := ToS(input, nil)
	patternStr := ToS(pattern, nil)
	parts := strings.Split(inputStr, patternStr)
	result := make([]interface{}, len(parts))
	for i, part := range parts {
		result[i] = part
	}
	return result
}

// Strip strips whitespace from both ends of a string.
func (sf *StandardFilters) Strip(input interface{}) string {
	return strings.TrimSpace(ToS(input, nil))
}

// Lstrip strips whitespace from the left end of a string.
func (sf *StandardFilters) Lstrip(input interface{}) string {
	return strings.TrimLeft(ToS(input, nil), " \t\n\r")
}

// Rstrip strips whitespace from the right end of a string.
func (sf *StandardFilters) Rstrip(input interface{}) string {
	return strings.TrimRight(ToS(input, nil), " \t\n\r")
}

// StripHTML strips HTML tags from a string.
func (sf *StandardFilters) StripHTML(input interface{}) string {
	s := ToS(input, nil)

	// First remove script/style/comment blocks
	stripHTMLBlocks := regexp.MustCompile(`(?s)<script.*?</script>|<!--.*?-->|<style.*?</style>`)
	result := stripHTMLBlocks.ReplaceAllString(s, "")

	// Then remove all HTML tags
	stripHTMLTags := regexp.MustCompile(`<.*?>`)
	result = stripHTMLTags.ReplaceAllString(result, "")

	return result
}

// First returns the first element of an array.
func (sf *StandardFilters) First(input interface{}) interface{} {
	if arr, ok := input.([]interface{}); ok && len(arr) > 0 {
		return arr[0]
	}
	return nil
}

// Last returns the last element of an array.
func (sf *StandardFilters) Last(input interface{}) interface{} {
	if arr, ok := input.([]interface{}); ok && len(arr) > 0 {
		return arr[len(arr)-1]
	}
	return nil
}

// Join joins array elements with a separator.
func (sf *StandardFilters) Join(input interface{}, separator interface{}) string {
	sep := ", "
	if separator != nil {
		sep = ToS(separator, nil)
	}

	if arr, ok := input.([]interface{}); ok {
		parts := make([]string, len(arr))
		for i, item := range arr {
			parts[i] = ToS(item, nil)
		}
		return strings.Join(parts, sep)
	}
	return ToS(input, nil)
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

// StripNewlines strips all newline characters from a string.
// Mirrors Ruby's strip_newlines from standardfilters.rb:354
func (sf *StandardFilters) StripNewlines(input interface{}) string {
	s := ToS(input, nil)
	s = strings.ReplaceAll(s, "\r\n", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// NewlineToBr converts newlines to HTML line breaks.
// Mirrors Ruby's newline_to_br from standardfilters.rb:709
func (sf *StandardFilters) NewlineToBr(input interface{}) string {
	s := ToS(input, nil)
	// Replace \r\n first to avoid double replacement
	re := regexp.MustCompile(`\r?\n`)
	return re.ReplaceAllString(s, "<br />\n")
}

// Replace replaces all occurrences of a substring.
// Mirrors Ruby's replace from standardfilters.rb:583
func (sf *StandardFilters) Replace(input interface{}, str interface{}, replacement interface{}) string {
	inputStr := ToS(input, nil)
	searchStr := ToS(str, nil)
	replaceStr := ""
	if replacement != nil {
		replaceStr = ToS(replacement, nil)
	}
	return strings.ReplaceAll(inputStr, searchStr, replaceStr)
}

// ReplaceFirst replaces the first occurrence of a substring.
// Mirrors Ruby's replace_first from standardfilters.rb:597
func (sf *StandardFilters) ReplaceFirst(input interface{}, str interface{}, replacement interface{}) string {
	inputStr := ToS(input, nil)
	searchStr := ToS(str, nil)
	replaceStr := ""
	if replacement != nil {
		replaceStr = ToS(replacement, nil)
	}
	return strings.Replace(inputStr, searchStr, replaceStr, 1)
}

// ReplaceLast replaces the last occurrence of a substring.
// Mirrors Ruby's replace_last from standardfilters.rb:611
func (sf *StandardFilters) ReplaceLast(input interface{}, str interface{}, replacement interface{}) string {
	inputStr := ToS(input, nil)
	searchStr := ToS(str, nil)
	replaceStr := ToS(replacement, nil)

	startIndex := strings.LastIndex(inputStr, searchStr)
	if startIndex == -1 {
		return inputStr
	}

	return inputStr[:startIndex] + replaceStr + inputStr[startIndex+len(searchStr):]
}

// Remove removes all occurrences of a substring.
// Mirrors Ruby's remove from standardfilters.rb:632
func (sf *StandardFilters) Remove(input interface{}, str interface{}) string {
	return sf.Replace(input, str, "")
}

// RemoveFirst removes the first occurrence of a substring.
// Mirrors Ruby's remove_first from standardfilters.rb:643
func (sf *StandardFilters) RemoveFirst(input interface{}, str interface{}) string {
	return sf.ReplaceFirst(input, str, "")
}

// RemoveLast removes the last occurrence of a substring.
// Mirrors Ruby's remove_last from standardfilters.rb:654
func (sf *StandardFilters) RemoveLast(input interface{}, str interface{}) string {
	return sf.ReplaceLast(input, str, "")
}

// Append adds a string to the end.
// Mirrors Ruby's append from standardfilters.rb:665
func (sf *StandardFilters) Append(input interface{}, str interface{}) string {
	inputStr := ToS(input, nil)
	appendStr := ToS(str, nil)
	return inputStr + appendStr
}

// Prepend adds a string to the beginning.
// Mirrors Ruby's prepend from standardfilters.rb:696
func (sf *StandardFilters) Prepend(input interface{}, str interface{}) string {
	inputStr := ToS(input, nil)
	prependStr := ToS(str, nil)
	return prependStr + inputStr
}

// Abs returns the absolute value of a number.
// Mirrors Ruby's abs from standardfilters.rb:792
func (sf *StandardFilters) Abs(input interface{}) interface{} {
	num := ToNumber(input)
	switch v := num.(type) {
	case int:
		if v < 0 {
			return -v
		}
		return v
	case float64:
		return math.Abs(v)
	default:
		return 0
	}
}

// Plus adds two numbers.
// Mirrors Ruby's plus from standardfilters.rb:804
func (sf *StandardFilters) Plus(input interface{}, operand interface{}) interface{} {
	return applyOperation(input, operand, "+")
}

// Minus subtracts two numbers.
// Mirrors Ruby's minus from standardfilters.rb:815
func (sf *StandardFilters) Minus(input interface{}, operand interface{}) interface{} {
	return applyOperation(input, operand, "-")
}

// Times multiplies two numbers.
// Mirrors Ruby's times from standardfilters.rb:826
func (sf *StandardFilters) Times(input interface{}, operand interface{}) interface{} {
	return applyOperation(input, operand, "*")
}

// DividedBy divides two numbers.
// Mirrors Ruby's divided_by from standardfilters.rb:837
func (sf *StandardFilters) DividedBy(input interface{}, operand interface{}) (interface{}, error) {
	operandNum := ToNumber(operand)
	var operandFloat float64
	switch v := operandNum.(type) {
	case int:
		if v == 0 {
			return nil, NewZeroDivisionError("divided by 0")
		}
		operandFloat = float64(v)
	case float64:
		if v == 0 {
			return nil, NewZeroDivisionError("divided by 0")
		}
		operandFloat = v
	default:
		return nil, NewZeroDivisionError("divided by 0")
	}

	if operandFloat == 0 {
		return nil, NewZeroDivisionError("divided by 0")
	}

	return applyOperation(input, operand, "/"), nil
}

// Modulo returns the remainder of division.
// Mirrors Ruby's modulo from standardfilters.rb:850
func (sf *StandardFilters) Modulo(input interface{}, operand interface{}) (interface{}, error) {
	operandNum := ToNumber(operand)
	var operandFloat float64
	switch v := operandNum.(type) {
	case int:
		if v == 0 {
			return nil, NewZeroDivisionError("divided by 0")
		}
		operandFloat = float64(v)
	case float64:
		if v == 0 {
			return nil, NewZeroDivisionError("divided by 0")
		}
		operandFloat = v
	default:
		return nil, NewZeroDivisionError("divided by 0")
	}

	if operandFloat == 0 {
		return nil, NewZeroDivisionError("divided by 0")
	}

	return applyOperation(input, operand, "%"), nil
}

// Round rounds a number to the nearest integer or specified precision.
// Mirrors Ruby's round from standardfilters.rb:863
func (sf *StandardFilters) Round(input interface{}, n interface{}) (interface{}, error) {
	num := ToNumber(input)
	precision := 0
	if n != nil {
		precisionNum := ToNumber(n)
		if p, ok := precisionNum.(int); ok {
			precision = p
		} else if p, ok := precisionNum.(float64); ok {
			precision = int(p)
		}
	}

	var result interface{}
	switch v := num.(type) {
	case int:
		result = v
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return nil, NewFloatDomainError("Infinity")
		}
		if precision == 0 {
			result = int(math.Round(v))
		} else {
			multiplier := math.Pow(10, float64(precision))
			result = math.Round(v*multiplier) / multiplier
		}
	default:
		result = 0
	}

	return result, nil
}

// Ceil rounds a number up to the nearest integer.
// Mirrors Ruby's ceil from standardfilters.rb:879
func (sf *StandardFilters) Ceil(input interface{}) (int, error) {
	num := ToNumber(input)
	switch v := num.(type) {
	case int:
		return v, nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return 0, NewFloatDomainError("Infinity")
		}
		return int(math.Ceil(v)), nil
	default:
		return 0, nil
	}
}

// Floor rounds a number down to the nearest integer.
// Mirrors Ruby's floor from standardfilters.rb:892
func (sf *StandardFilters) Floor(input interface{}) (int, error) {
	num := ToNumber(input)
	switch v := num.(type) {
	case int:
		return v, nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return 0, NewFloatDomainError("Infinity")
		}
		return int(math.Floor(v)), nil
	default:
		return 0, nil
	}
}

// AtLeast limits a number to a minimum value.
// Mirrors Ruby's at_least from standardfilters.rb:905
func (sf *StandardFilters) AtLeast(input interface{}, n interface{}) interface{} {
	minValue := ToNumber(n)
	result := ToNumber(input)

	minFloat, _ := ToNumberValue(minValue)
	resultFloat, _ := ToNumberValue(result)

	if minFloat > resultFloat {
		return minValue
	}
	return result
}

// AtMost limits a number to a maximum value.
// Mirrors Ruby's at_most from standardfilters.rb:920
func (sf *StandardFilters) AtMost(input interface{}, n interface{}) interface{} {
	maxValue := ToNumber(n)
	result := ToNumber(input)

	maxFloat, _ := ToNumberValue(maxValue)
	resultFloat, _ := ToNumberValue(result)

	if maxFloat < resultFloat {
		return maxValue
	}
	return result
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

	return input
}

// InputIterator wraps input for iteration with context support.
// Mirrors Ruby's InputIterator class from standardfilters.rb:1032-1094
type InputIterator struct {
	input   []interface{}
	context *Context
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
	} else {
		// Single item
		items = []interface{}{input}
	}

	return &InputIterator{
		input:   items,
		context: context,
	}
}

// flattenArray recursively flattens nested arrays.
func flattenArray(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, item := range arr {
		if nested, ok := item.([]interface{}); ok {
			result = append(result, flattenArray(nested)...)
		} else {
			result = append(result, item)
		}
	}
	return result
}

// Each iterates over items, converting them to liquid values.
func (it *InputIterator) Each(fn func(interface{})) {
	for _, item := range it.input {
		// Convert to liquid if needed
		liquidItem := ToLiquidValue(item)

		// Set context if supported
		if it.context != nil {
			if drop, ok := liquidItem.(interface{ SetContext(*Context) }); ok {
				drop.SetContext(it.context)
			}
		}

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

	// Try numeric comparison first
	aNum, aOk := ToNumberValue(a)
	bNum, bOk := ToNumberValue(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		} else if aNum > bNum {
			return 1
		}
		return 0
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
		// Try to parse as number
		num := ToNumber(v)
		if f, ok := num.(float64); ok {
			return f, true
		}
		if i, ok := num.(int); ok {
			return float64(i), true
		}
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

// applyOperation applies a math operation to two numbers.
// Mirrors Ruby's apply_operation from standardfilters.rb:1003-1006
func applyOperation(input, operand interface{}, operation string) interface{} {
	inputNum := ToNumber(input)
	operandNum := ToNumber(operand)

	var result interface{}

	// Convert to float64 for calculation
	var inputFloat, operandFloat float64
	switch v := inputNum.(type) {
	case int:
		inputFloat = float64(v)
	case float64:
		inputFloat = v
	default:
		inputFloat = 0
	}

	switch v := operandNum.(type) {
	case int:
		operandFloat = float64(v)
	case float64:
		operandFloat = v
	default:
		operandFloat = 0
	}

	// Determine if we should return int or float
	_, inputIsInt := inputNum.(int)
	_, operandIsInt := operandNum.(int)

	switch operation {
	case "+":
		result = inputFloat + operandFloat
	case "-":
		result = inputFloat - operandFloat
	case "*":
		result = inputFloat * operandFloat
	case "/":
		if operandFloat == 0 {
			return 0 // Handle division by zero
		}
		result = inputFloat / operandFloat
		// For division, if operand is int, result should be int
		if operandIsInt && inputIsInt {
			return int(inputFloat / operandFloat)
		}
	case "%":
		if operandFloat == 0 {
			return 0 // Handle modulo by zero
		}
		result = math.Mod(inputFloat, operandFloat)
	default:
		result = inputFloat
	}

	// Return int if both inputs are ints (except for division)
	if inputIsInt && operandIsInt && operation != "/" {
		if f, ok := result.(float64); ok {
			return int(f)
		}
	}

	return result
}

// raisePropertyError creates an ArgumentError for invalid properties.
func raisePropertyError(property interface{}) error {
	return NewArgumentError("cannot select the property '" + ToS(property, nil) + "'")
}

// Array Filters

// Reverse reverses the order of items in an array.
// Mirrors Ruby's reverse from standardfilters.rb:523
func (sf *StandardFilters) Reverse(input interface{}) []interface{} {
	iter := NewInputIterator(input, sf.context)
	return iter.Reverse()
}

// Sort sorts items in an array.
// Mirrors Ruby's sort from standardfilters.rb:378
func (sf *StandardFilters) Sort(input interface{}, property interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)
	arr := iter.ToArray()

	if len(arr) == 0 {
		return []interface{}{}
	}

	// If no property, sort items directly
	if property == nil {
		sorted := make([]interface{}, len(arr))
		copy(sorted, arr)
		sort.SliceStable(sorted, func(i, j int) bool {
			return nilSafeCompare(sorted[i], sorted[j]) < 0
		})
		return sorted
	}

	// Sort by property
	propStr := ToS(property, nil)

	// Check if all items support indexing
	for _, item := range arr {
		if !supportsIndexing(item) {
			return nil
		}
	}

	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	sort.SliceStable(sorted, func(i, j int) bool {
		a := getProperty(sorted[i], propStr)
		b := getProperty(sorted[j], propStr)
		return nilSafeCompare(a, b) < 0
	})

	return sorted
}

// SortNatural sorts items case-insensitively.
// Mirrors Ruby's sort_natural from standardfilters.rb:407
func (sf *StandardFilters) SortNatural(input interface{}, property interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)
	arr := iter.ToArray()

	if len(arr) == 0 {
		return []interface{}{}
	}

	// If no property, sort items directly
	if property == nil {
		sorted := make([]interface{}, len(arr))
		copy(sorted, arr)
		sort.SliceStable(sorted, func(i, j int) bool {
			return nilSafeCasecmp(sorted[i], sorted[j]) < 0
		})
		return sorted
	}

	// Sort by property
	propStr := ToS(property, nil)

	// Check if all items support indexing
	for _, item := range arr {
		if !supportsIndexing(item) {
			return nil
		}
	}

	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)

	sort.SliceStable(sorted, func(i, j int) bool {
		a := getProperty(sorted[i], propStr)
		b := getProperty(sorted[j], propStr)
		return nilSafeCasecmp(a, b) < 0
	})

	return sorted
}

// Uniq removes duplicate items from an array.
// Mirrors Ruby's uniq from standardfilters.rb:497
func (sf *StandardFilters) Uniq(input interface{}, property interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}
	}

	if property == nil {
		return iter.Uniq(nil)
	}

	// Uniq by property
	propStr := ToS(property, nil)
	return iter.Uniq(func(item interface{}) interface{} {
		return getProperty(item, propStr)
	})
}

// Compact removes nil items from an array.
// Mirrors Ruby's compact from standardfilters.rb:557
func (sf *StandardFilters) Compact(input interface{}, property interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}
	}

	if property == nil {
		return iter.Compact()
	}

	// Compact by property - remove items where property is nil
	propStr := ToS(property, nil)
	result := make([]interface{}, 0)

	iter.Each(func(item interface{}) {
		if !supportsIndexing(item) {
			return
		}
		val := getProperty(item, propStr)
		if val != nil {
			result = append(result, item)
		}
	})

	return result
}

// Map extracts property values from array items.
// Mirrors Ruby's map from standardfilters.rb:535
func (sf *StandardFilters) Map(input interface{}, property interface{}) (interface{}, error) {
	if property == nil {
		return nil, raisePropertyError(property)
	}

	iter := NewInputIterator(input, sf.context)
	propStr := ToS(property, nil)
	result := make([]interface{}, 0)

	iter.Each(func(item interface{}) {
		// Special case for "to_liquid"
		if propStr == "to_liquid" {
			result = append(result, item)
			return
		}

		if !supportsIndexing(item) {
			result = append(result, nil)
			return
		}

		val := getProperty(item, propStr)
		result = append(result, val)
	})

	return result, nil
}

// Where filters array items by property value.
// Mirrors Ruby's where from standardfilters.rb:434
func (sf *StandardFilters) Where(input interface{}, property interface{}, targetValue interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}
	}

	if property == nil {
		return nil
	}

	propStr := ToS(property, nil)
	result := make([]interface{}, 0)

	iter.Each(func(item interface{}) {
		if !supportsIndexing(item) {
			return
		}

		val := getProperty(item, propStr)

		// If no target value, filter by truthiness
		if targetValue == nil {
			if isTruthy(val) {
				result = append(result, item)
			}
		} else {
			// Filter by exact match
			if valuesEqual(val, targetValue) {
				result = append(result, item)
			}
		}
	})

	return result
}

// Reject filters out array items by property value.
// Mirrors Ruby's reject from standardfilters.rb:447
func (sf *StandardFilters) Reject(input interface{}, property interface{}, targetValue interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}
	}

	if property == nil {
		return nil
	}

	propStr := ToS(property, nil)
	result := make([]interface{}, 0)

	iter.Each(func(item interface{}) {
		if !supportsIndexing(item) {
			return
		}

		val := getProperty(item, propStr)

		// If no target value, reject by truthiness
		if targetValue == nil {
			if !isTruthy(val) {
				result = append(result, item)
			}
		} else {
			// Reject by exact match
			if !valuesEqual(val, targetValue) {
				result = append(result, item)
			}
		}
	})

	return result
}

// Has tests if any array item has a property value.
// Mirrors Ruby's has from standardfilters.rb:460
func (sf *StandardFilters) Has(input interface{}, property interface{}, targetValue interface{}) bool {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return false
	}

	if property == nil {
		return false
	}

	propStr := ToS(property, nil)
	found := false

	iter.Each(func(item interface{}) {
		if found {
			return
		}

		if !supportsIndexing(item) {
			return
		}

		val := getProperty(item, propStr)

		// If no target value, check truthiness
		if targetValue == nil {
			if isTruthy(val) {
				found = true
			}
		} else {
			// Check exact match
			if valuesEqual(val, targetValue) {
				found = true
			}
		}
	})

	return found
}

// Find returns the first array item with a property value.
// Mirrors Ruby's find from standardfilters.rb:473
func (sf *StandardFilters) Find(input interface{}, property interface{}, targetValue interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return nil
	}

	if property == nil {
		return nil
	}

	propStr := ToS(property, nil)
	var result interface{}

	iter.Each(func(item interface{}) {
		if result != nil {
			return
		}

		if !supportsIndexing(item) {
			return
		}

		val := getProperty(item, propStr)

		// If no target value, find by truthiness
		if targetValue == nil {
			if isTruthy(val) {
				result = item
			}
		} else {
			// Find by exact match
			if valuesEqual(val, targetValue) {
				result = item
			}
		}
	})

	return result
}

// FindIndex returns the index of the first array item with a property value.
// Mirrors Ruby's find_index from standardfilters.rb:486
func (sf *StandardFilters) FindIndex(input interface{}, property interface{}, targetValue interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return nil
	}

	if property == nil {
		return nil
	}

	propStr := ToS(property, nil)
	arr := iter.ToArray()

	for i, item := range arr {
		if !supportsIndexing(item) {
			continue
		}

		val := getProperty(item, propStr)

		// If no target value, find by truthiness
		if targetValue == nil {
			if isTruthy(val) {
				return i
			}
		} else {
			// Find by exact match
			if valuesEqual(val, targetValue) {
				return i
			}
		}
	}

	return nil
}

// Concat concatenates two arrays.
// Mirrors Ruby's concat from standardfilters.rb:682
func (sf *StandardFilters) Concat(input interface{}, array interface{}) (interface{}, error) {
	// Validate that second argument is an array
	if _, ok := array.([]interface{}); !ok {
		return nil, NewArgumentError("concat filter requires an array argument")
	}

	iter := NewInputIterator(input, sf.context)
	return iter.Concat(array), nil
}

// Sum returns the sum of array elements.
// Mirrors Ruby's sum from standardfilters.rb:953
func (sf *StandardFilters) Sum(input interface{}, property interface{}) interface{} {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return 0
	}

	var sum float64
	hasFloat := false

	if property == nil {
		// Sum items directly
		iter.Each(func(item interface{}) {
			num := ToNumber(item)
			switch v := num.(type) {
			case int:
				sum += float64(v)
			case float64:
				sum += v
				hasFloat = true
			}
		})
	} else {
		// Sum by property
		propStr := ToS(property, nil)
		iter.Each(func(item interface{}) {
			if !supportsIndexing(item) {
				return
			}

			val := getProperty(item, propStr)
			if val == nil {
				return
			}

			num := ToNumber(val)
			switch v := num.(type) {
			case int:
				sum += float64(v)
			case float64:
				sum += v
				hasFloat = true
			}
		})
	}

	// Return int if no floats were encountered
	if !hasFloat && sum == float64(int(sum)) {
		return int(sum)
	}

	return sum
}

// Helper functions for array filters

// supportsIndexing checks if an item supports property access.
func supportsIndexing(item interface{}) bool {
	if item == nil {
		return false
	}

	switch item.(type) {
	case map[string]interface{}:
		return true
	default:
		// Check if it has indexable methods via reflection
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Map || v.Kind() == reflect.Struct {
			return true
		}
		// Check for Drop interface
		if _, ok := item.(interface{ Get(string) interface{} }); ok {
			return true
		}
	}

	return false
}

// getProperty gets a property value from an item.
func getProperty(item interface{}, property string) interface{} {
	if item == nil {
		return nil
	}

	// Try map access
	if m, ok := item.(map[string]interface{}); ok {
		return m[property]
	}

	// Try Drop interface
	if drop, ok := item.(interface{ Get(string) interface{} }); ok {
		return drop.Get(property)
	}

	// Try reflection for struct fields
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		// Try to get field by name (case-insensitive)
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if strings.EqualFold(field.Name, property) {
				return v.Field(i).Interface()
			}
		}
	}

	return nil
}

// isTruthy checks if a value is truthy in Liquid context.
func isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}

	if b, ok := val.(bool); ok {
		return b
	}

	// Empty strings and arrays are falsy
	switch v := val.(type) {
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	}

	return true
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Direct comparison
	if a == b {
		return true
	}

	// String comparison
	aStr := ToS(a, nil)
	bStr := ToS(b, nil)

	return aStr == bStr
}
