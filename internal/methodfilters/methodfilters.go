// Package methodfilters declares the exported methods of filter structs as
// liquid filters, as liquidgo called them before filters were declared: for
// the tests ported from Ruby, whose filters are modules of methods, and for
// the benchmark. The library declares its filters with liquid.Filter.
package methodfilters

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// Of returns the exported methods of each filter, a pointer to a struct, as
// filters named in snake_case (MoneyWithCurrency is money_with_currency). A
// call passes nil as the zero value of its parameter, pads the missing ones
// with zero values, and gives an ArgumentError for an argument too many; the
// error a method returns renders nothing.
func Of(filters ...interface{}) []liquid.FilterDef {
	var defs []liquid.FilterDef
	for _, f := range filters {
		v := reflect.ValueOf(f)
		t := v.Type()
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			method := v.Method(i)
			if method.Type().NumIn() < 1 {
				continue
			}
			defs = append(defs, liquid.Filter(snake(m.Name), call(method)))
		}
	}
	return defs
}

func call(method reflect.Value) liquid.FilterFunc {
	mt := method.Type()
	return func(_ *liquid.Context, input interface{}, args ...interface{}) (interface{}, error) {
		all := append([]interface{}{input}, args...)
		numIn := mt.NumIn()
		variadic := mt.IsVariadic()
		required := numIn
		if variadic {
			required--
		}
		if !variadic && len(all) > numIn {
			return nil, liquid.NewArgumentError(fmt.Sprintf("wrong number of arguments (given %d, expected %d)", len(all), numIn))
		}
		in := make([]reflect.Value, 0, len(all))
		for i, a := range all {
			pt := paramType(mt, i, variadic)
			if a == nil {
				in = append(in, reflect.Zero(pt))
			} else {
				in = append(in, reflect.ValueOf(a))
			}
		}
		for i := len(in); i < required; i++ {
			in = append(in, reflect.Zero(mt.In(i)))
		}
		out := method.Call(in)
		if len(out) == 0 {
			return nil, nil
		}
		return out[0].Interface(), nil
	}
}

func paramType(mt reflect.Type, i int, variadic bool) reflect.Type {
	if variadic && i >= mt.NumIn()-1 {
		return mt.In(mt.NumIn() - 1).Elem()
	}
	return mt.In(i)
}

var (
	lowerUpper = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	acronym    = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
)

func snake(name string) string {
	s := lowerUpper.ReplaceAllString(name, "${1}_${2}")
	s = acronym.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}
