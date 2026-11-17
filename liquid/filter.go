package liquid

import (
	"fmt"
	"regexp"
	"sort"
)

// FilterFunc is a filter: it gets the context of the render, the value it is
// applied to and its arguments (keyword arguments as a last
// map[string]interface{}), and returns the result. An error gives nil and is
// recorded in Result.Errors, without being written; with strict filters it is
// raised. A Go error that is not a Liquid error is recorded as an
// InternalError that unwraps to it.
type FilterFunc func(ctx *Context, input interface{}, args ...interface{}) (interface{}, error)

// FilterDef declares a filter: its name in templates and its function.
type FilterDef struct {
	name string
	fn   FilterFunc
}

// Filter declares the filter name, which templates write {{ x | name }}.
func Filter(name string, fn FilterFunc) FilterDef {
	return FilterDef{name: name, fn: fn}
}

// Name returns the name of the filter.
func (f FilterDef) Name() string { return f.name }

// check returns an error for a filter with no name or no function.
func (f FilterDef) check() error {
	if f.fn == nil || !filterName.MatchString(f.name) {
		return fmt.Errorf("liquid: a filter needs a name a template can write and a function (got %q)", f.name)
	}
	return nil
}

// filterName is a name a template can write after a pipe, the identifier of
// the lexer.
var filterName = regexp.MustCompile(`^[a-zA-Z_][\w-]*\??$`)

// filterArg returns argument i of args, nil when it is missing.
func filterArg(args []interface{}, i int) interface{} {
	if i < len(args) {
		return args[i]
	}
	return nil
}

// tooManyFilterArgs is the error of a filter given more arguments than it
// takes, as Ruby's ArgumentError.
func tooManyFilterArgs(given, expected int) error {
	return NewArgumentError(fmt.Sprintf("wrong number of arguments (given %d, expected %d)", given, expected))
}

// tooFewFilterArgs is Ruby's ArgumentError for a filter given fewer
// arguments than it requires, the input included: "expected 2", or "expected
// 2..3" when it takes optional ones.
func tooFewFilterArgs(given, required, most int) error {
	expected := fmt.Sprint(required)
	if most > required {
		expected = fmt.Sprintf("%d..%d", required, most)
	}
	return NewArgumentError(fmt.Sprintf("wrong number of arguments (given %d, expected %s)", given, expected))
}

// filterNames returns the names of the filters of the table, sorted.
func filterNames(tables ...map[string]FilterFunc) []string {
	seen := map[string]bool{}
	var names []string
	for _, t := range tables {
		for name := range t {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names
}
