package liquid

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	expressionRangesRegex = regexp.MustCompile(`^\(\s*(\S+)\s*\.\.\s*(\S+)\s*\)$`)
	expressionIntegerRegex = regexp.MustCompile(`^(-?\d+)$`)
	expressionFloatRegex   = regexp.MustCompile(`^(-?\d+)\.\d+$`)
)

// globalExprCache provides thread-safe caching of parsed expressions across templates.
// Optimization: Avoids re-parsing the same expressions repeatedly when cache parameter is nil.
var globalExprCache sync.Map // map[string]interface{}

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

// SafeParse parses an expression from a parser.
func SafeParse(parser *Parser, ss *StringScanner, cache map[string]interface{}) interface{} {
	expr, err := parser.Expression()
	if err != nil {
		return nil
	}
	return Parse(expr, ss, cache)
}

// Parse parses a markup string into an expression value.
// Optimization: Uses global cache when local cache is nil for better performance across templates.
func Parse(markup string, ss *StringScanner, cache map[string]interface{}) interface{} {
	if markup == "" {
		return nil
	}

	markup = strings.TrimSpace(markup)

	// Handle quoted strings (fast path, don't cache)
	if (strings.HasPrefix(markup, `"`) && strings.HasSuffix(markup, `"`)) ||
		(strings.HasPrefix(markup, `'`) && strings.HasSuffix(markup, `'`)) {
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

	// Use global cache when local cache is nil (cross-template caching)
	if cached, ok := globalExprCache.Load(markup); ok {
		return cached
	}

	result := innerParse(markup, ss, nil)
	globalExprCache.Store(markup, result)
	return result
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
	byte := ss.PeekByte()
	if byte == 0 {
		return nil
	}

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
		peekedByte := ss.PeekByte()
		ss.ScanByte() // consume dash
		peekedByte = ss.PeekByte()
		// If it starts with a dash, the next byte must be a digit
		if peekedByte == 0 || peekedByte < zero || peekedByte > nine {
			return nil
		}
		ss.SetPos(ss.Pos() - 1) // back up
	}

	firstDotPos := -1
	numEndPos := -1

	for {
		byte := ss.ScanByte()
		if byte == 0 {
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

	if numEndPos >= 0 {
		// Number ends with a number "123.123"
		numStr := markup[0:numEndPos]
		if val, err := strconv.ParseFloat(numStr, 64); err == nil {
			return val
		}
	} else if firstDotPos >= 0 {
		// Number ends with a dot "123."
		numStr := markup[0:firstDotPos]
		if val, err := strconv.ParseFloat(numStr, 64); err == nil {
			return val
		}
	}

	return nil
}

