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

// envConfig returns the config of an environment with the standard tags, the
// given error mode ("" for the default) and resource limits.
func envConfig(mode liquid.ErrorMode, l limits) liquid.EnvironmentConfig {
	config := liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: mode}
	if l.renderLength > 0 {
		config.Limits.RenderLengthLimit = &l.renderLength
	}
	if l.renderScore > 0 {
		config.Limits.RenderScoreLimit = &l.renderScore
	}
	if l.assignScore > 0 {
		config.Limits.AssignScoreLimit = &l.assignScore
	}
	return config
}

// newEnv returns the environment of envConfig.
func newEnv(mode liquid.ErrorMode, l limits) *liquid.Environment {
	return mustEnv(envConfig(mode, l))
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

// stopped reports whether an error ended the render (a limit, a
// cancellation, a Go panic): the Ruby helper then writes its message as
// the output.
func (r result) stopped() bool {
	for _, err := range r.errors {
		switch err.(type) {
		case *liquid.MemoryError, *liquid.CanceledError, *liquid.PanicError:
			return true
		}
	}
	return false
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

// outOfRangeTag reads past the end of a list when it renders: a bug of a node.
type outOfRangeTag struct{ *liquid.Tag }

func (n *outOfRangeTag) RenderToOutput(liquid.TagContext, *liquid.OutputBuffer) {
	var items []int
	_ = items[len(items)]
}

// The fuzz and scaling checks see a render that a Go panic ended.
func TestInternalErrorSeesAGoPanic(t *testing.T) {
	config := envConfig("", limits{})
	config.Tags["outofrange"] = func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &outOfRangeTag{liquid.NewTag(name, markup, pc)}, nil
	}
	env := mustEnv(config)
	if r := renderWith(env, "a{% outofrange %}b", nil); r.internalError() == nil {
		t.Errorf("a render a Go panic ended gives no internal error (output %q, errors %v)", r.output, r.errors)
	}
	if r := renderWith(env, "a{{ x }}b", hostileData()); r.internalError() != nil {
		t.Errorf("a render without a panic gives the internal error %v", r.internalError())
	}
}
