package liquid

import (
	"regexp"
	"strings"
)

var (
	variableFilterMarkupRegex        = regexp.MustCompile(`\|\s*(.*)`)
	variableFilterParser             = regexp.MustCompile(`(?:\s+|"[^"]*"|'[^']*'|(?:[^\s,\|'"]|"[^"]*"|'[^']*')+)+`)
	variableFilterArgsRegex          = regexp.MustCompile(`(?::|,)\s*((?:\w+\s*:\s*)?(?:[^\s,\|'"]|"[^"]*"|'[^']*')+)`)
	variableJustTagAttributes        = regexp.MustCompile(`^(\w[\w-]*)\s*:\s*((?:"[^"]*"|'[^']*'|(?:[^\s,\|'"]|"[^"]*"|'[^']*')+)+)$`)
	variableMarkupWithQuotedFragment = regexp.MustCompile(`^([^\|]+)(.*)$`)
)

// ParseContextInterface interface for parsing expressions
// This interface is implemented by ParseContext struct
type ParseContextInterface interface {
	ParseExpression(markup string) interface{}
	SafeParseExpression(parser *Parser) interface{}
	NewParser(markup string) *Parser
	LineNumber() *int
	SetLineNumber(*int)
	Environment() *Environment
	TrimWhitespace() bool
	SetTrimWhitespace(bool)
	Depth() int
	IncrementDepth()
	DecrementDepth()
	NewBlockBody() *BlockBody
	NewTokenizer(source string, lineNumbers bool, startLineNumber *int, forLiquidTag bool) *Tokenizer
}

// Variable represents a Liquid variable with optional filters.
type Variable struct {
	name         interface{}
	parseContext ParseContextInterface
	lineNumber   *int
	markup       string
	filters      [][]interface{}
}

// NewVariable creates a new Variable from markup.
func NewVariable(markup string, parseContext ParseContextInterface) *Variable {
	v := &Variable{
		markup:       markup,
		parseContext: parseContext,
		lineNumber:   parseContext.LineNumber(),
	}

	// Use parser switching based on error mode
	// Create a wrapper that implements the ParserSwitching interface
	var psContext interface {
		ErrorMode() string
		AddWarning(error)
	}
	if pc, ok := parseContext.(*ParseContext); ok {
		psContext = pc
	} else {
		// Fallback: create a minimal wrapper
		psContext = &parseContextWrapper{
			errorMode: "lax", // Default to lax if we can't determine
		}
	}

	ps := &ParserSwitching{
		parseContext:  psContext,
		lineNumber:    parseContext.LineNumber(),
		markupContext: v.markupContext,
	}

	err := ps.StrictParseWithErrorModeFallback(
		markup,
		func(m string) error {
			// Catch panics from SafeParseExpression and convert to errors
			var parseErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							parseErr = e
						} else {
							panic(r) // Re-panic non-error panics
						}
					}
				}()
				parseErr = v.strictParse(m)
			}()
			return parseErr
		},
		func(m string) error {
			v.laxParse(m)
			return nil
		},
		func(m string) error {
			// Catch panics from SafeParseExpression and convert to errors
			var parseErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							parseErr = e
						} else {
							panic(r) // Re-panic non-error panics
						}
					}
				}()
				parseErr = v.rigidParse(m)
			}()
			return parseErr
		},
	)

	// If there was an error, it's already been handled by parser switching
	// But in strict/rigid mode, we need to panic it
	if err != nil {
		panic(err)
	}

	return v
}

// markupContext returns a context string for markup.
func (v *Variable) markupContext(markup string) string {
	return "in \"{{" + markup + "}}\""
}

// Raw returns the raw markup.
func (v *Variable) Raw() string {
	return v.markup
}

// Name returns the variable name expression.
func (v *Variable) Name() interface{} {
	return v.name
}

// Filters returns the filters.
func (v *Variable) Filters() [][]interface{} {
	return v.filters
}

// LineNumber returns the line number.
func (v *Variable) LineNumber() *int {
	return v.lineNumber
}

func (v *Variable) laxParse(markup string) {
	v.filters = [][]interface{}{}
	matches := variableMarkupWithQuotedFragment.FindStringSubmatch(markup)
	if len(matches) < 2 {
		return
	}

	nameMarkup := strings.TrimSpace(matches[1])
	filterMarkup := matches[2]

	v.name = v.parseContext.ParseExpression(nameMarkup)

	if filterMarkup != "" {
		filterMatches := variableFilterMarkupRegex.FindStringSubmatch(filterMarkup)
		if len(filterMatches) > 1 {
			filters := variableFilterParser.FindAllString(filterMatches[1], -1)
			for _, f := range filters {
				f = strings.TrimSpace(f)
				if f == "" {
					continue
				}
				// Extract filter name (first word)
				parts := strings.Fields(f)
				if len(parts) == 0 {
					continue
				}
				filterName := parts[0]
				// Extract filter args
				filterArgs := variableFilterArgsRegex.FindAllString(f, -1)
				v.filters = append(v.filters, v.laxParseFilterExpressions(filterName, filterArgs))
			}
		}
	}
}

func (v *Variable) strictParse(markup string) error {
	v.filters = [][]interface{}{}
	p := v.parseContext.NewParser(markup)

	if p.Look(":end_of_string", 0) {
		return nil
	}

	v.name = v.parseContext.SafeParseExpression(p)
	for {
		_, ok := p.ConsumeOptional(":pipe")
		if !ok {
			break
		}
		filterName, err := p.Consume(":id")
		if err != nil {
			return err
		}
		var filterArgs []string
		if _, ok := p.ConsumeOptional(":colon"); ok {
			filterArgs = v.parseFilterArgs(p)
		} else {
			filterArgs = []string{}
		}
		v.filters = append(v.filters, v.laxParseFilterExpressions(filterName, filterArgs))
	}
	_, err := p.Consume(":end_of_string")
	if err != nil {
		return err
	}
	return nil
}

func (v *Variable) rigidParse(markup string) error {
	v.filters = [][]interface{}{}
	p := v.parseContext.NewParser(markup)

	if p.Look(":end_of_string", 0) {
		return nil
	}

	v.name = v.parseContext.SafeParseExpression(p)
	for {
		_, ok := p.ConsumeOptional(":pipe")
		if !ok {
			break
		}
		v.filters = append(v.filters, v.rigidParseFilterExpressions(p))
	}
	_, err := p.Consume(":end_of_string")
	if err != nil {
		return err
	}
	return nil
}

func (v *Variable) parseFilterArgs(p *Parser) []string {
	arg, err := p.Argument()
	if err != nil {
		return []string{}
	}
	filterArgs := []string{arg}
	for {
		_, ok := p.ConsumeOptional(":comma")
		if !ok {
			break
		}
		arg, err := p.Argument()
		if err == nil {
			filterArgs = append(filterArgs, arg)
		}
	}
	return filterArgs
}

func (v *Variable) laxParseFilterExpressions(filterName string, unparsedArgs []string) []interface{} {
	filterArgs := []interface{}{}
	var keywordArgs map[string]interface{}

	for _, a := range unparsedArgs {
		matches := variableJustTagAttributes.FindStringSubmatch(a)
		if len(matches) == 3 {
			if keywordArgs == nil {
				keywordArgs = make(map[string]interface{})
			}
			keywordArgs[matches[1]] = v.parseContext.ParseExpression(matches[2])
		} else {
			filterArgs = append(filterArgs, v.parseContext.ParseExpression(a))
		}
	}

	result := []interface{}{filterName, filterArgs}
	if keywordArgs != nil {
		result = append(result, keywordArgs)
	}
	return result
}

func (v *Variable) rigidParseFilterExpressions(p *Parser) []interface{} {
	filterName, _ := p.Consume(":id")
	filterArgs := []interface{}{}
	keywordArgs := make(map[string]interface{})

	if _, ok := p.ConsumeOptional(":colon"); ok {
		// Parse first argument (no leading comma)
		if !v.endOfArguments(p) {
			v.argument(p, &filterArgs, keywordArgs)
		}

		// Parse remaining arguments (with leading commas)
		for {
			_, ok := p.ConsumeOptional(":comma")
			if !ok || v.endOfArguments(p) {
				break
			}
			v.argument(p, &filterArgs, keywordArgs)
		}
	}

	result := []interface{}{filterName, filterArgs}
	if len(keywordArgs) > 0 {
		result = append(result, keywordArgs)
	}
	return result
}

func (v *Variable) argument(p *Parser, positionalArgs *[]interface{}, keywordArgs map[string]interface{}) {
	if p.Look(":id", 0) && p.Look(":colon", 1) {
		key, _ := p.Consume(":id")
		_, _ = p.Consume(":colon")
		value := v.parseContext.SafeParseExpression(p)
		keywordArgs[key] = value
	} else {
		expr, err := p.Argument()
		if err == nil {
			*positionalArgs = append(*positionalArgs, v.parseContext.ParseExpression(expr))
		}
	}
}

func (v *Variable) endOfArguments(p *Parser) bool {
	return p.Look(":pipe", 0) || p.Look(":end_of_string", 0)
}

// Render renders the variable.
// Evaluate evaluates the variable with filters, used when Variable appears in conditions.
// This ensures that Variables with filters are properly evaluated in if/unless/case conditions.
func (v *Variable) Evaluate(context *Context) interface{} {
	// Context implements TagContext, so we can use it directly
	return v.Render(context)
}

func (v *Variable) Render(context TagContext) interface{} {
	// Evaluate the variable name expression directly (like Ruby: context.evaluate(@name))
	nameExpr := v.Name()
	value := context.Evaluate(nameExpr)

	// Apply filters
	for _, filter := range v.Filters() {
		if len(filter) == 0 {
			continue
		}
		filterName := ToS(filter[0], nil)
		var filterArgs []interface{}
		if len(filter) > 1 {
			if args, ok := filter[1].([]interface{}); ok {
				filterArgs = args
			}
		}

		// Evaluate filter arguments
		evaluatedArgs := make([]interface{}, len(filterArgs))
		for i, arg := range filterArgs {
			evaluatedArgs[i] = context.Evaluate(arg)
		}

		// Invoke filter
		value = context.Invoke(filterName, value, evaluatedArgs...)
	}

	// Apply global filter (like Ruby: context.apply_global_filter(obj))
	ctx := context.Context().(*Context)
	value = ctx.ApplyGlobalFilter(value)

	return value
}

// RenderToOutputBuffer renders the variable to the output buffer.
func (v *Variable) RenderToOutputBuffer(context TagContext, output *string) {
	val := v.Render(context)
	*output += ToS(val, nil)
}

// parseContextWrapper is a minimal wrapper for ParseContextInterface when it doesn't implement ErrorMode/AddWarning.
type parseContextWrapper struct {
	errorMode string
}

func (p *parseContextWrapper) ErrorMode() string {
	return p.errorMode
}

func (p *parseContextWrapper) AddWarning(err error) {
	// No-op for wrapper
	_ = err // no-op to register coverage
}
