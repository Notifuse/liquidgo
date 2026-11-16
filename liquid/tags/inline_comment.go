package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

var inlineCommentNewlinePattern = regexp.MustCompile(`\n\s*[^#\s]`)

// InlineCommentTag represents an inline comment tag.
type InlineCommentTag struct {
	*liquid.Tag
}

// NewInlineCommentTag creates a new InlineCommentTag.
func NewInlineCommentTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*InlineCommentTag, error) {
	// Check if markup has newlines without # on subsequent lines
	if inlineCommentNewlinePattern.MatchString(markup) {
		// Get locale from parse context if it's a ParseContext struct
		return nil, liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.inline_comment_invalid", nil)
	}

	return &InlineCommentTag{
		Tag: liquid.NewTag(tagName, markup, parseContext),
	}, nil
}

// RenderToOutput renders nothing for inline comments.
func (i *InlineCommentTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	// Do nothing - comments don't render
	_ = context // no-op to register coverage
}

// Blank returns true since comments are blank.
func (i *InlineCommentTag) Blank() bool {
	return true
}
