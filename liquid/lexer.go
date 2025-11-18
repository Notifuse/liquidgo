package liquid

import (
	"regexp"
)

// Token represents a lexer token with type and value.
type Token [2]interface{} // [type, value]

var (
	// Lexer constants
	lexerCloseRound                   = Token{":close_round", ")"}
	lexerCloseSquare                  = Token{":close_square", "]"}
	lexerColon                        = Token{":colon", ":"}
	lexerComma                        = Token{":comma", ","}
	lexerComparisonNotEqual           = Token{":comparison", "!="}
	lexerComparisonContains           = Token{":comparison", "contains"}
	lexerComparisonEqual              = Token{":comparison", "=="}
	lexerComparisonGreaterThan        = Token{":comparison", ">"}
	lexerComparisonGreaterThanOrEqual = Token{":comparison", ">="}
	lexerComparisonLessThan           = Token{":comparison", "<"}
	lexerComparisonLessThanOrEqual    = Token{":comparison", "<="}
	lexerComparisonNotEqualAlt        = Token{":comparison", "<>"}
	lexerDash                         = Token{":dash", "-"}
	lexerDot                          = Token{":dot", "."}
	lexerDotDot                       = Token{":dotdot", ".."}
	lexerEOS                          = Token{":end_of_string", nil}
	lexerPipe                         = Token{":pipe", "|"}
	lexerQuestion                     = Token{":question", "?"}
	lexerOpenRound                    = Token{":open_round", "("}
	lexerOpenSquare                   = Token{":open_square", "["}
)

var (
	lexerDoubleStringLiteral = regexp.MustCompile(`"[^"]*"`)
	lexerIdentifier          = regexp.MustCompile(`[a-zA-Z_][\w-]*\??`)
	lexerNumberLiteral       = regexp.MustCompile(`-?\d+(\.\d+)?`)
	lexerSingleStringLiteral = regexp.MustCompile(`'[^']*'`)
	lexerWhitespaceOrNothing = regexp.MustCompile(`\s*`)
)

// Lexer tokenizes expressions.
type Lexer struct{}

// Tokenize tokenizes the input string scanner and returns a slice of tokens.
func (l *Lexer) Tokenize(ss *StringScanner) ([]Token, error) {
	var output []Token

	for !ss.EOS() {
		ss.Skip(lexerWhitespaceOrNothing)

		if ss.EOS() {
			break
		}

		startPos := ss.Pos()
		peeked := ss.PeekByte()

		// Check special characters
		if special := getSpecialToken(peeked); special[0] != nil {
			ss.ScanByte()
			// Special case for ".."
			if special[0] == ":dot" && ss.PeekByte() == '.' {
				ss.ScanByte()
				output = append(output, lexerDotDot)
			} else if special[0] == ":dash" {
				// Special case for negative numbers
				if peekedByte := ss.PeekByte(); isNumberByte(peekedByte) {
					ss.SetPos(ss.Pos() - 1)
					if match := ss.Scan(lexerNumberLiteral); match != "" {
						output = append(output, Token{":number", match})
					}
				} else {
					output = append(output, special)
				}
			} else {
				output = append(output, special)
			}
		} else if subTable := getTwoCharsComparisonToken(peeked); subTable != nil {
			ss.ScanByte()
			peekedByte := ss.PeekByte()
			if peekedByte != 0 {
				if found, ok := subTable[peekedByte]; ok && found[0] != nil {
					output = append(output, found)
					ss.ScanByte()
				} else {
					return nil, raiseSyntaxError(startPos, ss)
				}
			} else {
				return nil, raiseSyntaxError(startPos, ss)
			}
		} else if subTable := getComparisonToken(peeked); subTable != nil {
			ss.ScanByte()
			peekedByte := ss.PeekByte()
			if peekedByte != 0 {
				if found, ok := subTable[peekedByte]; ok && found[0] != nil {
					output = append(output, found)
					ss.ScanByte()
				} else {
					singleToken := getSingleComparisonToken(peeked)
					if singleToken[0] != nil {
						output = append(output, singleToken)
					}
				}
			} else {
				singleToken := getSingleComparisonToken(peeked)
				if singleToken[0] != nil {
					output = append(output, singleToken)
				}
			}
		} else {
			typeAndPattern := getNextMatcherToken(peeked)
			if len(typeAndPattern) > 0 {
				tokenType := typeAndPattern[0].(string)
				pattern := typeAndPattern[1].(*regexp.Regexp)
				if t := ss.Scan(pattern); t != "" {
					// Special case for "contains" - it's a comparison operator unless preceded by a dot
					if tokenType == ":id" && t == "contains" {
						isAfterDot := len(output) > 0 && output[len(output)-1][0] == ":dot"
						if !isAfterDot {
							output = append(output, lexerComparisonContains)
						} else {
							output = append(output, Token{tokenType, t})
						}
					} else {
						output = append(output, Token{tokenType, t})
					}
				} else {
					return nil, raiseSyntaxError(startPos, ss)
				}
			} else {
				return nil, raiseSyntaxError(startPos, ss)
			}
		}
	}

	output = append(output, lexerEOS)
	return output, nil
}

// Tokenize is a convenience function that tokenizes a string scanner.
func Tokenize(ss *StringScanner) ([]Token, error) {
	lexer := &Lexer{}
	return lexer.Tokenize(ss)
}

func getSpecialToken(b byte) Token {
	switch b {
	case '|':
		return lexerPipe
	case '.':
		return lexerDot
	case ':':
		return lexerColon
	case ',':
		return lexerComma
	case '[':
		return lexerOpenSquare
	case ']':
		return lexerCloseSquare
	case '(':
		return lexerOpenRound
	case ')':
		return lexerCloseRound
	case '?':
		return lexerQuestion
	case '-':
		return lexerDash
	default:
		return Token{}
	}
}

func getTwoCharsComparisonToken(b byte) map[byte]Token {
	switch b {
	case '=':
		return map[byte]Token{
			'=': lexerComparisonEqual,
		}
	case '!':
		return map[byte]Token{
			'=': lexerComparisonNotEqual,
		}
	default:
		return nil
	}
}

func getComparisonToken(b byte) map[byte]Token {
	switch b {
	case '<':
		return map[byte]Token{
			'=': lexerComparisonLessThanOrEqual,
			'>': lexerComparisonNotEqualAlt,
		}
	case '>':
		return map[byte]Token{
			'=': lexerComparisonGreaterThanOrEqual,
		}
	default:
		return nil
	}
}

func getSingleComparisonToken(b byte) Token {
	switch b {
	case '<':
		return lexerComparisonLessThan
	case '>':
		return lexerComparisonGreaterThan
	default:
		return Token{}
	}
}

func getNextMatcherToken(b byte) []interface{} {
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' {
		return []interface{}{":id", lexerIdentifier}
	}
	if (b >= '0' && b <= '9') || b == '-' {
		return []interface{}{":number", lexerNumberLiteral}
	}
	if b == '\'' {
		return []interface{}{":string", lexerSingleStringLiteral}
	}
	if b == '"' {
		return []interface{}{":string", lexerDoubleStringLiteral}
	}
	return nil
}

func isNumberByte(b byte) bool {
	return b >= '0' && b <= '9'
}

func raiseSyntaxError(startPos int, ss *StringScanner) error {
	ss.SetPos(startPos)
	char := ss.Getch()
	return NewSyntaxError("Unexpected character " + char)
}
