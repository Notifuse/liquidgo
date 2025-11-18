package liquid

import (
	"reflect"
	"regexp"
	"strings"
)

var (
	blockBodyLiquidTagToken      = regexp.MustCompile(`^\s*(` + TagName.String() + `)\s*(.*?)$`)
	blockBodyFullToken           = regexp.MustCompile(`^` + TagStart.String() + `-?(\s*)(` + TagName.String() + `)(\s*)(.*?)-?` + TagEnd.String() + `$`)
	blockBodyContentOfVariable   = regexp.MustCompile(`^` + VariableStart.String() + `-?(.*?)-?` + VariableEnd.String() + `$`)
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
		nodelist: []interface{}{},
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
		if str, ok := node.(string); ok {
			// Raw strings are not profiled
			*output += str
		} else if variable, ok := node.(*Variable); ok {
			// Handle variables separately (before generic RenderToOutputBuffer check)
			// Render variable
			if profiler {
				// Profile variable rendering
				code := variable.Raw()
				lineNumber := variable.LineNumber()
				ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
					variable.RenderToOutputBuffer(context, output)
				})
			} else {
				variable.RenderToOutputBuffer(context, output)
			}
			// Check for interrupts
			if ctx, ok := context.(*Context); ok && ctx.Interrupt() {
				break
			}
		} else {
			// Check if this node is a Block (or subtype) with a custom Render method
			// First check if it's a Block directly
			var block *Block
			var blockBody *BlockBody
			var tag *Tag
			var tagCode string
			var tagLineNumber *int

			// Use reflection to check if node is or embeds Block
			// (can't use type assertion because *TestBlockTag is not *Tag)
			nodeValue := reflect.ValueOf(node)
			if nodeValue.Kind() == reflect.Ptr {
				actualType := nodeValue.Type()
				// First check if it's directly a *Block
				if b, ok := node.(*Block); ok {
					block = b
					blockBody = b.body
					tag = b.Tag
					tagCode = b.Raw()
					tagLineNumber = b.LineNumber()
				} else {
					// Check if it embeds Block using reflection (recursively)
					elemType := actualType.Elem()
					if elemType.Kind() == reflect.Struct {
						// Helper function to find Block recursively
						var findBlock func(reflect.Value, reflect.Type) bool
						findBlock = func(val reflect.Value, typ reflect.Type) bool {
							if typ.Kind() == reflect.Ptr {
								typ = typ.Elem()
							}
							if typ.Kind() == reflect.Struct {
								for i := 0; i < typ.NumField(); i++ {
									field := typ.Field(i)
									if field.Type == reflect.TypeOf((*Block)(nil)) {
										// Found embedded Block field
										if val.IsValid() {
											var blockField reflect.Value
											if val.Kind() == reflect.Ptr {
												blockField = val.Elem().Field(i)
											} else {
												blockField = val.Field(i)
											}
											if blockField.IsValid() && !blockField.IsNil() {
												if b, ok := blockField.Interface().(*Block); ok {
													block = b
													blockBody = b.body
													// Also set tag info from Block
													tag = b.Tag
													tagCode = b.Raw()
													tagLineNumber = b.LineNumber()
													return true
												}
											}
										}
									} else if field.Type.Kind() == reflect.Ptr {
										// Recursively check embedded types
										if val.IsValid() {
											var embeddedVal reflect.Value
											if val.Kind() == reflect.Ptr {
												embeddedVal = val.Elem().Field(i)
											} else {
												embeddedVal = val.Field(i)
											}
											if embeddedVal.IsValid() && !embeddedVal.IsNil() {
												if findBlock(embeddedVal, field.Type) {
													return true
												}
											}
										}
									}
								}
							}
							return false
						}
						findBlock(nodeValue, actualType)
					}
					// If we didn't find Block, try to find Tag using reflection
					if tag == nil {
						// Helper function to find Tag recursively
						var findTag func(reflect.Value, reflect.Type) bool
						findTag = func(val reflect.Value, typ reflect.Type) bool {
							if typ.Kind() == reflect.Ptr {
								typ = typ.Elem()
							}
							if typ.Kind() == reflect.Struct {
								for i := 0; i < typ.NumField(); i++ {
									field := typ.Field(i)
									if field.Type == reflect.TypeOf((*Tag)(nil)) {
										// Found embedded Tag field
										if val.IsValid() {
											var tagField reflect.Value
											if val.Kind() == reflect.Ptr {
												tagField = val.Elem().Field(i)
											} else {
												tagField = val.Field(i)
											}
											if tagField.IsValid() && !tagField.IsNil() {
												if t, ok := tagField.Interface().(*Tag); ok {
													tag = t
													tagCode = t.Raw()
													tagLineNumber = t.LineNumber()
													return true
												}
											}
										}
									} else if field.Type.Kind() == reflect.Ptr {
										// Recursively check embedded types
										if val.IsValid() {
											var embeddedVal reflect.Value
											if val.Kind() == reflect.Ptr {
												embeddedVal = val.Elem().Field(i)
											} else {
												embeddedVal = val.Field(i)
											}
											if embeddedVal.IsValid() && !embeddedVal.IsNil() {
												if findTag(embeddedVal, field.Type) {
													return true
												}
											}
										}
									}
								}
							}
							return false
						}
						findTag(nodeValue, actualType)
					}
					// If we still didn't find Tag, try to get Tag info via interface methods
					if tag == nil {
						if renderable, ok := node.(interface{ Raw() string }); ok {
							tagCode = renderable.Raw()
						}
						if lineNumberer, ok := node.(interface{ LineNumber() *int }); ok {
							tagLineNumber = lineNumberer.LineNumber()
						}
					}
				}
			}

			// If we found a Block or Tag, check if Render has been overridden
			if block != nil || blockBody != nil || tag != nil {
				// Call Render on the actual node type (not through Block/Tag)
				// Use reflection to call Render on the actual type
				nodeValue := reflect.ValueOf(node)
				// fmt.Printf("DEBUG block_body: node type=%T, block=%v, tag=%v\n", node, block != nil, tag != nil)
				if nodeValue.Kind() == reflect.Ptr {
					actualType := nodeValue.Type()
					// Check if actual type is not *Block (meaning it's a subtype)
					isActualBlock := actualType == reflect.TypeOf((*Block)(nil))
					isActualTag := actualType == reflect.TypeOf((*Tag)(nil))
					// fmt.Printf("DEBUG block_body: isActualBlock=%v, isActualTag=%v\n", isActualBlock, isActualTag)

					// If it's a tag subtype (not block subtype, not actual Block or Tag), check Render first
					if !isActualBlock && !isActualTag && block == nil {
						// Check if this tag is blank (like capture, assign, etc.)
						// Blank tags should not output their rendered content
						isBlank := false
						if blanker, ok := node.(interface{ Blank() bool }); ok {
							isBlank = blanker.Blank()
						}

						// Only use Render result if the tag is NOT blank
						if !isBlank {
							renderMethod := nodeValue.MethodByName("Render")
							if renderMethod.IsValid() {
								// Call Render on the actual type (e.g., TestTag1.Render)
								results := renderMethod.Call([]reflect.Value{reflect.ValueOf(context)})
								if len(results) > 0 {
									renderResult := results[0].String()
									// For subtypes, always use Render result if it's non-empty
									// (it overrides Block.Render which renders body)
									if renderResult != "" {
										// Use the Render result
										if profiler {
											code := tagCode
											lineNumber := tagLineNumber
											ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
												*output += renderResult
											})
										} else {
											*output += renderResult
										}
										// Check for interrupts
										if ctx, ok := context.(*Context); ok && ctx.Interrupt() {
											break
										}
										continue
									}
								}
							}
						}
					} else if block != nil {
						// For blocks (actual Block or subtypes), check if Render returns something different from body
						// This handles backwards compatibility: if a custom block overrides Render(), use that.
						// Otherwise, fall through to RenderToOutputBuffer.
						renderMethod := nodeValue.MethodByName("Render")
						if renderMethod.IsValid() {
							results := renderMethod.Call([]reflect.Value{reflect.ValueOf(context)})
							if len(results) > 0 {
								renderResult := results[0].String()
								bodyResult := ""
								if blockBody != nil {
									bodyResult = blockBody.Render(context)
								}
								// Only use Render if it's different from body
								// This means Render() was overridden (backwards compat case like TestBlockTag)
								if renderResult != bodyResult {
									if profiler {
										code := tagCode
										lineNumber := tagLineNumber
										ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
											*output += renderResult
										})
									} else {
										*output += renderResult
									}
									if ctx, ok := context.(*Context); ok && ctx.Interrupt() {
										break
									}
									continue
								}
								// If Render() returns same as body, fall through to RenderToOutputBuffer
								// This handles modern blocks like DisableCustomBlock that override RenderToOutputBuffer
							}
						}
					}
				}
			}

			// Default: check if node implements RenderToOutputBuffer first (for subtypes like CaptureTag)
			// This ensures we call the most specific implementation (e.g., CaptureTag.RenderToOutputBuffer)
			// instead of Tag.RenderToOutputBuffer
			if renderable, ok := node.(interface{ RenderToOutputBuffer(TagContext, *string) }); ok {
				// Handle tags that implement RenderToOutputBuffer directly
				if profiler {
					code := tagCode
					lineNumber := tagLineNumber
					if code == "" {
						if t, ok := renderable.(interface{ Raw() string }); ok {
							code = t.Raw()
						}
					}
					if tagLineNumber == nil {
						if t, ok := renderable.(interface{ LineNumber() *int }); ok {
							lineNumber = t.LineNumber()
						}
					}
					ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
						renderable.RenderToOutputBuffer(context, output)
					})
				} else {
					renderable.RenderToOutputBuffer(context, output)
				}
			} else if tag != nil {
				// Fallback: render using Tag's RenderToOutputBuffer
				if profiler {
					code := tagCode
					lineNumber := tagLineNumber
					ctx.Profiler().ProfileNode(ctx.TemplateName(), code, lineNumber, func() {
						tag.RenderToOutputBuffer(context, output)
					})
				} else {
					tag.RenderToOutputBuffer(context, output)
				}
			}
			// Check for interrupts (break, continue)
			// If we get an Interrupt that means the block must stop processing.
			// An Interrupt is any command that stops block execution such as {% break %}
			// or {% continue %}. These tags may also occur through Block or Include tags.
			if ctx, ok := context.(*Context); ok && ctx.Interrupt() {
				break
			}
		}

		// Increment write score after each node (like Ruby: context.resource_limits.increment_write_score(output))
		// This tracks assign_score when lastCaptureLength is set (during capture)
		if ctx != nil {
			rl := ctx.ResourceLimits()
			if rl != nil {
				rl.IncrementWriteScore(*output)
			}
		}
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
	bb.parseForLiquidTag(liquidTagTokenizer, parseContext, func(endTagName, _endTagMarkup string) bool {
		if endTagName != "" {
			// Unknown tag in liquid tag - raise error
			// This would call Block.raise_unknown_tag in Ruby
			// For now, we'll raise a syntax error
			panic(NewSyntaxError("Unknown tag '" + endTagName + "' in liquid tag"))
		}
		return true
	})
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
