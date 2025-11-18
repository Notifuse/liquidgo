package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
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

// RenderToOutputBuffer renders the decrement tag.
func (d *DecrementTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Get counter environment (first environment)
	ctx := context.Context().(*liquid.Context)
	environments := ctx.Scopes()
	if len(environments) == 0 {
		environments = []map[string]interface{}{make(map[string]interface{})}
	}

	counterEnv := environments[0]
	value, ok := counterEnv[d.variableName]
	if !ok {
		value = 0
	}

	// Convert to int
	var intValue int
	switch v := value.(type) {
	case int:
		intValue = v
	case float64:
		intValue = int(v)
	default:
		intValue = 0
	}

	// Decrement first, then output
	intValue--
	counterEnv[d.variableName] = intValue
	*output += liquid.ToS(intValue, nil)
}
