package tags

import (
	"github.com/Notifuse/liquidgo/liquid"
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

// RenderToOutputBuffer renders the ifchanged tag.
// Only renders if the block output is different from the last rendered output.
func (i *IfchangedTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Render block body to temporary buffer
	blockOutput := ""
	i.Block.RenderToOutputBuffer(context, &blockOutput)

	// Get registers
	registers := context.Registers()

	// Get last output from registers
	var lastOutput string
	if lastVal := registers.Get("ifchanged"); lastVal != nil {
		if str, ok := lastVal.(string); ok {
			lastOutput = str
		}
	}

	// Only output if different from last output
	if blockOutput != lastOutput {
		registers.Set("ifchanged", blockOutput)
		*output += blockOutput
	}
}
