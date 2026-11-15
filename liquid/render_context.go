package liquid

import (
	gocontext "context"
	"fmt"
	"runtime/debug"
)

// Result is what a render produced.
type Result struct {
	// Output is the rendered text. When the render ends with an error, it is
	// the text written before the render stopped.
	Output string
	// Errors are the errors the render recorded, as Ruby's Template#errors:
	// each one either wrote its message in Output ("Liquid error: ...") or
	// wrote nothing (an undefined variable or filter in strict mode).
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

// WithLimits gives the render its own resource limits, in place of the
// template's.
func WithLimits(limits ResourceLimitsConfig) RenderOption {
	return func(c *renderConfig) { c.limits = &limits }
}

// RaiseErrors makes the first error the render records end it, as Ruby's
// render!: RenderContext returns that error.
func RaiseErrors() RenderOption {
	return func(c *renderConfig) { c.raise = true }
}

// RenderContext renders the template for data in a render state of its own,
// and changes nothing in the template: renders of one template, concurrent
// or not, never see each other's assigns, errors, limits or profiler.
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
		limits = NewResourceLimits(*cfg.limits)
	}
	registers := NewRegisters(t.registers)
	if cfg.fileSystem != nil {
		registers.Set("file_system", cfg.fileSystem)
	}
	if data == nil {
		data = map[string]interface{}{}
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
	if ctx != nil {
		c.SetCancellation(ctx)
	}
	if t.profiling {
		c.SetProfiler(NewProfiler())
	}

	buffer := NewOutputBufferFor(c)
	defer func() {
		res.Output = buffer.String()
		res.Errors = c.Errors()
		if r := recover(); r != nil {
			err = renderContextError(r)
		}
	}()
	t.root.RenderToOutput(c, buffer)
	return res, nil
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
	if p, ok := r.(*PanicError); ok {
		return p
	}
	c, ok := context.(*Context)
	if !ok {
		if w, isWrapper := context.(interface{ Context() interface{} }); isWrapper {
			c, ok = w.Context().(*Context)
		}
	}
	if !ok || c == nil || !c.fatalPanics {
		return nil
	}
	switch r.(type) {
	case LiquidError, *Error:
		return nil
	}
	return &PanicError{Value: r, Stack: debug.Stack()}
}
