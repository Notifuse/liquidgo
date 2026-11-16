package tags

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// ContinueTag represents a continue tag that causes a for loop to skip to the next iteration.
type ContinueTag struct {
	*liquid.Tag
}

// NewContinueTag creates a new ContinueTag.
func NewContinueTag(tagName, markup string, parseContext liquid.ParseContextInterface) *ContinueTag {
	return &ContinueTag{
		Tag: liquid.NewTag(tagName, markup, parseContext),
	}
}

// RenderToOutput renders the continue tag by pushing a continue interrupt.
func (c *ContinueTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	interrupt := liquid.NewContinueInterrupt()
	context.PushInterrupt(interrupt)
}
