package integration

import (
	gocontext "context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/Notifuse/liquidgo/v5/internal/methodfilters"
	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func rcEnv() *liquid.Environment {
	return mustEnv(rcConfig())
}

// rcConfig is the config of rcEnv.
func rcConfig() liquid.EnvironmentConfig {
	return liquid.EnvironmentConfig{Tags: tags.Standard()}
}

func rcParse(t *testing.T, env *liquid.Environment, src string) *liquid.Template {
	t.Helper()
	tpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return tpl
}

// Concurrent renders of one template each get their own output, errors and
// limits, and the template keeps none of them.
func TestRenderIsolatesConcurrentRenders(t *testing.T) {
	config := rcConfig()
	config.Limits.RenderLengthLimit = intPtr(200)
	env := mustEnv(config)
	tpl := rcParse(t, env, `{% if admin %}{% assign secret = token %}{% endif %}[{{ secret }}]{% if fail %}{{ errors.standard_error }}{% endif %}{% for i in (1..n) %}x{% endfor %}`)
	var wg sync.WaitGroup
	failures := make(chan string, 2000)
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			admin, long, fail := i%2 == 0, i%3 == 0, i%5 == 0
			n := 1
			if long {
				n = 500
			}
			data := map[string]interface{}{"admin": admin, "token": fmt.Sprint("T", i), "fail": fail, "errors": NewErrorDrop(), "n": n}
			res, err := tpl.Render(gocontext.Background(), data)
			var limit *liquid.MemoryError
			wantErrors := 0
			if fail {
				wantErrors = 1
			}
			switch {
			case !long && len(res.Errors) != wantErrors:
				failures <- fmt.Sprintf("render %d: %d recorded errors, want %d", i, len(res.Errors), wantErrors)
			case long && !errors.As(err, &limit):
				failures <- fmt.Sprintf("render %d: over the limit, err = %v", i, err)
			case !long && err != nil:
				failures <- fmt.Sprintf("render %d: err = %v", i, err)
			case !long && admin && !strings.HasPrefix(res.Output, fmt.Sprint("[T", i, "]")):
				failures <- fmt.Sprintf("render %d: output %q", i, res.Output)
			case !long && !admin && !strings.HasPrefix(res.Output, "[]"):
				failures <- fmt.Sprintf("render %d: output %q, another render's secret", i, res.Output)
			}
		}(i)
	}
	wg.Wait()
	close(failures)
	for f := range failures {
		t.Error(f)
	}
}

// Render changes nothing in the template: no assigns, errors or registers,
// and the template's assigns and registers are only read.
func TestRenderLeavesTheTemplateAlone(t *testing.T) {
	env := rcEnv()
	tpl, err := liquid.ParseTemplate(`{% assign kept = "x" %}[{{ kept }}{{ greeting }}{{ errors.standard_error }}{% render "p" %}]`,
		&liquid.TemplateOptions{Environment: env, LineNumbers: true})
	if err != nil {
		t.Fatal(err)
	}
	tpl.Assigns()["greeting"] = "hi"
	tpl.Registers()["marker"] = "template register"
	fs := &mapFileSystem{templates: map[string]string{"p": "p"}}
	res, err := tpl.Render(gocontext.Background(), map[string]interface{}{"errors": NewErrorDrop()}, liquid.WithFileSystem(fs))
	if err != nil {
		t.Fatal(err)
	}
	if want := "[xhiLiquid error (line 1): standard errorp]"; res.Output != want {
		t.Errorf("output %q, want %q", res.Output, want)
	}
	if len(res.Errors) != 1 {
		t.Errorf("Result.Errors = %v, want the standard error", res.Errors)
	}
	if _, set := tpl.Registers()["file_system"]; set || len(tpl.Registers()) != 1 {
		t.Errorf("the template's registers changed: %v", tpl.Registers())
	}
	if len(tpl.Assigns()) != 1 {
		t.Errorf("the template's assigns changed: %v", tpl.Assigns())
	}
}

// A template register stays readable next to WithFileSystem, by a tag that
// reads the registers.
func TestRenderKeepsTemplateRegisters(t *testing.T) {
	config := rcConfig()

	config.Tags["marker"] = tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &markerTag{liquid.NewTag(name, markup, pc)}, nil
	})
	env := mustEnv(config)
	tpl := rcParse(t, env, `[{% marker %}{% render "p" %}]`)
	tpl.Registers()["marker"] = "M"
	fs := &mapFileSystem{templates: map[string]string{"p": "{% marker %}"}}
	res, err := tpl.Render(gocontext.Background(), nil, liquid.WithFileSystem(fs))
	if want := "[MM]"; err != nil || res.Output != want {
		t.Errorf("output %q, err %v, want %q", res.Output, err, want)
	}
}

type markerTag struct{ *liquid.Tag }

func (m *markerTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	if v, ok := context.Registers().Get("marker").(string); ok {
		output.WriteString(v)
	}
}

// Concurrent renders with different partials each read their own, and leave
// the template's registers alone (detected under -race).
func TestRenderConcurrentFileSystems(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `[{% render "p" %}]`)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprint("p", i)
			res, err := tpl.Render(gocontext.Background(), nil, liquid.WithFileSystem(&mapFileSystem{templates: map[string]string{"p": body}}))
			if err != nil || res.Output != "["+body+"]" {
				t.Errorf("render %d: output %q, err %v", i, res.Output, err)
			}
		}(i)
	}
	wg.Wait()
}

// A func in data or in the template's assigns is called by each render and
// stays a func: neither map is written, so concurrent renders do not crash.
func TestRenderLeavesFuncsAlone(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `{{ stamp }}-{{ local }}`)
	var n int64
	var mu sync.Mutex
	tpl.Assigns()["stamp"] = func() interface{} {
		mu.Lock()
		defer mu.Unlock()
		n++
		return n
	}
	data := map[string]interface{}{"local": func() interface{} { return "l" }}
	first, _ := tpl.Render(gocontext.Background(), data)
	second, _ := tpl.Render(gocontext.Background(), data)
	if first.Output != "1-l" || second.Output != "2-l" {
		t.Errorf("renders %q and %q, want each to call the func", first.Output, second.Output)
	}
	// A func read twice in one render, also from a partial, is called once,
	// as Ruby calls a proc once per render
	twice := rcParse(t, rcEnv(), `{{ stamp }}{{ stamp }}{% render "p", s: stamp %}`)
	twice.Assigns()["stamp"] = tpl.Assigns()["stamp"]
	res, _ := twice.Render(gocontext.Background(), nil, liquid.WithFileSystem(&mapFileSystem{templates: map[string]string{"p": "{{ s }}"}}))
	if res.Output != "333" {
		t.Errorf("a func read three times in one render gives %q, want %q", res.Output, "333")
	}
	if _, isFunc := tpl.Assigns()["stamp"].(func() interface{}); !isFunc {
		t.Errorf("the template's assign became %v", tpl.Assigns()["stamp"])
	}
	if _, isFunc := data["local"].(func() interface{}); !isFunc {
		t.Errorf("the data became %v", data["local"])
	}
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = tpl.Render(gocontext.Background(), data)
		}()
	}
	wg.Wait()
}

// An error message carries the template's name and line.
func TestRenderErrorsNameTheTemplate(t *testing.T) {
	tpl, err := liquid.ParseTemplate("a\n{{ errors.standard_error }}", &liquid.TemplateOptions{Environment: rcEnv(), LineNumbers: true})
	if err != nil {
		t.Fatal(err)
	}
	tpl.SetName("welcome")
	res, err := tpl.Render(gocontext.Background(), map[string]interface{}{"errors": NewErrorDrop()})
	if err != nil {
		t.Fatal(err)
	}
	if want := "a\nLiquid error (welcome line 2): standard error"; res.Output != want {
		t.Errorf("output %q, want %q", res.Output, want)
	}
}

// A recorded error leaves a complete output and comes back in Result.Errors;
// with RaiseErrors it ends the render.
func TestRenderRecordedErrors(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `a{{ errors.standard_error }}b`)
	data := map[string]interface{}{"errors": NewErrorDrop()}
	res, err := tpl.Render(gocontext.Background(), data)
	if err != nil {
		t.Fatalf("err = %v, want a recorded error", err)
	}
	if res.Output != "aLiquid error: standard errorb" || len(res.Errors) != 1 {
		t.Errorf("output %q, errors %v", res.Output, res.Errors)
	}
	res, err = tpl.Render(gocontext.Background(), data, liquid.RaiseErrors())
	var standard *liquid.StandardError
	if !errors.As(err, &standard) {
		t.Errorf("with RaiseErrors err = %v, want the StandardError", err)
	}
	if res.Output != "a" {
		t.Errorf("with RaiseErrors output %q, want what was written before the error", res.Output)
	}
	nested := rcParse(t, rcEnv(), `{% if true %}{% for i in (1..1) %}{% if true %}{{ errors.standard_error }}{% endif %}{% endfor %}{% endif %}`)
	res, err = nested.Render(gocontext.Background(), data, liquid.RaiseErrors())
	if err == nil || len(res.Errors) != 1 {
		t.Errorf("RaiseErrors through nested blocks: err %v, errors %v, want the error once", err, res.Errors)
	}
	// The errors recorded before a fatal error come back with it
	before := rcParse(t, rcEnv(), `{{ errors.standard_error }}{% for i in (1..100000) %}xxxxxxxxxx{% endfor %}`)
	res, err = before.Render(gocontext.Background(), data, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: intPtr(1000)}))
	var limit *liquid.MemoryError
	if !errors.As(err, &limit) || len(res.Errors) != 1 {
		t.Errorf("a limit after a recorded error: err %v, errors %v", err, res.Errors)
	}
}

// WithLimits replaces the limits it sets and keeps the template's others.
func TestRenderWithLimitsKeepsTheOthers(t *testing.T) {
	config := rcConfig()
	config.Limits.RenderScoreLimit = intPtr(50)
	env := mustEnv(config)
	tpl := rcParse(t, env, `{% for i in (1..1000) %}{% endfor %}`)
	_, err := tpl.Render(gocontext.Background(), nil, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: intPtr(1000000)}))
	var limit *liquid.MemoryError
	if !errors.As(err, &limit) {
		t.Errorf("err %v, want the template's score limit", err)
	}
}

// A resource limit and a cancellation end the render with their error.
func TestRenderLimitsAndCancellation(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `{% for i in (1..100000) %}xxxxxxxxxx{% endfor %}`)
	_, err := tpl.Render(gocontext.Background(), nil, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: intPtr(1000)}))
	var limit *liquid.MemoryError
	if !errors.As(err, &limit) {
		t.Errorf("over the limit err = %v, want a MemoryError", err)
	}
	ctx, cancel := gocontext.WithCancel(gocontext.Background())
	cancel()
	_, err = tpl.Render(ctx, nil)
	var canceled *liquid.CanceledError
	if !errors.As(err, &canceled) {
		t.Errorf("cancelled err = %v, want a CanceledError", err)
	}
}

type rcPanicFilters struct{}

func (f *rcPanicFilters) Explode(input interface{}) interface{} {
	items := []string{}
	return items[len(fmt.Sprint(input))] // a Go panic: index out of range
}

type rcPanicDrop struct{}

// LiquidGet returns the key a template reads.
func (d *rcPanicDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "name":
		return d.Name(), true, nil
	}
	return nil, false, nil
}

func (d *rcPanicDrop) Name() interface{} {
	var s []int
	return s[3] // a Go panic: index out of range
}

// A Go panic in a filter, a drop or a partial ends the render with its
// stack, and the output written before.
func TestRenderGoPanicIsFatal(t *testing.T) {
	config := rcConfig()
	config.Filters = methodfilters.Of(&rcPanicFilters{})
	config.FileSystem = &mapFileSystem{templates: map[string]string{"p": "{{ d.name }}"}}
	env := mustEnv(config)
	data := map[string]interface{}{"d": &rcPanicDrop{}}
	for _, c := range []struct{ src, frame string }{
		{`a{{ "x" | explode }}b`, "Explode"},
		{`a{{ d.name }}b`, "Name"},
		{`a{% render "p", d: d %}b`, "Name"},
	} {
		tpl := rcParse(t, env, c.src)
		res, err := tpl.Render(gocontext.Background(), data)
		var p *liquid.PanicError
		if !errors.As(err, &p) {
			t.Errorf("%s: err = %v, want a PanicError", c.src, err)
			continue
		}
		if !strings.Contains(string(p.Stack), c.frame) {
			t.Errorf("%s: the stack does not show %s:\n%s", c.src, c.frame, p.Stack)
		}
		if res.Output != "a" {
			t.Errorf("%s: output %q, want what was written before the panic", c.src, res.Output)
		}
	}
}

// WithFileSystem gives the partials of one render, nested ones included.
func TestRenderWithFileSystem(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `[{% render "p" %}]`)
	for _, body := range []string{"one", "two"} {
		res, err := tpl.Render(gocontext.Background(), nil, liquid.WithFileSystem(&mapFileSystem{templates: map[string]string{"p": body}}))
		if err != nil || res.Output != "["+body+"]" {
			t.Errorf("with partial %q: output %q, err %v", body, res.Output, err)
		}
	}
	nested := rcParse(t, rcEnv(), `[{% render "outer", name: "x" %}|{% include "outer" %}]`)
	fs := &mapFileSystem{templates: map[string]string{
		"outer": `<{% render "inner", name: name %}>`,
		"inner": `{% render "leaf", name: name %}`,
		"leaf":  `leaf {{ name }}`,
	}}
	res, err := nested.Render(gocontext.Background(), map[string]interface{}{"name": "y"}, liquid.WithFileSystem(fs))
	if want := "[<leaf x>|<leaf y>]"; err != nil || res.Output != want {
		t.Errorf("nested partials: output %q, err %v, want %q", res.Output, err, want)
	}
}

func intPtr(n int) *int { return &n }

type rcShoutFilters struct{}

func (f *rcShoutFilters) Shout(input interface{}) string { return fmt.Sprint(input) + "!" }

// The render options reach the render: registers read by a tag, local
// filters, and strict variables and filters, which record their error.
func TestRenderOptions(t *testing.T) {
	config := rcConfig()

	config.Tags["marker"] = tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &markerTag{liquid.NewTag(name, markup, pc)}, nil
	})
	env := mustEnv(config)
	tpl := rcParse(t, env, `[{% marker %}]`)
	res, err := tpl.Render(gocontext.Background(), nil, liquid.WithRegisters(map[string]interface{}{"marker": "R"}))
	if err != nil || res.Output != "[R]" {
		t.Errorf("WithRegisters: output %q, err %v", res.Output, err)
	}
	tpl.Registers()["marker"] = "T"
	res, _ = tpl.Render(gocontext.Background(), nil, liquid.WithRegisters(map[string]interface{}{"marker": "R"}))
	if res.Output != "[R]" {
		t.Errorf("WithRegisters over a template register: output %q, want %q", res.Output, "[R]")
	}

	shout := rcParse(t, env, `{{ "a" | shout }}`)
	if res, err := shout.Render(gocontext.Background(), nil, liquid.WithFilters(methodfilters.Of(&rcShoutFilters{})...)); err != nil || res.Output != "a!" {
		t.Errorf("WithFilters: output %q, err %v", res.Output, err)
	}
	if res, _ := shout.Render(gocontext.Background(), nil); res.Output != "a" {
		t.Errorf("without the local filter: output %q, want %q", res.Output, "a")
	}

	missing := rcParse(t, env, `[{{ user.nope }}]`)
	user := map[string]interface{}{"user": map[string]interface{}{"name": "n"}}
	if res, _ := missing.Render(gocontext.Background(), user); len(res.Errors) != 0 {
		t.Errorf("lax: errors %v, want none", res.Errors)
	}
	if res, _ := missing.Render(gocontext.Background(), user, liquid.WithStrictVariables()); len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), "undefined variable nope") {
		t.Errorf("WithStrictVariables: errors %v, want the undefined variable", res.Errors)
	}
	if res, _ := shout.Render(gocontext.Background(), nil, liquid.WithStrictFilters()); len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), "undefined filter shout") {
		t.Errorf("WithStrictFilters: errors %v, want the undefined filter", res.Errors)
	}

	// The scores a render counted come with its result
	loop := rcParse(t, env, `{% for i in (1..10) %}{% assign x = i %}{% endfor %}`)
	if res, _ := loop.Render(gocontext.Background(), nil); res.RenderScore == 0 || res.AssignScore == 0 {
		t.Errorf("scores %d and %d, want both counted", res.RenderScore, res.AssignScore)
	}
}

// lazyTag renders its body in a scope of its own whose "lazy" is a func
// giving the number of the render of the tag, as a host tag may do.
type lazyTag struct {
	*liquid.Block
	n *int
}

func (l *lazyTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	*l.n++
	v := *l.n
	runtime.GC() // a freed scope must not lend its address to a later one
	context.Context().(*liquid.Context).Stack(map[string]interface{}{"lazy": func() interface{} { return v }}, func() {
		l.Block.RenderToOutput(context, output)
	})
}

// A func in a scope of the render is called for the value it has there, as
// Ruby calls a proc and writes its value back in the scope: a loop variable
// or an assign that holds a func, and a scope a tag pushes, give each its
// own func's value.
func TestRenderCallsTheFuncOfEachScope(t *testing.T) {
	funcs := map[string]interface{}{"funcs": []interface{}{
		func() interface{} { return 1 }, func() interface{} { return 2 },
	}}
	tpl := rcParse(t, rcEnv(), `{% for f in funcs %}{{ f }}{% endfor %}|{% assign g = funcs[0] %}{{ g }}{% assign g = funcs[1] %}{{ g }}`)
	if res, _ := tpl.Render(gocontext.Background(), funcs); res.Output != "12|12" {
		t.Errorf("funcs in a loop and in assigns render %q, want %q", res.Output, "12|12")
	}

	config := rcConfig()
	n := 0
	config.Tags["lazy"] = func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &lazyTag{Block: liquid.NewBlock(name, markup, pc), n: &n}, nil
	}
	env := mustEnv(config)
	loop := rcParse(t, env, `{% for i in (1..200) %}{% lazy %}{{ lazy }}{% endlazy %},{% endfor %}`)
	res, err := loop.Render(gocontext.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	for i := 1; i <= 200; i++ {
		fmt.Fprintf(&want, "%d,", i)
	}
	if res.Output != want.String() {
		t.Errorf("a func of a pushed scope renders a value of an earlier scope:\n got %q\nwant %q", res.Output, want.String())
	}
}

// The registers of a render reach a partial of render and include, as Ruby
// writes them in the static registers, which the isolated context of a
// render inherits; and a file system given as a register serves nested
// partials too.
func TestRenderRegistersReachPartials(t *testing.T) {
	config := rcConfig()

	config.Tags["marker"] = func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &markerTag{liquid.NewTag(name, markup, pc)}, nil
	}
	env := mustEnv(config)
	fs := &mapFileSystem{templates: map[string]string{"p": "<{% marker %}>", "outer": "[{% render 'inner' %}]", "inner": "I"}}
	tpl := rcParse(t, env, `<{% marker %}>{% render 'p' %}{% include 'p' %}`)
	res, err := tpl.Render(gocontext.Background(), nil, liquid.WithFileSystem(fs), liquid.WithRegisters(map[string]interface{}{"marker": "me"}))
	if want := "<me><me><me>"; err != nil || res.Output != want {
		t.Errorf("output %q (%v), want %q", res.Output, err, want)
	}
	nested := rcParse(t, env, `{% render 'outer' %}`)
	res, err = nested.Render(gocontext.Background(), nil, liquid.WithRegisters(map[string]interface{}{"file_system": fs}))
	if want := "[I]"; err != nil || res.Output != want {
		t.Errorf("a file system given as a register: output %q (%v), want %q", res.Output, err, want)
	}
	if len(tpl.Registers()) != 0 {
		t.Errorf("the template's registers changed: %v", tpl.Registers())
	}
}

// sameErrorDrop raises the same error value each time it is read.
type sameErrorDrop struct{}

// LiquidGet returns the key a template reads.
func (d *sameErrorDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "boom":
		return d.Boom(), true, nil
	}
	return nil, false, nil
}

var sameError = liquid.NewStandardError("same")

func (d *sameErrorDrop) Boom() interface{} { panic(sameError) }

// Each error a render records is in Result.Errors, as Ruby records each:
// the same error value raised twice is listed twice; and with RaiseErrors,
// an error raised through enclosing blocks and partials is listed once.
func TestRenderListsEachRecordedError(t *testing.T) {
	data := map[string]interface{}{"d": &sameErrorDrop{}}
	tpl := rcParse(t, rcEnv(), `{{ d.boom }}|{{ d.boom }}`)
	res, err := tpl.Render(gocontext.Background(), data)
	if err != nil || len(res.Errors) != 2 {
		t.Errorf("errors %v (%v), want the error twice", res.Errors, err)
	}
	nested := rcParse(t, rcEnv(), `{% if true %}{% for i in (1..2) %}{% render 'p', errors: errors %}{% endfor %}{% endif %}`)
	fs := &mapFileSystem{templates: map[string]string{"p": "{% if true %}{{ errors.standard_error }}{% endif %}"}}
	res, err = nested.Render(gocontext.Background(), map[string]interface{}{"errors": NewErrorDrop()}, liquid.WithFileSystem(fs), liquid.RaiseErrors())
	if err == nil || len(res.Errors) != 1 {
		t.Errorf("with RaiseErrors: errors %v (%v), want the raised error once", res.Errors, err)
	}
}

// A func that takes the render's Context is called with it, from the data
// and from a scope.
func TestRenderCallsAFuncOfTheContext(t *testing.T) {
	data := map[string]interface{}{
		"name": func(c *liquid.Context) interface{} { return c.TemplateName() },
		"list": []interface{}{func(c *liquid.Context) interface{} { return "item" }},
	}
	tpl := rcParse(t, rcEnv(), `{{ name }}|{% for f in list %}{{ f }}{% endfor %}`)
	tpl.SetName("page")
	if res, _ := tpl.Render(gocontext.Background(), data); res.Output != "page|item" {
		t.Errorf("output %q, want %q", res.Output, "page|item")
	}
}

// A func read from a scope is called once there: its value is written back
// in the scope, as Ruby does, and read again from it.
func TestAFuncOfAScopeIsCalledOnceThere(t *testing.T) {
	calls := 0
	data := map[string]interface{}{"funcs": []interface{}{func() interface{} { calls++; return calls }}}
	tpl := rcParse(t, rcEnv(), `{% for f in funcs %}{{ f }}{{ f }}{% endfor %}`)
	if res, _ := tpl.Render(gocontext.Background(), data); res.Output != "11" || calls != 1 {
		t.Errorf("output %q after %d calls, want %q after 1", res.Output, calls, "11")
	}
}

// WithFileSystem takes precedence over a file_system register, whatever the
// order of the options.
func TestWithFileSystemTakesPrecedenceOverTheRegister(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `{% render 'p' %}`)
	mine := &mapFileSystem{templates: map[string]string{"p": "mine"}}
	theirs := &mapFileSystem{templates: map[string]string{"p": "register"}}
	reg := liquid.WithRegisters(map[string]interface{}{"file_system": theirs})
	for _, opts := range [][]liquid.RenderOption{{reg, liquid.WithFileSystem(mine)}, {liquid.WithFileSystem(mine), reg}} {
		if res, err := tpl.Render(gocontext.Background(), nil, opts...); err != nil || res.Output != "mine" {
			t.Errorf("output %q (%v), want %q", res.Output, err, "mine")
		}
	}
}

// A func read by a key of a nested map or an index of a list is called, as
// Ruby's lookup_and_evaluate calls a proc at every key or index of a lookup,
// once per render: it does not render as a Go value.
func TestAFuncInAListOrANestedMapIsCalled(t *testing.T) {
	calls := 0
	count := func() interface{} { calls++; return calls }
	data := map[string]interface{}{
		"funcs": []interface{}{func() interface{} { return "one" }},
		"h":     map[string]interface{}{"f": func() interface{} { return "v" }, "n": count},
	}
	tpl := rcParse(t, rcEnv(), `{{ funcs[0] }}|{{ h.f }}|{{ h.n }}{{ h.n }}`)
	if res, _ := tpl.Render(gocontext.Background(), data); res.Output != "one|v|11" || calls != 1 {
		t.Errorf("output %q after %d calls, want %q after 1", res.Output, calls, "one|v|11")
	}
	if _, isFunc := data["funcs"].([]interface{})[0].(func() interface{}); !isFunc {
		t.Error("the list was written")
	}
	if _, isFunc := data["h"].(map[string]interface{})["f"].(func() interface{}); !isFunc {
		t.Error("the data was written")
	}
}

// uncomparableError is a Liquid error whose value holds a slice, so that
// two of its values cannot be compared with ==.
type uncomparableError struct {
	*liquid.StandardError
	detail interface{}
}

type uncomparableTag struct{ *liquid.Tag }

func (u *uncomparableTag) RenderToOutput(liquid.TagContext, *liquid.OutputBuffer) {
	panic(uncomparableError{StandardError: liquid.NewStandardError("uncomparable"), detail: []int{1}})
}

// With RaiseErrors, an error whose value cannot be compared is raised as
// itself through enclosing blocks.
func TestRaiseErrorsWithAnUncomparableError(t *testing.T) {
	config := rcConfig()

	config.Tags["uncomparable"] = func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return &uncomparableTag{liquid.NewTag(name, markup, pc)}, nil
	}
	env := mustEnv(config)
	tpl := rcParse(t, env, `{% if true %}{% uncomparable %}{% endif %}`)
	_, err := tpl.Render(gocontext.Background(), nil, liquid.RaiseErrors())
	var raised uncomparableError
	if !errors.As(err, &raised) {
		t.Errorf("err %v (%T), want the uncomparable error", err, err)
	}
}
