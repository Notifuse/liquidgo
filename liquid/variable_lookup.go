package liquid

import (
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
		
		// Try to access as map/array
		if m, ok := obj.(map[string]interface{}); ok {
			if k, ok := key.(string); ok {
				if val, exists := m[k]; exists {
					obj = val
					continue
				}
			}
		}
		
		if arr, ok := obj.([]interface{}); ok {
			idx, _ := ToInteger(key)
			if idx >= 0 && idx < len(arr) {
				obj = arr[idx]
				continue
			}
		}
		
		// Try command method
		if vl.LookupCommand(i) {
			if _, ok := key.(string); ok {
				// Try to call method on object
				// For now, return nil
				return nil
			}
		}
		
		// Try drop method invocation
		if keyStr, ok := key.(string); ok {
			if IsInvokable(obj, keyStr) {
				dropResult := InvokeDropOn(obj, keyStr)
				// Even if result is nil, we found the method, so use it
				// (nil is a valid return value)
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

