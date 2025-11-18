package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	docNoUnexpectedArgs         = regexp.MustCompile(`^\s*$`)
	docFullTokenPossiblyInvalid = regexp.MustCompile(`^{%-?(\s*)(\w+)(\s*)(.*?)-?%}$`)
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
		var locale *liquid.I18n
		if pc, ok := parseContext.(*liquid.ParseContext); ok {
			locale = pc.Locale()
			msg := locale.Translate("errors.syntax.block_tag_unexpected_args", map[string]interface{}{"tag": tagName})
			return nil, liquid.NewSyntaxError(msg)
		}
		return nil, liquid.NewSyntaxError("Liquid syntax error: block tag unexpected args")
	}

	block := liquid.NewBlock(tagName, markup, parseContext)
	return &DocTag{
		Block: block,
		body:  "",
	}, nil
}

// RenderToOutputBuffer renders the doc tag (does nothing - docs don't render).
func (d *DocTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Docs don't render anything
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

	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}

		// Extract tag name from token
		var foundTagName string
		var leadingWhitespace string

		if strings.HasPrefix(token, "{%") {
			matches := docFullTokenPossiblyInvalid.FindStringSubmatch(token)
			if len(matches) > 0 {
				leadingWhitespace = matches[1]
				foundTagName = matches[2]
			}
		}

		// Raise error if nested doc tag found
		if foundTagName == tagName {
			var locale *liquid.I18n
			if pc, ok := d.ParseContext().(*liquid.ParseContext); ok {
				locale = pc.Locale()
				msg := locale.Translate("errors.syntax.doc_invalid_nested", map[string]interface{}{})
				return liquid.NewSyntaxError(msg)
			}
			return liquid.NewSyntaxError("Liquid syntax error: doc invalid nested")
		}

		if foundTagName == blockDelimiter {
			// Handle whitespace trimming
			if len(token) >= 3 && token[len(token)-3] == '-' {
				d.ParseContext().SetTrimWhitespace(true)
			}
			// Include leading whitespace in body if present
			if leadingWhitespace != "" {
				d.body += leadingWhitespace
			}
			return nil
		}

		if token != "" {
			d.body += token
		}
	}

	// Tag never closed
	return d.RaiseTagNeverClosed()
}

// RaiseTagNeverClosed raises an error for a tag that was never closed.
func (d *DocTag) RaiseTagNeverClosed() error {
	return liquid.NewSyntaxError("tag " + d.BlockName() + " was never closed")
}
