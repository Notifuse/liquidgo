package liquid

import (
	gocontext "context"
	"math"
	"reflect"
	"sync"
	"unicode/utf8"
)

// cloneStringMap creates a shallow copy of a map for thread-safe concurrent rendering.
func cloneStringMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return make(map[string]interface{})
	}
	cloned := make(map[string]interface{}, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}

// cloneResourceLimits creates a new ResourceLimits with the same limits for thread-safe concurrent rendering.
func cloneResourceLimits(rl *ResourceLimits) *ResourceLimits {
	if rl == nil {
		return NewResourceLimits(ResourceLimitsConfig{})
	}
	return NewResourceLimits(ResourceLimitsConfig{
		RenderLengthLimit: rl.renderLengthLimit,
		RenderScoreLimit:  rl.renderScoreLimit,
		AssignScoreLimit:  rl.assignScoreLimit,
	})
}

// Template represents a compiled Liquid template.
// Templates are central to liquid. Interpreting templates is a two step process.
// First you compile the source code you got. During compile time some extensive
// error checking is performed. Your code should expect to get some SyntaxErrors.
//
// After you have a compiled template you can then render it.
// You can use a compiled template over and over again and keep it cached.
//
// Template is safe for concurrent use. When rendered concurrently, each render
// gets its own isolated scope. State updates (errors, assigns) are synchronized
// with a mutex.
//
// Example:
//
//	template := liquid.ParseTemplate(source)
//	result := template.Render(map[string]interface{}{"user_name": "bob"})
type Template struct {
	mu              sync.Mutex // Protects concurrent access to mutable state
	environment     *Environment
	resourceLimits  *ResourceLimits
	root            *Document
	profiler        *Profiler
	registers       map[string]interface{}
	assigns         map[string]interface{}
	instanceAssigns map[string]interface{}
	name            string
	warnings        []error
	errors          []error
	lineNumbers     bool
	profiling       bool
	// persistInstanceAssigns: see TemplateOptions.PersistInstanceAssigns.
	persistInstanceAssigns bool
}

// TemplateOptions contains options for parsing a template.
type TemplateOptions struct {
	Environment       *Environment
	GlobalFilter      func(interface{}) interface{}
	ExceptionRenderer func(error) interface{}
	Registers         map[string]interface{}
	Filters           []interface{}
	Profile           bool
	LineNumbers       bool
	StrictVariables   bool
	StrictFilters     bool
	// PersistInstanceAssigns keeps what a render assigns (assign, capture) in
	// the template and gives it to the next render, as Ruby's Template does.
	// Off by default: each render starts from its own data only, so a value
	// assigned while rendering for one recipient never reaches the next.
	PersistInstanceAssigns bool
}

// ParseTemplate creates a new Template and parses the source code.
// To enable profiling, pass in Profile: true as an option.
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
		env = NewEnvironment()
	}

	var resourceLimits *ResourceLimits
	config := ResourceLimitsConfig{}
	if env.defaultResourceLimits != nil {
		// Convert map[string]interface{} to ResourceLimitsConfig
		config.RenderLengthLimit = limitValue(env.defaultResourceLimits["render_length_limit"])
		config.RenderScoreLimit = limitValue(env.defaultResourceLimits["render_score_limit"])
		config.AssignScoreLimit = limitValue(env.defaultResourceLimits["assign_score_limit"])
	}
	resourceLimits = NewResourceLimits(config)

	return &Template{
		persistInstanceAssigns: options != nil && options.PersistInstanceAssigns,
		environment:            env,
		resourceLimits:         resourceLimits,
		registers:              make(map[string]interface{}, 4), // Pre-allocate map
		assigns:                make(map[string]interface{}, 8), // Pre-allocate for typical assigns
		instanceAssigns:        make(map[string]interface{}, 8), // Pre-allocate for instance assigns
		errors:                 make([]error, 0, 2),             // Pre-allocate errors
		warnings:               make([]error, 0, 2),             // Pre-allocate warnings
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

// ResourceLimits returns the resource limits.
func (t *Template) ResourceLimits() *ResourceLimits {
	return t.resourceLimits
}

// SetResourceLimits sets the resource limits.
func (t *Template) SetResourceLimits(rl *ResourceLimits) {
	t.resourceLimits = rl
}

// Warnings returns the warnings.
func (t *Template) Warnings() []error {
	return t.warnings
}

// Profiler returns the profiler (if profiling was enabled).
func (t *Template) Profiler() *Profiler {
	return t.profiler
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

// SetPersistInstanceAssigns turns on or off Ruby's persistence of a render's
// assigns into the next render (see TemplateOptions.PersistInstanceAssigns).
// Turning it off forgets what earlier renders kept.
func (t *Template) SetPersistInstanceAssigns(on bool) {
	t.mu.Lock()
	t.persistInstanceAssigns = on
	if !on {
		t.instanceAssigns = make(map[string]interface{}, 8)
	}
	t.mu.Unlock()
}

// outerScope returns the outer scope a render starts from: a copy of what
// earlier renders kept with PersistInstanceAssigns, empty otherwise.
func (t *Template) outerScope() map[string]interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.persistInstanceAssigns {
		return make(map[string]interface{})
	}
	return cloneStringMap(t.instanceAssigns)
}

// InstanceAssigns returns the instance assigns map.
func (t *Template) InstanceAssigns() map[string]interface{} {
	if t.instanceAssigns == nil {
		t.instanceAssigns = make(map[string]interface{})
	}
	return t.instanceAssigns
}

// Errors returns the errors.
// With concurrent renders, they are those of the render that ended last.
func (t *Template) Errors() []error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.errors == nil {
		t.errors = []error{}
	}
	return t.errors
}

// Render renders the template with the given assigns.
// Render takes a hash with local variables.
//
// Following options can be passed via RenderOptions:
//
//   - Filters: array with local filters
//   - Registers: hash with register variables. Those can be accessed from
//     filters and tags and might be useful to integrate liquid more with its host application
func (t *Template) Render(assigns interface{}, options *RenderOptions) string {
	return t.render(assigns, options, false)
}

// render renders the template; rethrow makes an error raise out of the
// render, as Ruby's render!, for this render only.
func (t *Template) render(assigns interface{}, options *RenderOptions, rethrow bool) (output string) {
	if t.root == nil {
		return ""
	}

	// A Context of the caller keeps its own cancellation after this render.
	if ctx, ok := assigns.(*Context); ok && ctx != nil && options != nil && options.Context != nil {
		cancel, done := ctx.cancel, ctx.done
		defer func() { ctx.cancel, ctx.done = cancel, done }()
	}

	context := t.buildContext(assigns, options, rethrow)

	// Track whether we should merge back state (only when we create the context, not when user passes one)
	_, userProvidedContext := assigns.(*Context)

	// Create a cloned ResourceLimits for this render to avoid race conditions
	// when the same template is rendered concurrently
	if !userProvidedContext {
		if ctx, ok := context.(*Context); ok {
			ctx.SetResourceLimits(cloneResourceLimits(t.resourceLimits))
		}
	}

	// Reset resource usage for this render
	context.ResourceLimits().Reset()

	// Handle profiling
	if t.profiling {
		if ctx, ok := context.(*Context); ok && ctx.Profiler() == nil {
			t.profiler = NewProfiler()
			ctx.SetProfiler(t.profiler)
		}
	}

	// Cast to *Context to access TemplateName
	if ctx, ok := context.(*Context); ok {
		if ctx.TemplateName() == "" {
			ctx.SetTemplateName(t.name)
		}
	}

	// Use output from options if provided
	buffer := NewOutputBufferFor(context)
	if options != nil && options.Output != nil {
		buffer.buf = append(buffer.buf, *options.Output...)
	}

	defer func() {
		if r := recover(); r != nil {
			// Handle Liquid errors by converting them to error messages
			var err error
			var handled bool

			switch e := r.(type) {
			case *MemoryError:
				errorMsg := context.HandleError(e, nil)
				output = errorMsg
				if output == "" {
					output = "Liquid error: Memory limits exceeded"
				}
				handled = true
			case LiquidError:
				err = e
			case *Error:
				err = e
			case error:
				// Non-Liquid errors should be wrapped as InternalError
				err = NewInternalError("internal")
			default:
				// Non-error panics should be wrapped as InternalError
				err = NewInternalError("internal")
			}

			if !handled && err != nil {
				// Get context to handle the error
				if ctx, ok := context.(*Context); ok {
					errorMsg := ctx.HandleError(err, nil)
					output = errorMsg
					if output == "" {
						output = "Liquid error: internal"
					}
				} else {
					// Fallback if we can't get context
					output = "Liquid error: internal error"
				}
			} else if !handled {
				// Re-panic non-Liquid panics
				panic(r)
			}
		}
		// Update template state with mutex protection for thread-safe concurrent rendering
		if ctx, ok := context.(*Context); ok {
			t.mu.Lock()
			// Always capture errors from the render
			t.errors = ctx.Errors()
			// Only merge back instance assigns and resource limits when we created the context,
			// not when user passed their own Context
			if !userProvidedContext {
				// Merge back instance assigns from the render scope to persist across
				// renders, only when the template asks for Ruby's behaviour
				if t.persistInstanceAssigns && len(ctx.Scopes()) > 0 {
					lastScope := ctx.Scopes()[len(ctx.Scopes())-1]
					for k, v := range lastScope {
						if k != "__drop__" { // Don't persist internal drop reference
							t.instanceAssigns[k] = v
						}
					}
				}
				// Update template's resource limits from context's resource limits
				if ctx.ResourceLimits() != nil && t.resourceLimits != nil {
					ctxRL := ctx.ResourceLimits()
					t.resourceLimits.assignScore = ctxRL.AssignScore()
					t.resourceLimits.renderScore = ctxRL.RenderScore()
					t.resourceLimits.reachedLimit = ctxRL.Reached()
				}
			}
			t.mu.Unlock()
		}
		// Update output in options if provided
		if options != nil && options.Output != nil {
			*options.Output = output
		}
	}()

	t.root.RenderToOutput(context, buffer)
	output = buffer.String()
	return output
}

// RenderOptions contains options for rendering a template.
type RenderOptions struct {
	// Context, when set, stops the render once it is done: the render ends
	// with a CanceledError, which wraps the error of Context.
	Context           gocontext.Context
	Output            *string
	Registers         map[string]interface{}
	GlobalFilter      func(interface{}) interface{}
	ExceptionRenderer func(error) interface{}
	Filters           []interface{}
	StrictVariables   bool
	StrictFilters     bool
}

// RenderBang renders the template with rethrow_errors enabled: an error
// raises out of the render, as Ruby's render!. Unlike Ruby, this applies to
// this render only, and a later Render of the template records its errors.
func (t *Template) RenderBang(assigns interface{}, options *RenderOptions) string {
	return t.render(assigns, options, true)
}

// RenderToOutputBuffer renders the template with context and appends the
// result to output. As in Ruby, it resets the resource usage of context, and
// a MemoryError, or a CanceledError, is recorded rather than raised.
func (t *Template) RenderToOutputBuffer(context TagContext, output *string) {
	if t.root == nil {
		return
	}

	// Cast to *Context to access methods
	if ctx, ok := context.(*Context); ok {
		// Retrying a render resets resource usage
		ctx.ResourceLimits().Reset()

		if ctx.TemplateName() == "" {
			ctx.SetTemplateName(t.name)
		}

		defer func() {
			if r := recover(); r != nil {
				switch err := r.(type) {
				case *MemoryError:
					ctx.HandleError(err, nil)
				case *CanceledError:
					ctx.HandleError(err, nil)
				default:
					panic(r)
				}
			}
			t.errors = ctx.Errors()
		}()

		buffer := NewOutputBufferFor(context)
		defer func() { *output += buffer.String() }()
		t.root.RenderToOutput(context, buffer)
	} else {
		// Fallback: use Render method
		_ = t.Render(context, &RenderOptions{Output: output})
	}
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

// buildContext builds a Context from assigns and options.
func (t *Template) buildContext(assigns interface{}, options *RenderOptions, rethrow bool) TagContext {
	var ctx *Context

	switch v := assigns.(type) {
	case *Context:
		ctx = v
		if rethrow {
			ctx.SetExceptionRenderer(func(err error) interface{} {
				panic(err)
			})
		}
	case map[string]interface{}:
		ctx = BuildContext(ContextConfig{
			Environments:   []map[string]interface{}{v, t.assigns},
			OuterScope:     t.outerScope(),
			Registers:      NewRegisters(t.registers),
			ResourceLimits: t.resourceLimits,
			Environment:    t.environment,
			RethrowErrors:  rethrow,
		})
	case nil:
		ctx = BuildContext(ContextConfig{
			Environments:   []map[string]interface{}{t.assigns},
			OuterScope:     t.outerScope(),
			Registers:      NewRegisters(t.registers),
			ResourceLimits: t.resourceLimits,
			Environment:    t.environment,
			RethrowErrors:  rethrow,
		})
	default:
		// Check if it's a drop - if so, we need to make it accessible for variable lookups
		// In Ruby Liquid, drops can be passed as context and their methods become available as variables
		// assigns is not nil here (nil case handled above)
		dropToStore := assigns
		// Wrap the drop in the outer scope so it's accessible
		// The drop itself will be the context for variable lookups
		outerScope := map[string]interface{}{"__drop__": assigns}

		ctx = BuildContext(ContextConfig{
			Environments:   []map[string]interface{}{t.assigns},
			OuterScope:     outerScope,
			Registers:      NewRegisters(t.registers),
			ResourceLimits: t.resourceLimits,
			Environment:    t.environment,
			RethrowErrors:  rethrow,
		})

		// If assigns is a drop, we need special handling for variable lookups
		// Store it in a special way so FindVariable can access it
		if drop, ok := dropToStore.(interface{ SetContext(*Context) }); ok {
			// Set the context on the drop
			drop.SetContext(ctx)
			// Make the drop available as the primary lookup source
			// by putting it in the outer scope with a special key
			// The drop belongs to this render only: it is not kept for the next
			// one, as in Ruby, where it is one of the render's environments
			ctx.Scopes()[len(ctx.Scopes())-1]["__drop__"] = dropToStore
		}
	}

	// Apply options
	if options != nil {
		// Set registers
		if options.Registers != nil {
			for key, value := range options.Registers {
				ctx.Registers().Set(key, value)
			}
		}

		// Apply other options
		if options.Filters != nil {
			ctx.AddFilters(options.Filters)
		}
		if options.GlobalFilter != nil {
			ctx.SetGlobalFilter(options.GlobalFilter)
		}
		if options.ExceptionRenderer != nil {
			ctx.SetExceptionRenderer(options.ExceptionRenderer)
		}
		if options.StrictVariables {
			ctx.SetStrictVariables(true)
		}
		if options.StrictFilters {
			ctx.SetStrictFilters(true)
		}
		if options.Context != nil {
			ctx.SetCancellation(options.Context)
		}
	}

	return ctx
}

// configureOptions configures parse options and returns a ParseContext.
func (t *Template) configureOptions(options *TemplateOptions) ParseContextInterface {
	if options == nil {
		options = &TemplateOptions{}
	}

	if options.Environment != nil {
		t.environment = options.Environment
	} else if t.environment == nil {
		t.environment = NewEnvironment()
	}

	t.profiling = options.Profile
	t.lineNumbers = options.LineNumbers || t.profiling
	if options.PersistInstanceAssigns {
		t.persistInstanceAssigns = true
	}

	// Create parse context
	templateOpts := make(map[string]interface{})
	if options.StrictVariables {
		templateOpts["strict_variables"] = true
	}
	if options.StrictFilters {
		templateOpts["strict_filters"] = true
	}
	if t.lineNumbers {
		templateOpts["line_numbers"] = true
	}

	parseContextOpts := ParseContextOptions{
		Environment:     t.environment,
		TemplateOptions: templateOpts,
	}

	return NewParseContext(parseContextOpts)
}

// limitValue returns a resource limit given as any integer type, or as a
// whole float such as a number decoded from JSON, saturated to the int range,
// or nil when v is not an integer.
func limitValue(v interface{}) *int {
	var limit int
	switch n := v.(type) {
	case int:
		limit = n
	case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		var err error
		if limit, err = ToInteger(n); err != nil {
			return nil
		}
	case float64, float32:
		f := reflect.ValueOf(n).Float()
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
			return nil
		}
		switch {
		case f >= math.MaxInt:
			limit = math.MaxInt
		case f <= math.MinInt:
			limit = math.MinInt
		default:
			limit = int(f)
		}
	default:
		return nil
	}
	return &limit
}
