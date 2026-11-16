package tags

import (
	"regexp"
	"strings"

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
		return nil, liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.tag_unexpected_args", map[string]interface{}{"tag": tagName})
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
	var body strings.Builder
	blockDelimiter := r.BlockDelimiter()

	for !tokenizer.Exhausted() {
		token := tokenizer.Shift()
		if before, tagName, ok := liquid.MatchTagPossiblyInvalid(token); ok && tagName == blockDelimiter {
			r.ParseContext().SetTrimWhitespace(liquid.TrimsAfter(token))
			body.WriteString(before)
			r.body = body.String()
			return nil
		}
		body.WriteString(token)
	}

	// Tag never closed
	r.body = body.String()
	return liquid.NewSyntaxErrorFor(r.ParseContext(), "errors.syntax.tag_never_closed", map[string]interface{}{"block_name": r.BlockName()})
}

// RenderToOutput renders the raw tag.
func (r *RawTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString(r.body)
}

// Nodelist returns the nodelist (just the body as a string).
func (r *RawTag) Nodelist() []interface{} {
	return []interface{}{r.body}
}

// Blank returns true if body is empty.
func (r *RawTag) Blank() bool {
	return r.body == ""
}
