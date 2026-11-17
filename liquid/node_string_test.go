package liquid

import ()

// renderToString renders node to a buffer of the render of context and
// appends what it wrote to output, as the RenderToOutputBuffer method with a
// *string that nodes had before Node. A Variable's handled an error as a
// block body does, and so does this.
func renderToString(node Node, context TagContext, output *string) {
	buf := NewOutputBufferFor(context)
	func() {
		if v, ok := node.(*Variable); ok {
			defer func() {
				if r := recover(); r != nil {
					rescueRenderNode(context, buf, v, r, false)
				}
			}()
		}
		node.RenderToOutput(context, buf)
	}()
	*output += buf.String()
}

// renderString renders node to a buffer of the render of context and returns
// what it wrote, as the Render method that blocks and documents had.
func renderString(node Node, context TagContext) string {
	buf := NewOutputBufferFor(context)
	node.RenderToOutput(context, buf)
	return buf.String()
}

// silentTag is a tag that writes nothing, as a Ruby tag that does not
// override render: Tag has no render method of its own.
type silentTag struct{ *Tag }

func (s *silentTag) RenderToOutput(TagContext, *OutputBuffer) {}
