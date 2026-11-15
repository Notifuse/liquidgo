package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

var (
	// captureSyntax matches variable names, including quoted strings
	// Matches: var, 'var', "var", this-thing, etc.
	// Capture group 1: the variable name (with or without quotes)
	captureSyntax = regexp.MustCompile(`^("(?:"|[^"])*"|'[^']*'|[\w\-\.\[\]]+)$`)
)

// CaptureTag represents a capture block tag that captures rendered content into a variable.
type CaptureTag struct {
	*liquid.Block
	to string
}

// NewCaptureTag creates a new CaptureTag.
func NewCaptureTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*CaptureTag, error) {
	// Trim whitespace from markup
	markup = strings.TrimSpace(markup)

	// Parse variable name from markup
	matches := captureSyntax.FindStringSubmatch(markup)
	if len(matches) == 0 {
		return nil, liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.capture", nil)
	}

	block := liquid.NewBlock(tagName, markup, parseContext)

	// Extract variable name, stripping quotes if present
	varName := matches[1]
	// Strip surrounding quotes if present (both single and double)
	if len(varName) >= 2 {
		if (varName[0] == '"' && varName[len(varName)-1] == '"') ||
			(varName[0] == '\'' && varName[len(varName)-1] == '\'') {
			varName = varName[1 : len(varName)-1]
		}
	}

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

// RenderToOutputBuffer renders the tag and appends its output to output.
func (c *CaptureTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	liquid.RenderToString(context, output, func(o *liquid.OutputBuffer) { c.RenderToOutput(context, o) })
}
