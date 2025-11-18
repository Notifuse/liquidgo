package liquid

import "strings"

// ParserSwitching provides methods for switching between different parsing modes.
// This is typically embedded in types that need parsing functionality.
type ParserSwitching struct {
	parseContext interface {
		ErrorMode() string
		AddWarning(error)
	}
	lineNumber    *int
	markupContext func(string) string
}

// ParseWithSelectedParser parses markup using the parser selected by error mode.
func (p *ParserSwitching) ParseWithSelectedParser(markup string, strictParse, laxParse, rigidParse func(string) error) error {
	errorMode := p.parseContext.ErrorMode()

	switch errorMode {
	case "rigid":
		return p.rigidParseWithErrorContext(markup, rigidParse)
	case "strict":
		return p.strictParseWithErrorContext(markup, strictParse)
	case "lax":
		return laxParse(markup)
	case "warn":
		err := p.rigidParseWithErrorContext(markup, rigidParse)
		if err != nil {
			if syntaxErr, ok := err.(*SyntaxError); ok {
				p.parseContext.AddWarning(syntaxErr)
				return laxParse(markup)
			}
			return err
		}
		return nil
	default:
		return laxParse(markup)
	}
}

// StrictParseWithErrorModeFallback is deprecated. Use ParseWithSelectedParser instead.
func (p *ParserSwitching) StrictParseWithErrorModeFallback(markup string, strictParse, laxParse, rigidParse func(string) error) error {
	if p.parseContext.ErrorMode() == "rigid" {
		return p.rigidParseWithErrorContext(markup, rigidParse)
	}

	err := p.strictParseWithErrorContext(markup, strictParse)
	if err != nil {
		if syntaxErr, ok := err.(*SyntaxError); ok {
			errorMode := p.parseContext.ErrorMode()
			switch errorMode {
			case "rigid", "strict":
				return err
			case "warn":
				p.parseContext.AddWarning(syntaxErr)
			}
			return laxParse(markup)
		}
		return err
	}
	return nil
}

// RigidMode returns true if error mode is rigid.
func (p *ParserSwitching) RigidMode() bool {
	return p.parseContext.ErrorMode() == "rigid"
}

func (p *ParserSwitching) rigidParseWithErrorContext(markup string, rigidParse func(string) error) error {
	err := rigidParse(markup)
	if err != nil {
		if syntaxErr, ok := err.(*SyntaxError); ok {
			if p.lineNumber != nil {
				syntaxErr.Err.LineNumber = p.lineNumber
			}
			if p.markupContext != nil {
				syntaxErr.Err.MarkupContext = p.markupContext(markup)
			}
			return syntaxErr
		}
		return err
	}
	return nil
}

func (p *ParserSwitching) strictParseWithErrorContext(markup string, strictParse func(string) error) error {
	err := strictParse(markup)
	if err != nil {
		if syntaxErr, ok := err.(*SyntaxError); ok {
			if p.lineNumber != nil {
				syntaxErr.Err.LineNumber = p.lineNumber
			}
			if p.markupContext != nil {
				syntaxErr.Err.MarkupContext = p.markupContext(markup)
			}
			return syntaxErr
		}
		return err
	}
	return nil
}

// MarkupContext returns a context string for markup.
func MarkupContext(markup string) string {
	return "in \"" + strings.TrimSpace(markup) + "\""
}
