package integration

import (
	gocontext "context"
	"errors"
	"sync"
	"testing"
	"weak"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// rubyOptions are the render options of Ruby's Template#render, which the
// ported tests give.
type rubyOptions struct {
	Context           gocontext.Context
	Output            *string
	Registers         map[string]interface{}
	GlobalFilter      func(interface{}) interface{}
	ExceptionRenderer func(error) interface{}
	Filters           []interface{}
	StrictVariables   bool
	StrictFilters     bool
	Limits            *liquid.ResourceLimitsConfig
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

// renderRuby renders tmpl as Ruby's Template#render does, which the tests
// were ported from: an error that ends the render (a resource limit, a
// cancellation) gives its "Liquid error: ..." message as the output. assigns
// is a map, nil, or a *liquid.Context the test built, whose state the render
// uses and keeps.
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
	if c, ok := assigns.(*liquid.Context); ok {
		return renderRubyContext(tmpl, c, o, raise, prefix)
	}
	var data map[string]interface{}
	switch v := assigns.(type) {
	case nil:
	case map[string]interface{}:
		data = v
	default:
		panic("renderRuby takes a map, nil or a *liquid.Context")
	}
	var opts []liquid.RenderOption
	if o.Registers != nil {
		opts = append(opts, liquid.WithRegisters(o.Registers))
	}
	if len(o.Filters) > 0 {
		opts = append(opts, liquid.WithFilters(o.Filters...))
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
	if o.GlobalFilter != nil || o.ExceptionRenderer != nil {
		// Render has no such options: the Ruby hooks are only on a Context
		c := liquid.BuildContext(liquid.ContextConfig{Environments: []map[string]interface{}{data}, Environment: tmpl.Environment()})
		return renderRubyContext(tmpl, c, o, raise, prefix)
	}
	ctx := o.Context
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

// renderRubyContext renders tmpl with a Context the test built, as a partial
// of that Context's render.
func renderRubyContext(tmpl *liquid.Template, c *liquid.Context, o *rubyOptions, raise bool, prefix string) rubyResult {
	for k, v := range o.Registers {
		c.Registers().Set(k, v)
	}
	if len(o.Filters) > 0 {
		c.AddFilters(o.Filters)
	}
	if o.GlobalFilter != nil {
		c.SetGlobalFilter(o.GlobalFilter)
	}
	if o.ExceptionRenderer != nil {
		c.SetExceptionRenderer(o.ExceptionRenderer)
	}
	if o.StrictVariables {
		c.SetStrictVariables(true)
	}
	if o.StrictFilters {
		c.SetStrictFilters(true)
	}
	if o.Context != nil {
		// the cancellation applies to this render only, as Ruby's
		c.SetCancellation(o.Context)
		defer c.SetCancellation(nil)
	}
	if raise {
		c.SetExceptionRenderer(func(err error) interface{} { panic(err) })
	}
	out := prefix
	tmpl.RenderToOutputBuffer(c, &out)
	rl := c.ResourceLimits()
	return rubyResult{Output: out, Errors: c.Errors(), Reached: rl.Reached(), RenderScore: rl.RenderScore(), AssignScore: rl.AssignScore()}
}

// rubyErrorText is the text Ruby writes for an error that ends a render.
func rubyErrorText(err error) string {
	var liquidErr liquid.LiquidError
	if errors.As(err, &liquidErr) {
		return liquidErr.Error()
	}
	return "Liquid error: internal"
}

// mustRenderRuby fails the test when the render records an error.
func mustRenderRuby(t *testing.T, tmpl *liquid.Template, assigns interface{}) string {
	t.Helper()
	res := renderRubyResult(tmpl, assigns, nil, false)
	if len(res.Errors) > 0 {
		t.Fatalf("render recorded %v", res.Errors)
	}
	return res.Output
}
