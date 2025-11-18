package tags

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
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
		return liquid.NewSyntaxError("invalid table_row tag syntax")
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
			value := strings.TrimSpace(match[2])
			// Remove quotes if present
			if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
				value = value[1 : len(value)-1]
			}

			// Validate attribute
			if !allowedAttributes[key] {
				return liquid.NewSyntaxError(fmt.Sprintf("invalid table_row attribute: %s", key))
			}

			t.attributes[key] = parseContext.ParseExpression(value)
		}
	}

	return nil
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

// RenderToOutputBuffer renders the table_row tag.
func (t *TableRowTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Evaluate collection
	collection := context.Evaluate(t.collectionName)
	if collection == nil {
		*output += "<tr class=\"row1\">\n</tr>\n"
		return
	}

	// Calculate from (offset)
	var from int
	if offsetAttr, ok := t.attributes["offset"]; ok {
		offsetValue := context.Evaluate(offsetAttr)
		if offsetValue != nil {
			offsetInt, err := liquid.ToInteger(offsetValue)
			if err != nil {
				errorMsg := context.HandleError(liquid.NewArgumentError("invalid integer"), nil)
				*output += errorMsg
				return
			}
			from = offsetInt
		}
	}

	// Calculate to (limit)
	var to *int
	if limitAttr, ok := t.attributes["limit"]; ok {
		limitValue := context.Evaluate(limitAttr)
		if limitValue != nil {
			limitInt, err := liquid.ToInteger(limitValue)
			if err != nil {
				errorMsg := context.HandleError(liquid.NewArgumentError("invalid integer"), nil)
				*output += errorMsg
				return
			}
			toVal := from + limitInt
			to = &toVal
		}
	}

	// Slice collection
	segment := liquid.SliceCollection(collection, from, to)
	length := len(segment)

	// Calculate cols
	var cols int
	if colsAttr, ok := t.attributes["cols"]; ok {
		colsValue := context.Evaluate(colsAttr)
		if colsValue != nil {
			colsInt, err := liquid.ToInteger(colsValue)
			if err != nil {
				errorMsg := context.HandleError(liquid.NewArgumentError("invalid integer"), nil)
				*output += errorMsg
				return
			}
			cols = colsInt
		} else {
			cols = length
		}
	} else {
		cols = length
	}

	// Start first row
	*output += "<tr class=\"row1\">\n"

	// Get underlying Context for Stack
	ctx := context.Context().(*liquid.Context)

	// Create new scope and iterate
	ctx.Stack(make(map[string]interface{}), func() {
		// Create tablerowloop drop
		tablerowloop := liquid.NewTablerowloopDrop(length, cols)
		ctx.Set("tablerowloop", tablerowloop)

		// Iterate over segment
		for _, item := range segment {
			// Set variable
			ctx.Set(t.variableName, item)

			// Output <td> tag
			*output += fmt.Sprintf("<td class=\"col%d\">", tablerowloop.Col())

			// Render block body
			t.Block.RenderToOutputBuffer(context, output)

			// Close </td>
			*output += "</td>"

			// Handle interrupts
			if ctx.Interrupt() {
				interrupt := ctx.PopInterrupt()
				if _, ok := interrupt.(*liquid.BreakInterrupt); ok {
					break
				}
			}

			// Check if we need to close row and start new one
			if tablerowloop.ColLast() && !tablerowloop.Last() {
				*output += fmt.Sprintf("</tr>\n<tr class=\"row%d\">", tablerowloop.Row()+1)
			}

			// Increment loop
			tablerowloop.Increment()
		}
	})

	// Close last row
	*output += "</tr>\n"
}
