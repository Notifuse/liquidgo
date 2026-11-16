package liquid

import (
	"fmt"
	"sort"
)

// FilterFunc is a filter: it gets the context of the render, the value it is
// applied to and its arguments, and returns the result. An error is raised
// with strict filters, and gives nil otherwise.
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
	if f.name == "" || f.fn == nil {
		return fmt.Errorf("liquid: a filter needs a name and a function (got %q)", f.name)
	}
	return nil
}

// filterArgs returns the first n arguments of args, nil for the missing ones.
func filterArgs(args []interface{}, n int) []interface{} {
	if len(args) >= n {
		return args
	}
	padded := make([]interface{}, n)
	copy(padded, args)
	return padded
}

// tooManyFilterArgs is the error of a filter given more arguments than it
// takes, as Ruby's ArgumentError.
func tooManyFilterArgs(given, expected int) error {
	return NewArgumentError(fmt.Sprintf("wrong number of arguments (given %d, expected %d)", given, expected))
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
