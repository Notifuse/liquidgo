package liquid

import "reflect"

// LiquidGetter is a value whose keys a template reads, as a Liquid::Drop in
// Ruby: {{ product.title }} reads the key title of product through
// LiquidGet, with the context of the render. It returns the value and
// whether the key exists; an error is raised in the render. A template reads
// nothing else of a Go value but maps, lists, numbers, strings, times and
// what ToLiquid gives: no method or field is found by reflection.
type LiquidGetter interface {
	LiquidGet(ctx *Context, key string) (interface{}, bool, error)
}

var liquidGetterType = reflect.TypeOf((*LiquidGetter)(nil)).Elem()

// asDrop returns obj as a drop, if a template may read its keys: a struct
// value whose pointer is a drop, as the items of a []T whose *T declares
// LiquidGet, is read through a copy.
func asDrop(obj interface{}) (LiquidGetter, bool) {
	if g, ok := obj.(LiquidGetter); ok {
		return g, true
	}
	if obj == nil {
		return nil, false
	}
	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Struct || !reflect.PointerTo(t).Implements(liquidGetterType) {
		return nil, false
	}
	p := reflect.New(t)
	p.Elem().Set(reflect.ValueOf(obj))
	return p.Interface().(LiquidGetter), true
}

// isDrop reports whether a template may read the keys of obj.
func isDrop(obj interface{}) bool {
	if _, ok := obj.(LiquidGetter); ok {
		return true
	}
	if obj == nil {
		return false
	}
	t := reflect.TypeOf(obj)
	return t.Kind() == reflect.Struct && reflect.PointerTo(t).Implements(liquidGetterType)
}

// dropGet reads key of g for the render of ctx. A key g does not have is
// nil, or with strict variables an UndefinedDropMethod, as Ruby's
// liquid_method_missing; an error g returns is raised.
func dropGet(ctx *Context, g LiquidGetter, key string) interface{} {
	value, found, err := g.LiquidGet(ctx, key)
	if err != nil {
		panic(asLiquidError(err))
	}
	if !found {
		if ctx != nil && ctx.StrictVariables() {
			panic(NewUndefinedDropMethod("undefined method " + key))
		}
		return nil
	}
	return value
}
