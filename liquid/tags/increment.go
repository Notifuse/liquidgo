package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// IncrementTag represents an increment tag that creates a counter variable.
type IncrementTag struct {
	*liquid.Tag
	variableName string
}

// NewIncrementTag creates a new IncrementTag.
func NewIncrementTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*IncrementTag, error) {
	if err := checkVariableName(markup, markup, parseContext); err != nil {
		return nil, err
	}
	return &IncrementTag{
		Tag:          liquid.NewTag(tagName, markup, parseContext),
		variableName: strings.TrimSpace(markup),
	}, nil
}

// VariableName returns the variable name.
func (i *IncrementTag) VariableName() string {
	return i.variableName
}

// RenderToOutput renders the increment tag.
func (i *IncrementTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	ctx := context.Context().(*liquid.Context)
	value := ctx.Counter(i.variableName)
	output.WriteString(liquid.ToS(value, nil))
	ctx.SetCounter(i.variableName, value+1)
}
