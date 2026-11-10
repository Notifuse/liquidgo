package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

var rawSyntax = regexp.MustCompile(`^\s*$`)

// RawTag represents a raw tag that outputs Liquid code as text.
type RawTag struct {
	*liquid.Block
	body string
}

// NewRawTag creates a new RawTag.
func NewRawTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*RawTag, error) {
	// Validate markup is empty
	if !rawSyntax.MatchString(markup) {
		var locale *liquid.I18n
		if pc, ok := parseContext.(*liquid.ParseContext); ok {
			locale = pc.Locale()
			msg := locale.Translate("errors.syntax.tag_unexpected_args", map[string]interface{}{"tag": tagName})
			return nil, liquid.NewSyntaxError(msg)
		}
		return nil, liquid.NewSyntaxError("Liquid syntax error: tag unexpected args")
	}

	block := liquid.NewBlock(tagName, markup, parseContext)
	block.SetBlockDelimiter("endraw")

	return &RawTag{
		Block: block,
		body:  "",
	}, nil
}

// Parse parses the raw tag body, up to the first tag named endraw, as Ruby: text before
// that tag in the same token belongs to the body, and the body is kept as written.
func (r *RawTag) Parse(tokenizer *liquid.Tokenizer) error {
	r.body = ""
	blockDelimiter := r.BlockDelimiter()

	for !tokenizer.Exhausted() {
		token := tokenizer.Shift()
		if before, tagName, ok := liquid.MatchTagPossiblyInvalid(token); ok && tagName == blockDelimiter {
			r.ParseContext().SetTrimWhitespace(liquid.TrimsAfter(token))
			r.body += before
			return nil
		}
		r.body += token
	}

	// Tag never closed
	return liquid.NewSyntaxError("Liquid syntax error: " + r.BlockName() + " tag was never closed")
}

// RenderToOutputBuffer renders the raw tag.
func (r *RawTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	*output += r.body
}

// Nodelist returns the nodelist (just the body as a string).
func (r *RawTag) Nodelist() []interface{} {
	return []interface{}{r.body}
}

// Blank returns true if body is empty.
func (r *RawTag) Blank() bool {
	return r.body == ""
}
