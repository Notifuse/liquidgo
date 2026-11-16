package liquid

import "reflect"

// LiquidGetter is a value whose keys a template reads, as a Liquid::Drop in
// Ruby: {{ product.title }} reads the key title of product through
// LiquidGet, with the context of the render. It returns the value and
// whether the key exists: a missing key is nil, or an UndefinedDropMethod
// with strict variables. An error it returns is recorded, a Go error as an
// InternalError that unwraps to it. A nil pointer is nil, not a drop.
//
// No method or field of a Go value is found by reflection. Besides the keys
// of drops, a template reads maps, lists, numbers, strings, booleans, times
// and what ToLiquid gives; it writes the text of a String or Error method;
// it calls a func() interface{} held in a map, and, on any value, the
// methods Each(func(interface{})) and Count() int of a collection,
// LoadSlice, ToNumber and ToLiquidValue.
type LiquidGetter interface {
	LiquidGet(ctx *Context, key string) (interface{}, bool, error)
}

var liquidGetterType = reflect.TypeOf((*LiquidGetter)(nil)).Elem()

// asDrop returns obj as a drop, if a template may read its keys: a struct
// value whose pointer is a drop, as the items of a []T whose *T declares
// LiquidGet, is read through a copy.
func asDrop(obj interface{}) (LiquidGetter, bool) {
	if g, ok := obj.(LiquidGetter); ok {
		if isNilPointer(obj) {
			return nil, false
		}
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

// isDrop reports whether a template may read the keys of obj, without the
// copy asDrop makes.
func isDrop(obj interface{}) bool {
	if _, ok := obj.(LiquidGetter); ok {
		return !isNilPointer(obj)
	}
	if obj == nil {
		return false
	}
	t := reflect.TypeOf(obj)
	return t.Kind() == reflect.Struct && reflect.PointerTo(t).Implements(liquidGetterType)
}

// isNilPointer reports whether obj is a nil pointer: a nil drop is nil, as
// Ruby's nil.
func isNilPointer(obj interface{}) bool {
	switch obj.(type) {
	case nil, string, int, float64, bool, map[string]interface{}, []interface{}:
		return false // the common values: no reflection
	}
	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Ptr && v.IsNil()
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

// itemAt returns item i of v, a slice or an array. An item whose pointer is a
// drop is given by its address when it has one (the items of a slice), so that
// the keys of the drop read the item itself, and what a key changes stays.
func itemAt(v reflect.Value, i int) interface{} {
	e := v.Index(i)
	if e.Kind() == reflect.Struct && e.CanAddr() && reflect.PointerTo(e.Type()).Implements(liquidGetterType) {
		return e.Addr().Interface()
	}
	return e.Interface()
}
