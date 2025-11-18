package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
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

// Parse parses the raw tag body.
func (r *RawTag) Parse(tokenizer *liquid.Tokenizer) error {
	r.body = ""
	blockDelimiter := r.BlockDelimiter()

	for {
		token := tokenizer.Shift()
		if token == "" {
			// Tag never closed
			return liquid.NewSyntaxError("Liquid syntax error: " + r.BlockName() + " tag was never closed")
		}

		// Check if this is the end tag
		if strings.HasPrefix(token, "{%") && strings.HasSuffix(token, "%}") {
			// Extract tag name from token
			// Token format: {% endraw %} or {%- endraw -%}
			tokenContent := strings.TrimPrefix(token, "{%")
			tokenContent = strings.TrimSuffix(tokenContent, "%}")
			tokenContent = strings.TrimSpace(tokenContent)

			// Check if it matches the block delimiter
			if tokenContent == blockDelimiter || strings.HasPrefix(tokenContent, blockDelimiter+" ") {
				// This is the end tag
				// Check for whitespace control
				if len(token) > 2 && token[2] == '-' {
					// Trim whitespace from body
					r.body = strings.TrimRight(r.body, " \t\n\r")
				}
				return nil
			}
		}

		if token != "" {
			r.body += token
		}
	}
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
