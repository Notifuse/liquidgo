package liquid

import "strings"

// ErrorMode is how templates parse (EnvironmentConfig.ErrorMode).
type ErrorMode string

const (
	// ErrorModeLax tolerates malformed markup, as Ruby's lax parser; the
	// default.
	ErrorModeLax ErrorMode = "lax"
	// ErrorModeStrict makes malformed markup a syntax error.
	ErrorModeStrict ErrorMode = "strict"
	// ErrorModeStrict2 is stricter still, as Shopify Liquid 5.14's strict2.
	ErrorModeStrict2 ErrorMode = "strict2"
	// ErrorModeRigid is strict2's other name, as Ruby's :rigid.
	ErrorModeRigid ErrorMode = "rigid"
	// ErrorModeWarn records a syntax error as a warning, and parses the
	// markup as lax.
	ErrorModeWarn ErrorMode = "warn"
)

// rigidMode reports whether mode is strict2 or rigid, its other name.
func rigidMode(mode ErrorMode) bool {
	return mode == ErrorModeRigid || mode == ErrorModeStrict2
}

// parseWithErrorModeFallback parses markup as Ruby's
// strict_parse_with_error_mode_fallback, which Shopify Liquid's Variable
// still uses: in strict2 the strict2 parse alone; otherwise the strict parse,
// whose syntax error is raised in strict mode, recorded as a warning in warn
// mode, and in either warn or lax mode followed by the lax parse. A syntax
// error gets the line number and the markup context.
func parseWithErrorModeFallback(pc interface {
	ErrorMode() ErrorMode
	AddWarning(error)
}, markup string, lineNumber *int, markupContext func(string) string, strictParse, laxParse, rigidParse func(string) error) error {
	withContext := func(err error) error {
		if syntaxErr, ok := err.(*SyntaxError); ok {
			if lineNumber != nil {
				syntaxErr.Err.LineNumber = lineNumber
			}
			syntaxErr.Err.MarkupContext = markupContext(markup)
		}
		return err
	}
	mode := pc.ErrorMode()
	if rigidMode(mode) {
		return withContext(rigidParse(markup))
	}
	err := withContext(strictParse(markup))
	if err == nil {
		return nil
	}
	if _, ok := err.(*SyntaxError); !ok {
		return err
	}
	switch mode {
	case ErrorModeStrict:
		return err
	case ErrorModeWarn:
		pc.AddWarning(err)
	}
	return laxParse(markup)
}

// MarkupContext returns a context string for markup.
func MarkupContext(markup string) string {
	return "in \"" + strings.TrimSpace(markup) + "\""
}
