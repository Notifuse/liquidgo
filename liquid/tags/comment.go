package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	commentLiquidTagToken      = regexp.MustCompile(`^\s*(\w+)\s*(.*?)$`)
	commentFullToken           = regexp.MustCompile(`^{%-?(\s*)(\w+)(\s*)(.*?)-?%}$`)
	commentWhitespaceOrNothing = regexp.MustCompile(`^\s*$`)
)

// CommentTag represents a comment block tag that prevents content from being rendered.
type CommentTag struct {
	*liquid.Block
}

// NewCommentTag creates a new CommentTag.
func NewCommentTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*CommentTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)
	return &CommentTag{
		Block: block,
	}, nil
}

// RenderToOutputBuffer renders the comment tag (does nothing - comments don't render).
func (c *CommentTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Comments don't render anything
	_ = context // no-op to register coverage
}

// UnknownTag handles unknown tags (comments ignore all tags except endcomment).
func (c *CommentTag) UnknownTag(tagName, markup string, tokenizer *liquid.Tokenizer) error {
	// Comments ignore unknown tags
	return nil
}

// Blank returns true since comment tags are blank.
func (c *CommentTag) Blank() bool {
	return true
}

// Parse parses the comment block with special handling for nested comments and raw tags.
func (c *CommentTag) Parse(tokenizer *liquid.Tokenizer) error {
	parseContext := c.ParseContext()

	// Check depth (blockMaxDepth is 100, defined in liquid/block.go)
	if parseContext.Depth() >= 100 {
		return liquid.NewStackLevelError("Nesting too deep")
	}

	parseContext.IncrementDepth()
	defer parseContext.DecrementDepth()

	commentTagDepth := 1

	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}

		// Extract tag name from token
		var tagName string
		if tokenizer.ForLiquidTag() {
			if token == "" || commentWhitespaceOrNothing.MatchString(token) {
				continue
			}

			matches := commentLiquidTagToken.FindStringSubmatch(token)
			if len(matches) == 0 {
				continue
			}
			tagName = matches[1]
		} else {
			if !strings.HasPrefix(token, "{%") {
				continue
			}

			matches := commentFullToken.FindStringSubmatch(token)
			if len(matches) == 0 {
				continue
			}
			tagName = matches[2]
		}

		switch tagName {
		case "raw":
			err := c.parseRawTagBody(tokenizer)
			if err != nil {
				return err
			}
		case "comment":
			commentTagDepth++
		case "endcomment":
			commentTagDepth--
			if commentTagDepth == 0 {
				// Handle whitespace trimming
				if !tokenizer.ForLiquidTag() && len(token) >= 3 {
					if token[len(token)-3] == '-' {
						parseContext.SetTrimWhitespace(true)
					}
				}
				return nil
			}
		}
	}

	// Tag never closed
	return c.RaiseTagNeverClosed()
}

// parseRawTagBody parses the body of a raw tag within a comment.
func (c *CommentTag) parseRawTagBody(tokenizer *liquid.Tokenizer) error {
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}

		// Check for endraw
		if strings.HasPrefix(token, "{%") && strings.HasSuffix(token, "%}") {
			tokenContent := strings.TrimPrefix(token, "{%")
			tokenContent = strings.TrimSuffix(tokenContent, "%}")
			tokenContent = strings.TrimSpace(tokenContent)

			if tokenContent == "endraw" || strings.HasPrefix(tokenContent, "endraw ") {
				return nil
			}
		}
	}

	return liquid.NewSyntaxError("tag raw was never closed")
}
