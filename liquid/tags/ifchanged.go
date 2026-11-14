package tags

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// IfchangedTag represents an ifchanged block tag that only renders when content changes.
type IfchangedTag struct {
	*liquid.Block
}

// NewIfchangedTag creates a new IfchangedTag.
func NewIfchangedTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*IfchangedTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)
	return &IfchangedTag{
		Block: block,
	}, nil
}

// RenderToOutput renders the ifchanged tag: its body is written, and kept
// only when it differs from the last body an ifchanged rendered.
func (i *IfchangedTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	start := output.Len()
	i.Block.RenderToOutput(context, output)
	blockOutput := output.Since(start)

	registers := context.Registers()
	var lastOutput string
	if lastVal := registers.Get("ifchanged"); lastVal != nil {
		if str, ok := lastVal.(string); ok {
			lastOutput = str
		}
	}
	if blockOutput == lastOutput {
		output.Truncate(start)
		return
	}
	registers.Set("ifchanged", blockOutput)
}

// RenderToOutputBuffer renders the tag and appends its output to output.
func (i *IfchangedTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	liquid.RenderToString(context, output, func(o *liquid.OutputBuffer) { i.RenderToOutput(context, o) })
}
