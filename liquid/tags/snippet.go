package tags

import (
	"github.com/Notifuse/liquidgo/liquid"
)

// SnippetTag represents a snippet block tag that creates an inline snippet.
type SnippetTag struct {
	*liquid.Block
	to string // Variable name to assign snippet to
}

// NewSnippetTag creates a new SnippetTag.
func NewSnippetTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*SnippetTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)

	// Parse markup - should be just a variable name
	// Ruby: p.consume(:id)
	// For now, use simple parsing - expect just an identifier
	if markup == "" {
		return nil, liquid.NewSyntaxError("snippet tag requires a variable name")
	}

	// Extract variable name (simple for now - can be enhanced with parser later)
	to := markup

	return &SnippetTag{
		Block: block,
		to:    to,
	}, nil
}

// To returns the variable name.
func (s *SnippetTag) To() string {
	return s.to
}

// Blank returns true (snippet tag is always blank).
func (s *SnippetTag) Blank() bool {
	return true
}

// RenderToOutputBuffer renders the snippet tag.
func (s *SnippetTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	ctx := context.Context().(*liquid.Context)

	// Render block body to get snippet content
	bodyOutput := ""
	s.Block.RenderToOutputBuffer(context, &bodyOutput)

	// Get template name
	templateName := ctx.TemplateName()

	// Create snippet drop
	snippetDrop := liquid.NewSnippetDrop(bodyOutput, s.to, templateName)

	// Assign to variable in current scope (last scope)
	// Use Set which sets in the first scope (current/local scope)
	ctx.Set(s.to, snippetDrop)

	// Increment assign score in resource limits
	rl := context.ResourceLimits()
	if rl != nil {
		// Calculate assign score (sum of body bytes)
		assignScore := len(bodyOutput)
		rl.IncrementAssignScore(assignScore)
	}
}
