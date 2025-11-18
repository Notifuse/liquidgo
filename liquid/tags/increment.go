package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

// IncrementTag represents an increment tag that creates a counter variable.
type IncrementTag struct {
	*liquid.Tag
	variableName string
}

// NewIncrementTag creates a new IncrementTag.
func NewIncrementTag(tagName, markup string, parseContext liquid.ParseContextInterface) *IncrementTag {
	return &IncrementTag{
		Tag:          liquid.NewTag(tagName, markup, parseContext),
		variableName: strings.TrimSpace(markup),
	}
}

// VariableName returns the variable name.
func (i *IncrementTag) VariableName() string {
	return i.variableName
}

// RenderToOutputBuffer renders the increment tag.
func (i *IncrementTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Get counter environment (first environment)
	ctx := context.Context().(*liquid.Context)
	environments := ctx.Scopes()
	if len(environments) == 0 {
		environments = []map[string]interface{}{make(map[string]interface{})}
	}

	counterEnv := environments[0]
	value, ok := counterEnv[i.variableName]
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

	// Output current value, then increment
	*output += liquid.ToS(intValue, nil)
	counterEnv[i.variableName] = intValue + 1
}
