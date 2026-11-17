package tags

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

var (
	// tableRowSyntax matches: variable in collection [attributes...]
	// Note: QuotedFragment already has +, so we don't add another
	tableRowSyntax    = regexp.MustCompile(`^(\w+)\s+in\s+(` + liquid.QuotedFragment.String() + `)`)
	allowedAttributes = map[string]bool{
		"cols":   true,
		"limit":  true,
		"offset": true,
		"range":  true,
	}
)

// TableRowTag represents a table_row block tag that generates HTML table rows.
type TableRowTag struct {
	*liquid.Block
	variableName   string
	collectionName interface{}            // Expression
	attributes     map[string]interface{} // Attribute expressions
}

// NewTableRowTag creates a new TableRowTag.
func NewTableRowTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*TableRowTag, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)

	tag := &TableRowTag{
		Block:      block,
		attributes: make(map[string]interface{}),
	}

	// Parse markup
	err := tag.parseMarkup(markup, parseContext)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// parseMarkup parses the table_row tag markup.
func (t *TableRowTag) parseMarkup(markup string, parseContext liquid.ParseContextInterface) error {
	matches := tableRowSyntax.FindStringSubmatch(markup)
	if len(matches) == 0 {
		return liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.table_row", nil)
	}

	t.variableName = matches[1]
	collectionNameStr := matches[2]

	// Parse collection name as expression
	t.collectionName = parseContext.ParseExpression(collectionNameStr)

	// Parse attributes (cols, limit, offset, range)
	attributeMatches := liquid.TagAttributes.FindAllStringSubmatch(markup, -1)
	for _, match := range attributeMatches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			// the value is parsed as Ruby parses it: "a" is a string, not
			// the variable a
			value := strings.TrimSpace(match[2])

			// Validate attribute
			if !allowedAttributes[key] {
				return liquid.NewSyntaxError(fmt.Sprintf("invalid table_row attribute: %s", key))
			}

			t.attributes[key] = parseContext.ParseExpression(value)
		}
	}

	return nil
}

// Parse parses the block body from tokens.
// TableRowTag uses the default Block.Parse() method which handles body parsing
func (t *TableRowTag) Parse(tokenizer *liquid.Tokenizer) error {
	return t.Block.Parse(tokenizer)
}

// VariableName returns the variable name.
func (t *TableRowTag) VariableName() string {
	return t.variableName
}

// CollectionName returns the collection name expression.
func (t *TableRowTag) CollectionName() interface{} {
	return t.collectionName
}

// Attributes returns the attributes map.
func (t *TableRowTag) Attributes() map[string]interface{} {
	return t.attributes
}

// RenderToOutput renders the table_row tag.
func (t *TableRowTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	// Evaluate collection
	// As Ruby, a tablerow over nil or false renders nothing
	collection := context.Evaluate(t.collectionName)
	if collection == nil || collection == false {
		return
	}

	// As Ruby, the attributes are read with to_i: nil is 0, so a nil limit
	// keeps no item and nil cols put every item in one row
	var from int
	if offsetAttr, ok := t.attributes["offset"]; ok {
		from = tablerowInteger(context.Evaluate(offsetAttr))
	}
	var to *int
	if limitAttr, ok := t.attributes["limit"]; ok {
		toVal := addSaturated(from, tablerowInteger(context.Evaluate(limitAttr)))
		to = &toVal
	}

	// Slice collection; a range is iterated, not built
	segment := sliceSegment(collection, from, to)
	length := segment.Len()

	cols := length
	if colsAttr, ok := t.attributes["cols"]; ok {
		cols = tablerowInteger(context.Evaluate(colsAttr))
	}

	// Start first row
	output.WriteString("<tr class=\"row1\">\n")

	// Get underlying Context for Stack
	ctx := context.Context().(*liquid.Context)

	// Create new scope and iterate
	ctx.Stack(make(map[string]interface{}), func() {
		// Create tablerowloop drop
		tablerowloop := liquid.NewTablerowloopDrop(length, cols)
		ctx.Set("tablerowloop", tablerowloop)

		// Iterate over segment
		for i := 0; i < length; i++ {
			ctx.CheckCanceled()
			segment.chargeItem(ctx, t.Nodelist())

			// Set variable
			ctx.Set(t.variableName, segment.At(i))

			// Output <td> tag
			output.WriteString(fmt.Sprintf("<td class=\"col%d\">", tablerowloop.Col()))

			// Render block body (equivalent to Ruby's 'super')
			t.Block.RenderToOutput(context, output)

			// Close </td>
			output.WriteString("</td>")

			// Handle interrupts
			if ctx.Interrupt() {
				interrupt := ctx.PopInterrupt()
				if _, ok := interrupt.(*liquid.BreakInterrupt); ok {
					break
				}
			}

			// Check if we need to close row and start new one
			if tablerowloop.ColLast() && !tablerowloop.Last() {
				output.WriteString(fmt.Sprintf("</tr>\n<tr class=\"row%d\">", tablerowloop.Row()+1))
			}

			// Increment loop
			tablerowloop.Increment()
		}
	})

	// Close last row
	output.WriteString("</tr>\n")
}

// tablerowLeadingInteger is the integer a string starts with, as Ruby's
// String#to_i reads it.
var tablerowLeadingInteger = regexp.MustCompile(`^[ \t\n\v\f\r]*([+-]?[0-9]+(?:_[0-9]+)*)`)

// tablerowInteger converts an attribute of tablerow as Ruby's to_i, which
// the tag calls: nil is 0, a float is truncated, a string gives the integer
// it starts with or 0, and a value without to_i (true, a list...) raises
// invalid integer.
func tablerowInteger(value interface{}) int {
	switch v := value.(type) {
	case nil:
		return 0
	case string:
		m := tablerowLeadingInteger.FindStringSubmatch(v)
		if m == nil {
			return 0
		}
		n, err := strconv.ParseInt(strings.ReplaceAll(m[1], "_", ""), 10, 0)
		if err != nil { // out of range: the bound, as Ruby's integers never overflow
			if strings.HasPrefix(m[1], "-") {
				return math.MinInt
			}
			return math.MaxInt
		}
		return int(n)
	case float64:
		return truncateFloat(v)
	case float32:
		return truncateFloat(float64(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		n, err := liquid.ToInteger(v)
		if err == nil {
			return n
		}
	}
	liquid.Raise(liquid.NewArgumentError("invalid integer"))
	return 0
}

// truncateFloat is Ruby's Float#to_i, saturated to the int range; NaN and the
// infinities raise FloatDomainError in Ruby, an invalid integer here.
func truncateFloat(f float64) int {
	switch {
	case math.IsNaN(f) || math.IsInf(f, 0):
		liquid.Raise(liquid.NewArgumentError("invalid integer"))
	case f >= math.MaxInt:
		return math.MaxInt
	case f <= math.MinInt:
		return math.MinInt
	}
	return int(f)
}
