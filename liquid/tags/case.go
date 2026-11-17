package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

var (
	// caseSyntax matches: variable
	caseSyntax = regexp.MustCompile(`^(` + liquid.QuotedFragment.String() + `)$`)
	// whenSyntax matches: value [or value2] or value1, value2, value3
	whenSyntax = regexp.MustCompile(`(?s)(` + liquid.QuotedFragment.String() + `)(?:(?:\s+or\s+|\s*,\s*)(` + liquid.QuotedFragment.String() + `.*))?`)
)

// CaseBlock represents a case condition block (when or else).
type CaseBlock interface {
	Evaluate(context liquid.ConditionContext) (bool, error)
	Attachment() interface{}
	Attach(attachment interface{})
	IsElse() bool
}

// caseCondition wraps Condition to implement CaseBlock.
type caseCondition struct {
	*liquid.Condition
}

func (c *caseCondition) IsElse() bool {
	return false
}

// caseElseCondition wraps ElseCondition to implement CaseBlock.
type caseElseCondition struct {
	*liquid.ElseCondition
}

func (c *caseElseCondition) IsElse() bool {
	return true
}

// CaseTag represents a case block tag with when and else blocks.
type CaseTag struct {
	*liquid.Block
	left   interface{} // Expression to compare
	blocks []CaseBlock
	// blank is Ruby's @blank: whether every body parsed, the text before
	// the first when included, is blank.
	blank bool
}

// NewCaseTag creates a new CaseTag.
func NewCaseTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*CaseTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)

	tag := &CaseTag{
		Block:  block,
		blocks: []CaseBlock{},
		blank:  true,
	}

	// Parse markup to get left expression
	err := tag.parseMarkup(markup, parseContext)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// parseMarkup parses the case tag markup.
func (c *CaseTag) parseMarkup(markup string, parseContext liquid.ParseContextInterface) error {
	// Trim whitespace to match Ruby behavior
	markup = strings.TrimSpace(markup)

	matches := caseSyntax.FindStringSubmatch(markup)
	if len(matches) == 0 {
		return liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.case", nil)
	}

	c.left = parseContext.ParseExpression(matches[1])
	return nil
}

// Left returns the left expression.
func (c *CaseTag) Left() interface{} {
	return c.left
}

// Blocks returns the condition blocks.
func (c *CaseTag) Blocks() []CaseBlock {
	return c.blocks
}

// Nodelist returns the nodelist from all blocks.
func (c *CaseTag) Nodelist() []interface{} {
	nodelist := []interface{}{}
	for _, block := range c.blocks {
		if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
			nodelist = append(nodelist, attachment.Nodelist()...)
		}
	}
	return nodelist
}

// UnknownTag handles when and else tags.
func (c *CaseTag) UnknownTag(tagName, markup string, tokenizer *liquid.Tokenizer) error {
	switch tagName {
	case "when":
		return c.recordWhenCondition(markup)
	case "else":
		return c.recordElseCondition(markup)
	default:
		return c.Block.UnknownTag(tagName, markup, tokenizer)
	}
}

// Parse parses the case block with support for when and else.
func (c *CaseTag) Parse(tokenizer *liquid.Tokenizer) error {
	// Ruby: body = case_body = new_body; body = @blocks.last.attachment while parse_body(body, tokens)
	// We start with a dummy body, but when we encounter when/else, we create a new body for that block
	var currentBody *liquid.BlockBody

	for {
		// If we don't have a current body yet, create one (for the first when block)
		if currentBody == nil {
			currentBody = liquid.NewBlockBody()
		}

		currentBlockCount := len(c.blocks)

		shouldContinue, err := c.parseBodyForBlock(tokenizer, currentBody)
		if err != nil {
			return err
		}

		// If a new block was created (when/else), continue parsing with its body
		if len(c.blocks) > currentBlockCount {
			// Get the last block's attachment
			if len(c.blocks) > 0 {
				if attachment, ok := c.blocks[len(c.blocks)-1].Attachment().(*liquid.BlockBody); ok {
					currentBody = attachment
					continue
				}
			}
		}

		// If shouldContinue is false, we found endcase, so stop
		if !shouldContinue {
			break
		}
	}

	// Remove blank strings if block is blank
	if c.Blank() {
		for _, block := range c.blocks {
			if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
				attachment.RemoveBlankStrings()
			}
		}
	}

	return nil
}

// parseBodyForBlock parses the body for a specific block.
func (c *CaseTag) parseBodyForBlock(tokenizer *liquid.Tokenizer, body *liquid.BlockBody) (bool, error) {
	parseContext := c.ParseContext()

	// Check depth during parsing to prevent stack overflow
	if parseContext.Depth() >= 100 {
		return false, liquid.NewStackLevelError("Nesting too deep")
	}

	parseContext.IncrementDepth()
	defer parseContext.DecrementDepth()

	foundEndTag := false
	unknownTagHandler := func(endTagName, endTagMarkup string) bool {
		if endTagName == c.BlockDelimiter() {
			foundEndTag = true
			return false // Stop parsing - found endcase
		}
		if endTagName == "" {
			// Tag never closed - raise error (matches Ruby: raise_tag_never_closed)
			panic(liquid.NewSyntaxErrorFor(c.ParseContext(), "errors.syntax.tag_never_closed", map[string]interface{}{"block_name": c.BlockName()}))
		}

		// Handle when and else
		if endTagName == "when" || endTagName == "else" {
			if err := c.UnknownTag(endTagName, endTagMarkup, tokenizer); err != nil {
				panic(err)
			}
			return false // Stop parsing current block
		}

		// Unknown tag: as Ruby's unknown_tag, a syntax error in every mode
		if err := c.UnknownTag(endTagName, endTagMarkup, tokenizer); err != nil {
			panic(err)
		}
		return true
	}

	err := body.Parse(tokenizer, parseContext, unknownTagHandler)
	if err != nil {
		return false, err
	}
	c.blank = c.blank && body.Blank()

	// Return true if we should continue (didn't find endcase)
	return !foundEndTag, nil
}

// recordWhenCondition records a when condition.
func (c *CaseTag) recordWhenCondition(markup string) error {
	// Parse when conditions (can have multiple values separated by "or" or comma)
	// Ruby: parse_lax_when creates one body and multiple conditions attached to it
	body := liquid.NewBlockBody()

	// As Ruby's parse_lax_when, a when without a condition is a syntax error
	remainingMarkup := strings.TrimSpace(markup)
	for first := true; first || remainingMarkup != ""; first = false {
		matches := whenSyntax.FindStringSubmatch(remainingMarkup)
		if len(matches) == 0 {
			return liquid.NewSyntaxErrorFor(c.ParseContext(), "errors.syntax.case_invalid_when", nil)
		}

		// Parse the value expression
		// blank and empty are compared by their method, as Ruby's
		// Condition.parse_expression
		valueExpr := c.Block.ParseContext().ParseExpression(matches[1])
		if ml := liquid.MethodLiteralOf(matches[1]); ml != nil {
			valueExpr = ml
		}

		// Create condition: left == value
		condition := liquid.NewCondition(c.left, "==", valueExpr)
		condition.Attach(body)

		// Wrap in caseCondition to implement CaseBlock
		caseCond := &caseCondition{Condition: condition}

		// Add to blocks
		c.blocks = append(c.blocks, caseCond)

		// Get remaining markup
		if len(matches) > 2 && matches[2] != "" {
			remainingMarkup = strings.TrimSpace(matches[2])
		} else {
			remainingMarkup = ""
		}
	}

	return nil
}

// recordElseCondition records an else condition.
func (c *CaseTag) recordElseCondition(markup string) error {
	if strings.TrimSpace(markup) != "" {
		return liquid.NewSyntaxErrorFor(c.ParseContext(), "errors.syntax.case_invalid_else", nil)
	}

	block := liquid.NewElseCondition()
	body := liquid.NewBlockBody()
	block.Attach(body)

	// Wrap in caseElseCondition to implement CaseBlock
	caseElse := &caseElseCondition{ElseCondition: block}
	c.blocks = append(c.blocks, caseElse)

	return nil
}

// RenderToOutput renders the case tag.
func (c *CaseTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	ctx := context.Context().(*liquid.Context)
	executeElseBlock := true

	for _, block := range c.blocks {
		if block.IsElse() {
			if executeElseBlock {
				if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
					attachment.RenderToOutput(context, output)
				}
			}
			continue
		}

		result, err := block.Evaluate(ctx)
		if err != nil {
			liquid.Raise(err) // as Ruby: the block body handles it
		}

		// as Ruby, every when that matches renders, and the else only when
		// none before it did
		if result {
			executeElseBlock = false
			if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
				attachment.RenderToOutput(context, output)
			}
		}
	}
}

// Blank returns true if every body of the case is blank, as Ruby's blank?.
func (c *CaseTag) Blank() bool {
	return c.blank
}
