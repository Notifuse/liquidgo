package integration

import (
	gocontext "context"
	"errors"
	"sync"
	"weak"

	"github.com/Notifuse/liquidgo/v5/internal/methodfilters"
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// rubyOptions are the render options of Ruby's Template#render, which the
// ported tests give.
type rubyOptions struct {
	Context         gocontext.Context
	Output          *string
	Registers       map[string]interface{}
	Filters         []interface{}
	StrictVariables bool
	StrictFilters   bool
	Limits          *liquid.ResourceLimitsConfig
}

// rubyResult is what a Ruby-style render gives a test besides its output.
type rubyResult struct {
	Output      string
	Errors      []error
	Reached     bool // a resource limit ended the render
	RenderScore int
	AssignScore int
}

// rubyState is the state Ruby's Template keeps between renders, which the
// tests ported from Ruby set and read: the resource limits of the next
// renders, and the result of the last one. liquidgo's Template keeps no such
// state; the tests keep it here, per template.
type rubyState struct {
	mu     sync.Mutex
	limits *liquid.ResourceLimitsConfig
	last   rubyResult
}

// rubyStates holds the state of each template a test rendered, keyed by a
// weak pointer so that it keeps no template alive (tests check what the
// library retains), and at most maxRubyStates of them.
var (
	rubyStatesMu sync.Mutex
	rubyStates   = map[weak.Pointer[liquid.Template]]*rubyState{}
	rubyOrder    []weak.Pointer[liquid.Template]
)

const maxRubyStates = 256

func stateOf(tmpl *liquid.Template) *rubyState {
	key := weak.Make(tmpl)
	rubyStatesMu.Lock()
	defer rubyStatesMu.Unlock()
	if st, ok := rubyStates[key]; ok {
		return st
	}
	if len(rubyOrder) >= maxRubyStates {
		delete(rubyStates, rubyOrder[0])
		rubyOrder = rubyOrder[1:]
	}
	st := &rubyState{}
	rubyStates[key] = st
	rubyOrder = append(rubyOrder, key)
	return st
}

// setRubyLimits sets the resource limits of the next renders of tmpl, as
// Ruby's template.resource_limits.
func setRubyLimits(tmpl *liquid.Template, limits liquid.ResourceLimitsConfig) {
	st := stateOf(tmpl)
	st.mu.Lock()
	st.limits = &limits
	st.mu.Unlock()
}

// rubyLimitsResult is what Ruby's template.resource_limits tells of the
// last render.
type rubyLimitsResult struct{ last rubyResult }

func (r rubyLimitsResult) Reached() bool    { return r.last.Reached }
func (r rubyLimitsResult) RenderScore() int { return r.last.RenderScore }
func (r rubyLimitsResult) AssignScore() int { return r.last.AssignScore }

// rubyLimits returns the limits of the last render of tmpl.
func rubyLimits(tmpl *liquid.Template) rubyLimitsResult {
	st := stateOf(tmpl)
	st.mu.Lock()
	defer st.mu.Unlock()
	return rubyLimitsResult{st.last}
}

// rubyErrors returns the errors of the last render of tmpl, as Ruby's
// template.errors.
func rubyErrors(tmpl *liquid.Template) []error {
	st := stateOf(tmpl)
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.last.Errors == nil {
		return []error{}
	}
	return st.last.Errors
}

// rubyContext is the Liquid::Context a ported test builds and renders with,
// with the fields of liquid.ContextConfig. liquidgo renders from data and
// options only (Template.Render), so the helpers render with what the
// context holds: its environments and static environments as the data,
// merged so that a name resolves as in Ruby's lookup order; its
// static registers; its limits; and RethrowErrors as RaiseErrors. The
// template's environment renders: Environment is the test's record of it.
// Errors and ResourceLimits tell of the last render, as Ruby's
// context.errors and context.resource_limits.
type rubyContext struct {
	Environment        *liquid.Environment
	Environments       []map[string]interface{}
	StaticEnvironments []map[string]interface{}
	Registers          *liquid.Registers
	Limits             *liquid.ResourceLimitsConfig
	RethrowErrors      bool

	cancel gocontext.Context
	last   rubyResult
}

// newRubyContext returns a context for env, as liquid.BuildContext with
// only an environment.
func newRubyContext(env *liquid.Environment) *rubyContext {
	return &rubyContext{Environment: env}
}

// SetCancellation sets the context.Context the renders with c stop with.
func (c *rubyContext) SetCancellation(ctx gocontext.Context) { c.cancel = ctx }

// Errors returns the errors of the last render with c.
func (c *rubyContext) Errors() []error {
	if c.last.Errors == nil {
		return []error{}
	}
	return c.last.Errors
}

// ResourceLimits returns the limits of the last render with c.
func (c *rubyContext) ResourceLimits() rubyLimitsResult { return rubyLimitsResult{c.last} }

// data returns the names c resolves, each to the value Ruby's lookup finds
// first: the environments, then the static ones.
func (c *rubyContext) data() map[string]interface{} {
	data := map[string]interface{}{}
	for i := len(c.StaticEnvironments) - 1; i >= 0; i-- {
		for k, v := range c.StaticEnvironments[i] {
			data[k] = v
		}
	}
	for i := len(c.Environments) - 1; i >= 0; i-- {
		for k, v := range c.Environments[i] {
			data[k] = v
		}
	}
	return data
}

// renderRuby renders tmpl as Ruby's Template#render does, which the tests
// were ported from: an error that ends the render (a resource limit, a
// cancellation) gives its "Liquid error: ..." message as the output. assigns
// is a map, nil, or a *rubyContext the test built.
func renderRuby(tmpl *liquid.Template, assigns interface{}, o *rubyOptions) string {
	return renderRubyResult(tmpl, assigns, o, false).Output
}

// renderRubyBang renders tmpl as Ruby's Template#render! does: the first
// error the render records raises (panics) out of it.
func renderRubyBang(tmpl *liquid.Template, assigns interface{}, o *rubyOptions) string {
	return renderRubyResult(tmpl, assigns, o, true).Output
}

// renderRubyResult is renderRuby, or renderRubyBang with raise, with the
// errors the render recorded and whether a limit ended it.
func renderRubyResult(tmpl *liquid.Template, assigns interface{}, o *rubyOptions, raise bool) (out rubyResult) {
	if o == nil {
		o = &rubyOptions{}
	}
	st := stateOf(tmpl)
	st.mu.Lock()
	if o.Limits == nil && st.limits != nil {
		copied := *o
		copied.Limits = st.limits
		o = &copied
	}
	st.mu.Unlock()
	defer func() {
		st.mu.Lock()
		st.last = out
		st.mu.Unlock()
		// Ruby's output: option is the buffer the render writes to
		if o.Output != nil {
			*o.Output = out.Output
		}
	}()
	prefix := ""
	if o.Output != nil {
		prefix = *o.Output
	}
	var data map[string]interface{}
	var opts []liquid.RenderOption
	registers := o.Registers
	cancel := o.Context
	switch v := assigns.(type) {
	case nil:
	case map[string]interface{}:
		data = v
	case *rubyContext:
		defer func() { v.last = out }()
		data = v.data()
		if v.Registers != nil {
			// the options' registers add to the context's
			registers = map[string]interface{}{}
			for k, r := range v.Registers.Static() {
				registers[k] = r
			}
			for k, r := range o.Registers {
				registers[k] = r
			}
		}
		if v.Limits != nil && o.Limits == nil {
			opts = append(opts, liquid.WithLimits(*v.Limits))
		}
		raise = raise || v.RethrowErrors
		if cancel == nil {
			cancel = v.cancel
		}
	default:
		panic("renderRuby takes a map, nil or a *rubyContext")
	}
	if registers != nil {
		opts = append(opts, liquid.WithRegisters(registers))
	}
	if len(o.Filters) > 0 {
		opts = append(opts, liquid.WithFilters(methodfilters.Of(o.Filters...)...))
	}
	if o.StrictVariables {
		opts = append(opts, liquid.WithStrictVariables())
	}
	if o.StrictFilters {
		opts = append(opts, liquid.WithStrictFilters())
	}
	if o.Limits != nil {
		opts = append(opts, liquid.WithLimits(*o.Limits))
	}
	if raise {
		opts = append(opts, liquid.RaiseErrors())
	}
	ctx := cancel
	if ctx == nil {
		ctx = gocontext.Background()
	}
	res, err := tmpl.Render(ctx, data, opts...)
	out = rubyResult{Output: prefix + res.Output, Errors: res.Errors, RenderScore: res.RenderScore, AssignScore: res.AssignScore}
	if err != nil {
		var limit *liquid.MemoryError
		out.Reached = errors.As(err, &limit)
		var liquidErr liquid.LiquidError
		if raise && errors.As(err, &liquidErr) && !out.Reached {
			panic(err)
		}
		// Ruby writes an error that ends the render in place of the output,
		// and records it with the template's errors
		out.Output = prefix + rubyErrorText(err)
		out.Errors = append(append([]error{}, out.Errors...), err)
	}
	return out
}

// rubyErrorText is the text Ruby writes for an error that ends a render.
func rubyErrorText(err error) string {
	var liquidErr liquid.LiquidError
	if errors.As(err, &liquidErr) {
		return liquidErr.Error()
	}
	return "Liquid error: internal"
}
