package integration

import (
	gocontext "context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func renderB(t *testing.T, src string, data map[string]interface{}, opts ...liquid.RenderOption) liquid.Result {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	res, err := tmpl.Render(gocontext.Background(), data, opts...)
	if err != nil {
		t.Fatalf("render %q: %v", src, err)
	}
	return res
}

// countingDrop counts the reads of its key calls.
type countingDrop struct{ calls int }

func (d *countingDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "calls":
		d.calls++
		return d.calls, true, nil
	case "nothing":
		return nil, true, nil
	}
	return nil, false, nil
}

// A nil pointer to a drop is nil: its keys are nil, as Ruby's nil.name.
func TestANilDropIsNil(t *testing.T) {
	var nd *countingDrop
	res := renderB(t, `[{{ nd }}|{{ nd.calls }}]|after`, map[string]interface{}{"nd": nd})
	if res.Output != "[|]|after" || len(res.Errors) != 0 {
		t.Errorf("got %q %v", res.Output, res.Errors)
	}
}

// A key a drop declares with a nil value is not an undefined method with
// strict variables; the parentloop of an outer loop is nil.
func TestADeclaredNilKeyIsDefined(t *testing.T) {
	res := renderB(t, `[{{ d.nothing }}]`, map[string]interface{}{"d": &countingDrop{}}, liquid.WithStrictVariables())
	if res.Output != "[]" || len(res.Errors) != 0 {
		t.Errorf("got %q %v", res.Output, res.Errors)
	}
	res = renderB(t, `{% for i in (1..1) %}[{{ forloop.parentloop.index }}]{% endfor %}`, nil)
	if res.Output != "[]" || len(res.Errors) != 0 {
		t.Errorf("parentloop: %q %v", res.Output, res.Errors)
	}
}

// An integer key on a drop is a missing key: an UndefinedDropMethod with
// strict variables, as Ruby's invoke_drop(0).
func TestAnIntegerKeyOnADropIsAMissingKey(t *testing.T) {
	res := renderB(t, `[{{ d[0] }}]`, map[string]interface{}{"d": &countingDrop{}}, liquid.WithStrictVariables())
	var undefined *liquid.UndefinedDropMethod
	if res.Output != "[]" || len(res.Errors) != 1 || !errors.As(res.Errors[0], &undefined) {
		t.Errorf("got %q %v, want an UndefinedDropMethod", res.Output, res.Errors)
	}
}

// The items of a []T whose *T is a drop are the items themselves: what a key
// changes stays for the next key.
func TestTheItemsOfASliceOfValueDropsKeepTheirState(t *testing.T) {
	data := map[string]interface{}{"ds": []countingDrop{{}}}
	res := renderB(t, `{{ ds[0].calls }}{{ ds[0].calls }}|{% for d in ds %}{{ d.calls }}{{ d.calls }}{% endfor %}|{{ ds | map: "calls" | join }}`, data)
	if res.Output != "12|34|5" {
		t.Errorf("got %q, want %q", res.Output, "12|34|5")
	}
}

// The local filters of a render reach the partials it renders.
func TestLocalFiltersReachAPartial(t *testing.T) {
	shout := liquid.Filter("shout", func(_ *liquid.Context, in interface{}, _ ...interface{}) (interface{}, error) {
		return strings.ToUpper(fmt.Sprint(in)) + "!", nil
	})
	fs := &mapFileSystem{templates: map[string]string{"p": "{{ 'b' | shout }}"}}
	res := renderB(t, `{{ 'a' | shout }}{% render 'p' %}`, nil, liquid.WithFilters(shout), liquid.WithFileSystem(fs))
	if res.Output != "A!B!" {
		t.Errorf("got %q", res.Output)
	}
}

// A standard filter given an argument too many, append and raw among them,
// gives nil and records an ArgumentError.
func TestTheArityOfTheStandardFilters(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: true})
	for _, src := range []string{`[{{ "a" | append: "b", "c" }}]`, `[{{ "a" | raw: 1 }}]`, `[{{ "a" | upcase: 1 }}]`} {
		tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		res, _ := tmpl.Render(gocontext.Background(), nil)
		var argument *liquid.ArgumentError
		if res.Output != "[]" || len(res.Errors) != 1 || !errors.As(res.Errors[0], &argument) {
			t.Errorf("%s: %q %v", src, res.Output, res.Errors)
		}
	}
}

// A filter with no function is refused, as a render option as in an
// environment; so is a name no template can write.
func TestAFilterMustBeWritable(t *testing.T) {
	spaced := liquid.Filter("has space", func(*liquid.Context, interface{}, ...interface{}) (interface{}, error) { return nil, nil })
	if _, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Filters: []liquid.FilterDef{spaced}}); err == nil {
		t.Error("a filter named with a space was declared")
	}
	tmpl, err := liquid.ParseTemplate(`{{ "A" | downcase }}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpl.Render(gocontext.Background(), nil, liquid.WithFilters(liquid.Filter("downcase", nil))); err == nil {
		t.Error("a render took a filter with no function")
	}
}

// The visitor walks the attributes of render and include, by name, each run
// the same; a callback that gives no context keeps its parent's.
func TestTheVisitorWalksTheAttributesOfPartials(t *testing.T) {
	src := `{% render 'p', e: e, a: a, d: d, c: c, b: b %}{% include 'q', z: z, y: y %}`
	first := visitVariables(t, src)
	if !reflect.DeepEqual(first, []interface{}{"a", "b", "c", "d", "e", "y", "z"}) {
		t.Errorf("got %v, want the attributes by name", first)
	}
	for i := 0; i < 20; i++ {
		if got := visitVariables(t, src); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d gave %v, run 0 %v", i, got, first)
		}
	}
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{% if a %}{{ b }}{% endif %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	var contexts []interface{}
	liquid.ForParseTreeVisitor(tmpl.Root(), nil).
		AddCallbackFor([]interface{}{(*liquid.VariableLookup)(nil)}, func(node, context interface{}) (interface{}, interface{}) {
			contexts = append(contexts, context)
			return nil, nil
		}).Visit("root")
	for _, c := range contexts {
		if c != "root" {
			t.Errorf("a callback got the context %v, want the parent's", c)
		}
	}
}
