package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	// forSyntax matches: variable in collection [reversed]
	// VariableSegment is a single character pattern, so we use + after it
	// QuotedFragment already has +, so we don't add another
	forSyntax = regexp.MustCompile(`^(` + liquid.VariableSegment.String() + `+)\s+in\s+(` + liquid.QuotedFragment.String() + `)\s*(reversed)?`)
)

// ForTag represents a for loop tag.
type ForTag struct {
	*liquid.Block
	variableName   string
	collectionName interface{} // Expression
	limit          interface{} // Expression or nil
	from           interface{} // Expression, :continue, or nil
	reversed       bool
	name           string // "#{variable_name}-#{collection_name}"
	forBlock       *liquid.BlockBody
	elseBlock      *liquid.BlockBody
}

// NewForTag creates a new ForTag.
func NewForTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*ForTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)

	tag := &ForTag{
		Block:     block,
		from:      nil,
		limit:     nil,
		reversed:  false,
		forBlock:  liquid.NewBlockBody(),
		elseBlock: nil,
	}

	// Parse markup
	err := tag.parseMarkup(markup, parseContext)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// parseMarkup parses the for tag markup.
func (f *ForTag) parseMarkup(markup string, parseContext liquid.ParseContextInterface) error {
	matches := forSyntax.FindStringSubmatch(markup)
	if len(matches) == 0 {
		return liquid.NewSyntaxError("invalid for tag syntax")
	}

	f.variableName = matches[1]
	collectionNameStr := matches[2]
	if len(matches) > 3 && matches[3] == "reversed" {
		f.reversed = true
	}

	// Parse collection name as expression
	f.collectionName = parseContext.ParseExpression(collectionNameStr)
	f.name = f.variableName + "-" + collectionNameStr

	// Parse attributes (limit, offset)
	attributeMatches := liquid.TagAttributes.FindAllStringSubmatch(markup, -1)
	for _, match := range attributeMatches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			// Remove quotes if present
			if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
				value = value[1 : len(value)-1]
			}
			err := f.setAttribute(key, value, parseContext)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// setAttribute sets an attribute (limit or offset).
func (f *ForTag) setAttribute(key, expr string, parseContext liquid.ParseContextInterface) error {
	switch key {
	case "offset":
		if expr == "continue" {
			f.from = "continue" // Special marker
		} else {
			f.from = parseContext.ParseExpression(expr)
		}
	case "limit":
		f.limit = parseContext.ParseExpression(expr)
	}
	return nil
}

// Nodelist returns the nodelist (for_block and optional else_block).
func (f *ForTag) Nodelist() []interface{} {
	if f.elseBlock != nil {
		return []interface{}{f.forBlock, f.elseBlock}
	}
	return []interface{}{f.forBlock}
}

// UnknownTag handles else tags.
func (f *ForTag) UnknownTag(tagName, markup string, tokenizer *liquid.Tokenizer) error {
	if tagName == "else" {
		f.elseBlock = liquid.NewBlockBody()
		return nil
	}
	return f.Block.UnknownTag(tagName, markup, tokenizer)
}

// Parse parses the for block.
func (f *ForTag) Parse(tokenizer *liquid.Tokenizer) error {
	// Parse for block body
	shouldContinue, err := f.parseBody(tokenizer, f.forBlock)
	if err != nil {
		return err
	}

	// If we didn't find endfor, parse else block
	if shouldContinue && f.elseBlock != nil {
		_, err = f.parseBody(tokenizer, f.elseBlock)
		if err != nil {
			return err
		}
	}

	// Remove blank strings if block is blank
	if f.Blank() {
		if f.elseBlock != nil {
			f.elseBlock.RemoveBlankStrings()
		}
		f.forBlock.RemoveBlankStrings()
	}

	return nil
}

// parseBody parses a block body (for_block or else_block).
// Returns (shouldContinue, error) where shouldContinue is true if we should continue parsing (didn't find endfor).
func (f *ForTag) parseBody(tokenizer *liquid.Tokenizer, body *liquid.BlockBody) (bool, error) {
	parseContext := f.ParseContext()

	// Check depth during parsing to prevent stack overflow
	if parseContext.Depth() >= 100 {
		return false, liquid.NewStackLevelError("Nesting too deep")
	}

	parseContext.IncrementDepth()
	defer parseContext.DecrementDepth()

	foundEndTag := false
	unknownTagHandler := func(endTagName, endTagMarkup string) bool {
		if endTagName == f.BlockDelimiter() {
			foundEndTag = true
			return false // Stop parsing
		}
		if endTagName == "" {
			// Tag never closed - raise error (matches Ruby: raise_tag_never_closed)
			panic(liquid.NewSyntaxError("'" + f.BlockName() + "' tag was never closed"))
		}
		if endTagName == "else" {
			// Handle else - UnknownTag will create elseBlock if needed
			err := f.UnknownTag(endTagName, endTagMarkup, tokenizer)
			if err != nil {
				return false
			}
			return false // Stop parsing current block, continue with else block
		}
		// Unknown tag - let block handle it
		return f.UnknownTag(endTagName, endTagMarkup, tokenizer) == nil
	}

	err := body.Parse(tokenizer, parseContext, unknownTagHandler)
	if err != nil {
		return false, err
	}

	// Return true if we should continue (didn't find endfor)
	return !foundEndTag, nil
}

// RenderToOutputBuffer renders the for tag.
func (f *ForTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	segment := f.collectionSegment(context)

	if len(segment) == 0 {
		f.renderElse(context, output)
	} else {
		f.renderSegment(context, output, segment)
	}
}

// collectionSegment gets the segment of the collection to iterate over.
func (f *ForTag) collectionSegment(context liquid.TagContext) []interface{} {
	registers := context.Registers()

	// Get or create offsets map
	var offsets map[string]interface{}
	if offsetsVal := registers.Get("for"); offsetsVal != nil {
		if m, ok := offsetsVal.(map[string]interface{}); ok {
			offsets = m
		} else {
			offsets = make(map[string]interface{})
			registers.Set("for", offsets)
		}
	} else {
		offsets = make(map[string]interface{})
		registers.Set("for", offsets)
	}

	// Calculate from (offset)
	var from int
	if f.from == "continue" {
		// Get from offsets
		if offsetVal, ok := offsets[f.name]; ok {
			if offsetInt, ok := offsetVal.(int); ok {
				from = offsetInt
			}
		}
	} else if f.from != nil {
		fromValue := context.Evaluate(f.from)
		if fromValue == nil {
			from = 0
		} else {
			var err error
			from, err = liquid.ToInteger(fromValue)
			if err != nil {
				from = 0
			}
		}
	}

	// Evaluate collection
	collection := context.Evaluate(f.collectionName)
	if collection == nil {
		return []interface{}{}
	}

	// Convert Range to array if needed
	if r, ok := collection.(*liquid.Range); ok {
		// Convert range to array
		arr := []interface{}{}
		for i := r.Start; i <= r.End; i++ {
			arr = append(arr, i)
		}
		collection = arr
	}

	// Calculate to (limit)
	var to *int
	if f.limit != nil {
		limitValue := context.Evaluate(f.limit)
		if limitValue != nil {
			limitInt, err := liquid.ToInteger(limitValue)
			if err == nil {
				toVal := from + limitInt
				to = &toVal
			}
		}
	}

	// Slice collection
	segment := liquid.SliceCollection(collection, from, to)

	// Reverse if needed
	if f.reversed {
		for i, j := 0, len(segment)-1; i < j; i, j = i+1, j-1 {
			segment[i], segment[j] = segment[j], segment[i]
		}
	}

	// Store offset for continue
	offsets[f.name] = from + len(segment)

	return segment
}

// renderSegment renders the segment.
func (f *ForTag) renderSegment(context liquid.TagContext, output *string, segment []interface{}) {
	registers := context.Registers()

	// Get or create for_stack
	var forStack []*liquid.ForloopDrop
	if stackVal := registers.Get("for_stack"); stackVal != nil {
		if stack, ok := stackVal.([]*liquid.ForloopDrop); ok {
			forStack = stack
		} else {
			forStack = []*liquid.ForloopDrop{}
			registers.Set("for_stack", forStack)
		}
	} else {
		forStack = []*liquid.ForloopDrop{}
		registers.Set("for_stack", forStack)
	}

	length := len(segment)

	// Get parent loop (if any)
	var parentLoop *liquid.ForloopDrop
	if len(forStack) > 0 {
		parentLoop = forStack[len(forStack)-1]
	}

	// Create forloop drop
	loopVars := liquid.NewForloopDrop(f.name, length, parentLoop)

	// Push to stack
	forStack = append(forStack, loopVars)
	registers.Set("for_stack", forStack)

	// Get underlying Context for Stack
	ctx := context.Context().(*liquid.Context)

	// Create new scope and iterate
	ctx.Stack(make(map[string]interface{}), func() {
		// Set forloop in context
		ctx.Set("forloop", loopVars)

		// Iterate over segment
	forLoop:
		for _, item := range segment {
			// Set variable
			ctx.Set(f.variableName, item)

			// Render for block
			f.forBlock.RenderToOutputBuffer(context, output)

			// Increment loop
			loopVars.Increment()

			// Handle interrupts
			if ctx.Interrupt() {
				interrupt := ctx.PopInterrupt()
				switch interrupt.(type) {
				case *liquid.BreakInterrupt:
					break forLoop
				case *liquid.ContinueInterrupt:
					continue forLoop
				}
			}
		}
	})

	// Pop from stack
	forStack = forStack[:len(forStack)-1]
	registers.Set("for_stack", forStack)
}

// renderElse renders the else block if collection is empty.
func (f *ForTag) renderElse(context liquid.TagContext, output *string) {
	if f.elseBlock != nil {
		f.elseBlock.RenderToOutputBuffer(context, output)
	}
}

// VariableName returns the variable name.
func (f *ForTag) VariableName() string {
	return f.variableName
}

// CollectionName returns the collection name expression.
func (f *ForTag) CollectionName() interface{} {
	return f.collectionName
}

// Limit returns the limit expression.
func (f *ForTag) Limit() interface{} {
	return f.limit
}

// From returns the from/offset expression.
func (f *ForTag) From() interface{} {
	return f.from
}
