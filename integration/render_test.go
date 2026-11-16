package integration

import (
	"bytes"
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func rcEnv() *liquid.Environment {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	return env
}

func rcParse(t *testing.T, env *liquid.Environment, src string) *liquid.Template {
	t.Helper()
	tpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return tpl
}

func copyData(data map[string]interface{}) map[string]interface{} {
	var copied map[string]interface{}
	raw, _ := json.Marshal(data)
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	_ = dec.Decode(&copied)
	return withIntegers(copied).(map[string]interface{})
}

// Concurrent renders of one template each get their own output, errors and
// limits, and the template keeps none of them.
func TestRenderIsolatesConcurrentRenders(t *testing.T) {
	env := rcEnv()
	env.SetDefaultResourceLimits(map[string]interface{}{"render_length_limit": 200})
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

// Render changes nothing in the template, even when the template keeps
// Ruby's instance assigns for Render: no assigns, errors, registers or
// profiler, and the template's assigns and registers are only read.
func TestRenderLeavesTheTemplateAlone(t *testing.T) {
	env := rcEnv()
	tpl, err := liquid.ParseTemplate(`{% assign kept = "x" %}[{{ kept }}{{ greeting }}{{ errors.standard_error }}{% render "p" %}]`,
		&liquid.TemplateOptions{Environment: env, Profile: true})
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
	env := rcEnv()
	env.RegisterTag("marker", tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		return &markerTag{liquid.NewTag(name, markup, pc)}, nil
	}))
	tpl := rcParse(t, env, `[{% marker %}{% render "p" %}]`)
	tpl.Registers()["marker"] = "M"
	fs := &mapFileSystem{templates: map[string]string{"p": "{% marker %}"}}
	res, err := tpl.Render(gocontext.Background(), nil, liquid.WithFileSystem(fs))
	if want := "[MM]"; err != nil || res.Output != want {
		t.Errorf("output %q, err %v, want %q", res.Output, err, want)
	}
}

type markerTag struct{ *liquid.Tag }

func (m *markerTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	if v, ok := context.Registers().Get("marker").(string); ok {
		*output += v
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
	env := rcEnv()
	env.SetDefaultResourceLimits(map[string]interface{}{"render_score_limit": 50})
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

type rcPanicDrop struct{ *liquid.Drop }

func (d *rcPanicDrop) Name() interface{} {
	var s []int
	return s[3] // a Go panic: index out of range
}

// A Go panic in a filter, a drop or a partial ends the render with its stack;
// Render writes it as "Liquid error: internal", as before.
func TestRenderGoPanicIsFatal(t *testing.T) {
	env := rcEnv()
	if err := env.RegisterFilter(&rcPanicFilters{}); err != nil {
		t.Fatal(err)
	}
	env.SetFileSystem(&mapFileSystem{templates: map[string]string{"p": "{{ d.name }}"}})
	data := map[string]interface{}{"d": &rcPanicDrop{Drop: liquid.NewDrop()}}
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
		if got := renderRuby(tpl, data, nil); !strings.Contains(got, "Liquid error: internal") {
			t.Errorf("%s: Render wrote %q, want the inline internal error as before", c.src, got)
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
	env := rcEnv()
	env.RegisterTag("marker", tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		return &markerTag{liquid.NewTag(name, markup, pc)}, nil
	}))
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
	if res, err := shout.Render(gocontext.Background(), nil, liquid.WithFilters(&rcShoutFilters{})); err != nil || res.Output != "a!" {
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
