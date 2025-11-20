package liquid

import (
	"reflect"
	"strings"
	"sync"
)

var (
	variableLookupCommandMethods = []string{"size", "first", "last"}
)

// globalVariableLookupCache provides thread-safe caching of parsed variable lookups.
// Optimization: Avoids re-parsing the same variable lookups repeatedly.
var globalVariableLookupCache sync.Map // map[string]*VariableLookup

// VariableLookup represents a variable lookup expression.
type VariableLookup struct {
	name         interface{}
	lookups      []interface{}
	commandFlags uint
}

// VariableLookupParse parses a markup string into a VariableLookup.
// Optimization: Uses global cache for better performance across templates.
func VariableLookupParse(markup string, ss *StringScanner, cache map[string]interface{}) *VariableLookup {
	// Try global cache first for simple variable lookups (no dynamic parts)
	// We can only cache if the markup doesn't contain dynamic expressions in brackets
	canCache := !strings.Contains(markup, "[")

	if canCache {
		if cached, ok := globalVariableLookupCache.Load(markup); ok {
			return cached.(*VariableLookup)
		}
	}

	// Scan for variable parts using VariableParser pattern
	// This matches [brackets] or identifier? patterns
	matches := VariableParser.FindAllString(markup, -1)
	if len(matches) == 0 {
		vl := &VariableLookup{name: markup, lookups: []interface{}{}}
		if canCache {
			globalVariableLookupCache.Store(markup, vl)
		}
		return vl
	}

	lookups := make([]interface{}, len(matches))
	for i, match := range matches {
		lookups[i] = match
	}

	name := lookups[0]
	lookups = lookups[1:]

	// Parse name if it's in brackets
	if nameStr, ok := name.(string); ok {
		if strings.HasPrefix(nameStr, "[") && strings.HasSuffix(nameStr, "]") {
			name = Parse(nameStr[1:len(nameStr)-1], ss, cache)
			canCache = false // Can't cache variable lookups with dynamic names
		}
	}

	vl := &VariableLookup{
		name:         name,
		lookups:      lookups,
		commandFlags: 0,
	}

	// Process lookups
	for i, lookup := range vl.lookups {
		if lookupStr, ok := lookup.(string); ok {
			if strings.HasPrefix(lookupStr, "[") && strings.HasSuffix(lookupStr, "]") {
				// Parse bracket expression
				vl.lookups[i] = Parse(lookupStr[1:len(lookupStr)-1], ss, cache)
				canCache = false // Can't cache variable lookups with dynamic expressions
			} else if isCommandMethod(lookupStr) {
				// Mark as command method
				vl.commandFlags |= 1 << i
			}
		}
	}

	// Cache the result if it's cacheable (no dynamic parts)
	if canCache {
		globalVariableLookupCache.Store(markup, vl)
	}

	return vl
}

func isCommandMethod(method string) bool {
	for _, cmd := range variableLookupCommandMethods {
		if cmd == method {
			return true
		}
	}
	return false
}

// tryMapAccess attempts to access a map value using reflection.
// This handles custom map type aliases (e.g., type MapOfAny map[string]any)
// that don't match the concrete type map[string]interface{}.
// Returns (value, true) if successful, (nil, false) if not a map or key not found.
func tryMapAccess(obj interface{}, key string) (interface{}, bool) {
	if obj == nil {
		return nil, false
	}

	v := reflect.ValueOf(obj)

	// Check if it's a map type
	if v.Kind() != reflect.Map {
		return nil, false
	}

	// Get the map's key type
	mapKeyType := v.Type().Key()

	// Try to convert our string key to the map's key type
	var keyValue reflect.Value

	switch mapKeyType.Kind() {
	case reflect.String:
		keyValue = reflect.ValueOf(key)
	case reflect.Interface:
		// For map[interface{}]interface{} or map[any]any
		keyValue = reflect.ValueOf(key)
	default:
		// Key type is not string or interface{}, can't look up
		return nil, false
	}

	// Look up the value in the map
	mapValue := v.MapIndex(keyValue)
	if !mapValue.IsValid() {
		// Key not found in map
		return nil, false
	}

	return mapValue.Interface(), true
}

// LookupCommand returns true if the lookup at the given index is a command method.
func (vl *VariableLookup) LookupCommand(lookupIndex int) bool {
	return (vl.commandFlags & (1 << lookupIndex)) != 0
}

// Name returns the variable name.
func (vl *VariableLookup) Name() interface{} {
	return vl.name
}

// Lookups returns the lookups.
func (vl *VariableLookup) Lookups() []interface{} {
	return vl.lookups
}

// Evaluate evaluates the variable lookup in the given context.
func (vl *VariableLookup) Evaluate(context *Context) interface{} {
	name := context.Evaluate(vl.name)
	obj := context.FindVariable(ToString(name, nil), false)

	for i, lookup := range vl.lookups {
		key := context.Evaluate(lookup)
		key = ToLiquidValue(key)

		// Ruby logic: Try to access as hash/array first
		// If object is a hash- or array-like object we look for the presence of the key
		// Fast path: Check for map[string]interface{} first (most common case)
		if m, ok := obj.(map[string]interface{}); ok {
			if k, ok := key.(string); ok {
				if val, exists := m[k]; exists {
					obj = val
					continue
				}
			}
		}

		// Fallback: Use reflection for custom map types (e.g., type MapOfAny map[string]any)
		// This matches Ruby's duck-typing behavior: respond_to?(:[])
		if k, ok := key.(string); ok {
			if val, found := tryMapAccess(obj, k); found {
				obj = val
				continue
			}
		}

		// Array index access
		if arr, ok := obj.([]interface{}); ok {
			if idx, err := ToInteger(key); err == nil {
				if idx >= 0 && idx < len(arr) {
					obj = arr[idx]
					continue
				}
			}
		}

		// Ruby logic: Some special cases. If the part wasn't in square brackets and
		// no key with the same name was found we interpret following calls
		// as commands and call them on the current object
		// (This is lines 67-71 in Ruby)
		if keyStr, ok := key.(string); ok {
			// Check if it's a command method AND the object responds to it
			if vl.LookupCommand(i) {
				// Command methods for arrays
				if arr, ok := obj.([]interface{}); ok {
					switch keyStr {
					case "size":
						obj = len(arr)
						continue
					case "first":
						if len(arr) > 0 {
							obj = arr[0]
							continue
						}
					case "last":
						if len(arr) > 0 {
							obj = arr[len(arr)-1]
							continue
						}
					}
				}
				// Command methods for strings
				if str, ok := obj.(string); ok && keyStr == "size" {
					obj = len(str)
					continue
				}
			}

			// Try drop method invocation (for drops like forloop.last)
			// This handles objects that respond to methods
			if IsInvokable(obj, keyStr) {
				dropResult := InvokeDropOn(obj, keyStr)
				// Even if result is nil, we found the method, so use it
				obj = dropResult
				continue
			}
		}

		// Not found
		if context.StrictVariables() {
			panic(NewUndefinedVariable("undefined variable " + ToString(key, nil)))
		}
		return nil
	}

	return ToLiquid(obj)
}

// ToString converts a value to string.
func ToString(v interface{}, ctx *Context) string {
	return ToS(v, nil)
}
