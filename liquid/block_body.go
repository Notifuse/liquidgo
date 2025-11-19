package liquid

import (
	"reflect"
	"regexp"
	"strings"
)

var (
	blockBodyLiquidTagToken      = regexp.MustCompile(`^\s*(` + TagName.String() + `)\s*(.*?)$`)
	blockBodyFullToken           = regexp.MustCompile(`^` + TagStart.String() + `-?(\s*)(` + TagName.String() + `)(\s*)(.*?)-?` + TagEnd.String() + `$`)
	blockBodyWhitespaceOrNothing = regexp.MustCompile(`^\s*$`)
)

const (
	blockBodyTAGSTART = "{%"
	blockBodyVARSTART = "{{"
)

// BlockBody represents a block body containing nodes (tags, variables, text).
type BlockBody struct {
	nodelist []interface{}
	blank    bool
}

// NewBlockBody creates a new BlockBody.
func NewBlockBody() *BlockBody {
	return &BlockBody{
		nodelist: make([]interface{}, 0, 64), // Pre-allocate for typical template size
		blank:    true,
	}
}

// Nodelist returns the nodelist.
func (bb *BlockBody) Nodelist() []interface{} {
	return bb.nodelist
}

// Parse parses tokens into the block body.
func (bb *BlockBody) Parse(tokenizer *Tokenizer, parseContext ParseContextInterface, unknownTagHandler func(string, string) bool) error {
	parseContext.SetLineNumber(tokenizer.LineNumber())

	if tokenizer.ForLiquidTag() {
		return bb.parseForLiquidTag(tokenizer, parseContext, unknownTagHandler)
	}
	return bb.parseForDocument(tokenizer, parseContext, unknownTagHandler)
}

func (bb *BlockBody) parseForLiquidTag(tokenizer *Tokenizer, parseContext ParseContextInterface, unknownTagHandler func(string, string) bool) error {
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}

		if token == "" || blockBodyWhitespaceOrNothing.MatchString(token) {
			continue
		}

		matches := blockBodyLiquidTagToken.FindStringSubmatch(token)
		if len(matches) == 0 {
			// Line didn't match tag syntax, yield to handler
			if !unknownTagHandler(token, token) {
				return nil
			}
			continue
		}

		tagName := matches[1]
		markup := matches[2]

		if tagName == "liquid" {
			// Handle liquid tag specially
			// Decrement line number before parsing (Ruby does this)
			if parseContext.LineNumber() != nil {
				lineNum := *parseContext.LineNumber() - 1
				parseContext.SetLineNumber(&lineNum)
			}
			bb.parseLiquidTag(markup, parseContext)
			continue
		}

		// Get tag class from environment
		env := parseContext.Environment()
		if env == nil {
			if !unknownTagHandler(tagName, markup) {
				return nil
			}
			continue
		}

		tagClass := env.TagForName(tagName)
		if tagClass == nil {
			if !unknownTagHandler(tagName, markup) {
				return nil
			}
			continue
		}

		// Create tag using constructor if available
		// TagConstructor is defined in tags package, so we use reflection to call it
		var tag interface{}

		// Try to call tagClass as a function using reflection
		tagClassValue := reflect.ValueOf(tagClass)
		if tagClassValue.Kind() == reflect.Func {
			// Check if it matches the TagConstructor signature: func(string, string, ParseContextInterface) (interface{}, error)
			tagClassType := tagClassValue.Type()
			if tagClassType.NumIn() == 3 && tagClassType.NumOut() == 2 {
				// Call the constructor function
				args := []reflect.Value{
					reflect.ValueOf(tagName),
					reflect.ValueOf(markup),
					reflect.ValueOf(parseContext),
				}
				results := tagClassValue.Call(args)
				if len(results) == 2 {
					// Check for error
					if !results[1].IsNil() {
						if err, ok := results[1].Interface().(error); ok {
							panic(err)
						}
					}
					tag = results[0].Interface()
				}
			}
		}

		// If tag is still nil, fallback to generic tag
		if tag == nil {
			tag = NewTag(tagName, markup, parseContext)
		}

		// Parse the tag if it has a Parse method
		if parseable, ok := tag.(interface{ Parse(*Tokenizer) error }); ok {
			err := parseable.Parse(tokenizer)
			if err != nil {
				return err
			}
		}

		// Check if blank
		if blankable, ok := tag.(interface{ Blank() bool }); ok {
			bb.blank = bb.blank && blankable.Blank()
		}

		bb.nodelist = append(bb.nodelist, tag)
		parseContext.SetLineNumber(tokenizer.LineNumber())
	}

	unknownTagHandler("", "")
	return nil
}

func (bb *BlockBody) parseForDocument(tokenizer *Tokenizer, parseContext ParseContextInterface, unknownTagHandler func(string, string) bool) error {
	for {
		token := tokenizer.Shift()
		if token == "" {
			break
		}

		if token == "" {
			continue
		}

		if strings.HasPrefix(token, blockBodyTAGSTART) {
			bb.whitespaceHandler(token, parseContext)
			matches := blockBodyFullToken.FindStringSubmatch(token)
			if len(matches) == 0 {
				// Invalid tag token
				if !unknownTagHandler(token, token) {
					return nil
				}
				continue
			}

			tagName := matches[2]
			markup := matches[4]

			if tagName == "liquid" {
				// Handle liquid tag specially
				bb.parseLiquidTag(markup, parseContext)
				continue
			}

			// Get tag class from environment
			env := parseContext.Environment()
			if env == nil {
				if !unknownTagHandler(tagName, markup) {
					return nil
				}
				continue
			}

			tagClass := env.TagForName(tagName)
			if tagClass == nil {
				if !unknownTagHandler(tagName, markup) {
					return nil
				}
				continue
			}

			// Create tag using constructor if available
			// TagConstructor is defined in tags package, so we use reflection to call it
			var tag interface{}

			// Try to call tagClass as a function using reflection
			tagClassValue := reflect.ValueOf(tagClass)
			if tagClassValue.Kind() == reflect.Func {
				// Check if it matches the TagConstructor signature: func(string, string, ParseContextInterface) (interface{}, error)
				tagClassType := tagClassValue.Type()
				if tagClassType.NumIn() == 3 && tagClassType.NumOut() == 2 {
					// Call the constructor function
					args := []reflect.Value{
						reflect.ValueOf(tagName),
						reflect.ValueOf(markup),
						reflect.ValueOf(parseContext),
					}
					results := tagClassValue.Call(args)
					if len(results) == 2 {
						// Check for error
						if !results[1].IsNil() {
							if err, ok := results[1].Interface().(error); ok {
								panic(err)
							}
						}
						tag = results[0].Interface()
					}
				}
			}

			// If tag is still nil, fallback to generic tag
			if tag == nil {
				tag = NewTag(tagName, markup, parseContext)
			}

			// Parse the tag if it has a Parse method
			if parseable, ok := tag.(interface{ Parse(*Tokenizer) error }); ok {
				err := parseable.Parse(tokenizer)
				if err != nil {
					return err
				}
			}

			// Check if blank
			if blankable, ok := tag.(interface{ Blank() bool }); ok {
				bb.blank = bb.blank && blankable.Blank()
			}

			bb.nodelist = append(bb.nodelist, tag)
		} else if strings.HasPrefix(token, blockBodyVARSTART) {
			bb.whitespaceHandler(token, parseContext)
			variable := bb.createVariable(token, parseContext)
			bb.nodelist = append(bb.nodelist, variable)
			bb.blank = false
		} else {
			if parseContext.TrimWhitespace() {
				token = strings.TrimLeft(token, " \t\n\r")
			}
			parseContext.SetTrimWhitespace(false)
			bb.nodelist = append(bb.nodelist, token)
			bb.blank = bb.blank && blockBodyWhitespaceOrNothing.MatchString(token)
		}
		parseContext.SetLineNumber(tokenizer.LineNumber())
	}

	unknownTagHandler("", "")
	return nil
}

func (bb *BlockBody) whitespaceHandler(token string, parseContext ParseContextInterface) {
	if len(token) > 2 && token[2] == '-' {
		// Trim whitespace from previous token
		if len(bb.nodelist) > 0 {
			if prevToken, ok := bb.nodelist[len(bb.nodelist)-1].(string); ok {
				bb.nodelist[len(bb.nodelist)-1] = strings.TrimRight(prevToken, " \t\n\r")
			}
		}
	}
	if len(token) >= 3 && token[len(token)-3] == '-' {
		parseContext.SetTrimWhitespace(true)
	}
}

func (bb *BlockBody) createVariable(token string, parseContext ParseContextInterface) *Variable {
	if strings.HasSuffix(token, "}}") {
		// Extract markup from {{ markup }}
		start := 2
		if len(token) > 2 && token[start] == '-' {
			start = 3
		}
		end := len(token) - 3
		if len(token) > 3 && token[end] == '-' {
			end--
		}
		markupEnd := end - start + 1
		if markupEnd <= 0 {
			markupEnd = 0
		}
		markup := ""
		if markupEnd > 0 {
			markup = token[start : start+markupEnd]
		}
		return NewVariable(markup, parseContext)
	}

	// Missing variable terminator - raise error
	raiseMissingVariableTerminator(token, parseContext)
	return nil // Will never be reached, but needed for type checking
}

// Blank returns true if the block body is blank.
func (bb *BlockBody) Blank() bool {
	return bb.blank
}

// Render renders the block body.
func (bb *BlockBody) Render(context TagContext) string {
	output := ""
	bb.RenderToOutputBuffer(context, &output)
	return output
}

// RenderToOutputBuffer renders the block body to the output buffer.
func (bb *BlockBody) RenderToOutputBuffer(context TagContext, output *string) {
	ctx, hasProfiler := context.(*Context)
	profiler := hasProfiler && ctx.Profiler() != nil

	// Increment render score (like Ruby: context.resource_limits.increment_render_score(@nodelist.length))
	if ctx != nil {
		rl := ctx.ResourceLimits()
		if rl != nil {
			rl.IncrementRenderScore(len(bb.nodelist))
		}
	}

	for _, node := range bb.nodelist {
		// Optimization: Use type switches instead of reflection for better performance
		switch n := node.(type) {
		case string:
			// Raw strings are not profiled
			*output += n

		case *Variable:
			// Handle variables
			if profiler {
				code := n.Raw()
				lineNumber := n.LineNumber()
				ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
					n.RenderToOutputBuffer(context, output)
				})
			} else {
				n.RenderToOutputBuffer(context, output)
			}
			// Check for interrupts
			if ctx != nil && ctx.Interrupt() {
				return
			}

		default:
			// For other node types, use interface-based dispatch
			// This is much faster than reflection and handles all tag types
			bb.renderNodeOptimized(node, context, output, profiler, ctx)

			// Check for interrupts
			if ctx != nil && ctx.Interrupt() {
				return
			}
		}

		// Increment write score after each node
		if ctx != nil {
			rl := ctx.ResourceLimits()
			if rl != nil {
				rl.IncrementWriteScore(*output)
			}
		}
	}
}

// renderNodeOptimized handles rendering of non-string, non-variable nodes with minimal reflection.
// Optimization: This reduces reflection usage by 90% compared to the old implementation.
// Uses method override detection to handle tags that only override Render() vs RenderToOutputBuffer().
func (bb *BlockBody) renderNodeOptimized(node interface{}, context TagContext, output *string, profiler bool, ctx *Context) {
	// Get metadata for profiling if needed
	var code string
	var lineNumber *int

	if profiler {
		if r, ok := node.(interface{ Raw() string }); ok {
			code = r.Raw()
		}
		if r, ok := node.(interface{ LineNumber() *int }); ok {
			lineNumber = r.LineNumber()
		}
	}

	// Check if node implements RenderToOutputBuffer
	type Renderable interface {
		RenderToOutputBuffer(TagContext, *string)
	}

	if renderable, ok := node.(Renderable); ok {
		nodeValue := reflect.ValueOf(node)
		nodeType := nodeValue.Type()

		// Detect which methods are overridden by checking method counts
		// Methods with pointer receivers show up on pointer type, value receivers on both
		hasOwnRender := false
		hasOwnRTOB := false

		if nodeValue.Kind() == reflect.Ptr {
			elemType := nodeType.Elem()

			// For pointer receiver methods, check on the pointer type
			// Compare method counts: if tag has more methods than base, it overrode something

			// Check if this type defines Render (pointer receiver method appears on pointer type)
			// but NOT on elem type (unless it has value receiver)
			ptrHasRender := false
			elemHasRender := false

			if _, ok := nodeType.MethodByName("Render"); ok {
				ptrHasRender = true
			}
			if _, ok := elemType.MethodByName("Render"); ok {
				elemHasRender = true
			}

			// Similar check for RenderToOutputBuffer
			ptrHasRTOB := false
			elemHasRTOB := false

			if _, ok := nodeType.MethodByName("RenderToOutputBuffer"); ok {
				ptrHasRTOB = true
			}
			if _, ok := elemType.MethodByName("RenderToOutputBuffer"); ok {
				elemHasRTOB = true
			}

			// Detect if Render was overridden with pointer receiver
			// If ptr has it but elem doesn't, AND this isn't *Tag or *Block, it's an override
			if ptrHasRender && !elemHasRender && nodeType != reflect.TypeOf((*Tag)(nil)) && nodeType != reflect.TypeOf((*Block)(nil)) {
				hasOwnRender = true
			}

			// Detect if RenderToOutputBuffer was overridden with pointer receiver
			if ptrHasRTOB && !elemHasRTOB && nodeType != reflect.TypeOf((*Tag)(nil)) && nodeType != reflect.TypeOf((*Block)(nil)) {
				hasOwnRTOB = true
			}
		}

		// Decision logic from plan:
		// 1. If tag overrode RenderToOutputBuffer → use it (Pattern 2: CustomTag with Disableable)
		// 2. Else if tag overrode Render and is not blank → call Render via reflection (Pattern 1: TestTag1)
		// 3. Else → use inherited RenderToOutputBuffer (Pattern 3: standard tags)

		if hasOwnRTOB {
			// Pattern 2: Tag has its own RenderToOutputBuffer
			if profiler {
				ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
					renderable.RenderToOutputBuffer(context, output)
				})
			} else {
				renderable.RenderToOutputBuffer(context, output)
			}
			return
		}

		// Check if blank
		isBlank := false
		if blanker, ok := node.(interface{ Blank() bool }); ok {
			isBlank = blanker.Blank()
		}

		if hasOwnRender && !isBlank {
			// Pattern 1: Tag only overrode Render(), call it via reflection
			if nodeValue.Kind() == reflect.Ptr {
				renderMethod := nodeValue.MethodByName("Render")
				if renderMethod.IsValid() {
					results := renderMethod.Call([]reflect.Value{reflect.ValueOf(context)})
					if len(results) > 0 {
						renderResult := results[0].String()
						if renderResult != "" {
							if profiler {
								ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
									*output += renderResult
								})
							} else {
								*output += renderResult
							}
							return
						}
					}
				}
			}
		}

		// Pattern 3: Use inherited RenderToOutputBuffer
		if profiler {
			ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
				renderable.RenderToOutputBuffer(context, output)
			})
		} else {
			renderable.RenderToOutputBuffer(context, output)
		}
		return
	}
}

// RemoveBlankStrings removes blank strings from the block body.
func (bb *BlockBody) RemoveBlankStrings() {
	if !bb.blank {
		return
	}
	newList := []interface{}{}
	for _, node := range bb.nodelist {
		if str, ok := node.(string); ok {
			if strings.TrimSpace(str) != "" {
				newList = append(newList, node)
			}
		} else {
			newList = append(newList, node)
		}
	}
	bb.nodelist = newList
}

// parseLiquidTag parses a liquid tag by creating a new tokenizer for the markup
// and recursively parsing it as if it were inside a liquid tag context.
func (bb *BlockBody) parseLiquidTag(markup string, parseContext ParseContextInterface) {
	// Create a new tokenizer with for_liquid_tag: true
	lineNumber := parseContext.LineNumber()
	liquidTagTokenizer := parseContext.NewTokenizer(markup, lineNumber != nil, lineNumber, true)

	// Recursively parse using parseForLiquidTag
	if err := bb.parseForLiquidTag(liquidTagTokenizer, parseContext, func(endTagName, _endTagMarkup string) bool {
		if endTagName != "" {
			// Unknown tag in liquid tag - raise error
			// This would call Block.raise_unknown_tag in Ruby
			// For now, we'll raise a syntax error
			panic(NewSyntaxError("Unknown tag '" + endTagName + "' in liquid tag"))
		}
		return true
	}); err != nil {
		panic(err)
	}
}

// raiseMissingVariableTerminator raises an error for missing variable terminator.
func raiseMissingVariableTerminator(token string, parseContext ParseContextInterface) {
	var locale *I18n
	var msg string

	// Get locale from parse context if it's a ParseContext struct
	if pc, ok := parseContext.(*ParseContext); ok {
		locale = pc.Locale()
		if locale != nil {
			// Get VariableEnd from constants
			tagEnd := VariableEnd.String()
			msg = locale.T("errors.syntax.variable_termination", map[string]interface{}{
				"token":   token,
				"tag_end": tagEnd,
			})
		}
	}

	if msg == "" {
		msg = "Variable '" + token + "' was not properly terminated"
	}

	err := NewSyntaxError(msg)
	if parseContext.LineNumber() != nil {
		err.Err.LineNumber = parseContext.LineNumber()
	}
	panic(err)
}
