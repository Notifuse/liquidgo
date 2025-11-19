package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var (
	// includeSyntax matches: 'filename' [with|for expression] [as alias] [attributes...]
	// Note: QuotedFragment already has +, so we don't add another
	includeSyntax = regexp.MustCompile(`(` + liquid.QuotedFragment.String() + `)(\s+(?:with|for)\s+(` + liquid.QuotedFragment.String() + `))?(\s+(?:as)\s+(` + liquid.VariableSegment.String() + `+))?`)
)

// IncludeTag represents an include tag that renders a partial template.
// Note: This is deprecated in favor of render tag.
type IncludeTag struct {
	*liquid.Tag
	templateNameExpr interface{}            // Expression
	variableNameExpr interface{}            // Expression or nil
	aliasName        string                 // Alias name or ""
	attributes       map[string]interface{} // Attribute expressions
}

// NewIncludeTag creates a new IncludeTag.
func NewIncludeTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*IncludeTag, error) {
	tag := liquid.NewTag(tagName, markup, parseContext)

	includeTag := &IncludeTag{
		Tag:        tag,
		attributes: make(map[string]interface{}),
	}

	// Parse markup
	err := includeTag.parseMarkup(markup, parseContext)
	if err != nil {
		return nil, err
	}

	return includeTag, nil
}

// parseMarkup parses the include tag markup.
func (i *IncludeTag) parseMarkup(markup string, parseContext liquid.ParseContextInterface) error {
	matches := includeSyntax.FindStringSubmatch(markup)
	if len(matches) == 0 {
		return liquid.NewSyntaxError("invalid include tag syntax")
	}

	// Template name
	templateNameStr := matches[1]
	i.templateNameExpr = parseContext.ParseExpression(templateNameStr)

	// Variable name (with/for)
	if len(matches) > 3 && matches[3] != "" {
		i.variableNameExpr = parseContext.ParseExpression(matches[3])
	}

	// Alias name (as)
	if len(matches) > 5 && matches[5] != "" {
		i.aliasName = matches[5]
	}

	// Parse attributes
	attributeMatches := liquid.TagAttributes.FindAllStringSubmatch(markup, -1)
	for _, match := range attributeMatches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			// Let ParseExpression handle quoted strings correctly
			i.attributes[key] = parseContext.ParseExpression(value)
		}
	}

	return nil
}

// Parse parses the include tag (no-op for include).
func (i *IncludeTag) Parse(tokenizer *liquid.Tokenizer) error {
	// Include tag doesn't parse tokens
	return nil
}

// RenderToOutputBuffer renders the include tag.
func (i *IncludeTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	// Evaluate template name
	templateName := context.Evaluate(i.templateNameExpr)
	templateNameStr, ok := templateName.(string)
	if !ok {
		var locale *liquid.I18n
		if pc, ok := i.ParseContext().(*liquid.ParseContext); ok {
			locale = pc.Locale()
		}
		var msg string
		if locale != nil {
			msg = locale.T("errors.argument.include", nil)
		} else {
			msg = "include tag requires a string template name"
		}
		errorMsg := context.HandleError(liquid.NewArgumentError(msg), nil)
		*output += errorMsg
		return
	}

	// Load partial from cache
	partial, err := liquid.LoadPartial(templateNameStr, context, i.ParseContext())
	if err != nil {
		errorMsg := context.HandleError(err, nil)
		*output += errorMsg
		return
	}

	// Get partial as Template
	partialTemplate, ok := partial.(*liquid.Template)
	if !ok {
		errorMsg := context.HandleError(liquid.NewFileSystemError("partial is not a template"), nil)
		*output += errorMsg
		return
	}

	// Determine context variable name
	contextVariableName := i.aliasName
	if contextVariableName == "" {
		// Use last part of template name (split by '/')
		parts := strings.Split(templateNameStr, "/")
		contextVariableName = parts[len(parts)-1]
	}

	// Get variable value
	var variable interface{}
	if i.variableNameExpr != nil {
		variable = context.Evaluate(i.variableNameExpr)
	} else {
		// Find variable by template name
		if ctx, ok := context.(*liquid.Context); ok {
			variable = ctx.FindVariable(templateNameStr, false)
		}
	}

	// Get context as *Context for state management
	ctx, ok := context.(*liquid.Context)
	if !ok {
		errorMsg := context.HandleError(liquid.NewInternalError("context is not a liquid.Context"), nil)
		*output += errorMsg
		return
	}

	// Save old state
	oldTemplateName := ctx.TemplateName()
	oldPartial := ctx.Partial()

	// Set new state
	partialName := partialTemplate.Name()
	if partialName == "" {
		partialName = templateNameStr
	}
	ctx.SetTemplateName(partialName)
	ctx.SetPartial(true)

	// Use stack to create isolated scope
	ctx.Stack(make(map[string]interface{}), func() {
		// Set attributes
		for key, valueExpr := range i.attributes {
			value := context.Evaluate(valueExpr)
			ctx.Set(key, value)
		}

		// Render partial with variable
		if arr, ok := variable.([]interface{}); ok {
			// Array: render once for each item
			for _, varItem := range arr {
				ctx.Set(contextVariableName, varItem)
				partialTemplate.RenderToOutputBuffer(ctx, output)
			}
		} else {
			// Single value
			ctx.Set(contextVariableName, variable)
			partialTemplate.RenderToOutputBuffer(ctx, output)
		}
	})

	// Restore old state
	ctx.SetTemplateName(oldTemplateName)
	ctx.SetPartial(oldPartial)
}

// TemplateNameExpr returns the template name expression.
func (i *IncludeTag) TemplateNameExpr() interface{} {
	return i.templateNameExpr
}

// VariableNameExpr returns the variable name expression.
func (i *IncludeTag) VariableNameExpr() interface{} {
	return i.variableNameExpr
}

// AliasName returns the alias name.
func (i *IncludeTag) AliasName() string {
	return i.aliasName
}

// Attributes returns the attributes map.
func (i *IncludeTag) Attributes() map[string]interface{} {
	return i.attributes
}
