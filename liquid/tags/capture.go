package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
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
		var locale *liquid.I18n
		if pc, ok := parseContext.(*liquid.ParseContext); ok {
			locale = pc.Locale()
			msg := locale.Translate("errors.syntax.capture", map[string]interface{}{})
			return nil, liquid.NewSyntaxError(msg)
		}
		return nil, liquid.NewSyntaxError("Liquid syntax error: capture")
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

// RenderToOutputBuffer renders the capture tag.
// Following Ruby implementation: always uses resource_limits.with_capture,
// renders the block body, and assigns to context.scopes.last
func (c *CaptureTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	ctx := context.Context().(*liquid.Context)
	rl := context.ResourceLimits()

	// Use WithCapture to track resource limits during capture (like Ruby: context.resource_limits.with_capture)
	// Ruby always calls with_capture, so we do too (resource_limits should always exist)
	if rl != nil {
		rl.WithCapture(func() {
			captureOutput := c.Render(context)
			// Increment write score with captured output (like Ruby: increment_write_score is called in block_body)
			// This will increment assign_score by the byte difference since lastCaptureLength is set
			rl.IncrementWriteScore(captureOutput)
			// Set in the last scope (outermost scope, matching Ruby's context.scopes.last[@to] = capture_output)
			ctx.SetLast(c.to, captureOutput)
		})
	} else {
		// Fallback if resource_limits is nil (shouldn't happen in normal usage)
		captureOutput := c.Render(context)
		ctx.SetLast(c.to, captureOutput)
	}

	// Ruby returns output unchanged (doesn't modify it)
	// In Go, we don't modify output, so this is correct
}

// Blank returns true since capture tags are blank (they don't output).
func (c *CaptureTag) Blank() bool {
	return true
}
