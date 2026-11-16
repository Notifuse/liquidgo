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
	env := NewEnvironment()
	if err := env.RegisterFilter(Filter("upcase", constant("env"))); err != nil {
		t.Fatal(err)
	}
	res, _ := renderFilters(t, env, "{{ x | upcase }}|{{ x | downcase }}")
	if res.Output != "env|a" {
		t.Errorf("environment filter: %q", res.Output)
	}
	res, _ = renderFilters(t, env, "{{ x | upcase }}", WithFilters(Filter("upcase", constant("first")), Filter("upcase", constant("last"))))
	if res.Output != "last" {
		t.Errorf("local filters: %q, want the last given", res.Output)
	}
	if res, _ := renderFilters(t, NewEnvironment(), "{{ x | upcase }}"); res.Output != "A" {
		t.Errorf("standard filter: %q", res.Output)
	}
}

// An unknown filter passes its input through, or is an UndefinedFilter with
// strict filters.
func TestAnUnknownFilter(t *testing.T) {
	if res, _ := renderFilters(t, NewEnvironment(), "{{ x | nope }}"); res.Output != "a" || len(res.Errors) != 0 {
		t.Errorf("lax: %q %v", res.Output, res.Errors)
	}
	res, _ := renderFilters(t, NewEnvironment(), "{{ x | nope }}", WithStrictFilters())
	var undefined *UndefinedFilter
	if len(res.Errors) != 1 || !errors.As(res.Errors[0], &undefined) {
		t.Errorf("strict: errors %v, want an UndefinedFilter", res.Errors)
	}
}

// An argument too many, or the error a filter returns, gives nil, and is
// raised with strict filters; a missing argument is nil.
func TestFilterArgumentsAndErrors(t *testing.T) {
	if res, _ := renderFilters(t, NewEnvironment(), "[{{ x | upcase: 1 }}]"); res.Output != "[]" {
		t.Errorf("an argument too many: %q", res.Output)
	}
	res, _ := renderFilters(t, NewEnvironment(), "[{{ x | upcase: 1 }}]", WithStrictFilters())
	if len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), "wrong number of arguments (given 2, expected 1)") {
		t.Errorf("strict: errors %v", res.Errors)
	}
	failing := Filter("fail", func(*Context, interface{}, ...interface{}) (interface{}, error) {
		return "kept?", NewArgumentError("bad")
	})
	if res, _ := renderFilters(t, NewEnvironment(), "[{{ x | fail }}]", WithFilters(failing)); res.Output != "[]" {
		t.Errorf("a failing filter: %q", res.Output)
	}
	if res, _ := renderFilters(t, NewEnvironment(), "{{ x | append }}|{{ x | default }}"); res.Output != "a|a" {
		t.Errorf("missing arguments: %q", res.Output)
	}
}

// A filter needs a name and a function.
func TestRegisterFilterChecksTheFilter(t *testing.T) {
	env := NewEnvironment()
	if err := env.RegisterFilter(Filter("", rawFilter)); err == nil {
		t.Error("a filter with no name was registered")
	}
	if err := env.RegisterFilter(Filter("x", nil)); err == nil {
		t.Error("a filter with no function was registered")
	}
	if names := env.FilterNames(); !sort.StringsAreSorted(names) || containsName(names, "x") {
		t.Errorf("FilterNames %v", names)
	}
}

// The standard filter table has a filter for each exported method of
// StandardFilters, named in snake_case, and nothing else.
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
