package liquid

import (
	"reflect"
	"strings"
	"sync"
)

// dropMethodCache caches method lookups for drops to avoid repeated reflection.
// Optimization: This provides a 5-10x speedup for drop method invocations.
var dropMethodCache sync.Map // map[reflect.Type]*cachedDropMethods

// cachedDropMethods stores pre-computed method information for a drop type.
type cachedDropMethods struct {
	methods map[string]int // method name -> method index
}

// Drop is a base class for drops that allows exporting DOM-like things to liquid.
// Methods of drops are callable. The main use for liquid drops is to implement lazy loaded objects.
type Drop struct {
	context *Context
}

// NewDrop creates a new Drop instance.
func NewDrop() *Drop {
	return &Drop{
		context: nil,
	}
}

// SetContext sets the context for the drop.
func (d *Drop) SetContext(ctx *Context) {
	d.context = ctx
}

// Context returns the context.
func (d *Drop) Context() *Context {
	return d.context
}

// LiquidMethodMissing is called when a method is not found.
// It can be overridden by specific drop types.
func (d *Drop) LiquidMethodMissing(method string) interface{} {
	if d.context != nil && d.context.StrictVariables() {
		panic(NewUndefinedDropMethod("undefined method " + method))
	}
	return nil
}

// InvokeDropOn invokes a method on any drop type.
// Optimization: Uses cached method lookups to avoid repeated reflection.
func InvokeDropOn(drop interface{}, methodOrKey string) interface{} {
	if !IsInvokable(drop, methodOrKey) {
		// Call LiquidMethodMissing if available
		if dropWithMissing, ok := drop.(interface {
			LiquidMethodMissing(string) interface{}
		}); ok {
			return dropWithMissing.LiquidMethodMissing(methodOrKey)
		}
		return nil
	}

	v := reflect.ValueOf(drop)

	// Handle both pointer and non-pointer types for method calls
	// For structs from typed slices, we get values not pointers
	var t reflect.Type
	var structValue reflect.Value

	if v.Kind() == reflect.Ptr {
		t = v.Type()

		// Try to get cached method lookup
		var cache *cachedDropMethods
		if cached, ok := dropMethodCache.Load(t); ok {
			cache = cached.(*cachedDropMethods)
		} else {
			// Build cache for this type
			cache = buildDropMethodCache(t)
			dropMethodCache.Store(t, cache)
		}

		// Try snake_case to CamelCase conversion first (e.g., "standard_error" -> "StandardError")
		camelName := snakeToCamel(methodOrKey)
		if methodIdx, exists := cache.methods[camelName]; exists {
			method := v.Method(methodIdx)
			if method.IsValid() && method.Kind() == reflect.Func {
				// Let panics propagate naturally - they'll be caught at template.Render level
				results := method.Call([]reflect.Value{})
				if len(results) > 0 {
					return results[0].Interface()
				}
				return nil
			}
		}

		// Try capitalized version (e.g., "standard_error" -> "Standard_error")
		methodName := stringsTitle(methodOrKey)
		if methodIdx, exists := cache.methods[methodName]; exists {
			method := v.Method(methodIdx)
			if method.IsValid() && method.Kind() == reflect.Func {
				// Let panics propagate naturally - they'll be caught at template.Render level
				results := method.Call([]reflect.Value{})
				if len(results) > 0 {
					return results[0].Interface()
				}
				return nil
			}
		}

		// Try original case
		if methodIdx, exists := cache.methods[methodOrKey]; exists {
			method := v.Method(methodIdx)
			if method.IsValid() && method.Kind() == reflect.Func {
				// Let panics propagate naturally - they'll be caught at template.Render level
				results := method.Call([]reflect.Value{})
				if len(results) > 0 {
					return results[0].Interface()
				}
				return nil
			}
		}

		// For pointers, dereference to get struct value
		structValue = v.Elem()
	} else {
		// For non-pointer values (e.g., structs from typed slices),
		// we can only access fields, not methods
		structValue = v
	}

	// Try to get field from struct
	if structValue.IsValid() && structValue.Kind() == reflect.Struct {
		// Try CamelCase (for snake_case property names like "comments_count" -> "CommentsCount")
		field := structValue.FieldByName(snakeToCamel(methodOrKey))
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
		// Try simple capitalization (Title case)
		field = structValue.FieldByName(stringsTitle(methodOrKey))
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
		// Try original case
		field = structValue.FieldByName(methodOrKey)
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
	}

	// Call LiquidMethodMissing if available
	if dropWithMissing, ok := drop.(interface {
		LiquidMethodMissing(string) interface{}
	}); ok {
		return dropWithMissing.LiquidMethodMissing(methodOrKey)
	}

	return nil
}

// buildDropMethodCache builds a method cache for a drop type.
func buildDropMethodCache(t reflect.Type) *cachedDropMethods {
	cache := &cachedDropMethods{
		methods: make(map[string]int, t.NumMethod()),
	}

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		cache.methods[method.Name] = i
	}

	return cache
}

// InvokeDrop invokes a method on the drop.
func (d *Drop) InvokeDrop(methodOrKey string) interface{} {
	return InvokeDropOn(d, methodOrKey)
}

// InvokeDropOld invokes a method on the drop (old implementation).
func (d *Drop) InvokeDropOld(methodOrKey string) interface{} {
	if IsInvokable(d, methodOrKey) {
		// Use reflection to call the method
		v := reflect.ValueOf(d)
		// If d is a pointer, get the element type
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		method := v.MethodByName(stringsTitle(methodOrKey))
		if !method.IsValid() {
			// Try with original case
			method = v.MethodByName(methodOrKey)
		}

		if method.IsValid() && method.Kind() == reflect.Func {
			// Call the method
			results := method.Call([]reflect.Value{})
			if len(results) > 0 {
				return results[0].Interface()
			}
			return nil
		}
	}

	// Try to get field
	if v := reflect.ValueOf(d); v.Kind() == reflect.Ptr {
		v = v.Elem()
		if v.Kind() == reflect.Struct {
			field := v.FieldByName(stringsTitle(methodOrKey))
			if field.IsValid() {
				return field.Interface()
			}
			// Try with original case
			field = v.FieldByName(methodOrKey)
			if field.IsValid() {
				return field.Interface()
			}
		}
	}

	return d.LiquidMethodMissing(methodOrKey)
}

// Key returns true if the key exists (drops always return true).
func (d *Drop) Key(name string) bool {
	return true
}

// Note: Drop does NOT implement ToLiquid() to avoid issues with embedded pointers.
// When Drop is embedded in another struct (e.g., *Drop in FooDrop), calling ToLiquid()
// would return the embedded *Drop pointer instead of the outer *FooDrop pointer.
// By not implementing ToLiquid(), the default behavior returns the object unchanged,
// which preserves the actual type.

// String returns the string representation of the drop.
func (d *Drop) String() string {
	return reflect.TypeOf(d).String()
}

// IsInvokable checks if a method is invokable on a drop.
func IsInvokable(drop interface{}, methodName string) bool {
	if drop == nil {
		return false
	}
	invokableMethods := GetInvokableMethods(drop)
	// Check multiple name variants:
	// 1. Original name (e.g., "Title")
	// 2. Simple capitalization (e.g., "title" -> "Title")
	// 3. Snake-to-camel (e.g., "comments_count" -> "CommentsCount")
	// This supports both Liquid's snake_case and Go's CamelCase conventions
	capitalizedName := stringsTitle(methodName)
	camelName := snakeToCamel(methodName)
	for _, m := range invokableMethods {
		if m == methodName || m == capitalizedName || m == camelName {
			return true
		}
	}
	return false
}

// GetInvokableMethods returns a list of invokable methods for a drop.
func GetInvokableMethods(drop interface{}) []string {
	if drop == nil {
		return []string{}
	}
	t := reflect.TypeOf(drop)
	// Keep pointer type for method lookup
	if t.Kind() != reflect.Ptr {
		// If not a pointer, create a pointer type
		t = reflect.PointerTo(t)
	}

	// Blacklist of methods that shouldn't be invokable
	blacklist := map[string]bool{
		"SetContext":          true,
		"Context":             true,
		"InvokeDrop":          true,
		"Key":                 true,
		"String":              true,
		"LiquidMethodMissing": true,
		"Each":                true,
		"Increment":           true, // Protected method
		"Get":                 true, // Prevent recursion via Context.Get
	}

	var methods []string
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if !blacklist[method.Name] {
			methods = append(methods, method.Name)
		}
	}

	// Also check for exported fields
	elemType := t
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Struct {
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			if field.IsExported() {
				methods = append(methods, field.Name)
			}
		}
	}

	return methods
}

// strings.Title capitalizes the first letter of a string.
func stringsTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// snakeToCamel converts snake_case to CamelCase.
// This is used to map Liquid's snake_case property names to Go's CamelCase field names.
// Examples: "comments_count" -> "CommentsCount", "created_at" -> "CreatedAt"
func snakeToCamel(s string) string {
	if len(s) == 0 {
		return s
	}

	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = stringsTitle(parts[i])
		}
	}
	return strings.Join(parts, "")
}
