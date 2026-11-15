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
	tpl := rcParse(t, env, `{% if admin %}{% assign secret = token %}{% endif %}[{{ secret }}]{{ 1 | divided_by: d }}{% for i in (1..n) %}x{% endfor %}`)
	var wg sync.WaitGroup
	failures := make(chan string, 2000)
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			admin, long := i%2 == 0, i%3 == 0
			n := 1
			if long {
				n = 500
			}
			data := map[string]interface{}{"admin": admin, "token": fmt.Sprint("T", i), "d": 1, "n": n}
			res, err := tpl.RenderContext(gocontext.Background(), data)
			var limit *liquid.MemoryError
			switch {
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
// Ruby's instance assigns for Render.
func TestRenderContextLeavesTheTemplateAlone(t *testing.T) {
	tpl, err := liquid.ParseTemplate(`{% assign kept = "x" %}{{ nope | no_such_filter }}`, &liquid.TemplateOptions{Environment: rcEnv(), PersistInstanceAssigns: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tpl.RenderContext(gocontext.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if len(tpl.InstanceAssigns()) != 0 || len(tpl.Errors()) != 0 {
		t.Errorf("the template kept assigns %v and errors %v", tpl.InstanceAssigns(), tpl.Errors())
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

// WithFileSystem gives the partials of one render.
func TestRenderContextWithFileSystem(t *testing.T) {
	tpl := rcParse(t, rcEnv(), `[{% render "p" %}]`)
	for _, body := range []string{"one", "two"} {
		res, err := tpl.RenderContext(gocontext.Background(), nil, liquid.WithFileSystem(&mapFileSystem{templates: map[string]string{"p": body}}))
		if err != nil || res.Output != "["+body+"]" {
			t.Errorf("with partial %q: output %q, err %v", body, res.Output, err)
		}
	}
}

func intPtr(n int) *int { return &n }
