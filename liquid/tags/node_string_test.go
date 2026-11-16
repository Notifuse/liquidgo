package tags

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// renderToString renders node to a buffer of the render of context and
// appends what it wrote to output, as the RenderToOutputBuffer method with a
// *string that nodes had before liquid.Node.
func renderToString(node liquid.Node, context liquid.TagContext, output *string) {
	buf := liquid.NewOutputBufferFor(context)
	node.RenderToOutput(context, buf)
	*output += buf.String()
}
