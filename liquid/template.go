package liquid

import (
	"sync/atomic"
	"unicode/utf8"
)

// Template represents a compiled Liquid template.
// Templates are central to liquid. Interpreting templates is a two step process.
// First you compile the source code you got. During compile time some extensive
// error checking is performed. Your code should expect to get some SyntaxErrors.
//
// After you have a compiled template you can then render it, over and over
// again, from any number of goroutines: a render changes nothing in the
// template (see Render).
//
// Example:
//
//	tmpl, err := liquid.ParseTemplate(source, nil)
//	res, err := tmpl.Render(ctx, map[string]interface{}{"user_name": "bob"})
type Template struct {
	environment *Environment
	// limits are the resource limits of a render, from the environment's
	// defaults; WithLimits changes them for one render.
	limits      ResourceLimitsConfig
	root        *Document
	registers   map[string]interface{}
	assigns     map[string]interface{}
	name        string
	warnings    []error
	lineNumbers bool
	// outputSize is the length of the output of the last render, so that
	// the next render makes its buffer at once.
	outputSize atomic.Int64
}

// TemplateOptions contains options for parsing a template.
type TemplateOptions struct {
	Environment *Environment
	LineNumbers bool
}

// ParseTemplate creates a new Template and parses the source code.
func ParseTemplate(source string, options *TemplateOptions) (*Template, error) {
	template := NewTemplate(options)
	err := template.Parse(source, options)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// NewTemplate creates a new Template instance.
func NewTemplate(options *TemplateOptions) *Template {
	var env *Environment
	if options != nil && options.Environment != nil {
		env = options.Environment
	} else {
		env = newEnvironment()
	}

	return &Template{
		environment: env,
		limits:      copyLimits(env.limits),
		registers:   make(map[string]interface{}, 4),
		assigns:     make(map[string]interface{}, 8),
	}
}

// Parse parses source code.
// Returns self for easy chaining.
func (t *Template) Parse(source string, options *TemplateOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	parseContext := t.configureOptions(options)

	// Convert source to string
	sourceStr := source

	// Validate encoding (Go strings are UTF-8 by default, but we should validate)
	// In Ruby: unless source.valid_encoding?
	if !isValidUTF8(sourceStr) {
		var locale *I18n
		if pc, ok := parseContext.(*ParseContext); ok {
			locale = pc.Locale()
		}
		var msg string
		if locale != nil {
			msg = locale.T("errors.syntax.invalid_template_encoding", nil)
		} else {
			msg = "Invalid template encoding"
		}
		return NewTemplateEncodingError(msg)
	}

	// Create tokenizer
	var startLineNumber *int
	if t.lineNumbers {
		lineNum := 1
		startLineNumber = &lineNum
	}
	tokenizer := parseContext.NewTokenizer(sourceStr, false, startLineNumber, false)

	// Parse document
	root, err := ParseDocument(tokenizer, parseContext)
	if err != nil {
		return err
	}

	t.root = root
	// Get warnings from parse context if available
	if pc, ok := parseContext.(*ParseContext); ok {
		t.warnings = pc.Warnings()
	}

	return nil
}

// isValidUTF8 checks if a string is valid UTF-8.
func isValidUTF8(s string) bool {
	return utf8.ValidString(s)
}

// Environment returns the environment the template was parsed with.
func (t *Template) Environment() *Environment {
	return t.environment
}

// Root returns the root document.
func (t *Template) Root() *Document {
	return t.root
}

// SetRoot sets the root document.
func (t *Template) SetRoot(root *Document) {
	t.root = root
}

// Name returns the template name.
func (t *Template) Name() string {
	return t.name
}

// SetName sets the template name.
func (t *Template) SetName(name string) {
	t.name = name
}

// Warnings returns the warnings.
func (t *Template) Warnings() []error {
	if t.warnings == nil {
		return []error{}
	}
	return t.warnings
}

// Registers returns the registers map.
func (t *Template) Registers() map[string]interface{} {
	if t.registers == nil {
		t.registers = make(map[string]interface{})
	}
	return t.registers
}

// Assigns returns the assigns map.
func (t *Template) Assigns() map[string]interface{} {
	if t.assigns == nil {
		t.assigns = make(map[string]interface{})
	}
	return t.assigns
}

// RenderToOutput renders the template with context to output, as a partial
// of the render that context belongs to: the resource usage goes on from
// where it is, and an error that ends the render, such as a MemoryError, goes
// on up.
func (t *Template) RenderToOutput(context TagContext, output *OutputBuffer) {
	if t.root == nil {
		return
	}
	if ctx, ok := context.(*Context); ok && ctx.TemplateName() == "" {
		ctx.SetTemplateName(t.name)
	}
	t.root.RenderToOutput(context, output)
}

// configureOptions configures parse options and returns a ParseContext.
func (t *Template) configureOptions(options *TemplateOptions) ParseContextInterface {
	if options == nil {
		options = &TemplateOptions{}
	}

	if options.Environment != nil {
		t.environment = options.Environment
	} else if t.environment == nil {
		t.environment = newEnvironment()
	}

	t.lineNumbers = options.LineNumbers

	// Create parse context
	templateOpts := make(map[string]interface{})
	if t.lineNumbers {
		templateOpts["line_numbers"] = true
	}

	parseContextOpts := ParseContextOptions{
		Environment:     t.environment,
		TemplateOptions: templateOpts,
	}

	return NewParseContext(parseContextOpts)
}
