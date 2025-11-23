package liquid

import (
	"fmt"
	"reflect"
)

const (
	blockMaxDepth = 100
)

// Block represents a block tag (tag with a body).
type Block struct {
	*Tag
	body           *BlockBody
	blockDelimiter string
	blank          bool
}

// NewBlock creates a new block tag.
func NewBlock(tagName, markup string, parseContext ParseContextInterface) *Block {
	return &Block{
		Tag:            NewTag(tagName, markup, parseContext),
		blank:          true,
		blockDelimiter: "end" + tagName,
	}
}

// ParseBlock parses a block tag from tokenizer.
func ParseBlock(tagName, markup string, tokenizer *Tokenizer, parseContext ParseContextInterface) (*Block, error) {
	block := NewBlock(tagName, markup, parseContext)
	err := block.Parse(tokenizer)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// Parse parses the block body.
func (b *Block) Parse(tokenizer *Tokenizer) error {
	b.body = b.newBody()

	for {
		shouldContinue, err := b.parseBody(tokenizer)
		if err != nil {
			return err
		}
		if !shouldContinue {
			break
		}
	}

	return nil
}

// Render renders the block body.
func (b *Block) Render(context TagContext) string {
	if b.body == nil {
		return ""
	}
	return b.body.Render(context)
}

// RenderToOutputBuffer renders the block body to the output buffer.
// If the block has a custom Render method that returns a non-empty string,
// it uses that instead of rendering the body (for backwards compatibility).
// This matches Ruby's behavior where Block#render_to_output_buffer calls render
// if it's been overridden, otherwise renders the body.
func (b *Block) RenderToOutputBuffer(context TagContext, output *string) {
	// Use Tag.RenderToOutputBuffer which calls Render and checks if it returns non-empty
	// Tag.RenderToOutputBuffer calls t.Render(context). If t is *TestBlockTag,
	// it will call TestBlockTag.Render, not Block.Render, because Go's method resolution
	// finds the most specific method.
	//
	// But when Block.RenderToOutputBuffer is called, b is *Block, so calling b.Render
	// will call Block.Render. We need to call Render on the actual type.
	//
	// Solution: Use reflection to get the actual type and call its Render method.
	// The actual type might be *TestBlockTag, which has its own Render method.
	v := reflect.ValueOf(b)
	if v.Kind() == reflect.Ptr {
		// Check if this type has a Render method (it should, since Block has one)
		// MethodByName will find the most specific Render method for the actual type
		renderMethod := v.MethodByName("Render")
		if renderMethod.IsValid() {
			// Call Render - this will call the most specific Render method for the actual type
			results := renderMethod.Call([]reflect.Value{reflect.ValueOf(context)})
			if len(results) > 0 {
				renderResult := results[0].String()
				// Get what body would render for comparison
				bodyResult := ""
				if b.body != nil {
					bodyResult = b.body.Render(context)
				}
				// If Render returns something different from body, it's been overridden
				if renderResult != bodyResult {
					*output += renderResult
					return
				}
			}
		}
	}

	// No override detected or Render returns same as body, render body
	if b.body == nil {
		return
	}
	b.body.RenderToOutputBuffer(context, output)
}

// Blank returns true if the block is blank.
func (b *Block) Blank() bool {
	return b.blank
}

// Nodelist returns the nodelist from the body.
func (b *Block) Nodelist() []interface{} {
	if b.body == nil {
		return []interface{}{}
	}
	return b.body.Nodelist()
}

// BlockName returns the block name.
func (b *Block) BlockName() string {
	return b.TagName()
}

// BlockDelimiter returns the block delimiter (e.g., "endif" for "if").
func (b *Block) BlockDelimiter() string {
	return b.blockDelimiter
}

// SetBlockDelimiter sets the block delimiter.
func (b *Block) SetBlockDelimiter(delimiter string) {
	b.blockDelimiter = delimiter
}

// UnknownTag handles unknown tags encountered during parsing.
func (b *Block) UnknownTag(tagName, markup string, tokenizer *Tokenizer) error {
	return RaiseUnknownTag(tagName, b.BlockName(), b.BlockDelimiter(), b.ParseContext())
}

// RaiseUnknownTag raises an error for an unknown tag.
func RaiseUnknownTag(tag, blockName, blockDelimiter string, parseContext ParseContextInterface) error {
	var locale *I18n
	if pc, ok := parseContext.(*ParseContext); ok {
		locale = pc.Locale()
	} else {
		// Use default locale path
		locale = NewI18n(DefaultLocalePath)
	}

	if tag == "else" {
		msg := locale.Translate("errors.syntax.unexpected_else", map[string]interface{}{
			"block_name": blockName,
		})
		// If translation failed (returns key), use fallback
		if msg == "errors.syntax.unexpected_else" {
			msg = fmt.Sprintf("%s tag does not expect 'else' tag", blockName)
		}
		return NewSyntaxError(msg)
	} else if len(tag) >= 3 && tag[:3] == "end" {
		msg := locale.Translate("errors.syntax.invalid_delimiter", map[string]interface{}{
			"tag":             tag,
			"block_name":      blockName,
			"block_delimiter": blockDelimiter,
		})
		// If translation failed (returns key), use fallback
		if msg == "errors.syntax.invalid_delimiter" {
			msg = fmt.Sprintf("'%s' is not a valid delimiter for %s tags. use %s", tag, blockName, blockDelimiter)
		}
		return NewSyntaxError(msg)
	} else {
		msg := locale.Translate("errors.syntax.unknown_tag", map[string]interface{}{
			"tag": tag,
		})
		// If translation failed (returns key), use fallback
		if msg == "errors.syntax.unknown_tag" {
			msg = fmt.Sprintf("Unknown tag '%s'", tag)
		}
		return NewSyntaxError(msg)
	}
}

// RaiseTagNeverClosed raises an error for a tag that was never closed.
func (b *Block) RaiseTagNeverClosed() error {
	return NewSyntaxError("'" + b.BlockName() + "' tag was never closed")
}

func (b *Block) newBody() *BlockBody {
	return NewBlockBody()
}

func (b *Block) parseBody(tokenizer *Tokenizer) (bool, error) {
	parseContext := b.ParseContext()

	// Check depth
	if parseContext.Depth() >= blockMaxDepth {
		return false, NewStackLevelError("Nesting too deep")
	}

	parseContext.IncrementDepth()
	defer parseContext.DecrementDepth()

	foundEndTag := false
	unknownTagHandler := func(endTagName, endTagMarkup string) bool {
		b.blank = b.blank && b.body.Blank()

		if endTagName == b.blockDelimiter {
			foundEndTag = true
			return false // Stop parsing
		}

		if endTagName == "" {
			// Tag never closed - raise error (matches Ruby: raise_tag_never_closed)
			panic(b.RaiseTagNeverClosed())
		}

		// Unknown tag - let block handle it
		err := b.UnknownTag(endTagName, endTagMarkup, tokenizer)
		return err == nil
	}

	err := b.body.Parse(tokenizer, parseContext, unknownTagHandler)
	if err != nil {
		return false, err
	}

	// If we found the end tag, stop parsing (return false)
	// Otherwise, continue parsing (return true)
	return !foundEndTag, nil
}
