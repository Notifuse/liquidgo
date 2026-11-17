package liquid

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var (
	expressionRangesRegex  = regexp.MustCompile(`^\(\s*(\S+)\s*\.\.\s*(\S+)\s*\)$`)
	expressionIntegerRegex = regexp.MustCompile(`^(-?\d+)$`)
	expressionFloatRegex   = regexp.MustCompile(`^(-?\d+)\.\d+$`)
)

// Expression literals map
var expressionLiterals = map[string]interface{}{
	"nil":   nil,
	"null":  nil,
	"":      nil,
	"true":  true,
	"false": false,
	"blank": "",
	"empty": "",
}

// Expression represents a parsed Liquid expression.
type Expression struct{}

// rubyStrip is Ruby's String#strip: it removes the ASCII whitespace and the
// NUL bytes at both ends, and no other character (Go's strings.TrimSpace also
// removes Unicode spaces and keeps NUL bytes).
func rubyStrip(s string) string {
	return strings.Trim(s, " \t\n\v\f\r\x00")
}

// Parse parses a markup string into an expression value. cache, when set,
// keeps the expressions of one parse; nothing is kept across parses, so that
// the memory a template takes goes with it.
func Parse(markup string, ss *StringScanner, cache map[string]interface{}) interface{} {
	if markup == "" {
		return nil
	}

	markup = rubyStrip(markup)

	// Handle quoted strings (fast path, don't cache)
	if (strings.HasPrefix(markup, `"`) && strings.HasSuffix(markup, `"`)) ||
		(strings.HasPrefix(markup, `'`) && strings.HasSuffix(markup, `'`)) {
		// A lone quote is an empty string, as Ruby's markup[1..-2]
		if len(markup) < 2 {
			return ""
		}
		return markup[1 : len(markup)-1]
	}

	// Check literals (fast path, don't cache)
	if val, ok := expressionLiterals[markup]; ok {
		return val
	}

	// Try local cache first (template-specific)
	if cache != nil {
		if cached, ok := cache[markup]; ok {
			return cached
		}
		result := innerParse(markup, ss, cache)
		cache[markup] = result
		return result
	}

	return innerParse(markup, ss, nil)
}

func innerParse(markup string, ss *StringScanner, cache map[string]interface{}) interface{} {
	// Check for range expressions: (start..end)
	if strings.HasPrefix(markup, "(") && strings.HasSuffix(markup, ")") {
		matches := expressionRangesRegex.FindStringSubmatch(markup)
		if len(matches) == 3 {
			return RangeLookupParse(matches[1], matches[2], ss, cache)
		}
	}

	// Try to parse as number
	if num := parseNumber(markup, ss); num != nil {
		return num
	}

	// Otherwise parse as variable lookup
	return VariableLookupParse(markup, ss, cache)
}

func parseNumber(markup string, ss *StringScanner) interface{} {
	// Check if it's a simple integer or float
	if matches := expressionIntegerRegex.FindStringSubmatch(markup); len(matches) > 0 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			return val
		}
		// beyond an int, an integer stays exact, as Ruby's Integer
		if val, ok := new(big.Int).SetString(matches[1], 10); ok {
			return val
		}
	}

	if matches := expressionFloatRegex.FindStringSubmatch(markup); len(matches) > 0 {
		if val, err := strconv.ParseFloat(markup, 64); err == nil {
			return val
		}
	}

	// More complex number parsing
	if ss == nil {
		ss = NewStringScanner(markup)
	} else {
		ss.SetString(markup)
	}

	// Check first byte
	if ss.EOS() {
		return nil
	}
	byte := ss.PeekByte()

	const (
		dash = '-'
		dot  = '.'
		zero = '0'
		nine = '9'
	)

	// First byte must be a digit or dash
	if byte != dash && (byte < zero || byte > nine) {
		return nil
	}

	if byte == dash {
		ss.ScanByte() // consume dash
		peekedByte := ss.PeekByte()
		// If it starts with a dash, the next byte must be a digit
		if ss.EOS() || peekedByte < zero || peekedByte > nine {
			return nil
		}
		ss.SetPos(ss.Pos() - 1) // back up
	}

	firstDotPos := -1
	numEndPos := -1

	for {
		byte, ok := ss.next()
		if !ok { // the end of the markup; a NUL byte is a byte, which no number has
			break
		}

		if byte != dot && (byte < zero || byte > nine) {
			return nil
		}

		// If we already found the number end, just scan the rest
		if numEndPos >= 0 {
			continue
		}

		if byte == dot {
			if firstDotPos < 0 {
				firstDotPos = ss.Pos()
			} else {
				// Found another dot, number ends here
				numEndPos = ss.Pos() - 1
			}
		}
	}

	if ss.EOS() {
		numEndPos = len(markup)
	}

	// as Ruby, the number is String#to_f of the markup up to its end: the
	// float it starts with ("1..3" is 1.0, "1.2.3" is 1.2)
	if numEndPos >= 0 {
		// Number ends with a number "123.123"
		return rubyToF(markup[0:numEndPos])
	} else if firstDotPos >= 0 {
		// Number ends with a dot "123."
		return rubyToF(markup[0:firstDotPos])
	}

	return nil
}

// rubyLeadingFloat is the float a string starts with, as Ruby's String#to_f
// reads it.
var rubyLeadingFloat = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?`)

// rubyToF is Ruby's String#to_f for a string that starts with a number.
func rubyToF(s string) interface{} {
	val, err := strconv.ParseFloat(rubyLeadingFloat.FindString(s), 64)
	if err != nil {
		return nil
	}
	return val
}
