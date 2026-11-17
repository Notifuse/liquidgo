package liquid

import (
	gocontext "context"
	"errors"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func containsName(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// renderFilters parses source and renders it with data and options.
func renderFilters(t *testing.T, env *Environment, source string, opts ...RenderOption) (Result, error) {
	t.Helper()
	tmpl, err := ParseTemplate(source, &TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tmpl.Render(gocontext.Background(), map[string]interface{}{"x": "a"}, opts...)
}

func constant(v string) FilterFunc {
	return func(*Context, interface{}, ...interface{}) (interface{}, error) { return v, nil }
}

// A local filter of the render takes precedence over the environment's, the
// last given first, which takes precedence over the standard one.
func TestFilterPrecedence(t *testing.T) {
	env := mustEnv(EnvironmentConfig{Filters: []FilterDef{Filter("upcase", constant("env"))}})
	res, _ := renderFilters(t, env, "{{ x | upcase }}|{{ x | downcase }}")
	if res.Output != "env|a" {
		t.Errorf("environment filter: %q", res.Output)
	}
	res, _ = renderFilters(t, env, "{{ x | upcase }}", WithFilters(Filter("upcase", constant("first")), Filter("upcase", constant("last"))))
	if res.Output != "last" {
		t.Errorf("local filters: %q, want the last given", res.Output)
	}
	if res, _ := renderFilters(t, newEnvironment(), "{{ x | upcase }}"); res.Output != "A" {
		t.Errorf("standard filter: %q", res.Output)
	}
}

// An unknown filter passes its input through, or is an UndefinedFilter with
// strict filters.
func TestAnUnknownFilter(t *testing.T) {
	if res, _ := renderFilters(t, newEnvironment(), "{{ x | nope }}"); res.Output != "a" || len(res.Errors) != 0 {
		t.Errorf("lax: %q %v", res.Output, res.Errors)
	}
	res, _ := renderFilters(t, newEnvironment(), "{{ x | nope }}", WithStrictFilters())
	var undefined *UndefinedFilter
	if len(res.Errors) != 1 || !errors.As(res.Errors[0], &undefined) {
		t.Errorf("strict: errors %v, want an UndefinedFilter", res.Errors)
	}
}

// An argument too many, or the error a filter returns, gives nil, and is
// raised with strict filters; a missing argument is nil.
func TestFilterArgumentsAndErrors(t *testing.T) {
	if res, _ := renderFilters(t, newEnvironment(), "[{{ x | upcase: 1 }}]"); res.Output != "[]" {
		t.Errorf("an argument too many: %q", res.Output)
	}
	res, _ := renderFilters(t, newEnvironment(), "[{{ x | upcase: 1 }}]", WithStrictFilters())
	if len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), "wrong number of arguments (given 2, expected 1)") {
		t.Errorf("strict: errors %v", res.Errors)
	}
	failing := Filter("fail", func(*Context, interface{}, ...interface{}) (interface{}, error) {
		return "kept?", NewArgumentError("bad")
	})
	if res, _ := renderFilters(t, newEnvironment(), "[{{ x | fail }}]", WithFilters(failing)); res.Output != "[]" {
		t.Errorf("a failing filter: %q", res.Output)
	}
	// append requires its argument, as in Ruby, which raises (differential
	// case filters-errors-03): the error writes nothing; default does not
	if res, _ := renderFilters(t, newEnvironment(), "{{ x | append }}|{{ x | default }}"); res.Output != "|a" {
		t.Errorf("missing arguments: %q", res.Output)
	}
}

// A filter needs a name and a function.
func TestAnEnvironmentFilterIsChecked(t *testing.T) {
	if _, err := NewEnvironment(EnvironmentConfig{Filters: []FilterDef{Filter("", rawFilter)}}); err == nil {
		t.Error("a filter with no name was declared")
	}
	if _, err := NewEnvironment(EnvironmentConfig{Filters: []FilterDef{Filter("x", nil)}}); err == nil {
		t.Error("a filter with no function was declared")
	}
	if names := newEnvironment().FilterNames(); !sort.StringsAreSorted(names) || containsName(names, "x") {
		t.Errorf("FilterNames %v", names)
	}
}

// The standard filter table has a filter for each exported method of
// StandardFilters, named in snake_case, and nothing else. The names only are
// compared: a table whose calls no longer match the methods does not compile.
func TestTheStandardFilterTableFollowsStandardFilters(t *testing.T) {
	lowerUpper := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	acronym := regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	override := map[string]string{"TruncateWords": "truncatewords"}
	typ := reflect.TypeOf(&StandardFilters{})
	want := map[string]bool{}
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name
		snake, ok := override[name]
		if !ok {
			snake = strings.ToLower(acronym.ReplaceAllString(lowerUpper.ReplaceAllString(name, "${1}_${2}"), "${1}_${2}"))
		}
		want[snake] = true
		if standardFilters[snake] == nil {
			t.Errorf("StandardFilters.%s has no filter %q: run scripts/generate_filter_table.py", name, snake)
		}
	}
	for name := range standardFilters {
		if !want[name] {
			t.Errorf("the filter %q has no method of StandardFilters", name)
		}
	}
}

// An unknown filter gives its input back: it never reads a key of a drop, as
// in Ruby.
func TestAnUnknownFilterReadsNothingOfADrop(t *testing.T) {
	d := &keysDrop{}
	tmpl, err := ParseTemplate("[{{ d | a }}]", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, _ := tmpl.Render(gocontext.Background(), map[string]interface{}{"d": d})
	if res.Output != "[]" || d.calls != 0 {
		t.Errorf("got %q, %d reads of the drop, want %q and none", res.Output, d.calls, "[]")
	}
}

// A Go error a filter returns is, with strict filters, recorded as an
// internal error that keeps it as its cause, and the render goes on.
func TestAGoErrorOfAFilterIsAnInternalError(t *testing.T) {
	cause := errors.New("down")
	failing := Filter("fail", func(*Context, interface{}, ...interface{}) (interface{}, error) { return nil, cause })
	tmpl, err := ParseTemplate(`{{ x | fail }}|after`, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := tmpl.Render(gocontext.Background(), nil, WithFilters(failing), WithStrictFilters())
	if err != nil {
		t.Fatalf("the render ended: %v", err)
	}
	var internal *InternalError
	if len(res.Errors) != 1 || !errors.As(res.Errors[0], &internal) || !errors.Is(res.Errors[0], cause) {
		t.Errorf("errors %v, want an internal error caused by %v", res.Errors, cause)
	}
	if res.Output != "Liquid error: internal|after" {
		t.Errorf("output %q", res.Output)
	}
}

// An error of a filter, standard or not, writes nothing and is recorded; with
// strict filters it is raised, as an argument too many is.
func TestAFilterErrorIsRecorded(t *testing.T) {
	failing := Filter("fail", func(*Context, interface{}, ...interface{}) (interface{}, error) {
		return nil, NewArgumentError("bad")
	})
	src := `[{{ 1 | divided_by: 0 }}|{{ x | fail | default: "d" }}]`
	res, _ := renderFilters(t, newEnvironment(), src, WithFilters(failing))
	var zero *ZeroDivisionError
	var argument *ArgumentError
	if res.Output != "[|d]" || len(res.Errors) != 2 || !errors.As(res.Errors[0], &zero) || !errors.As(res.Errors[1], &argument) {
		t.Errorf("lax: %q %v, want %q and the two errors", res.Output, res.Errors, "[|d]")
	}
	res, _ = renderFilters(t, newEnvironment(), `[{{ 1 | divided_by: 0 }}]`, WithStrictFilters())
	if res.Output != "[Liquid error: divided by 0]" || len(res.Errors) != 1 {
		t.Errorf("strict: %q %v", res.Output, res.Errors)
	}
}

// A standard filter takes Ruby's default for an argument it is not given, and
// an argument given as nil as nil, as Ruby does: nil is an empty separator or
// ellipsis. (A length that is not an integer, nil included, keeps the input,
// where Ruby raises invalid integer.)
func TestAnArgumentGivenAsNilIsNotAMissingOne(t *testing.T) {
	tmpl, err := ParseTemplate(`{{ arr | join }}|{{ arr | join: sep }}|{{ long | truncate: 5 }}|{{ long | truncate: 5, nil }}|{{ long | truncatewords: 2 }}|{{ long | truncatewords: 2, nil }}|{{ "ab" | replace: "a" }}|{{ 1.6 | round }}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{"arr": []interface{}{"b", "a"}, "long": "0123456789 a b c"}
	res, _ := tmpl.Render(gocontext.Background(), data)
	want := "b a|ba|01...|01234|0123456789 a...|0123456789 a|b|2"
	if res.Output != want || len(res.Errors) != 0 {
		t.Errorf("got %q %v, want %q", res.Output, res.Errors, want)
	}
}

// A filter that takes the place of a standard one gets arguments of its
// own, which it may keep: only a standard filter shares the context's.
func TestAFilterKeepsItsArguments(t *testing.T) {
	var kept [][]interface{}
	keep := func(_ *Context, input interface{}, args ...interface{}) (interface{}, error) {
		kept = append(kept, args)
		return input, nil
	}
	for _, opts := range [][]RenderOption{{WithFilters(Filter("append", keep))}, nil} {
		env := newEnvironment()
		if opts == nil {
			env = mustEnv(EnvironmentConfig{Filters: []FilterDef{Filter("append", keep)}})
		}
		kept = nil
		if res, _ := renderFilters(t, env, `{{ x | append: "1" }}{{ x | upcase | append: "2" }}`, opts...); res.Output != "aA" {
			t.Fatalf("output %q", res.Output)
		}
		if len(kept) != 2 || kept[0][0] != "1" || kept[1][0] != "2" {
			t.Errorf("the arguments a filter kept changed: %v", kept)
		}
	}
}
