package integration

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func strictTemplate(t *testing.T, source string) *liquid.Template {
	t.Helper()
	return strictTemplateIn(t, source, "")
}

// strictTemplateIn parses source in errorMode, or the default mode.
func strictTemplateIn(t *testing.T, source string, errorMode liquid.ErrorMode) *liquid.Template {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: errorMode})
	tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tmpl
}

var undefinedVariablesData = map[string]interface{}{
	"x": 33,
	"z": map[string]interface{}{"a": 32, "c": map[string]interface{}{"e": 31}},
}

// Ported from: test_undefined_variables (test/integration/template_test.rb)
func TestTemplate_UndefinedVariables(t *testing.T) {
	tmpl := strictTemplate(t, "{{x}} {{y}} {{z.a}} {{z.b}} {{z.c.d}}")
	result := renderRuby(tmpl, undefinedVariablesData, &rubyOptions{StrictVariables: true})

	if result != "33  32  " {
		t.Errorf("result %q, want %q", result, "33  32  ")
	}
	errs := rubyErrors(tmpl)
	if len(errs) != 3 {
		t.Fatalf("errors %v, want 3", errs)
	}
	for i, want := range []string{
		"Liquid error: undefined variable y",
		"Liquid error: undefined variable b",
		"Liquid error: undefined variable d",
	} {
		if _, ok := errs[i].(*liquid.UndefinedVariable); !ok {
			t.Errorf("error %d is a %T, want an UndefinedVariable", i, errs[i])
		}
		if errs[i].Error() != want {
			t.Errorf("error %d is %q, want %q", i, errs[i].Error(), want)
		}
	}
}

// Ported from: test_nil_value_does_not_raise
func TestTemplate_NilValueDoesNotRaise(t *testing.T) {
	tmpl := strictTemplateIn(t, "some{{x}}thing", "strict")
	result := renderRubyBang(tmpl, map[string]interface{}{"x": nil}, &rubyOptions{StrictVariables: true})

	if n := len(rubyErrors(tmpl)); n != 0 {
		t.Errorf("errors %v, want none", rubyErrors(tmpl))
	}
	if result != "something" {
		t.Errorf("result %q, want %q", result, "something")
	}
}

// Ported from: test_undefined_variables_raise
func TestTemplate_UndefinedVariablesRaise(t *testing.T) {
	tmpl := strictTemplate(t, "{{x}} {{y}} {{z.a}} {{z.b}} {{z.c.d}}")
	defer func() {
		if _, ok := recover().(*liquid.UndefinedVariable); !ok {
			t.Error("render! did not raise an UndefinedVariable")
		}
	}()
	renderRubyBang(tmpl, undefinedVariablesData, &rubyOptions{StrictVariables: true})
}

// Render with strict variables records an undefined name of the first
// level, as Ruby's find_variable: it was recorded only below the first level.
func TestRenderWithStrictVariablesRecordsAnUndefinedName(t *testing.T) {
	tmpl := strictTemplate(t, "a{{ nope }}b")
	res, err := tmpl.Render(gocontext.Background(), nil, liquid.WithStrictVariables())
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "ab" {
		t.Errorf("output %q, want %q", res.Output, "ab")
	}
	if len(res.Errors) != 1 || res.Errors[0].Error() != "Liquid error: undefined variable nope" {
		t.Errorf("errors %v, want the undefined variable nope", res.Errors)
	}
	if res, _ := tmpl.Render(gocontext.Background(), nil); len(res.Errors) != 0 {
		t.Errorf("without strict variables: errors %v, want none", res.Errors)
	}
}

// An array index follows Ruby's VariableLookup: an Integer key reads the
// array as Array#[] (a JSON number comes to Go as an integral float64), and
// any other key is no index: nil, or an undefined variable with strict
// variables.
func TestAnArrayIndexFollowsRuby(t *testing.T) {
	data := map[string]interface{}{
		"arr":   []interface{}{1, 2, 3},
		"typed": []int{1, 2, 3},
		"fm1":   float64(-1), "f9": float64(9), "u1": uint(1), "half": 1.5,
		"i64m1": int64(-1), "i8": int8(0), "huge": uint64(1) << 40,
	}
	for _, c := range []struct {
		source, want, strictErr string
	}{
		{"{{ arr[fm1] }}", "3", ""},
		{"{{ typed[fm1] }}", "3", ""},
		{"{{ arr[u1] }}", "2", ""},
		{"{{ arr[i64m1] }}{{ typed[i8] }}", "31", ""},
		{"[{{ arr[huge] }}]", "[]", ""},
		{"[{{ arr[f9] }}]", "[]", ""},
		{"[{{ arr['1'] }}]", "[]", "Liquid error: undefined variable 1"},
		{"[{{ typed['1'] }}]", "[]", "Liquid error: undefined variable 1"},
		{"[{{ arr[half] }}]", "[]", "Liquid error: undefined variable 1.5"},
	} {
		tmpl := strictTemplate(t, c.source)
		if got := renderRuby(tmpl, data, nil); got != c.want {
			t.Errorf("%s renders %q, want %q", c.source, got, c.want)
		}
		renderRuby(tmpl, data, &rubyOptions{StrictVariables: true})
		errs := rubyErrors(tmpl)
		switch {
		case c.strictErr == "" && len(errs) != 0:
			t.Errorf("%s with strict variables: errors %v, want none", c.source, errs)
		case c.strictErr != "" && (len(errs) != 1 || errs[0].Error() != c.strictErr):
			t.Errorf("%s with strict variables: errors %v, want %q", c.source, errs, c.strictErr)
		}
	}
}

// With strict variables, a render given no data finds the template's
// assigns, as Ruby's render(nil) makes them its one environment; given a
// map, Ruby looks a name up in the map first, which then decides.
func TestStrictVariablesAndTheTemplateAssigns(t *testing.T) {
	tmpl := strictTemplate(t, "[{{ x }}]")
	tmpl.Assigns()["x"] = "from-assigns"
	res, err := tmpl.Render(gocontext.Background(), nil, liquid.WithStrictVariables())
	if err != nil || res.Output != "[from-assigns]" || len(res.Errors) != 0 {
		t.Errorf("no data: output %q, errors %v (%v), want %q and no error", res.Output, res.Errors, err, "[from-assigns]")
	}
	res, err = tmpl.Render(gocontext.Background(), map[string]interface{}{}, liquid.WithStrictVariables())
	if err != nil || res.Output != "[]" || len(res.Errors) != 1 {
		t.Errorf("a map: output %q, errors %v (%v), want %q and the undefined variable", res.Output, res.Errors, err, "[]")
	}
}

// A name that holds nil in the data is looked up in the template's assigns
// next, as Ruby's try_variable_find_in_environments goes on past a nil;
// with strict variables the data decides, nil included.
func TestANilInTheDataFallsThroughToTheAssigns(t *testing.T) {
	tmpl := strictTemplate(t, "[{{ x }}]")
	tmpl.Assigns()["x"] = "y"
	data := map[string]interface{}{"x": nil}
	if res, _ := tmpl.Render(gocontext.Background(), data); res.Output != "[y]" {
		t.Errorf("output %q, want %q", res.Output, "[y]")
	}
	if res, _ := tmpl.Render(gocontext.Background(), data, liquid.WithStrictVariables()); res.Output != "[]" || len(res.Errors) != 0 {
		t.Errorf("strict variables: output %q, errors %v, want %q and no error", res.Output, res.Errors, "[]")
	}
}

type namedInt int
type namedInt64 int64
type namedFloat float64

// Every Go integer kind, named or not, and an integral float of any float
// kind, is an Integer key of an array.
func TestEveryGoIntegerIsAnArrayIndex(t *testing.T) {
	data := map[string]interface{}{
		"arr": []interface{}{"a", "b", "c"},
		"ni":  namedInt(1), "ni64": namedInt64(1), "month": time.February,
		"f32": float32(1), "nf": namedFloat(1), "f32half": float32(1.5),
	}
	for src, want := range map[string]string{
		"{{ arr[ni] }}": "b", "{{ arr[ni64] }}": "b", "{{ arr[month] }}": "c",
		"{{ arr[f32] }}": "b", "{{ arr[nf] }}": "b", "[{{ arr[f32half] }}]": "[]",
	} {
		if got := renderRuby(strictTemplate(t, src), data, nil); got != want {
			t.Errorf("%s renders %q, want %q", src, got, want)
		}
	}
}

// liquidNil is a value whose liquid form is nil.
type liquidNil struct{}

func (liquidNil) ToLiquid() interface{} { return nil }

// A value of the data whose liquid form is nil is found, as Ruby tests the
// value itself before to_liquid: the assigns are not looked up.
func TestAValueWhoseLiquidFormIsNilIsFound(t *testing.T) {
	tmpl := strictTemplate(t, "[{{ x }}]")
	tmpl.Assigns()["x"] = "A"
	if res, _ := tmpl.Render(gocontext.Background(), map[string]interface{}{"x": liquidNil{}}); res.Output != "[]" {
		t.Errorf("output %q, want %q", res.Output, "[]")
	}
}

// A segment of digits after a dot is an index of an array, as Notifuse
// writes the paths of its display conditions (order.items.0.sku): a
// deliberate difference from Ruby, where it stays the string "0". A map
// still reads it as the key "0", and a string in brackets is no index.
func TestADottedSegmentOfDigitsIsAnIndex(t *testing.T) {
	data := map[string]interface{}{
		"order": map[string]interface{}{"items": []interface{}{
			map[string]interface{}{"sku": "A1"}, map[string]interface{}{"sku": "B2"},
		}},
		"tags": []interface{}{"new", "vip"},
		"h":    map[string]interface{}{"0": "zero"},
	}
	for src, want := range map[string]string{
		"{{ order.items.0.sku }}{{ order.items.1.sku }}": "A1B2",
		"{{ tags.1 }}":      "vip",
		"[{{ tags.5 }}]":    "[]",
		"{{ h.0 }}":         "zero",
		"[{{ tags['1'] }}]": "[]",
	} {
		if got := renderRuby(strictTemplate(t, src), data, nil); got != want {
			t.Errorf("%s renders %q, want %q", src, got, want)
		}
	}
}
