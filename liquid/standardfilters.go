package liquid

import (
	"encoding/base64"
	"html"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
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
	first, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}
	return string(unicode.ToTitle(first)) + strings.ToLower(s[size:])
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
// As Ruby's /["><']|&(?!([a-zA-Z]+|(#\d+));)/: an & starting a named entity (&amp;) or a
// decimal one (&#39;) is kept; any other & is escaped.
func (sf *StandardFilters) EscapeOnce(input interface{}) string {
	s := ToS(input, nil)
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			if isEntityAt(s, i+1) {
				b.WriteByte('&')
			} else {
				b.WriteString("&amp;")
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// isEntityAt reports whether s[i:] starts with letters or #digits followed by ";".
func isEntityAt(s string, i int) bool {
	j := i
	if j < len(s) && s[j] == '#' {
		j++
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		return j > i+1 && j < len(s) && s[j] == ';'
	}
	for j < len(s) && ((s[j] >= 'a' && s[j] <= 'z') || (s[j] >= 'A' && s[j] <= 'Z')) {
		j++
	}
	return j > i && j < len(s) && s[j] == ';'
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

// Slice returns a substring or series of array items, as Ruby's slice(offset, length):
// a negative offset counts from the end, and an offset or length out of range gives an
// empty result. Anything that is not a list is sliced as text, by characters.
func (sf *StandardFilters) Slice(input interface{}, offset interface{}, length interface{}) interface{} {
	offsetInt, _ := ToInteger(offset)
	lengthInt := 1
	if length != nil {
		lengthInt, _ = ToInteger(length)
	}

	if input != nil {
		if items, ok := input.([]interface{}); ok {
			start, end, ok := sliceBounds(len(items), offsetInt, lengthInt)
			if !ok {
				return []interface{}{}
			}
			return items[start:end]
		}
		val := reflect.ValueOf(input)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			start, end, ok := sliceBounds(val.Len(), offsetInt, lengthInt)
			if !ok {
				return []interface{}{}
			}
			result := make([]interface{}, end-start)
			for i := start; i < end; i++ {
				result[i-start] = val.Index(i).Interface()
			}
			return result
		}
	}

	// The string is walked only as far as the slice reaches: a positive
	// offset needs no count of its characters.
	str := ToS(input, nil)
	if offsetInt < 0 {
		start, end, ok := sliceBounds(utf8.RuneCountInString(str), offsetInt, lengthInt)
		if !ok {
			return ""
		}
		offsetInt, lengthInt = start, end-start
	}
	if lengthInt < 0 {
		return ""
	}
	start, ok := runeOffset(str, offsetInt)
	if !ok {
		return ""
	}
	size, _ := runeOffset(str[start:], lengthInt)
	return asRunes(str[start : start+size])
}

// runeOffset returns the byte index after the first n characters of s, as
// []rune(s) counts them, and whether s has n characters.
func runeOffset(s string, n int) (int, bool) {
	i := 0
	for k := 0; k < n; k++ {
		if i >= len(s) {
			return i, false
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
	}
	return i, true
}

// asRunes returns s as string([]rune(s)) does: each invalid byte becomes
// U+FFFD.
func asRunes(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return string([]rune(s))
}

// sliceBounds applies Ruby's slice(offset, length) rules to a sequence of size n.
func sliceBounds(n, offset, length int) (start, end int, ok bool) {
	if offset < 0 {
		offset += n
	}
	if offset < 0 || offset > n || length < 0 {
		return 0, 0, false
	}
	end = n
	if length < n-offset {
		end = offset + length
	}
	return offset, end, true
}

// Truncate shortens a string to length characters, the ellipsis included (50 by
// default). A length of 0 or less leaves only the ellipsis, and a length that is not an
// integer leaves the input unchanged.
func (sf *StandardFilters) Truncate(input interface{}, length interface{}, truncateString interface{}) string {
	if input == nil {
		return ""
	}
	inputStr := ToS(input, nil)
	lengthInt := 50
	if length != nil {
		var err error
		if lengthInt, err = ToInteger(length); err != nil {
			return inputStr
		}
	}

	truncateStr := "..."
	if truncateString != nil {
		truncateStr = ToS(truncateString, nil)
	}

	// The string is walked only as far as the length: a long input is not
	// copied to count its characters.
	if lengthInt >= 0 {
		if end, ok := runeOffset(inputStr, lengthInt); !ok || end == len(inputStr) {
			return inputStr
		}
	}

	// compared before subtracting, so that a length near the int minimum cannot wrap
	l := 0
	if ellipsis := utf8.RuneCountInString(truncateStr); lengthInt > ellipsis {
		l = lengthInt - ellipsis
	}

	end, _ := runeOffset(inputStr, l)
	return asRunes(inputStr[:end]) + truncateStr
}

// TruncateWords shortens a string to a number of words (15 by default, at least 1).
// Words are separated by ASCII whitespace, as in Ruby's split(" "). A count that is not
// an integer leaves the input unchanged.
func (sf *StandardFilters) TruncateWords(input interface{}, words interface{}, truncateString interface{}) string {
	if input == nil {
		return ""
	}
	inputStr := ToS(input, nil)
	wordsInt := 15
	if words != nil {
		var err error
		if wordsInt, err = ToInteger(words); err != nil {
			return inputStr
		}
	}
	if wordsInt <= 0 {
		wordsInt = 1
	}

	truncateStr := "..."
	if truncateString != nil {
		truncateStr = ToS(truncateString, nil)
	}

	// Ruby gives up on a count beyond a 32-bit integer and keeps the input
	if wordsInt >= math.MaxInt32 {
		return inputStr
	}
	wordList := splitASCIIWhitespaceLimit(inputStr, wordsInt+1)
	if len(wordList) <= wordsInt {
		return inputStr
	}

	result := strings.Join(wordList[:len(wordList)-1], " ")
	return result + truncateStr
}

// splitASCIIWhitespaceLimit is Ruby's split(" ", limit): at most limit fields, the
// last one holding the rest of the string. As in Ruby, a string ending in whitespace
// has an empty last field, so "a b " | truncatewords: 2 is "a b...".
func splitASCIIWhitespaceLimit(s string, limit int) []string {
	var fields []string
	beg, end, skip := 0, 0, true
	count := 1
	for i := 0; i < len(s); i++ {
		c := rune(s[i])
		switch {
		case skip && isASCIIWhitespace(c):
			beg = i + 1
		case skip:
			end, skip = i+1, false
			if limit <= count {
				return append(fields, s[beg:])
			}
		case isASCIIWhitespace(c):
			fields = append(fields, s[beg:end])
			skip, beg = true, i+1
			count++
		default:
			end = i + 1
		}
	}
	if len(s) > 0 {
		fields = append(fields, s[beg:])
	}
	return fields
}

func isASCIIWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r'
}

// splitASCIIWhitespace splits like Ruby's split(" "): on runs of ASCII whitespace,
// ignoring leading and trailing whitespace.
// asciiFieldCount returns the number of strings splitASCIIWhitespace gives.
func asciiFieldCount(s string) int {
	count, inField := 0, false
	for i := 0; i < len(s); i++ {
		space := isASCIIWhitespace(rune(s[i]))
		if !space && !inField {
			count++
		}
		inField = !space
	}
	return count
}

func splitASCIIWhitespace(s string) []string {
	return strings.FieldsFunc(s, isASCIIWhitespace)
}

// Split splits a string into an array of substrings based on a separator, as Ruby's
// split: " " splits on runs of whitespace, "" splits into characters, and trailing
// empty strings are dropped.
func (sf *StandardFilters) Split(input interface{}, pattern interface{}) []interface{} {
	inputStr := ToS(input, nil)
	patternStr := ToS(pattern, nil)
	var parts []string
	if patternStr == " " {
		sf.checkItems(asciiFieldCount(inputStr))
		parts = splitASCIIWhitespace(inputStr)
	} else if inputStr != "" {
		sf.checkItems(splitCount(inputStr, patternStr))
		parts = strings.Split(inputStr, patternStr)
		// Ruby drops trailing empty strings
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
	}
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

var (
	stripHTMLBlocks = regexp.MustCompile(`(?s)<script.*?</script>|<!--.*?-->|<style.*?</style>`)
	stripHTMLTags   = regexp.MustCompile(`(?s)<.*?>`)
)

// StripHTML strips HTML tags from a string.
func (sf *StandardFilters) StripHTML(input interface{}) string {
	s := ToS(input, nil)

	// First remove script/style/comment blocks, then all HTML tags
	result := stripHTMLBlocks.ReplaceAllString(s, "")
	result = stripHTMLTags.ReplaceAllString(result, "")

	return result
}

// First returns the first element of an array, or the start of a range, as
// Range#first, even when the range is empty.
func (sf *StandardFilters) First(input interface{}) interface{} {
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
			return v.Index(0).Interface()
		}
	}
	return nil
}

// Last returns the last element of an array, or the end of a range, as
// Range#last, even when the range is empty.
func (sf *StandardFilters) Last(input interface{}) interface{} {
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
			return v.Index(v.Len() - 1).Interface()
		}
	}
	return nil
}

// Join joins the items of the input with a separator, as Ruby's
// InputIterator#join: nested arrays are flattened, a range gives its
// integers, and any other input is a single item.
func (sf *StandardFilters) Join(input interface{}, separator interface{}) string {
	sep := " "
	if separator != nil {
		sep = ToS(separator, nil)
	}
	if r, ok := input.(*Range); ok {
		return sf.joinRange(r, sep)
	}
	iter := NewInputIterator(input, sf.context)
	parts := make([]string, 0, len(iter.input))
	iter.Each(func(item interface{}) {
		parts = append(parts, ToS(item, nil))
	})
	sf.checkLength(joinedLength(parts, sep))
	return strings.Join(parts, sep)
}

// joinRange joins the integers of a range without building them first, and
// stops at the render length limit.
func (sf *StandardFilters) joinRange(r *Range, sep string) string {
	var b strings.Builder
	for i, n := 0, r.Len(); i < n; i++ {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(strconv.Itoa(r.Start + i))
		sf.checkLength(b.Len())
	}
	return b.String()
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
	sf.checkLength(replacedLength(inputStr, searchStr, replaceStr, strings.Count(inputStr, searchStr)))
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
	if strings.Contains(inputStr, searchStr) {
		sf.checkLength(replacedLength(inputStr, searchStr, replaceStr, 1))
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

	sf.checkLength(replacedLength(inputStr, searchStr, replaceStr, 1))
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
	sf.checkLength(len(inputStr) + len(appendStr))
	return inputStr + appendStr
}

// Prepend adds a string to the beginning.
// Mirrors Ruby's prepend from standardfilters.rb:696
func (sf *StandardFilters) Prepend(input interface{}, str interface{}) string {
	inputStr := ToS(input, nil)
	prependStr := ToS(str, nil)
	sf.checkLength(len(inputStr) + len(prependStr))
	return prependStr + inputStr
}

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
	if divisor.isZero() {
		return nil, NewZeroDivisionError("divided by 0")
	}
	return arithmetic(toLiquidNumber(input), divisor, '/').value(), nil
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
	if num.kind == integerNumber && digits >= 0 {
		return num.value(), nil
	}
	// beyond a float64's range of digits the result is known without computing 10^digits
	if digits > maxRoundDigits {
		return num.value(), nil
	}
	if digits < -maxRoundDigits {
		if num.kind == integerNumber {
			return 0, nil
		}
		return 0.0, nil
	}
	rounded := decimalOf(roundRat(num.rat(), digits))
	if digits == 0 || num.kind == integerNumber {
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
func (sf *StandardFilters) Ceil(input interface{}) (int, error) {
	num := toLiquidNumber(input)
	if num.kind == specialNumber {
		return 0, NewFloatDomainError(ToS(num.f, nil))
	}
	return int(ceilRat(num.rat()).Int64()), nil
}

// Floor rounds a number down to the nearest integer.
// Mirrors Ruby's floor from standardfilters.rb:892
func (sf *StandardFilters) Floor(input interface{}) (int, error) {
	num := toLiquidNumber(input)
	if num.kind == specialNumber {
		return 0, NewFloatDomainError(ToS(num.f, nil))
	}
	return int(floorRat(num.rat()).Int64()), nil
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
	} else if r, ok := input.(*Range); ok {
		// A range is enumerable: its integers, within the limits
		n := r.Len()
		(&StandardFilters{context: context}).checkItems(n)
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
				items[i] = v.Index(i).Interface()
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
	} else if args != nil {
		// Use reflection to handle typed slices
		v := reflect.ValueOf(args)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			for i := 0; i < v.Len(); i++ {
				result = append(result, v.Index(i).Interface())
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

// Reverse reverses the order of items in an array.
// Mirrors Ruby's reverse from standardfilters.rb:523
func (sf *StandardFilters) Reverse(input interface{}) []interface{} {
	iter := NewInputIterator(input, sf.context)
	return iter.Reverse()
}

// Sort sorts items in an array.
// Mirrors Ruby's sort from standardfilters.rb:378
func (sf *StandardFilters) Sort(input interface{}, property interface{}) (interface{}, error) {
	return sf.sortBy(input, property, nilSafeCompare)
}

// SortNatural sorts items case-insensitively.
// Mirrors Ruby's sort_natural from standardfilters.rb:407
func (sf *StandardFilters) SortNatural(input interface{}, property interface{}) (interface{}, error) {
	return sf.sortBy(input, property, nilSafeCasecmp)
}

// sortBy sorts the items of input with compare, by a property when one is
// given. As Ruby, sorting by a property gives nil unless every item has a
// [] method, and an integer item raises the property error once two items
// are compared.
func (sf *StandardFilters) sortBy(input interface{}, property interface{}, compare func(a, b interface{}) int) (interface{}, error) {
	arr := NewInputIterator(input, sf.context).ToArray()
	if len(arr) == 0 {
		return []interface{}{}, nil
	}
	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)
	if property == nil {
		sort.SliceStable(sorted, func(i, j int) bool {
			return compare(sorted[i], sorted[j]) < 0
		})
		return sorted, nil
	}

	keys := make(map[int]interface{}, len(arr))
	failed := false
	for i, item := range arr {
		key, kind := indexItem(item, property)
		if kind == noIndex {
			return nil, nil
		}
		failed = failed || kind == indexError
		keys[i] = key
	}
	if failed && len(arr) > 1 {
		return nil, raisePropertyError(property)
	}
	order := make([]int, len(arr))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return compare(keys[order[i]], keys[order[j]]) < 0
	})
	for i, k := range order {
		sorted[i] = arr[k]
	}
	return sorted, nil
}

// Uniq removes duplicate items from an array.
// Mirrors Ruby's uniq from standardfilters.rb:497
func (sf *StandardFilters) Uniq(input interface{}, property interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}, nil
	}

	if property == nil {
		return iter.Uniq(nil), nil
	}

	// Uniq by property: as Ruby, an item without [] gives nil, and a single
	// item is kept without reading its property, as Array#uniq
	items := iter.ToArray()
	if len(items) == 1 {
		return items, nil
	}
	seen := make(map[string]bool)
	result := make([]interface{}, 0)
	for _, item := range items {
		key, kind := indexItem(ToLiquidValue(item), property)
		switch kind {
		case indexError:
			return nil, raisePropertyError(property)
		case noIndex:
			return nil, nil
		}
		keyStr := ""
		if key != nil {
			keyStr = ToS(key, nil)
		}
		if !seen[keyStr] {
			seen[keyStr] = true
			result = append(result, item)
		}
	}
	return result, nil
}

// Compact removes nil items from an array.
// Mirrors Ruby's compact from standardfilters.rb:557
func (sf *StandardFilters) Compact(input interface{}, property interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}, nil
	}

	if property == nil {
		return iter.Compact(), nil
	}

	// Compact by property - remove items where property is nil; as Ruby, an
	// item without [] gives nil
	result := make([]interface{}, 0)
	for _, item := range iter.ToArray() {
		val, kind := indexItem(item, property)
		switch kind {
		case indexError:
			return nil, raisePropertyError(property)
		case noIndex:
			return nil, nil
		}
		if val != nil {
			result = append(result, item)
		}
	}

	return result, nil
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

	for _, item := range iter.ToArray() {
		// Special case for "to_liquid"
		if propStr == "to_liquid" {
			result = append(result, item)
			continue
		}
		val, kind := indexItem(item, property)
		if kind == indexError {
			return nil, raisePropertyError(property)
		}
		result = append(result, val)
	}

	return result, nil
}

// Where filters array items by property value.
// Mirrors Ruby's where from standardfilters.rb:434
func (sf *StandardFilters) Where(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	result := make([]interface{}, 0)
	return sf.filterArray(input, property, targetValue, []interface{}{}, func(i int, item interface{}, match bool) bool {
		if match {
			result = append(result, item)
		}
		return false
	}, func() interface{} { return result })
}

// Reject filters out array items by property value.
// Mirrors Ruby's reject from standardfilters.rb:447
func (sf *StandardFilters) Reject(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	result := make([]interface{}, 0)
	return sf.filterArray(input, property, targetValue, []interface{}{}, func(i int, item interface{}, match bool) bool {
		if !match {
			result = append(result, item)
		}
		return false
	}, func() interface{} { return result })
}

// Has tests if any array item has a property value.
// Mirrors Ruby's has from standardfilters.rb:460
func (sf *StandardFilters) Has(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	found := false
	return sf.filterArray(input, property, targetValue, false, func(i int, item interface{}, match bool) bool {
		found = match
		return match
	}, func() interface{} { return found })
}

// Find returns the first array item with a property value.
// Mirrors Ruby's find from standardfilters.rb:473
func (sf *StandardFilters) Find(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	var found interface{}
	return sf.filterArray(input, property, targetValue, nil, func(i int, item interface{}, match bool) bool {
		if match {
			found = item
		}
		return match
	}, func() interface{} { return found })
}

// FindIndex returns the index of the first array item with a property value.
// Mirrors Ruby's find_index from standardfilters.rb:486
func (sf *StandardFilters) FindIndex(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	var found interface{}
	return sf.filterArray(input, property, targetValue, nil, func(i int, item interface{}, match bool) bool {
		if match {
			found = i
		}
		return match
	}, func() interface{} { return found })
}

// filterArray is Ruby's filter_array: it gives each item of input to visit,
// in order, with whether its property is truthy, or equal to targetValue
// when one is given, until visit returns true, and then returns result. An
// empty input gives empty. As Ruby, an integer item raises the property
// error, and an item without [] makes the filter give nil.
func (sf *StandardFilters) filterArray(input, property, targetValue, empty interface{}, visit func(i int, item interface{}, match bool) bool, result func() interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)
	if iter.Empty() {
		return empty, nil
	}
	if property == nil {
		return empty, nil
	}
	for i, item := range iter.ToArray() {
		val, kind := indexItem(item, property)
		switch kind {
		case indexError:
			return nil, raisePropertyError(property)
		case noIndex:
			return nil, nil
		}
		// Truthiness and equality stay liquidgo's: an empty string or
		// list is falsy, and a string matches the number it writes
		var match bool
		if targetValue == nil {
			match = isTruthy(val)
		} else {
			match = valuesEqual(val, targetValue)
		}
		if visit(i, item, match) {
			break
		}
	}
	return result(), nil
}

// Concat concatenates two arrays.
// Mirrors Ruby's concat from standardfilters.rb:682
func (sf *StandardFilters) Concat(input interface{}, array interface{}) (interface{}, error) {
	// Validate that second argument is an array
	isArray := false
	if _, ok := array.([]interface{}); ok {
		isArray = true
	} else if array != nil {
		// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
		v := reflect.ValueOf(array)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			isArray = true
		}
	}

	if !isArray {
		return nil, NewArgumentError("concat filter requires an array argument")
	}

	iter := NewInputIterator(input, sf.context)
	sf.checkItems(len(iter.input) + reflectLen(array))
	return iter.Concat(array), nil
}

// Sum returns the sum of array elements.
// Mirrors Ruby's sum from standardfilters.rb:953
func (sf *StandardFilters) Sum(input interface{}, property interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return 0, nil
	}

	sum := integerOf(0)
	for _, item := range iter.ToArray() {
		if property != nil {
			// As Ruby, an item without [] adds 0
			val, kind := indexItem(item, property)
			if kind == indexError {
				return nil, raisePropertyError(property)
			}
			item = val
		}
		sum = arithmetic(sum, toLiquidNumber(item), '+')
	}

	return sum.value(), nil
}

// Helper functions for array filters

// indexKind says how Ruby's item[property] went for an item.
type indexKind int

const (
	// indexed: the item has a [] method that gave a value.
	indexed indexKind = iota
	// noIndex: the item has no [] method, as a float, a boolean or nil.
	noIndex
	// indexError: the item's [] raised a TypeError, as an integer's for a
	// string property.
	indexError
)

// indexItem reads property on item as Ruby's item[property] in the array
// filters. A hash or a drop gives its property. A string gives the property
// when it contains it, as String#[], or its character at an integer
// property. An integer gives its bit at an integer property and raises a
// TypeError for any other property, as Integer#[].
func indexItem(item, property interface{}) (interface{}, indexKind) {
	switch v := item.(type) {
	case nil, bool, float32, float64:
		return nil, noIndex
	case string:
		if sub, ok := property.(string); ok {
			if strings.Contains(v, sub) {
				return sub, indexed
			}
			return nil, indexed
		}
		if n, ok := integerItem(property); ok {
			runes := []rune(v)
			if n < 0 {
				n += int64(len(runes))
			}
			if n < 0 || n >= int64(len(runes)) {
				return nil, indexed
			}
			return string(runes[n]), indexed
		}
		return nil, indexError
	}
	if n, ok := integerItem(item); ok {
		bit, ok := integerItem(property)
		if !ok {
			return nil, indexError
		}
		switch {
		case bit >= 64:
			if n < 0 {
				return 1, indexed
			}
			return 0, indexed
		case bit < 0:
			return 0, indexed
		}
		return int((n >> uint(bit)) & 1), indexed
	}
	if supportsIndexing(item) {
		return getProperty(item, ToS(property, nil)), indexed
	}
	if rv := reflect.ValueOf(item); rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		// Array#[] with a property that is not an integer
		return nil, indexError
	}
	return nil, noIndex
}

// integerItem returns the value of an item of an integer kind.
func integerItem(item interface{}) (int64, bool) {
	rv := reflect.ValueOf(item)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(rv.Uint()), true
	}
	return 0, false
}

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

	// For structs, use InvokeDropOn which handles:
	// - snake_case to CamelCase conversion (comments_count -> CommentsCount)
	// - Method calls
	// - Field access on both pointer and non-pointer structs
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		result := InvokeDropOn(item, property)
		if result != nil {
			return result
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

	// Reflection fallback for typed slices/arrays/maps
	// This matches Ruby's duck-typing behavior: empty collections are falsy
	if val != nil {
		v := reflect.ValueOf(val)
		kind := v.Kind()
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return v.Len() > 0
		}
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

	// Liquid equality: numbers by value, lists and maps element by element
	if liquidEqual(a, b) {
		return true
	}

	// String comparison
	aStr := ToS(a, nil)
	bStr := ToS(b, nil)

	return aStr == bStr
}
