package liquid

// Document represents the root node of a Liquid template parse tree.
type Document struct {
	parseContext ParseContextInterface
	body         *BlockBody
}

// ParseDocument parses tokens into a Document.
// Catches panics from parsing and converts them to errors to prevent application crashes.
func ParseDocument(tokenizer *Tokenizer, parseContext ParseContextInterface) (*Document, error) {
	doc := NewDocument(parseContext)

	// Catch panics from parsing and convert to errors
	var parseErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(*SyntaxError); ok {
					parseErr = err
				} else if err, ok := r.(error); ok {
					parseErr = err
				} else {
					// Re-panic non-error panics
					panic(r)
				}
			}
		}()
		err := doc.Parse(tokenizer, parseContext)
		if err != nil {
			parseErr = err
		}
	}()

	if parseErr != nil {
		return nil, parseErr
	}
	return doc, nil
}

// NewDocument creates a new Document.
func NewDocument(parseContext ParseContextInterface) *Document {
	return &Document{
		parseContext: parseContext,
		body:         parseContext.NewBlockBody(),
	}
}

// ParseContext returns the parse context.
func (d *Document) ParseContext() ParseContextInterface {
	return d.parseContext
}

// Body returns the body.
func (d *Document) Body() *BlockBody {
	return d.body
}

// Nodelist returns the nodelist from the body.
func (d *Document) Nodelist() []interface{} {
	return d.body.Nodelist()
}

// Parse parses tokens into the document.
func (d *Document) Parse(tokenizer *Tokenizer, parseContext ParseContextInterface) error {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(*SyntaxError); ok {
				if err.Err.LineNumber == nil {
					err.Err.LineNumber = parseContext.LineNumber()
				}
				panic(err)
			}
			panic(r)
		}
	}()

	for {
		shouldContinue := d.parseBody(tokenizer, parseContext)
		if !shouldContinue {
			break
		}
	}

	return nil
}

// UnknownTag handles unknown tags encountered during parsing.
func (d *Document) UnknownTag(tag, markup string, tokenizer *Tokenizer) error {
	var msg string
	switch tag {
	case "else", "end":
		// Get locale from parse context if it's a ParseContext struct
		var locale *I18n
		if pc, ok := d.parseContext.(*ParseContext); ok {
			locale = pc.Locale()
			msg = locale.Translate("errors.syntax.unexpected_outer_tag", map[string]interface{}{"tag": tag})
		} else {
			msg = "Liquid syntax error: unexpected outer tag " + tag
		}
		return NewSyntaxError(msg)
	default:
		// Get locale from parse context if it's a ParseContext struct
		var locale *I18n
		if pc, ok := d.parseContext.(*ParseContext); ok {
			locale = pc.Locale()
			msg = locale.Translate("errors.syntax.unknown_tag", map[string]interface{}{"tag": tag})
		} else {
			msg = "Liquid syntax error: unknown tag " + tag
		}
		return NewSyntaxError(msg)
	}
}

// RenderToOutputBuffer renders the document to an output buffer.
func (d *Document) RenderToOutputBuffer(context TagContext, output *string) {
	// Check if profiling is enabled
	if ctx, ok := context.(*Context); ok && ctx.Profiler() != nil {
		templateName := ctx.TemplateName()
		ctx.Profiler().Profile(templateName, func() {
			d.body.RenderToOutputBuffer(context, output)
		})
		return
	}

	d.body.RenderToOutputBuffer(context, output)
}

// Render renders the document and returns the result as a string.
func (d *Document) Render(context TagContext) string {
	var output string
	d.RenderToOutputBuffer(context, &output)
	return output
}

func (d *Document) parseBody(tokenizer *Tokenizer, parseContext ParseContextInterface) bool {
	unknownTagHandler := func(unknownTagName, unknownTagMarkup string) bool {
		if unknownTagName != "" {
			err := d.UnknownTag(unknownTagName, unknownTagMarkup, tokenizer)
			if err != nil {
				if d.parseContext.ErrorMode() == "warn" {
					d.parseContext.AddWarning(err)
					return true
				}
				panic(err)
			}
			return true
		}
		return false
	}

	err := d.body.Parse(tokenizer, parseContext, unknownTagHandler)
	if err != nil {
		panic(err)
	}

	// Return false to stop parsing (body.Parse handles the loop)
	return false
}
