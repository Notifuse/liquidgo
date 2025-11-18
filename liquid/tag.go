package liquid

import (
	"fmt"
	"strings"
)

// TagContext interface for tag rendering context
type TagContext interface {
	TagDisabled(tagName string) bool
	WithDisabledTags(tags []string, fn func())
	HandleError(err error, lineNumber *int) string
	ParseContext() ParseContextInterface
	Evaluate(expr interface{}) interface{}
	FindVariable(key string, raiseOnNotFound bool) interface{}
	Invoke(method string, obj interface{}, args ...interface{}) interface{}
	ApplyGlobalFilter(obj interface{}) interface{}
	Interrupt() bool
	PushInterrupt(interrupt interface{})
	ResourceLimits() *ResourceLimits
	Registers() *Registers
	Context() interface{}
}

// Tag represents a Liquid tag.
type Tag struct {
	tagName      string
	markup       string
	parseContext ParseContextInterface
	lineNumber   *int
	nodelist     []interface{} // For block tags
}

// NewTag creates a new tag instance.
func NewTag(tagName, markup string, parseContext ParseContextInterface) *Tag {
	lineNum := parseContext.LineNumber()
	return &Tag{
		tagName:      tagName,
		markup:       markup,
		parseContext: parseContext,
		lineNumber:   lineNum,
		nodelist:     []interface{}{},
	}
}

// ParseTag parses a tag from tokenizer.
func ParseTag(tagName, markup string, tokenizer *Tokenizer, parseContext ParseContextInterface) (*Tag, error) {
	tag := NewTag(tagName, markup, parseContext)
	err := tag.Parse(tokenizer)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// TagName returns the tag name.
func (t *Tag) TagName() string {
	return t.tagName
}

// Markup returns the tag markup.
func (t *Tag) Markup() string {
	return t.markup
}

// LineNumber returns the line number.
func (t *Tag) LineNumber() *int {
	return t.lineNumber
}

// ParseContext returns the parse context.
func (t *Tag) ParseContext() ParseContextInterface {
	return t.parseContext
}

// Raw returns the raw tag representation.
func (t *Tag) Raw() string {
	return fmt.Sprintf("%s %s", t.tagName, t.markup)
}

// Name returns the tag name (for backwards compatibility).
func (t *Tag) Name() string {
	return strings.ToLower(t.tagName)
}

// Parse parses the tag (can be overridden by subclasses).
func (t *Tag) Parse(tokenizer *Tokenizer) error {
	// Default implementation does nothing
	return nil
}

// Render renders the tag (returns empty string by default).
func (t *Tag) Render(context TagContext) string {
	return ""
}

// RenderToOutputBuffer renders the tag to the output buffer.
// Note: Due to Go's method dispatch with embedded pointers, this method cannot
// automatically call overridden Render() methods in subtypes. Tags that override
// Render() must be handled specially in the rendering code.
func (t *Tag) RenderToOutputBuffer(context TagContext, output *string) {
	renderResult := t.Render(context)
	if renderResult != "" {
		*output += renderResult
	}
}

// Blank returns true if the tag is blank (produces no output).
func (t *Tag) Blank() bool {
	return false
}

// Nodelist returns the nodelist (for block tags).
func (t *Tag) Nodelist() []interface{} {
	return t.nodelist
}

// SetNodelist sets the nodelist.
func (t *Tag) SetNodelist(nodelist []interface{}) {
	t.nodelist = nodelist
}

// SafeParseExpression safely parses an expression.
func (t *Tag) SafeParseExpression(parser *Parser) interface{} {
	return t.parseContext.SafeParseExpression(parser)
}

// ParseExpression parses an expression.
func (t *Tag) ParseExpression(markup string, safe bool) interface{} {
	return t.parseContext.ParseExpression(markup)
}
