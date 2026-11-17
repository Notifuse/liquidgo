package integration

import (
	gocontext "context"
	"errors"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// TestBlocksRenderTheirBodyOnce checks that tablerow and ifchanged apply the
// effects of their body once per item, as Ruby does.
func TestBlocksRenderTheirBodyOnce(t *testing.T) {
	for _, c := range []struct{ source, want string }{
		{"{% tablerow i in (1..2) %}{% increment c %}{% endtablerow %}",
			"<tr class=\"row1\">\n<td class=\"col1\">0</td><td class=\"col2\">1</td></tr>\n"},
		{"{% tablerow i in (1..3) %}{% cycle 'a', 'b' %}{% endtablerow %}",
			"<tr class=\"row1\">\n<td class=\"col1\">a</td><td class=\"col2\">b</td><td class=\"col3\">a</td></tr>\n"},
		{"{% tablerow i in (1..2) %}{% assign n = n | append: 'x' %}{% endtablerow %}{{ n }}",
			"<tr class=\"row1\">\n<td class=\"col1\"></td><td class=\"col2\"></td></tr>\nxx"},
		{"{% for i in (1..3) %}{% ifchanged %}{% increment c %}{% endifchanged %}{% endfor %}", "012"},
		{"{% for i in (1..2) %}{% ifchanged %}{% cycle 'a', 'b' %}{% endifchanged %}{% endfor %}", "ab"},
		{"{% ifchanged %}{% assign n = n | append: 'x' %}{% endifchanged %}{{ n }}", "x"},
	} {
		if got := renderWith(newEnv("", limits{}), c.source, nil); got.output != c.want || got.parseErr != nil {
			t.Errorf("%q renders %q (%v), want %q", c.source, got.output, got.parseErr, c.want)
		}
	}
}

// panicTag panics while it renders, with err when it is set, or with a
// runtime error.
type panicTag struct {
	*liquid.Tag
	err   error
	blank bool
}

func (p *panicTag) RenderToOutput(liquid.TagContext, *liquid.OutputBuffer) {
	if p.err != nil {
		panic(p.err)
	}
	var items []int
	_ = items[len(items)]
}

func (p *panicTag) Blank() bool { return p.blank }

// panicConfig is the config of an environment with the tags of panicTag and
// limits l.
func panicConfig(l limits) liquid.EnvironmentConfig {
	config := envConfig("", l)
	register := func(name string, err func() error, blank bool) {
		config.Tags[name] = tags.TagConstructor(func(tagName, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
			var e error
			if err != nil {
				e = err()
			}
			return &panicTag{Tag: liquid.NewTag(tagName, markup, pc), err: e, blank: blank}, nil
		})
	}
	register("crash", nil, false)
	register("fail", func() error { return liquid.NewArgumentError("bad argument") }, false)
	register("blankfail", func() error { return liquid.NewArgumentError("bad argument") }, true)
	register("undefined", func() error { return liquid.NewUndefinedVariable("undefined variable y") }, false)
	register("toomuch", func() error { return liquid.NewMemoryError("Memory limits exceeded") }, false)
	return config
}

// TestAnErrorInANodeKeepsTheRestOfTheOutput checks the rules of Ruby's
// rescue_render_node: the message of an error takes the place of its node,
// unless the node is a blank tag or the error is an undefined variable,
// drop method or filter, which write nothing; a resource limit ends the
// render, from any block or partial. A Go panic (crash), which Ruby writes
// as an internal error and goes on from, ends the render in liquidgo.
func TestAnErrorInANodeKeepsTheRestOfTheOutput(t *testing.T) {
	config := panicConfig(limits{})
	config.FileSystem = NewStubFileSystem(map[string]string{"p": "{% toomuch %}"})
	env := mustEnv(config)
	for _, c := range []struct {
		source, want, wantError string
	}{
		{"a{% fail %}b", "aLiquid error: bad argumentb", "Liquid error: bad argument"},
		{"a{% blankfail %}b", "ab", "Liquid error: bad argument"},
		{"a{% undefined %}b", "ab", "Liquid error: undefined variable y"},
		{"A{% cycle a: %}B", "AB", ""},
	} {
		r := renderWith(env, c.source, nil)
		if r.parseErr != nil || r.output != c.want {
			t.Errorf("%q renders %q (%v), want %q", c.source, r.output, r.parseErr, c.want)
		}
		if c.wantError == "" && len(r.errors) > 0 {
			t.Errorf("%q: unexpected errors %v", c.source, r.errors)
		}
		if c.wantError != "" && (len(r.errors) == 0 || !strings.Contains(r.errors[0].Error(), strings.TrimPrefix(c.wantError, "Liquid error: "))) {
			t.Errorf("%q: errors %v, want %q", c.source, r.errors, c.wantError)
		}
	}
	// A Go panic and a resource limit end the render, from any block or
	// partial, with the output written before
	var panicked *liquid.PanicError
	var limit *liquid.MemoryError
	for _, c := range []struct {
		source string
		want   interface{}
	}{
		{"a{% crash %}b", &panicked},
		{"a{% for i in (1..2) %}{% crash %}{% endfor %}b", &panicked},
		{"a{% toomuch %}b", &limit},
		{"a{% for i in (1..2) %}{% toomuch %}{% endfor %}b", &limit},
		{"a{% capture c %}{% toomuch %}{% endcapture %}b", &limit},
		{"a{% if true %}{% toomuch %}{% endif %}b", &limit},
		{"a{% render 'p' %}b", &limit},
		{"a{% include 'p' %}b", &limit},
	} {
		tmpl, err := liquid.ParseTemplate(c.source, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		res, err := tmpl.Render(gocontext.Background(), nil)
		if !errors.As(err, c.want) {
			t.Errorf("%q: err %v (%T), want a %T", c.source, err, err, c.want)
		}
		if res.Output != "a" || len(res.Errors) != 0 {
			t.Errorf("%q: output %q, errors %v, want %q and none", c.source, res.Output, res.Errors, "a")
		}
	}
}

// boomDrop raises a Liquid error when its method is called.
type boomDrop struct{}

// LiquidGet returns the key a template reads.
func (b *boomDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "boom":
		return b.Boom(), true, nil
	}
	return nil, false, nil
}

func (b *boomDrop) Boom() string {
	panic(liquid.NewStandardError("boom"))
}

// TestAnIncludeRestoresTheTemplateNameOnAnError checks that the template
// name and the partial flag are restored when an include ends on an error,
// as Ruby's ensure: a later error is reported for the template itself.
func TestAnIncludeRestoresTheTemplateNameOnAnError(t *testing.T) {
	config := envConfig("", limits{})
	config.FileSystem = NewStubFileSystem(map[string]string{"part": "p"})
	env := mustEnv(config)
	source := "a\n{% include 'part', v: d.boom %}\nb\n{{ 'x' | nofilter }}"
	tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env, LineNumbers: true})
	if err != nil {
		t.Fatal(err)
	}
	renderRuby(tmpl, map[string]interface{}{"d": &boomDrop{}}, &rubyOptions{StrictFilters: true})
	errs := rubyErrors(tmpl)
	if len(errs) == 0 {
		t.Fatal("no error recorded")
	}
	last := errs[len(errs)-1].Error()
	if last != "Liquid error (line 4): undefined filter nofilter" {
		t.Errorf("the last error is %q, want it on line 4 of the template itself", last)
	}
}
