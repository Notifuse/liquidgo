package liquid

// Extensions provides to_liquid methods for various Go types.
// In Go, we can't monkey-patch types like in Ruby, so these are utility functions
// that can be used when needed. The actual conversion happens in the context
// when rendering values.

// ToLiquid converts a value to its liquid representation.
// This is a helper function that checks if the value implements ToLiquidValue
// or returns the value as-is.
func ToLiquid(obj interface{}) interface{} {
	if _, ok := obj.(string); ok {
		return obj // the most common value, inlined
	}
	return toLiquidMethod(obj)
}

// toLiquidMethod is ToLiquid for the values that may have the method, apart
// so that ToLiquid is inlined.
func toLiquidMethod(obj interface{}) interface{} {
	switch obj.(type) {
	case nil, int, float64, bool, map[string]interface{}, []interface{}:
		return obj // the common values have no method: no method lookup
	}
	if v, ok := obj.(interface {
		ToLiquid() interface{}
	}); ok {
		return v.ToLiquid()
	}
	return obj
}
