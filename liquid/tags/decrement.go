package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// DecrementTag represents a decrement tag that creates a counter variable.
type DecrementTag struct {
	*liquid.Tag
	variableName string
}

// NewDecrementTag creates a new DecrementTag.
func NewDecrementTag(tagName, markup string, parseContext liquid.ParseContextInterface) *DecrementTag {
	return &DecrementTag{
		Tag:          liquid.NewTag(tagName, markup, parseContext),
		variableName: strings.TrimSpace(markup),
	}
}

// VariableName returns the variable name.
func (d *DecrementTag) VariableName() string {
	return d.variableName
}

// RenderToOutput renders the decrement tag.
func (d *DecrementTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	ctx := context.Context().(*liquid.Context)
	value := ctx.Counter(d.variableName) - 1
	ctx.SetCounter(d.variableName, value)
	output.WriteString(liquid.ToS(value, nil))
}
