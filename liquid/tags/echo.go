package tags

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// EchoTag represents an echo tag that outputs an expression.
type EchoTag struct {
	*liquid.Tag
	variable *liquid.Variable
}

// NewEchoTag creates a new EchoTag.
func NewEchoTag(tagName, markup string, parseContext liquid.ParseContextInterface) *EchoTag {
	return &EchoTag{
		Tag:      liquid.NewTag(tagName, markup, parseContext),
		variable: liquid.NewVariable(markup, parseContext),
	}
}

// RenderToOutput renders the echo tag by rendering the variable.
func (e *EchoTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	e.variable.RenderToOutput(context, output)
}
