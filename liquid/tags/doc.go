package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

var (
	docNoUnexpectedArgs = regexp.MustCompile(`^\s*$`)
)

// DocTag represents a doc block tag that stores documentation without rendering it.
type DocTag struct {
	*liquid.Block
	body string
}

// NewDocTag creates a new DocTag.
func NewDocTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*DocTag, error) {
	// Validate markup is empty
	if !docNoUnexpectedArgs.MatchString(markup) {
		return nil, liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.block_tag_unexpected_args", map[string]interface{}{"tag": tagName})
	}

	block := liquid.NewBlock(tagName, markup, parseContext)
	return &DocTag{
		Block: block,
		body:  "",
	}, nil
}

// RenderToOutput renders the doc tag (does nothing - docs don't render).
func (d *DocTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	// Docs don't render anything
	_ = context // no-op to register coverage
}

// Blank returns true if body is empty.
func (d *DocTag) Blank() bool {
	return d.body == ""
}

// Nodelist returns the nodelist (just the body as a string).
func (d *DocTag) Nodelist() []interface{} {
	return []interface{}{d.body}
}

// Parse parses the doc block body.
func (d *DocTag) Parse(tokenizer *liquid.Tokenizer) error {
	d.body = ""
	blockDelimiter := d.BlockDelimiter()
	tagName := d.TagName()

	for !tokenizer.Exhausted() {
		token := tokenizer.Shift()
		before, foundTagName, _ := liquid.MatchTagPossiblyInvalid(token)

		// Raise error if nested doc tag found
		if foundTagName == tagName {
			return liquid.NewSyntaxErrorFor(d.ParseContext(), "errors.syntax.doc_invalid_nested", nil)
		}

		if foundTagName == blockDelimiter {
			d.ParseContext().SetTrimWhitespace(liquid.TrimsAfter(token))
			d.body += before
			return nil
		}

		d.body += token
	}

	// Tag never closed
	return d.RaiseTagNeverClosed()
}

// RaiseTagNeverClosed raises an error for a tag that was never closed.
func (d *DocTag) RaiseTagNeverClosed() error {
	return liquid.NewSyntaxError("'" + d.BlockName() + "' tag was never closed")
}
