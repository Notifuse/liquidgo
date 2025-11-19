package liquid

// Extensions provides to_liquid methods for various Go types.
// In Go, we can't monkey-patch types like in Ruby, so these are utility functions
// that can be used when needed. The actual conversion happens in the context
// when rendering values.

// ToLiquid converts a value to its liquid representation.
// This is a helper function that checks if the value implements ToLiquidValue
// or returns the value as-is.
func ToLiquid(obj interface{}) interface{} {
	if toLiquid, ok := obj.(interface {
		ToLiquid() interface{}
	}); ok {
		return toLiquid.ToLiquid()
	}
	return obj
}
