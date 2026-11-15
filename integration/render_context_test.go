package integration

import (
	"bytes"
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

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

// RenderContext writes what Render writes, for every template of the Shopify
// differential corpus, in both modes.
func TestRenderContextWritesWhatRenderWrites(t *testing.T) {
	local := time.Local
	time.Local = time.UTC
	defer func() { time.Local = local }()
	raw, err := os.ReadFile(differentialDir + "/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		Results []differentialResult `json:"results"`
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	data := differentialData(t)
	compared := 0
	for _, r := range expected.Results {
		env := newEnv(r.Mode, limits{})
		tpl, err := liquid.ParseTemplate(r.Template, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			continue
		}
		want := tpl.Render(copyData(data), nil)
		res, err := tpl.RenderContext(gocontext.Background(), copyData(data))
		if err != nil {
			t.Errorf("%s: RenderContext error %v; Render wrote %q", r.key(), err, want)
			continue
		}
		if res.Output != want {
			t.Errorf("%s: RenderContext wrote %q, Render %q", r.key(), res.Output, want)
		}
		compared++
	}
	if compared < 1000 {
		t.Fatalf("compared %d results only", compared)
	}
}

// Concurrent renders of one template each get their own output, errors and
// limits, and the template keeps none of them.
func TestRenderContextIsolatesConcurrentRenders(t *testing.T) {
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
			res, err := tpl.RenderContext(gocontext.Background(), data)
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
	if len(tpl.Errors()) != 0 || len(tpl.InstanceAssigns()) != 0 {
		t.Errorf("the template kept state: errors %v, assigns %v", tpl.Errors(), tpl.InstanceAssigns())
	}
	if rl := tpl.ResourceLimits(); rl != nil && (rl.Reached() || rl.RenderScore() != 0) {
		t.Errorf("the template kept the limits of a render")
	}
}

// RenderContext changes nothing in the template, even when the template keeps
// Ruby's instance assigns for Render: no assigns, errors, registers or
// profiler, and the template's assigns and registers are only read.
func TestRenderContextLeavesTheTemplateAlone(t *testing.T) {
	env := rcEnv()
	tpl, err := liquid.ParseTemplate(`{% assign kept = "x" %}[{{ kept }}{{ greeting }}{{ errors.standard_error }}{% render "p" %}]`,
		&liquid.TemplateOptions{Environment: env, PersistInstanceAssigns: true, Profile: true})
	if err != nil {
		t.Fatal(err)
	}
	tpl.Assigns()["greeting"] = "hi"
	tpl.Registers()["marker"] = "template register"
	fs := &mapFileSystem{templates: map[string]string{"p": "p"}}
	res, err := tpl.RenderContext(gocontext.Background(), map[string]interface{}{"errors": NewErrorDrop()}, liquid.WithFileSystem(fs))
	if err != nil {
		t.Fatal(err)
	}
	if want := "[xhiLiquid error (line 1): standard errorp]"; res.Output != want {
		t.Errorf("output %q, want %q", res.Output, want)
	}
	if len(res.Errors) != 1 {
		t.Errorf("Result.Errors = %v, want the standard error", res.Errors)
	}
	if len(tpl.InstanceAssigns()) != 0 || len(tpl.Errors()) != 0 {
		t.Errorf("the template kept assigns %v and errors %v", tpl.InstanceAssigns(), tpl.Errors())
	}
	if _, set := tpl.Registers()["file_system"]; set || len(tpl.Registers()) != 1 {
		t.Errorf("the template's registers changed: %v", tpl.Registers())
	}
	if tpl.Profiler() != nil {
		t.Error("RenderContext set the template's profiler")
	}
	if len(tpl.Assigns()) != 1 {
		t.Errorf("the template's assigns changed: %v", tpl.Assigns())
	}
}

// A template register stays readable next to WithFileSystem, by a tag that
// reads the registers.
func TestRenderContextKeepsTemplateRegisters(t *testing.T) {
	env := rcEnv()
	env.RegisterTag("marker", tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		return &markerTag{liquid.NewTag(name, markup, pc)}, nil
	}))
	tpl := rcParse(t, env, `[{% marker %}{% render "p" %}]`)
	tpl.Registers()["marker"] = "M"
	fs := &mapFileSystem{templates: map[string]string{"p": "{% marker %}"}}
	res, err := tpl.RenderContext(gocontext.Background(), nil, liquid.WithFileSystem(fs))
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
func TestRenderContextConcurrentFileSystems(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `[{% render "p" %}]`)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprint("p", i)
			res, err := tpl.RenderContext(gocontext.Background(), nil, liquid.WithFileSystem(&mapFileSystem{templates: map[string]string{"p": body}}))
			if err != nil || res.Output != "["+body+"]" {
				t.Errorf("render %d: output %q, err %v", i, res.Output, err)
			}
		}(i)
	}
	wg.Wait()
}

// A func in data or in the template's assigns is called by each render and
// stays a func: neither map is written, so concurrent renders do not crash.
func TestRenderContextLeavesFuncsAlone(t *testing.T) {
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
	first, _ := tpl.RenderContext(gocontext.Background(), data)
	second, _ := tpl.RenderContext(gocontext.Background(), data)
	if first.Output != "1-l" || second.Output != "2-l" {
		t.Errorf("renders %q and %q, want each to call the func", first.Output, second.Output)
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
			_, _ = tpl.RenderContext(gocontext.Background(), data)
		}()
	}
	wg.Wait()
}

// An error message carries the template's name and line.
func TestRenderContextErrorsNameTheTemplate(t *testing.T) {
	tpl, err := liquid.ParseTemplate("a\n{{ errors.standard_error }}", &liquid.TemplateOptions{Environment: rcEnv(), LineNumbers: true})
	if err != nil {
		t.Fatal(err)
	}
	tpl.SetName("welcome")
	res, err := tpl.RenderContext(gocontext.Background(), map[string]interface{}{"errors": NewErrorDrop()})
	if err != nil {
		t.Fatal(err)
	}
	if want := "a\nLiquid error (welcome line 2): standard error"; res.Output != want {
		t.Errorf("output %q, want %q", res.Output, want)
	}
}

// A recorded error leaves a complete output and comes back in Result.Errors;
// with RaiseErrors it ends the render.
func TestRenderContextRecordedErrors(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `a{{ errors.standard_error }}b`)
	data := map[string]interface{}{"errors": NewErrorDrop()}
	res, err := tpl.RenderContext(gocontext.Background(), data)
	if err != nil {
		t.Fatalf("err = %v, want a recorded error", err)
	}
	if res.Output != "aLiquid error: standard errorb" || len(res.Errors) != 1 {
		t.Errorf("output %q, errors %v", res.Output, res.Errors)
	}
	res, err = tpl.RenderContext(gocontext.Background(), data, liquid.RaiseErrors())
	var standard *liquid.StandardError
	if !errors.As(err, &standard) {
		t.Errorf("with RaiseErrors err = %v, want the StandardError", err)
	}
	if res.Output != "a" {
		t.Errorf("with RaiseErrors output %q, want what was written before the error", res.Output)
	}
	nested := rcParse(t, rcEnv(), `{% if true %}{% for i in (1..1) %}{% if true %}{{ errors.standard_error }}{% endif %}{% endfor %}{% endif %}`)
	res, err = nested.RenderContext(gocontext.Background(), data, liquid.RaiseErrors())
	if err == nil || len(res.Errors) != 1 {
		t.Errorf("RaiseErrors through nested blocks: err %v, errors %v, want the error once", err, res.Errors)
	}
	// The errors recorded before a fatal error come back with it
	before := rcParse(t, rcEnv(), `{{ errors.standard_error }}{% for i in (1..100000) %}xxxxxxxxxx{% endfor %}`)
	res, err = before.RenderContext(gocontext.Background(), data, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: intPtr(1000)}))
	var limit *liquid.MemoryError
	if !errors.As(err, &limit) || len(res.Errors) != 1 {
		t.Errorf("a limit after a recorded error: err %v, errors %v", err, res.Errors)
	}
}

// WithLimits replaces the limits it sets and keeps the template's others.
func TestRenderContextWithLimitsKeepsTheOthers(t *testing.T) {
	env := rcEnv()
	env.SetDefaultResourceLimits(map[string]interface{}{"render_score_limit": 50})
	tpl := rcParse(t, env, `{% for i in (1..1000) %}{% endfor %}`)
	_, err := tpl.RenderContext(gocontext.Background(), nil, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: intPtr(1000000)}))
	var limit *liquid.MemoryError
	if !errors.As(err, &limit) {
		t.Errorf("err %v, want the template's score limit", err)
	}
}

// Render writes a PanicError raised by a filter in place of that node only, as
// any other panic: RenderContext's rule applies to RenderContext only.
type rcSubFilters struct{}

func (f *rcSubFilters) Sub(input interface{}) interface{} {
	panic(&liquid.PanicError{Value: "inner render"})
}

func TestRenderKeepsAPanicErrorToItsNode(t *testing.T) {
	env := rcEnv()
	if err := env.RegisterFilter(&rcSubFilters{}); err != nil {
		t.Fatal(err)
	}
	tpl := rcParse(t, env, `before {{ 1 | sub }} after`)
	if got, want := tpl.Render(map[string]interface{}{}, nil), "before Liquid error: internal after"; got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

// A resource limit and a cancellation end the render with their error.
func TestRenderContextLimitsAndCancellation(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `{% for i in (1..100000) %}xxxxxxxxxx{% endfor %}`)
	_, err := tpl.RenderContext(gocontext.Background(), nil, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: intPtr(1000)}))
	var limit *liquid.MemoryError
	if !errors.As(err, &limit) {
		t.Errorf("over the limit err = %v, want a MemoryError", err)
	}
	ctx, cancel := gocontext.WithCancel(gocontext.Background())
	cancel()
	_, err = tpl.RenderContext(ctx, nil)
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
func TestRenderContextGoPanicIsFatal(t *testing.T) {
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
		res, err := tpl.RenderContext(gocontext.Background(), data)
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
		if got := tpl.Render(data, nil); !strings.Contains(got, "Liquid error: internal") {
			t.Errorf("%s: Render wrote %q, want the inline internal error as before", c.src, got)
		}
	}
}

// WithFileSystem gives the partials of one render, nested ones included.
func TestRenderContextWithFileSystem(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `[{% render "p" %}]`)
	for _, body := range []string{"one", "two"} {
		res, err := tpl.RenderContext(gocontext.Background(), nil, liquid.WithFileSystem(&mapFileSystem{templates: map[string]string{"p": body}}))
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
	res, err := nested.RenderContext(gocontext.Background(), map[string]interface{}{"name": "y"}, liquid.WithFileSystem(fs))
	if want := "[<leaf x>|<leaf y>]"; err != nil || res.Output != want {
		t.Errorf("nested partials: output %q, err %v, want %q", res.Output, err, want)
	}
}

func intPtr(n int) *int { return &n }
