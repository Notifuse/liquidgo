package tags

import (
	"github.com/Notifuse/liquidgo/liquid"
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

// RenderToOutputBuffer renders the continue tag by pushing a continue interrupt.
func (c *ContinueTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	interrupt := liquid.NewContinueInterrupt()
	context.PushInterrupt(interrupt)
}
