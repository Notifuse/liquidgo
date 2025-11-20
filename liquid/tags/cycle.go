package tags

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	cycleSimpleSyntax   = regexp.MustCompile(`^` + liquid.QuotedFragment.String())
	cycleNamedSyntax    = regexp.MustCompile(`^(` + liquid.QuotedFragment.String() + `)\s*\:\s*(.*)`)
	cycleUnnamedPattern = regexp.MustCompile(`\w+:0x[0-9a-fA-F]{8}`)
)

// CycleTag represents a cycle tag that loops through a group of strings.
type CycleTag struct {
	*liquid.Tag
	variables []interface{}
	name      interface{}
	isNamed   bool
}

// NewCycleTag creates a new CycleTag.
func NewCycleTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*CycleTag, error) {
	ct := &CycleTag{
		Tag: liquid.NewTag(tagName, markup, parseContext),
	}

	// Parse markup
	err := ct.parseMarkup(markup, parseContext)
	if err != nil {
		return nil, err
	}

	return ct, nil
}

// Variables returns the cycle variables.
func (c *CycleTag) Variables() []interface{} {
	return c.variables
}

// Named returns true if this is a named cycle.
func (c *CycleTag) Named() bool {
	return c.isNamed
}

// RenderToOutputBuffer renders the cycle tag.
func (c *CycleTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	ctx := context.Context().(*liquid.Context)
	registers := ctx.Registers()

	// Get or initialize cycle register
	cycleReg := registers.Get("cycle")
	var cycleMap map[string]interface{}
	if cycleReg == nil {
		cycleMap = make(map[string]interface{})
		registers.Set("cycle", cycleMap)
	} else {
		cycleMap = cycleReg.(map[string]interface{})
	}

	// Evaluate cycle name/key
	key := liquid.ToS(context.Evaluate(c.name), nil)

	// Get current iteration
	iteration, ok := cycleMap[key]
	if !ok {
		iteration = 0
	}
	iter := 0
	switch v := iteration.(type) {
	case int:
		iter = v
	case float64:
		iter = int(v)
	}

	// Get value at current iteration
	if iter >= len(c.variables) {
		iter = 0
	}
	val := context.Evaluate(c.variables[iter])

	// Convert to string
	var valStr string
	if arr, ok := val.([]interface{}); ok {
		// Join array
		parts := make([]string, len(arr))
		for i, item := range arr {
			parts[i] = liquid.ToS(item, nil)
		}
		valStr = strings.Join(parts, "")
	} else if val != nil {
		// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			parts := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				parts[i] = liquid.ToS(v.Index(i).Interface(), nil)
			}
			valStr = strings.Join(parts, "")
		} else {
			valStr = liquid.ToS(val, nil)
		}
	} else {
		valStr = liquid.ToS(val, nil)
	}
	*output += valStr

	// Increment iteration
	iter++
	if iter >= len(c.variables) {
		iter = 0
	}
	cycleMap[key] = iter
}

func (c *CycleTag) parseMarkup(markup string, parseContext liquid.ParseContextInterface) error {
	// Try named syntax first
	if matches := cycleNamedSyntax.FindStringSubmatch(markup); len(matches) >= 3 {
		c.name = parseContext.ParseExpression(matches[1])
		c.isNamed = true
		c.variables = c.variablesFromString(matches[2], parseContext)
		return nil
	}

	// Try simple syntax
	if cycleSimpleSyntax.MatchString(markup) {
		c.variables = c.variablesFromString(markup, parseContext)
		// Generate name from variables
		c.name = liquid.ToS(c.variables, nil)
		c.isNamed = !cycleUnnamedPattern.MatchString(c.name.(string))
		return nil
	}

	// Syntax error
	var locale *liquid.I18n
	if pc, ok := parseContext.(*liquid.ParseContext); ok {
		locale = pc.Locale()
		msg := locale.Translate("errors.syntax.cycle", map[string]interface{}{})
		return liquid.NewSyntaxError(msg)
	}
	return liquid.NewSyntaxError("Liquid syntax error: cycle")
}

func (c *CycleTag) variablesFromString(markup string, parseContext liquid.ParseContextInterface) []interface{} {
	parts := strings.Split(markup, ",")
	variables := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Match quoted fragment - FindStringSubmatch returns full match and submatches
		matches := liquid.QuotedFragment.FindStringSubmatch(part)
		if len(matches) > 0 {
			// Use the first match (full match) or first submatch if available
			matchStr := part
			if len(matches) > 1 && matches[1] != "" {
				matchStr = matches[1]
			} else if matches[0] != "" {
				matchStr = matches[0]
			}
			expr := parseContext.ParseExpression(matchStr)
			variables = append(variables, expr)
		} else {
			// If no match, parse the part as-is
			expr := parseContext.ParseExpression(part)
			variables = append(variables, expr)
		}
	}

	return variables
}
