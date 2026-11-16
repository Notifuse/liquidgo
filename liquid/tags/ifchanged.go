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

// RenderToOutput renders the ifchanged tag: its body is rendered apart, as
// Ruby's block_output, and written only when it differs from the last body
// an ifchanged rendered, so that a body that is not kept counts in no limit.
func (i *IfchangedTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	body := output.Scratch()
	i.Block.RenderToOutput(context, body)
	blockOutput := body.String()

	registers := context.Registers()
	var lastOutput string
	if lastVal := registers.Get("ifchanged"); lastVal != nil {
		if str, ok := lastVal.(string); ok {
			lastOutput = str
		}
	}
	if blockOutput == lastOutput {
		return
	}
	registers.Set("ifchanged", blockOutput)
	output.WriteString(blockOutput)
}
