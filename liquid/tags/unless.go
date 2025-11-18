package tags

import (
	"github.com/Notifuse/liquidgo/liquid"
)

// UnlessTag represents an unless block tag (opposite of if).
type UnlessTag struct {
	*IfTag
}

// NewUnlessTag creates a new UnlessTag.
func NewUnlessTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*UnlessTag, error) {
	// Unless inherits from If, so we create an IfTag but with "unless" as the tag name
	// The Block needs to have the correct tag name for parsing
	block := liquid.NewBlock(tagName, markup, parseContext)

	// Parse the initial condition (same as if)
	condition, err := parseIfCondition(markup, parseContext)
	if err != nil {
		return nil, err
	}

	// Create attachment for the first block
	attachment := liquid.NewBlockBody()
	condition.Attach(attachment)

	ifTag := &IfTag{
		Block:  block,
		blocks: []ConditionBlock{condition},
	}

	return &UnlessTag{
		IfTag: ifTag,
	}, nil
}

// RenderToOutputBuffer renders the unless tag (negated if).
func (u *UnlessTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Get the underlying Context which implements ConditionContext
	ctx := context.Context().(*liquid.Context)

	// Render the first block (unless condition) if it evaluates to false
	// Then render elsif/else blocks normally
	for i, block := range u.Blocks() {
		result, err := block.Evaluate(ctx)
		if err != nil {
			// Handle error
			errorMsg := context.HandleError(err, nil)
			*output += errorMsg
			return
		}

		// Convert result to liquid value
		resultVal := liquid.ToLiquidValue(result)

		// For the first block (unless), negate the condition
		// For elsif/else blocks, use normal logic
		shouldRender := false
		if i == 0 {
			// First block: render if condition is false
			shouldRender = (resultVal == nil || resultVal == false || resultVal == "")
		} else {
			// Elsif/else blocks: render if condition is true
			shouldRender = (resultVal != nil && resultVal != false && resultVal != "")
		}

		if shouldRender {
			// Render the attachment
			if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
				attachment.RenderToOutputBuffer(context, output)
			}
			return
		}
	}
}
