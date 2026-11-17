package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// checkVariableName checks the name an assign, a capture, an increment or a
// decrement writes to, as Shopify Liquid 5.14 parses it in strict2: one
// identifier (my-var, not a.b, 1abc nor two words). In lax and strict any
// name is kept, as Ruby keeps it; in warn, a name that is not one identifier
// is a warning and the name is kept.
func checkVariableName(name, markup string, parseContext liquid.ParseContextInterface) error {
	mode := parseContext.ErrorMode()
	if mode != liquid.ErrorModeRigid && mode != liquid.ErrorModeStrict2 && mode != liquid.ErrorModeWarn {
		return nil
	}
	p := parseContext.NewParser(strings.TrimSpace(name))
	_, err := p.Consume(":id")
	if err == nil {
		_, err = p.Consume(":end_of_string")
	}
	if err == nil {
		return nil
	}
	if syntaxErr, ok := err.(*liquid.SyntaxError); ok {
		syntaxErr.Err.LineNumber = parseContext.LineNumber()
		syntaxErr.Err.MarkupContext = liquid.MarkupContext(markup)
	}
	if mode == liquid.ErrorModeWarn {
		parseContext.AddWarning(err)
		return nil
	}
	return err
}
