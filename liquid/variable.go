package liquid

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	variableFilterMarkupRegex = regexp.MustCompile(`(?s)\|\s*(.*)`)
	// Ruby's FilterParser, /(?:\s+|#{QuotedFragment}|#{ArgumentSeparator})+/o
	variableFilterParser = regexp.MustCompile(`(?:\s+|(?:` + QuotedFragment.String() + `)|,)+`)
	variableFilterName   = regexp.MustCompile(`\w+`)
	// Ruby's FilterArgsRegex: a quoted string is tried before a fragment,
	// as the alternation of QuotedFragment orders them
	variableFilterArgsRegex   = regexp.MustCompile(`(?::|,)\s*((?:\w+\s*:\s*)?(?:` + QuotedFragment.String() + `))`)
	variableJustTagAttributes = regexp.MustCompile(`^(\w[\w-]*)\s*:\s*((?:"[^"]*"|'[^']*'|(?:[^\s,\|'"]|"[^"]*"|'[^']*')+)+)$`)
	// Ruby's MarkupWithQuotedFragment, /(#{QuotedFragment})(.*)/om: the
	// name is the first fragment of the markup, wherever it starts
	variableMarkupWithQuotedFragment = regexp.MustCompile(`(?s)(` + QuotedFragment.String() + `)(.*)`)
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
	ErrorMode() ErrorMode
	AddWarning(error)
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
	var lineNum *int
	if parseContext.LineNumber() != nil {
		ln := *parseContext.LineNumber()
		lineNum = &ln
	}

	v := &Variable{
		markup:       markup,
		parseContext: parseContext,
		lineNumber:   lineNum,
	}

	err := parseWithErrorModeFallback(parseContext, markup, lineNum, v.markupContext,
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
				// The filter name is the first word of the filter, as
				// Ruby's f =~ /\w+/
				filterName := variableFilterName.FindString(f)
				if filterName == "" {
					continue
				}
				// The arguments are the captures of FilterArgsRegex, as
				// Ruby's f.scan(FilterArgsRegex).flatten: without the
				// separator before each
				var filterArgs []string
				for _, m := range variableFilterArgsRegex.FindAllStringSubmatch(f, -1) {
					filterArgs = append(filterArgs, m[1])
				}
				v.filters = append(v.filters, v.laxParseFilterExpressions(filterName, filterArgs))
			}
		}
	}
}

func (v *Variable) strictParse(markup string) error {
	v.filters = [][]interface{}{}
	p := v.parseContext.NewParser(markup)
	if err := p.Error(); err != nil {
		return err
	}

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
			if filterArgs, err = v.parseFilterArgs(p); err != nil {
				return err
			}
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
	if err := p.Error(); err != nil {
		return err
	}

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

func (v *Variable) parseFilterArgs(p *Parser) ([]string, error) {
	// as Ruby's parse_filterargs: an argument after the colon and after
	// each comma, or a syntax error
	arg, err := p.Argument()
	if err != nil {
		return nil, err
	}
	filterArgs := []string{arg}
	for {
		_, ok := p.ConsumeOptional(":comma")
		if !ok {
			break
		}
		arg, err := p.Argument()
		if err != nil {
			return nil, err
		}
		filterArgs = append(filterArgs, arg)
	}
	return filterArgs, nil
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
	return v.render(context, nil)
}

// render is Render; with safe set, it also reports whether the value is HTML that
// the output escape writes as it is (see escapeTracker).
func (v *Variable) render(context TagContext, safe *escapeTracker) interface{} {
	// Evaluate the variable name expression directly (like Ruby: context.evaluate(@name))
	nameExpr := v.Name()
	value := context.Evaluate(nameExpr)
	safe.start(value)

	// Apply filters
	for _, filter := range v.Filters() {
		if len(filter) == 0 {
			continue
		}
		filterName, ok := filter[0].(string)
		if !ok {
			filterName = ToS(filter[0], nil)
		}
		var filterArgs []interface{}
		if len(filter) > 1 {
			if args, ok := filter[1].([]interface{}); ok {
				filterArgs = args
			}
		}

		// Evaluate filter arguments; a filter without any gets none, and a
		// standard filter without keyword arguments the context's slice,
		// so that its call allocates nothing for them
		ctx, isContext := context.(*Context)
		var fn FilterFunc
		var standard bool
		if isContext {
			fn, standard = ctx.resolveFilter(filterName)
		}
		var evaluatedArgs, lent []interface{}
		if len(filterArgs) > 0 {
			if standard && len(filter) <= 2 {
				lent = ctx.lendFilterArgs(len(filterArgs))
				evaluatedArgs = lent
			}
			if evaluatedArgs == nil {
				evaluatedArgs = make([]interface{}, len(filterArgs), len(filterArgs)+1)
			}
			for i, arg := range filterArgs {
				evaluatedArgs[i] = context.Evaluate(arg)
			}
		}

		// Keyword arguments are passed as one trailing map, as in Ruby
		// (default: "x", allow_false: true calls default(input, "x", {"allow_false" => true}))
		if len(filter) > 2 {
			if keywordArgs, ok := filter[2].(map[string]interface{}); ok && len(keywordArgs) > 0 {
				evaluated := make(map[string]interface{}, len(keywordArgs))
				for key, arg := range keywordArgs {
					evaluated[key] = context.Evaluate(arg)
				}
				evaluatedArgs = append(evaluatedArgs, evaluated)
			}
		}

		// Invoke filter, resolved once above
		if isContext {
			value = ctx.invokeFilter(filterName, fn, value, evaluatedArgs)
		} else {
			value = context.Invoke(filterName, value, evaluatedArgs...)
		}
		if lent != nil {
			ctx.returnFilterArgs(lent)
		}
		safe.filter(filterName, filterArgs)
	}

	return value
}

// RenderToOutput writes the value of the variable to output. With the
// environment's output escape on (see EnvironmentConfig.OutputEscape), the
// written text is HTML-escaped unless it is HTML on purpose (see
// escapeTracker).
func (v *Variable) RenderToOutput(context TagContext, output *OutputBuffer) {
	ctx, ok := context.Context().(*Context)
	if !ok || !ctx.OutputEscape() {
		WriteValue(output, v.Render(context))
		return
	}
	safe := &escapeTracker{ctx: ctx}
	value := v.render(context, safe)
	if safe.safe {
		WriteValue(output, value)
		return
	}
	writeValue(escapingWriter{output}, value, nil)
}

// stringWriter is what writeValue writes to: the output buffer, or a writer
// that escapes on the way.
type stringWriter interface {
	WriteString(s string)
}

// WriteValue writes a value to output as Ruby's render_obj_to_output: nil
// writes nothing, and an array writes its items one after the other, nested
// arrays included. An array that holds itself is not written again inside
// itself.
func WriteValue(output *OutputBuffer, value interface{}) {
	writeValue(output, value, nil)
}

func writeValue(output stringWriter, value interface{}, seen map[uintptr]bool) {
	value = ToLiquid(value)
	switch v := value.(type) {
	case nil:
		return
	case string:
		output.WriteString(v)
		return
	case int:
		output.WriteString(strconv.Itoa(v)) // as ToS, without its method lookups
		return
	case []interface{}:
		if len(v) == 0 {
			return
		}
		ptr := reflect.ValueOf(v).Pointer()
		if seen[ptr] {
			return
		}
		if seen == nil {
			seen = make(map[uintptr]bool)
		}
		seen[ptr] = true
		defer delete(seen, ptr)
		for _, item := range v {
			writeValue(output, item, seen)
		}
		return
	}
	if _, ok := value.(fmt.Stringer); ok {
		// A value with a String method renders it, even a slice or an array
		output.WriteString(ToS(value, nil))
		return
	}
	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		for i := 0; i < rv.Len(); i++ {
			writeValue(output, rv.Index(i).Interface(), seen)
		}
		return
	}
	output.WriteString(ToS(value, nil))
}
