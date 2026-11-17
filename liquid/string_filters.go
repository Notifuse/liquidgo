package liquid

import (
	"math"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

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

// Slice returns a substring or series of array items, as Ruby's slice(offset, length):
// a negative offset counts from the end, and an offset or length out of range gives an
// empty result. Anything that is not a list is sliced as text, by characters.
func (sf *StandardFilters) Slice(input interface{}, offset interface{}, length interface{}) (interface{}, error) {
	// as Ruby's Utils.to_integer: an offset or length that is not an
	// integer is an error
	offsetInt, err := ToInteger(offset)
	if err != nil {
		return nil, err
	}
	lengthInt := 1
	if length != nil {
		if lengthInt, err = ToInteger(length); err != nil {
			return nil, err
		}
	}

	if input != nil {
		if items, ok := input.([]interface{}); ok {
			start, end, ok := sliceBounds(len(items), offsetInt, lengthInt)
			if !ok {
				return []interface{}{}, nil
			}
			return items[start:end], nil
		}
		val := reflect.ValueOf(input)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			start, end, ok := sliceBounds(val.Len(), offsetInt, lengthInt)
			if !ok {
				return []interface{}{}, nil
			}
			result := make([]interface{}, end-start)
			for i := start; i < end; i++ {
				result[i-start] = itemAt(val, i)
			}
			return result, nil
		}
	}

	// The string is walked only as far as the slice reaches: a positive
	// offset needs no count of its characters.
	str := ToS(input, nil)
	if offsetInt < 0 {
		start, end, ok := sliceBounds(utf8.RuneCountInString(str), offsetInt, lengthInt)
		if !ok {
			return "", nil
		}
		offsetInt, lengthInt = start, end-start
	}
	if lengthInt < 0 {
		return "", nil
	}
	start, ok := runeOffset(str, offsetInt)
	if !ok {
		return "", nil
	}
	size, _ := runeOffset(str[start:], lengthInt)
	return asRunes(str[start : start+size]), nil
}

// runeOffset returns the byte index after the first n characters of s, as
// []rune(s) counts them, and whether s has n characters.
func runeOffset(s string, n int) (int, bool) {
	i := 0
	for k := 0; k < n; k++ {
		if i >= len(s) {
			return i, false
		}
		if s[i] < utf8.RuneSelf { // an ASCII byte is a character
			i++
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
	}
	return i, true
}

// runePrefix returns the byte index after the first n characters of s, as
// []rune(s) counts them, whether s has n characters, and whether those
// characters are valid UTF-8, which asRunes then need not check again.
func runePrefix(s string, n int) (end int, ok, valid bool) {
	valid = true
	for k := 0; k < n; k++ {
		if end >= len(s) {
			return end, false, valid
		}
		if s[end] < utf8.RuneSelf { // an ASCII byte is a character
			end++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[end:])
		if r == utf8.RuneError && size == 1 {
			valid = false
		}
		end += size
	}
	return end, true, valid
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

	// compared before subtracting, so that a length near the int minimum cannot wrap
	l := 0
	if ellipsis := utf8.RuneCountInString(truncateStr); lengthInt > ellipsis {
		l = lengthInt - ellipsis
	}

	// The string is walked once, and only as far as the length: to the cut,
	// then to the length, which a shorter input keeps whole. A long input is
	// not copied to count its characters.
	end, ok, valid := runePrefix(inputStr, l)
	if lengthInt >= 0 {
		if !ok {
			return inputStr
		}
		if rest, ok := runeOffset(inputStr[end:], lengthInt-l); !ok || end+rest == len(inputStr) {
			return inputStr
		}
	}
	if valid {
		return inputStr[:end] + truncateStr
	}
	return string([]rune(inputStr[:end])) + truncateStr
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

// squishSpaces is Ruby's /\s+/: runs of ASCII whitespace.
var squishSpaces = regexp.MustCompile(`[ \t\n\v\f\r]+`)

// Squish removes the whitespace at both ends of the input and turns each run
// of whitespace in it into one space (Shopify Liquid 5.14). nil gives nil.
func (sf *StandardFilters) Squish(input interface{}) interface{} {
	if input == nil {
		return nil
	}
	return squishSpaces.ReplaceAllString(rubyStrip(ToS(input, nil)), " ")
}
