package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// captureSyntax is Ruby's /(#{VariableSignature}+)/o: the name is the first
// run of name characters in the markup, so 'var' captures into var.
var captureSyntax = regexp.MustCompile(`(?:\(?[\w\-\.\[\]]\)?)+`)

// CaptureTag represents a capture block tag that captures rendered content into a variable.
type CaptureTag struct {
	*liquid.Block
	to string
}

// NewCaptureTag creates a new CaptureTag.
func NewCaptureTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*CaptureTag, error) {
	varName := captureSyntax.FindString(markup)
	if varName == "" {
		return nil, liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.capture", nil)
	}
	if err := checkVariableName(markup, markup, parseContext); err != nil {
		return nil, err
	}

	block := liquid.NewBlock(tagName, markup, parseContext)

	return &CaptureTag{
		Block: block,
		to:    varName,
	}, nil
}

// To returns the variable name to capture to.
func (c *CaptureTag) To() string {
	return c.to
}

// RenderToOutput renders the capture tag: its body is written to a buffer of
// its own, whose writes count in the assign score, and assigned in the outer
// scope, as Ruby's context.scopes.last[@to] = capture_output.
func (c *CaptureTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	ctx := context.Context().(*liquid.Context)
	captured := liquid.RenderCaptured(context, func(o *liquid.OutputBuffer) { c.Block.RenderToOutput(context, o) })
	if ctx.OutputEscape() {
		// Its outputs are escaped already; the rest is the template's markup.
		ctx.MarkSafeHTML(captured)
	}
	ctx.SetLast(c.to, captured)
}

// Blank returns true since capture tags are blank (they don't output).
func (c *CaptureTag) Blank() bool {
	return true
}
