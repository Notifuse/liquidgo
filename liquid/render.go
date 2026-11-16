package liquid

import (
	gocontext "context"
	"fmt"
	"reflect"
	"runtime/debug"
)

// Result is what a render produced.
type Result struct {
	// Output is the rendered text. When the render ends with an error, it is
	// the text written before the render stopped.
	Output string
	// Errors are the errors the render recorded, as Ruby's Template#errors,
	// each once: each one wrote its message in Output ("Liquid error: ..."),
	// unless it was raised by RaiseErrors.
	Errors []error
	// RenderScore and AssignScore are the work and the assigned size the
	// render counted against its limits (see ResourceLimitsConfig).
	RenderScore int
	AssignScore int
}

// PanicError is a Go panic during a render: a bug in liquidgo or in a tag,
// filter or drop given to it. It ends the render, with the stack of the
// panic. A Liquid error a tag, filter or drop raises is not a
// PanicError: the render records it, as Ruby does.
type PanicError struct {
	Value interface{}
	Stack []byte
}

// Error implements error.
func (e *PanicError) Error() string {
	return fmt.Sprintf("liquid: panic during render: %v", e.Value)
}

// Unwrap returns the panic value when it is an error.
func (e *PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

// RenderOption configures one render.
type RenderOption func(*renderConfig)

type renderConfig struct {
	fileSystem      FileSystem
	limits          *ResourceLimitsConfig
	raise           bool
	registers       map[string]interface{}
	filters         []interface{}
	strictVariables bool
	strictFilters   bool
}

// WithRegisters gives the render registers, read by tags and filters next to
// the template's (Template.Registers), which they take precedence over. As
// in Ruby, they are static registers: a partial of render, which inherits
// only those, reads them too. WithFileSystem takes precedence over a
// "file_system" register.
func WithRegisters(registers map[string]interface{}) RenderOption {
	return func(c *renderConfig) { c.registers = registers }
}

// WithFilters gives the render local filters, next to the environment's,
// which they take precedence over for a name both define.
func WithFilters(filters ...interface{}) RenderOption {
	return func(c *renderConfig) { c.filters = append(c.filters, filters...) }
}

// WithStrictVariables makes an undefined variable a recorded error.
func WithStrictVariables() RenderOption {
	return func(c *renderConfig) { c.strictVariables = true }
}

// WithStrictFilters makes an undefined filter a recorded error.
func WithStrictFilters() RenderOption {
	return func(c *renderConfig) { c.strictFilters = true }
}

// WithFileSystem gives the render the file system its partials (render,
// include) are read from, in place of the environment's.
func WithFileSystem(fs FileSystem) RenderOption {
	return func(c *renderConfig) { c.fileSystem = fs }
}

// WithLimits gives the render its own resource limits: each field set replaces
// the template's limit of the same name, and a field left nil keeps it.
func WithLimits(limits ResourceLimitsConfig) RenderOption {
	return func(c *renderConfig) { c.limits = &limits }
}

// RaiseErrors makes the first error the render records end it, as Ruby's
// render!: Render returns that error.
func RaiseErrors() RenderOption {
	return func(c *renderConfig) { c.raise = true }
}

// Render renders the template for data in a render state of its own, and
// changes nothing in the template or in data: renders of one template,
// concurrent or not, never see each other's assigns, errors or limits. It
// reads the template's assigns (Template.Assigns) and registers. A drop is
// bound to the render that reads it (Drop.SetContext): give each render its
// own drop instances.
//
// The error is non-nil when the render could not complete, and Result.Output
// is then only what was written before it stopped:
//   - a *MemoryError: a resource limit was reached;
//   - a *CanceledError: ctx was done;
//   - a *PanicError: Go code panicked during the render, with its stack;
//   - with RaiseErrors, the first error the render recorded.
//
// Otherwise the error is nil, and Result.Errors holds the errors the render
// recorded while it went on, as Ruby does.
func (t *Template) Render(ctx gocontext.Context, data map[string]interface{}, opts ...RenderOption) (res Result, err error) {
	var cfg renderConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if t.root == nil {
		return res, nil
	}

	limitsConfig := t.limits
	if cfg.limits != nil {
		limitsConfig = mergeLimits(limitsConfig, *cfg.limits)
	}
	limits := NewResourceLimits(limitsConfig)
	// The registers of the render and the file system go in the static
	// registers, as Ruby's, which the isolated context of a render tag
	// inherits, so that partials find them too. The template's own are
	// copied, never written.
	static := t.registers
	if cfg.fileSystem != nil || len(cfg.registers) > 0 {
		static = make(map[string]interface{}, len(t.registers)+len(cfg.registers)+1)
		for k, v := range t.registers {
			static[k] = v
		}
		for k, v := range cfg.registers {
			static[k] = v
		}
		if cfg.fileSystem != nil {
			static["file_system"] = cfg.fileSystem
		}
	}
	registers := NewRegisters(static)
	// The render never writes data or the template's assigns: a func's value
	// is kept with the render (Context.funcValueOf), so concurrent renders may
	// share them.
	if data == nil {
		data = emptyData
	}
	c := BuildContext(ContextConfig{
		Environments:   []map[string]interface{}{data, t.assigns},
		OuterScope:     map[string]interface{}{},
		Registers:      registers,
		ResourceLimits: limits,
		Environment:    t.environment,
		RethrowErrors:  cfg.raise,
	})
	c.fatalPanics = true
	c.SetTemplateName(t.name)
	if len(cfg.filters) > 0 {
		c.AddFilters(cfg.filters)
	}
	if cfg.strictVariables {
		c.SetStrictVariables(true)
	}
	if cfg.strictFilters {
		c.SetStrictFilters(true)
	}
	if ctx != nil {
		c.SetCancellation(ctx)
	}

	buffer := NewOutputBufferFor(c)
	defer func() {
		res.Output = buffer.String()
		res.Errors = distinctErrors(c.Errors())
		res.RenderScore = c.ResourceLimits().RenderScore()
		res.AssignScore = c.ResourceLimits().AssignScore()
		if r := recover(); r != nil {
			err = renderContextError(r)
		}
	}()
	t.root.RenderToOutput(c, buffer)
	return res, nil
}

// mergeLimits returns base with the limits that set gives replacing its own.
func mergeLimits(base, set ResourceLimitsConfig) ResourceLimitsConfig {
	merged := base
	if set.RenderLengthLimit != nil {
		merged.RenderLengthLimit = set.RenderLengthLimit
	}
	if set.RenderScoreLimit != nil {
		merged.RenderScoreLimit = set.RenderScoreLimit
	}
	if set.AssignScoreLimit != nil {
		merged.AssignScoreLimit = set.AssignScoreLimit
	}
	return merged
}

// emptyData is the data of a render given none; a render never writes it.
var emptyData = map[string]interface{}{}

// distinctErrors returns errs with each error once: an error raised through
// enclosing blocks (RaiseErrors) is recorded by each of them.
func distinctErrors(errs []error) []error {
	if len(errs) < 2 {
		return errs
	}
	var out []error
	seen := make(map[error]bool, len(errs))
	for _, err := range errs {
		if comparable(err) {
			if seen[err] {
				continue
			}
			seen[err] = true
		}
		out = append(out, err)
	}
	return out
}

// comparable reports whether err can be a map key: an error type with an
// uncomparable value (a slice, a map) would make the lookup panic.
func comparable(err error) bool {
	return reflect.TypeOf(err).Comparable()
}

// renderContextError returns the error a panic out of a render
// ends it with.
func renderContextError(r interface{}) error {
	switch e := r.(type) {
	case *PanicError:
		return e
	case LiquidError:
		return e
	case *Error:
		return e
	}
	return &PanicError{Value: r, Stack: debug.Stack()}
}

// fatalPanic returns the PanicError a panic in a node raises when the render
// ends on Go panics (Render), or nil when the panic is a Liquid error,
// which the node records as Ruby does.
func fatalPanic(context TagContext, r interface{}) *PanicError {
	c, ok := context.(*Context)
	if !ok || c == nil || !c.fatalPanics {
		return nil
	}
	if p, ok := r.(*PanicError); ok {
		return p
	}
	switch r.(type) {
	case LiquidError, *Error:
		return nil
	}
	return &PanicError{Value: r, Stack: debug.Stack()}
}
