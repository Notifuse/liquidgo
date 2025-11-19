package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	ifSyntax           = regexp.MustCompile(`(` + liquid.QuotedFragment.String() + `)\s*([=!<>a-z_]+)?\s*(` + liquid.QuotedFragment.String() + `)?`)
	ifBooleanOperators = []string{"and", "or"}
)

// ConditionBlock represents a condition with its attachment.
type ConditionBlock interface {
	Evaluate(context liquid.ConditionContext) (bool, error)
	Attachment() interface{}
	Attach(attachment interface{})
}

// IfTag represents an if block tag with support for elsif and else.
type IfTag struct {
	*liquid.Block
	blocks []ConditionBlock
}

// NewIfTag creates a new IfTag.
func NewIfTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*IfTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)

	// Parse the initial condition
	condition, err := parseIfCondition(markup, parseContext)
	if err != nil {
		return nil, err
	}

	// Create attachment for the first block
	attachment := liquid.NewBlockBody()
	condition.Attach(attachment)

	return &IfTag{
		Block:  block,
		blocks: []ConditionBlock{condition},
	}, nil
}

// Blocks returns the condition blocks.
func (i *IfTag) Blocks() []ConditionBlock {
	return i.blocks
}

// Nodelist returns the nodelist from all blocks.
func (i *IfTag) Nodelist() []interface{} {
	nodelist := []interface{}{}
	for _, block := range i.blocks {
		if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
			nodelist = append(nodelist, attachment.Nodelist()...)
		}
	}
	return nodelist
}

// UnknownTag handles elsif and else tags.
func (i *IfTag) UnknownTag(tagName, markup string, tokenizer *liquid.Tokenizer) error {
	if tagName == "elsif" || tagName == "else" {
		return i.pushBlock(tagName, markup)
	}
	return i.Block.UnknownTag(tagName, markup, tokenizer)
}

// Parse parses the if block with support for elsif and else.
// Following Ruby: while parse_body(@blocks.last.attachment, tokens); end
func (i *IfTag) Parse(tokenizer *liquid.Tokenizer) error {
	// Parse blocks in sequence - when elsif/else is found, a new block is created
	// and parsing continues with that new block's attachment
	// Ruby: while parse_body(@blocks.last.attachment, tokens); end
	for {
		currentBlockCount := len(i.blocks)
		currentBlock := i.blocks[currentBlockCount-1]

		shouldContinue, err := i.parseBodyForBlock(tokenizer, currentBlock)
		if err != nil {
			return err
		}

		// If a new block was created (elsif/else), continue parsing it
		// The while loop will call parseBodyForBlock again with the new block
		if len(i.blocks) > currentBlockCount {
			continue
		}

		// If shouldContinue is false, we found endif, so stop
		if !shouldContinue {
			break
		}
	}

	// Remove blank strings if block is blank (Ruby: block.attachment.remove_blank_strings if blank?)
	if i.Blank() {
		for _, block := range i.blocks {
			if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
				attachment.RemoveBlankStrings()
			}
		}
	}

	return nil
}

// parseBodyForBlock parses the body for a specific condition block.
// Returns (shouldContinue, error) where shouldContinue is true if we should continue parsing
// (either more content in this block, or a new elsif/else block was created)
func (i *IfTag) parseBodyForBlock(tokenizer *liquid.Tokenizer, condition ConditionBlock) (bool, error) {
	parseContext := i.ParseContext()
	attachment, ok := condition.Attachment().(*liquid.BlockBody)
	if !ok {
		return false, liquid.NewSyntaxError("invalid attachment for condition block")
	}

	// Check depth
	if parseContext.Depth() >= 100 {
		return false, liquid.NewStackLevelError("Nesting too deep")
	}

	parseContext.IncrementDepth()
	defer parseContext.DecrementDepth()

	foundEndTag := false
	unknownTagHandler := func(endTagName, endTagMarkup string) bool {
		if endTagName == i.BlockDelimiter() {
			foundEndTag = true
			return false // Stop parsing - found endif
		}

		if endTagName == "" {
			// Tag never closed - raise error (matches Ruby: raise_tag_never_closed)
			panic(liquid.NewSyntaxError("Tag was never closed: " + i.BlockName()))
		}

		// Handle elsif and else (Ruby: unknown_tag handles these)
		if endTagName == "elsif" || endTagName == "else" {
			err := i.pushBlock(endTagName, endTagMarkup)
			if err != nil {
				return false
			}
			// New block created - Parse() will detect this and continue
			return false // Stop parsing current block
		}

		// Unknown tag - let block handle it
		err := i.UnknownTag(endTagName, endTagMarkup, tokenizer)
		if err != nil {
			// Raise the error (Ruby: raises SyntaxError immediately)
			panic(err)
		}
		return true
	}

	err := attachment.Parse(tokenizer, parseContext, unknownTagHandler)
	if err != nil {
		return false, err
	}

	// If we found endif, stop parsing (return false)
	// If we created a new block (elsif/else), parseBodyForBlock returns false
	// but Parse() will detect the new block and continue the loop
	// Otherwise, continue parsing this block (return true)
	return !foundEndTag, nil
}

// pushBlock adds a new condition block (elsif or else).
func (i *IfTag) pushBlock(tagName, markup string) error {
	var condition ConditionBlock

	if tagName == "else" {
		condition = liquid.NewElseCondition()
	} else {
		// Parse elsif condition
		elsifCondition, err := parseIfCondition(markup, i.ParseContext())
		if err != nil {
			return err
		}
		condition = elsifCondition
	}

	// Create attachment for the new block
	attachment := liquid.NewBlockBody()
	condition.Attach(attachment)

	i.blocks = append(i.blocks, condition)
	return nil
}

// RenderToOutputBuffer renders the if tag.
func (i *IfTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Get the underlying Context which implements ConditionContext
	ctx := context.Context().(*liquid.Context)

	for _, block := range i.blocks {
		result, err := block.Evaluate(ctx)
		if err != nil {
			// Handle error
			errorMsg := context.HandleError(err, nil)
			*output += errorMsg
			return
		}

		// Convert result to liquid value
		resultVal := liquid.ToLiquidValue(result)

		// Check if condition is true
		if resultVal != nil && resultVal != false && resultVal != "" {
			// Render the attachment
			if attachment, ok := block.Attachment().(*liquid.BlockBody); ok {
				attachment.RenderToOutputBuffer(context, output)
			}
			return
		}
	}
}

// Blank returns true if all blocks are blank.
func (i *IfTag) Blank() bool {
	return i.Block.Blank()
}

// parseIfCondition parses a condition from markup.
// This implements the lax parsing similar to Ruby's lax_parse method.
func parseIfCondition(markup string, parseContext liquid.ParseContextInterface) (*liquid.Condition, error) {
	markup = strings.TrimSpace(markup)

	// Split by 'and' and 'or' operators, keeping track of which operator was used
	// We need to parse from right to left like Ruby does
	parts := splitByBooleanOperators(markup)

	if len(parts) == 0 {
		return nil, liquid.NewSyntaxError("Syntax Error in 'if' - Valid syntax: if [expression]")
	}

	// Parse the rightmost condition first
	condition, err := parseSingleCondition(parts[len(parts)-1].expr, parseContext)
	if err != nil {
		return nil, err
	}

	// Build the condition chain from right to left
	for i := len(parts) - 2; i >= 0; i-- {
		part := parts[i]
		newCondition, err := parseSingleCondition(part.expr, parseContext)
		if err != nil {
			return nil, err
		}

		// Chain with the next operator
		switch part.nextOp {
		case "or":
			newCondition.Or(condition)
		case "and":
			newCondition.And(condition)
		}
		condition = newCondition
	}

	return condition, nil
}

type conditionPart struct {
	expr   string
	nextOp string // "and" or "or" - the operator that follows this expression
}

// splitByBooleanOperators splits markup by 'and' and 'or' operators
func splitByBooleanOperators(markup string) []conditionPart {
	var parts []conditionPart
	var currentExpr strings.Builder
	words := strings.Fields(markup)

	for i := 0; i < len(words); i++ {
		word := words[i]
		// Check if word is a boolean operator
		isBoolOp := false
		for _, op := range ifBooleanOperators {
			if word == op {
				isBoolOp = true
				break
			}
		}

		if isBoolOp {
			// Found an operator
			if currentExpr.Len() > 0 {
				parts = append(parts, conditionPart{
					expr:   strings.TrimSpace(currentExpr.String()),
					nextOp: word,
				})
				currentExpr.Reset()
			}
		} else {
			if currentExpr.Len() > 0 {
				currentExpr.WriteString(" ")
			}
			currentExpr.WriteString(word)
		}
	}

	// Add the last expression
	if currentExpr.Len() > 0 {
		parts = append(parts, conditionPart{
			expr:   strings.TrimSpace(currentExpr.String()),
			nextOp: "",
		})
	}

	return parts
}

// parseSingleCondition parses a single condition (without and/or)
func parseSingleCondition(expr string, parseContext liquid.ParseContextInterface) (*liquid.Condition, error) {
	matches := ifSyntax.FindStringSubmatch(expr)
	if len(matches) == 0 {
		// Try as simple expression
		parsedExpr := parseContext.ParseExpression(expr)
		return liquid.NewCondition(parsedExpr, "", nil), nil
	}

	left := parseContext.ParseExpression(matches[1])
	operator := matches[2]
	var right interface{}
	if len(matches) > 3 && matches[3] != "" {
		right = parseContext.ParseExpression(matches[3])
	}

	return liquid.NewCondition(left, operator, right), nil
}
