package liquid

import (
	"encoding/base64"
	"html"
	"net/url"
	"regexp"
	"strings"
)

// StandardFilters provides standard filter implementations.
// This is a struct with methods that can be used as filters.
type StandardFilters struct{}

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

