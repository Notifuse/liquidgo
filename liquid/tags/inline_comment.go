package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/liquid"
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
		var locale *liquid.I18n
		if pc, ok := parseContext.(*liquid.ParseContext); ok {
			locale = pc.Locale()
			msg := locale.Translate("errors.syntax.inline_comment_invalid", map[string]interface{}{})
			return nil, liquid.NewSyntaxError(msg)
		}
		return nil, liquid.NewSyntaxError("Liquid syntax error: inline comment invalid")
	}

	return &InlineCommentTag{
		Tag: liquid.NewTag(tagName, markup, parseContext),
	}, nil
}

// RenderToOutputBuffer renders nothing for inline comments.
func (i *InlineCommentTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Do nothing - comments don't render
}

// Blank returns true since comments are blank.
func (i *InlineCommentTag) Blank() bool {
	return true
}
