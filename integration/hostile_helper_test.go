package integration

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// limits holds the resource limits a test sets on the environment; a zero
// field leaves that limit unset.
type limits struct {
	renderLength int
	renderScore  int
	assignScore  int
}

// emailLimits is the configuration of an application that bounds only the
// length of the output, at 10 MB.
var emailLimits = limits{renderLength: 10_000_000}

// newEnv returns an environment with the standard tags, the given error mode
// ("" for the default) and resource limits.
func newEnv(mode string, l limits) *liquid.Environment {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	if mode != "" {
		env.SetErrorMode(mode)
	}
	set := map[string]interface{}{}
	if l.renderLength > 0 {
		set["render_length_limit"] = l.renderLength
	}
	if l.renderScore > 0 {
		set["render_score_limit"] = l.renderScore
	}
	if l.assignScore > 0 {
		set["assign_score_limit"] = l.assignScore
	}
	if len(set) > 0 {
		env.SetDefaultResourceLimits(set)
	}
	return env
}

// result is what one parse and render of a source gave.
type result struct {
	output   string
	parseErr error
	errors   []error // Template.Errors() after the render
}

// hasError reports whether the parse or the render failed with an error
// whose message contains text.
func (r result) hasError(text string) bool {
	if r.parseErr != nil && strings.Contains(r.parseErr.Error(), text) {
		return true
	}
	for _, err := range r.errors {
		if strings.Contains(err.Error(), text) {
			return true
		}
	}
	return false
}

// internalError returns the first error of the render that a bug of the
// library gives, or nil: the PanicError that ends a render a Go panic
// stopped, or an InternalError, a non-Liquid error of a node written as
// "Liquid error: internal".
func (r result) internalError() error {
	for _, err := range r.errors {
		switch err.(type) {
		case *liquid.PanicError, *liquid.InternalError:
			return err
		}
	}
	return nil
}

// renderWith parses source in env and renders it with data, as an
// application does: a parse error is returned, not rendered.
func renderWith(env *liquid.Environment, source string, data map[string]interface{}) result {
	tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		return result{parseErr: err}
	}
	out := renderRuby(tmpl, data, nil)
	return result{output: out, errors: append([]error(nil), rubyErrors(tmpl)...)}
}

// hostileData is the data every hostile template is rendered with.
func hostileData() map[string]interface{} {
	return map[string]interface{}{"a": []interface{}{1}, "x": "x", "long": rep("x", 100)}
}

// allocated returns the bytes fn allocated, counted by the runtime.
func allocated(fn func()) uint64 {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	fn()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// mb writes a byte count in MB for a test message.
func mb(n uint64) string { return fmt.Sprintf("%.1f MB", float64(n)/(1<<20)) }

// rep returns s repeated n times.
func rep(s string, n int) string { return strings.Repeat(s, n) }

// requireNoPanicEscapes runs fn and fails the test if a panic leaves it: the
// public API must return every failure as an error.
func requireNoPanicEscapes(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic left the public API: %v", r)
		}
	}()
	fn()
}

// newContext returns a rendering context of env with no data.
func newContext(env *liquid.Environment) *liquid.Context {
	return liquid.BuildContext(liquid.ContextConfig{Environment: env})
}

// nilMapTag writes to a nil map when it renders: a bug of a node.
type nilMapTag struct{ *liquid.Tag }

func (n *nilMapTag) RenderToOutput(liquid.TagContext, *liquid.OutputBuffer) {
	var m map[string]int
	m["x"] = 1
}

// The fuzz and scaling checks see a render that a Go panic ended.
func TestInternalErrorSeesAGoPanic(t *testing.T) {
	env := newEnv("", limits{})
	env.RegisterTag("nilmap", func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &nilMapTag{liquid.NewTag(name, markup, pc)}, nil
	})
	if r := renderWith(env, "a{% nilmap %}b", nil); r.internalError() == nil {
		t.Errorf("a render a Go panic ended gives no internal error (output %q, errors %v)", r.output, r.errors)
	}
	if r := renderWith(env, "a{{ x }}b", hostileData()); r.internalError() != nil {
		t.Errorf("a render without a panic gives the internal error %v", r.internalError())
	}
}
