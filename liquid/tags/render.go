package tags

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	// renderSyntax matches: 'filename'|variable [with|for expression] [as alias] [attributes...]
	// Note: QuotedFragment and QuotedString already have +, so we don't add another
	renderSyntax = regexp.MustCompile(`(` + liquid.QuotedString.String() + `|` + liquid.VariableSegment.String() + `+)(\s+(with|for)\s+(` + liquid.QuotedFragment.String() + `))?(\s+(?:as)\s+(` + liquid.VariableSegment.String() + `+))?`)
)

// RenderTag represents a render tag that renders a partial template with isolated context.
type RenderTag struct {
	*liquid.Tag
	templateNameExpr interface{}            // Expression
	variableNameExpr interface{}            // Expression or nil
	aliasName        string                 // Alias name or ""
	attributes       map[string]interface{} // Attribute expressions
	isForLoop        bool                   // True if "for" was used instead of "with"
}

// NewRenderTag creates a new RenderTag.
func NewRenderTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*RenderTag, error) {
	tag := liquid.NewTag(tagName, markup, parseContext)

	renderTag := &RenderTag{
		Tag:        tag,
		attributes: make(map[string]interface{}),
		isForLoop:  false,
	}

	// Parse markup
	err := renderTag.parseMarkup(markup, parseContext)
	if err != nil {
		return nil, err
	}

	return renderTag, nil
}

// parseMarkup parses the render tag markup.
func (r *RenderTag) parseMarkup(markup string, parseContext liquid.ParseContextInterface) error {
	matches := renderSyntax.FindStringSubmatch(markup)
	if len(matches) == 0 {
		return liquid.NewSyntaxError("invalid render tag syntax")
	}

	// Template name
	templateNameStr := matches[1]
	r.templateNameExpr = parseContext.ParseExpression(templateNameStr)

	// Variable name (with/for)
	if len(matches) > 4 && matches[4] != "" {
		r.variableNameExpr = parseContext.ParseExpression(matches[4])
		// Check if it's "for" (isForLoop) or "with"
		if len(matches) > 3 && matches[3] == "for" {
			r.isForLoop = true
		}
	}

	// Alias name (as)
	if len(matches) > 6 && matches[6] != "" {
		r.aliasName = matches[6]
	}

	// Parse attributes
	attributeMatches := liquid.TagAttributes.FindAllStringSubmatch(markup, -1)
	for _, match := range attributeMatches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			// Let ParseExpression handle quoted strings correctly
			r.attributes[key] = parseContext.ParseExpression(value)
		}
	}

	return nil
}

// Parse parses the render tag (no-op for render).
func (r *RenderTag) Parse(tokenizer *liquid.Tokenizer) error {
	// Render tag doesn't parse tokens
	return nil
}

// RenderToOutputBuffer renders the render tag.
func (r *RenderTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Evaluate template name
	template := context.Evaluate(r.templateNameExpr)

	// Get context as *Context
	ctx, ok := context.(*liquid.Context)
	if !ok {
		errorMsg := context.HandleError(liquid.NewInternalError("context is not a liquid.Context"), r.LineNumber())
		*output += errorMsg
		return
	}

	var partial *liquid.Template
	var templateName string
	var contextVariableName string

	// Check if template responds to to_partial returning a string (for SnippetDrop)
	if toPartialStr, ok := template.(interface{ ToPartial() string }); ok {
		// Parse the body string as a template
		body := toPartialStr.ToPartial()
		if filename, ok := template.(interface{ Filename() string }); ok {
			templateName = filename.Filename()
		}
		if name, ok := template.(interface{ Name() string }); ok {
			contextVariableName = r.aliasName
			if contextVariableName == "" {
				contextVariableName = name.Name()
			}
		}

		// Parse the body as a template
		parsedTemplate, err := liquid.ParseTemplate(body, &liquid.TemplateOptions{
			Environment: r.ParseContext().Environment(),
		})
		if err != nil {
			errorMsg := context.HandleError(err, r.LineNumber())
			*output += errorMsg
			return
		}
		partial = parsedTemplate
	} else if toPartial, ok := template.(interface{ ToPartial() *liquid.Template }); ok {
		// Check if template responds to to_partial (for template objects)
		partial = toPartial.ToPartial()
		if filename, ok := template.(interface{ Filename() string }); ok {
			templateName = filename.Filename()
		}
		if name, ok := template.(interface{ Name() string }); ok {
			contextVariableName = r.aliasName
			if contextVariableName == "" {
				contextVariableName = name.Name()
			}
		}
	} else if templateNameStr, ok := template.(string); ok {
		// String template name - load from cache
		partialInterface, err := liquid.LoadPartial(templateNameStr, context, r.ParseContext())
		if err != nil {
			errorMsg := context.HandleError(err, r.LineNumber())
			*output += errorMsg
			return
		}
		var ok bool
		partial, ok = partialInterface.(*liquid.Template)
		if !ok {
			errorMsg := context.HandleError(liquid.NewFileSystemError("partial is not a template"), r.LineNumber())
			*output += errorMsg
			return
		}
		templateName = partial.Name()
		if templateName == "" {
			templateName = templateNameStr
		}
		contextVariableName = r.aliasName
		if contextVariableName == "" {
			// Use last part of template name
			parts := strings.Split(templateNameStr, "/")
			contextVariableName = parts[len(parts)-1]
		}
	} else {
		errorMsg := context.HandleError(liquid.NewArgumentError("render tag requires a string template name or template object"), r.LineNumber())
		*output += errorMsg
		return
	}

	// Render partial function
	renderPartialFunc := func(varItem interface{}, forloop *liquid.ForloopDrop) {
		innerContext := ctx.NewIsolatedSubcontext()
		innerContext.SetTemplateName(templateName)
		innerContext.SetPartial(true)

		// Set forloop if provided
		if forloop != nil {
			innerContext.Set("forloop", forloop)
		}

		// Set attributes
		for key, valueExpr := range r.attributes {
			value := context.Evaluate(valueExpr)
			innerContext.Set(key, value)
		}

		// Set variable unless nil
		if varItem != nil {
			innerContext.Set(contextVariableName, varItem)
		}

		// Render partial
		partial.RenderToOutputBuffer(innerContext, output)

		// Increment forloop if provided
		if forloop != nil {
			forloop.Increment()
		}
	}

	// Get variable value
	var variable interface{}
	if r.variableNameExpr != nil {
		variable = context.Evaluate(r.variableNameExpr)
	}

	// Handle for loop or single render
	if r.isForLoop && variable != nil {
		// Check if variable is iterable (has Each and Count methods)
		if iterable, ok := variable.(interface {
			Each(func(interface{}))
			Count() int
		}); ok {
			count := iterable.Count()
			forloop := liquid.NewForloopDrop(templateName, count, nil)
			iterable.Each(func(item interface{}) {
				renderPartialFunc(item, forloop)
			})
		} else if arr, ok := variable.([]interface{}); ok {
			// Array fallback
			forloop := liquid.NewForloopDrop(templateName, len(arr), nil)
			for _, item := range arr {
				renderPartialFunc(item, forloop)
			}
		} else if variable != nil {
			// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
			// This matches Ruby's duck-typing behavior: arrays respond to iteration
			v := reflect.ValueOf(variable)
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				forloop := liquid.NewForloopDrop(templateName, v.Len(), nil)
				for i := 0; i < v.Len(); i++ {
					renderPartialFunc(v.Index(i).Interface(), forloop)
				}
			} else {
				// Not iterable, render once with variable
				renderPartialFunc(variable, nil)
			}
		} else {
			// Not iterable, render once with variable
			renderPartialFunc(variable, nil)
		}
	} else {
		// Single render
		renderPartialFunc(variable, nil)
	}
}

// TemplateNameExpr returns the template name expression.
func (r *RenderTag) TemplateNameExpr() interface{} {
	return r.templateNameExpr
}

// VariableNameExpr returns the variable name expression.
func (r *RenderTag) VariableNameExpr() interface{} {
	return r.variableNameExpr
}

// AliasName returns the alias name.
func (r *RenderTag) AliasName() string {
	return r.aliasName
}

// Attributes returns the attributes map.
func (r *RenderTag) Attributes() map[string]interface{} {
	return r.attributes
}

// IsForLoop returns true if this is a for loop render.
func (r *RenderTag) IsForLoop() bool {
	return r.isForLoop
}
