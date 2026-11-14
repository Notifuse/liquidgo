package liquid

import (
	"reflect"
	"regexp"
	"strings"
)

var (
	blockBodyLiquidTagToken      = regexp.MustCompile(`^\s*(` + TagName.String() + `)\s*(.*?)$`)
	blockBodyFullToken           = regexp.MustCompile(`(?s)^` + TagStart.String() + `-?(\s*)(` + TagName.String() + `)(\s*)(.*?)-?` + TagEnd.String() + `$`)
	blockBodyWhitespaceOrNothing = regexp.MustCompile(`^\s*$`)
	// blockBodyFullTokenPossiblyInvalid is Ruby's BlockBody::FullTokenPossiblyInvalid: a
	// token ending with a tag, with any text before it ("{% invalid {% endraw %}").
	blockBodyFullTokenPossiblyInvalid = regexp.MustCompile(`(?s)^(.*)` + TagStart.String() + `-?\s*(\w+)\s*(.*)?-?` + TagEnd.String() + `$`)
)

// MatchTagPossiblyInvalid reports whether token ends with a tag, and returns the text
// before that tag and its name. raw, comment and doc use it to find their closing tag.
func MatchTagPossiblyInvalid(token string) (before, tagName string, ok bool) {
	// most tokens of a raw or comment body are text; skip the regexp for them
	if !strings.HasSuffix(token, "%}") {
		return "", "", false
	}
	m := blockBodyFullTokenPossiblyInvalid.FindStringSubmatch(token)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

// TrimsAfter reports whether a tag token ends with whitespace control ("-%}").
func TrimsAfter(token string) bool {
	return len(token) >= 3 && token[len(token)-3] == '-'
}

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
	// An empty line is a token too: stop only when the tokenizer is exhausted.
	for !tokenizer.Exhausted() {
		token := tokenizer.Shift()

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

				func() {
					defer func() {
						if r := recover(); r != nil {
							if e, ok := r.(error); ok {
								if parseContext.ErrorMode() == "warn" {
									parseContext.AddWarning(e)
									tag = nil
								} else {
									panic(r)
								}
							} else {
								panic(r)
							}
						}
					}()
					results := tagClassValue.Call(args)
					if len(results) == 2 {
						// Check for error
						if !results[1].IsNil() {
							if err, ok := results[1].Interface().(error); ok {
								if parseContext.ErrorMode() == "warn" {
									parseContext.AddWarning(err)
									tag = nil
								} else {
									panic(err)
								}
							}
						} else {
							tag = results[0].Interface()
						}
					}
				}()
			}
		}

		// If tag creation failed in warn mode, skip it (liquid tags are just lines of code)
		if tag == nil && parseContext.ErrorMode() == "warn" {
			continue
		}

		// If tag is still nil, fallback to generic tag
		if tag == nil {
			tag = NewTag(tagName, markup, parseContext)
		}

		// Parse the tag if it has a Parse method
		if parseable, ok := tag.(interface{ Parse(*Tokenizer) error }); ok {
			err := parseable.Parse(tokenizer)
			if err != nil {
				if parseContext.ErrorMode() == "warn" {
					parseContext.AddWarning(err)
					continue
				}
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

			// Count newlines inside the tag token (matches[1] and matches[3] are whitespace captures)
			// This is particularly important for multiline {% liquid %} tags
			if parseContext.LineNumber() != nil && len(matches) >= 4 {
				whitespaceBefore := matches[1]
				whitespaceAfter := matches[3]
				newlineCount := strings.Count(whitespaceBefore, "\n") + strings.Count(whitespaceAfter, "\n")
				if newlineCount > 0 {
					lineNum := *parseContext.LineNumber() + newlineCount
					parseContext.SetLineNumber(&lineNum)
				}
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

					func() {
						defer func() {
							if r := recover(); r != nil {
								if e, ok := r.(error); ok {
									if parseContext.ErrorMode() == "warn" {
										parseContext.AddWarning(e)
										tag = nil
									} else {
										panic(r)
									}
								} else {
									panic(r)
								}
							}
						}()
						results := tagClassValue.Call(args)
						if len(results) == 2 {
							// Check for error
							if !results[1].IsNil() {
								if err, ok := results[1].Interface().(error); ok {
									if parseContext.ErrorMode() == "warn" {
										parseContext.AddWarning(err)
										tag = nil
									} else {
										panic(err)
									}
								}
							} else {
								tag = results[0].Interface()
							}
						}
					}()
				}
			}

			// If tag creation failed in warn mode, treat as text
			if tag == nil && parseContext.ErrorMode() == "warn" {
				bb.nodelist = append(bb.nodelist, token)
				continue
			}

			// If tag is still nil, fallback to generic tag
			if tag == nil {
				tag = NewTag(tagName, markup, parseContext)
			}

			// Parse the tag if it has a Parse method
			if parseable, ok := tag.(interface{ Parse(*Tokenizer) error }); ok {
				err := parseable.Parse(tokenizer)
				if err != nil {
					if parseContext.ErrorMode() == "warn" {
						parseContext.AddWarning(err)
						// If parsing failed, treat as text
						bb.nodelist = append(bb.nodelist, token)
						continue
					}
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
	// Always set trim_whitespace flag based on whether this token has a trailing dash
	// This matches Ruby's behavior: parse_context.trim_whitespace = (token[-3] == WhitespaceControl)
	hasTrailingDash := len(token) >= 3 && token[len(token)-3] == '-'
	parseContext.SetTrimWhitespace(hasTrailingDash)
}

func (bb *BlockBody) createVariable(token string, parseContext ParseContextInterface) interface{} {
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

		var variable *Variable
		var err error

		func() {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					} else {
						panic(r)
					}
				}
			}()
			variable = NewVariable(markup, parseContext)
		}()

		if err != nil {
			if parseContext.ErrorMode() == "warn" {
				parseContext.AddWarning(err)
				return token
			}
			panic(err)
		}
		return variable
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
	output := NewOutputBufferFor(context)
	bb.RenderToOutput(context, output)
	return output.String()
}

// RenderToOutputBuffer renders the block body and appends it to output.
func (bb *BlockBody) RenderToOutputBuffer(context TagContext, output *string) {
	RenderToString(context, output, func(o *OutputBuffer) { bb.RenderToOutput(context, o) })
}

// RenderToOutput renders the block body to output.
func (bb *BlockBody) RenderToOutput(context TagContext, output *OutputBuffer) {
	ctx, _ := context.(*Context)
	profiler := ctx != nil && ctx.Profiler() != nil

	// Like Ruby: context.resource_limits.increment_render_score(@nodelist.length),
	// but an empty body counts 1, so that a loop with an empty body is
	// bounded too.
	if rl := limitsOf(context); rl != nil {
		rl.IncrementRenderScore(max(1, len(bb.nodelist)))
	}
	ctx.checkCanceled()

	for _, node := range bb.nodelist {
		if s, ok := node.(string); ok {
			output.WriteString(s)
			continue
		}
		ctx.checkCanceled()
		bb.renderNode(context, ctx, output, node, profiler)
		if context != nil && context.Interrupt() {
			return
		}
	}
}

// renderNode renders one node that is not raw text. An error in the node is
// handled as Ruby's rescue_render_node does.
func (bb *BlockBody) renderNode(context TagContext, ctx *Context, output *OutputBuffer, node interface{}, profiler bool) {
	defer func() {
		if r := recover(); r != nil {
			rescueRenderNode(context, output, node, r)
		}
	}()
	render := func() {
		if v, ok := node.(*Variable); ok {
			v.RenderToOutput(context, output)
			return
		}
		renderWithPlan(planFor(node), node, context, output)
	}
	if !profiler {
		render()
		return
	}
	var code string
	var lineNumber *int
	if r, ok := node.(interface{ Raw() string }); ok {
		code = r.Raw()
	}
	if r, ok := node.(interface{ LineNumber() *int }); ok {
		lineNumber = r.LineNumber()
	}
	ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, render)
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
	// Each nested liquid tag parses the rest of the markup again: the depth
	// of blocks bounds them, so that the cost stays linear.
	if parseContext.Depth() >= blockMaxDepth {
		panic(NewStackLevelError("Nesting too deep"))
	}
	parseContext.IncrementDepth()
	defer parseContext.DecrementDepth()

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
