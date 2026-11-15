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
}

// PanicError is a Go panic during a render: a bug in liquidgo or in a tag,
// filter or drop given to it. It ends a RenderContext render, with the stack
// of the panic. A Liquid error a tag, filter or drop raises is not a
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

// RenderOption configures one RenderContext render.
type RenderOption func(*renderConfig)

type renderConfig struct {
	fileSystem FileSystem
	limits     *ResourceLimitsConfig
	raise      bool
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
// render!: RenderContext returns that error.
func RaiseErrors() RenderOption {
	return func(c *renderConfig) { c.raise = true }
}

// RenderContext renders the template for data in a render state of its own,
// and changes nothing in the template or in data: renders of one template,
// concurrent or not, never see each other's assigns, errors or limits. It
// reads the template's assigns (Template.Assigns) and registers, but not the
// instance assigns Render keeps with PersistInstanceAssigns. A drop is bound
// to the render that reads it (Drop.SetContext): give each render its own
// drop instances.
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
func (t *Template) RenderContext(ctx gocontext.Context, data map[string]interface{}, opts ...RenderOption) (res *Result, err error) {
	var cfg renderConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	res = &Result{}
	if t.root == nil {
		return res, nil
	}

	limits := cloneResourceLimits(t.ResourceLimits())
	if cfg.limits != nil {
		limits = NewResourceLimits(mergeLimits(limits, *cfg.limits))
	}
	// The file system goes in the static registers, which the isolated
	// context of a render tag inherits, so that nested partials find it too.
	static := t.registers
	if cfg.fileSystem != nil {
		static = make(map[string]interface{}, len(t.registers)+1)
		for k, v := range t.registers {
			static[k] = v
		}
		static["file_system"] = cfg.fileSystem
	}
	registers := NewRegisters(static)
	// A lookup writes the value of a func back into the map it read it from:
	// each render reads copies, so that neither data nor the template's
	// assigns change, and concurrent renders never write one map.
	c := BuildContext(ContextConfig{
		Environments:   []map[string]interface{}{shallowCopy(data), shallowCopy(t.assigns)},
		OuterScope:     map[string]interface{}{},
		Registers:      registers,
		ResourceLimits: limits,
		Environment:    t.environment,
		RethrowErrors:  cfg.raise,
	})
	c.fatalPanics = true
	c.SetTemplateName(t.name)
	if ctx != nil {
		c.SetCancellation(ctx)
	}

	buffer := NewOutputBufferFor(c)
	defer func() {
		res.Output = buffer.String()
		res.Errors = distinctErrors(c.Errors())
		if r := recover(); r != nil {
			err = renderContextError(r)
		}
	}()
	t.root.RenderToOutput(c, buffer)
	return res, nil
}

// mergeLimits returns base with the limits that set gives replacing its own.
func mergeLimits(base *ResourceLimits, set ResourceLimitsConfig) ResourceLimitsConfig {
	merged := ResourceLimitsConfig{
		RenderLengthLimit: base.renderLengthLimit,
		RenderScoreLimit:  base.renderScoreLimit,
		AssignScoreLimit:  base.assignScoreLimit,
	}
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

// shallowCopy returns a copy of m that a render may write to.
func shallowCopy(m map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{}, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

// distinctErrors returns errs with each error once: an error raised through
// enclosing blocks (RaiseErrors) is recorded by each of them.
func distinctErrors(errs []error) []error {
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

// renderContextError returns the error a panic out of a RenderContext render
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
// ends on Go panics (RenderContext), or nil when the panic is a Liquid error,
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
