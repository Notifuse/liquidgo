package liquid

import (
	"fmt"
	"strings"
)

// Parser parses expressions from tokens.
type Parser struct {
	tokens []Token
	p      int // pointer to current location
}

// NewParser creates a new parser from a string scanner or string.
func NewParser(input interface{}) *Parser {
	var ss *StringScanner
	switch v := input.(type) {
	case *StringScanner:
		ss = v
	case string:
		ss = NewStringScanner(v)
	default:
		ss = NewStringScanner(fmt.Sprintf("%v", v))
	}

	tokens, err := Tokenize(ss)
	if err != nil {
		// If tokenization fails, create parser with empty tokens
		tokens = []Token{lexerEOS}
	}

	return &Parser{
		tokens: tokens,
		p:      0,
	}
}

// Jump sets the parser position to the given point.
func (p *Parser) Jump(point int) {
	p.p = point
}

// Consume consumes a token of the given type (or any type if nil).
func (p *Parser) Consume(tokenType interface{}) (string, error) {
	if p.p >= len(p.tokens) {
		return "", NewSyntaxError("Unexpected end of expression")
	}

	token := p.tokens[p.p]
	if tokenType != nil {
		expectedType := tokenType.(string)
		if token[0] != expectedType {
			// Format error message: remove colon prefix from token types for user-friendly message
			expectedStr := expectedType
			if len(expectedStr) > 0 && expectedStr[0] == ':' {
				expectedStr = expectedStr[1:]
			}
			foundStr := fmt.Sprintf("%v", token[0])
			if len(foundStr) > 0 && foundStr[0] == ':' {
				foundStr = foundStr[1:]
			}
			return "", NewSyntaxError(fmt.Sprintf("Expected %s but found %s", expectedStr, foundStr))
		}
	}

	p.p++
	if token[1] == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", token[1]), nil
}

// ConsumeOptional consumes a token if it matches the type, returns false otherwise.
func (p *Parser) ConsumeOptional(tokenType string) (string, bool) {
	if p.p >= len(p.tokens) {
		return "", false
	}

	token := p.tokens[p.p]
	if token[0] != tokenType {
		return "", false
	}

	p.p++
	if token[1] == nil {
		return "", true
	}
	return fmt.Sprintf("%v", token[1]), true
}

// ID checks if the next token is an identifier with the given name.
func (p *Parser) ID(str string) (string, bool) {
	if p.p >= len(p.tokens) {
		return "", false
	}

	token := p.tokens[p.p]
	if token[0] != ":id" {
		return "", false
	}

	tokenValue := fmt.Sprintf("%v", token[1])
	if tokenValue != str {
		return "", false
	}

	p.p++
	return tokenValue, true
}

// Look checks if a token of the given type is at the current position (or ahead).
func (p *Parser) Look(tokenType string, ahead int) bool {
	pos := p.p + ahead
	if pos >= len(p.tokens) {
		return false
	}
	return p.tokens[pos][0] == tokenType
}

// Expression parses an expression.
func (p *Parser) Expression() (string, error) {
	if p.p >= len(p.tokens) {
		return "", NewSyntaxError("Unexpected end of expression")
	}

	token := p.tokens[p.p]
	tokenType := fmt.Sprintf("%v", token[0])

	switch tokenType {
	case ":id":
		str, err := p.Consume(":id")
		if err != nil {
			return "", err
		}
		lookups, err := p.VariableLookups()
		if err != nil {
			return "", err
		}
		return str + lookups, nil
	case ":open_square":
		str, err := p.Consume(":open_square")
		if err != nil {
			return "", err
		}
		expr, err := p.Expression()
		if err != nil {
			return "", err
		}
		str += expr
		closeSquare, err := p.Consume(":close_square")
		if err != nil {
			return "", err
		}
		str += closeSquare
		lookups, err := p.VariableLookups()
		if err != nil {
			return "", err
		}
		return str + lookups, nil
	case ":string", ":number":
		return p.Consume(tokenType)
	case ":open_round":
		_, err := p.Consume(":open_round")
		if err != nil {
			return "", err
		}
		first, err := p.Expression()
		if err != nil {
			return "", err
		}
		_, err = p.Consume(":dotdot")
		if err != nil {
			return "", err
		}
		last, err := p.Expression()
		if err != nil {
			return "", err
		}
		_, err = p.Consume(":close_round")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s..%s)", first, last), nil
	default:
		return "", NewSyntaxError(fmt.Sprintf("%v is not a valid expression", token))
	}
}

// Argument parses an argument (possibly a keyword argument).
func (p *Parser) Argument() (string, error) {
	var b strings.Builder

	// Check for keyword argument (identifier: expression)
	if p.Look(":id", 0) && p.Look(":colon", 1) {
		id, _ := p.Consume(":id")
		colon, _ := p.Consume(":colon")
		b.WriteString(id)
		b.WriteString(colon)
		b.WriteString(" ")
	}

	expr, err := p.Expression()
	if err != nil {
		return "", err
	}
	b.WriteString(expr)
	return b.String(), nil
}

// VariableLookups parses variable lookups (dots and brackets).
func (p *Parser) VariableLookups() (string, error) {
	var b strings.Builder

	for {
		if p.Look(":open_square", 0) {
			open, err := p.Consume(":open_square")
			if err != nil {
				return "", err
			}
			b.WriteString(open)
			expr, err := p.Expression()
			if err != nil {
				return "", err
			}
			b.WriteString(expr)
			closeSquare, err := p.Consume(":close_square")
			if err != nil {
				return "", err
			}
			b.WriteString(closeSquare)
		} else if p.Look(":dot", 0) {
			dot, err := p.Consume(":dot")
			if err != nil {
				return "", err
			}
			b.WriteString(dot)
			id, err := p.Consume(":id")
			if err != nil {
				return "", err
			}
			b.WriteString(id)
		} else {
			break
		}
	}

	return b.String(), nil
}

