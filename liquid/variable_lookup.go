package liquid

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	variableLookupCommandMethods = []string{"size", "first", "last", "empty"}
)

// VariableLookup represents a variable lookup expression.
type VariableLookup struct {
	name         interface{}
	lookups      []interface{}
	commandFlags uint
	// indexFlags marks the lookups written after a dot with digits only
	// (items.0): a list reads them as an index, as Notifuse writes the
	// paths of its display conditions, where Ruby keeps the string "0".
	indexFlags uint
}

// parseVariableTokens tokenizes a variable expression, properly handling nested brackets.
// For example: "[list[settings.zero]]" returns ["[list[settings.zero]]"]
// And: "list[settings.zero]" returns ["list", "[settings.zero]"]
func parseVariableTokens(markup string) []string {
	var tokens []string
	i := 0

	for i < len(markup) {
		// Check if we're at the start of a bracket expression
		switch {
		case markup[i] == '[':
			// Find the matching closing bracket, handling nesting
			bracketCount := 1
			start := i
			i++
			for i < len(markup) && bracketCount > 0 {
				switch markup[i] {
				case '[':
					bracketCount++
				case ']':
					bracketCount--
				}
				i++
			}
			if bracketCount > 0 {
				// never closed: as Ruby's VariableParser, which does not
				// match here, the bracket is skipped and the scan goes on
				i = start + 1
				continue
			}
			tokens = append(tokens, markup[start:i])
		case (markup[i] >= 'a' && markup[i] <= 'z') ||
			(markup[i] >= 'A' && markup[i] <= 'Z') ||
			(markup[i] >= '0' && markup[i] <= '9') ||
			markup[i] == '_' || markup[i] == '-':
			// Identifier or number
			start := i
			for i < len(markup) && ((markup[i] >= 'a' && markup[i] <= 'z') ||
				(markup[i] >= 'A' && markup[i] <= 'Z') ||
				(markup[i] >= '0' && markup[i] <= '9') ||
				markup[i] == '_' || markup[i] == '-') {
				i++
			}
			// Check for optional '?' suffix
			if i < len(markup) && markup[i] == '?' {
				i++
			}
			tokens = append(tokens, markup[start:i])
		case markup[i] == '.':
			// Skip dots between identifiers (they're separators, not tokens)
			i++
		default:
			// Unknown character, skip it
			i++
		}
	}

	return tokens
}

// bracketDepth returns the deepest nesting of brackets in markup, outside
// quoted strings.
func bracketDepth(markup string) int {
	depth, deepest := 0, 0
	for i := 0; i < len(markup); i++ {
		switch markup[i] {
		case '\'', '"':
			if end := strings.IndexByte(markup[i+1:], markup[i]); end >= 0 {
				i += end + 1
			}
		case '[':
			depth++
			deepest = max(deepest, depth)
		case ']':
			depth = max(depth-1, 0)
		}
	}
	return deepest
}

// VariableLookupParse parses a markup string into a VariableLookup.
func VariableLookupParse(markup string, ss *StringScanner, cache map[string]interface{}) *VariableLookup {
	// Each level of brackets parses the markup inside it again: the depth
	// of blocks bounds them, so that the cost stays linear.
	if bracketDepth(markup) > blockMaxDepth {
		panic(NewStackLevelError("Nesting too deep"))
	}

	// Tokenize the variable expression, properly handling nested brackets
	matches := parseVariableTokens(markup)
	if len(matches) == 0 {
		return &VariableLookup{name: markup, lookups: []interface{}{}}
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
			} else if isCommandMethod(lookupStr) {
				// Mark as command method
				vl.commandFlags |= 1 << i
			} else if isDigits(lookupStr) {
				vl.indexFlags |= 1 << i
			}
		}
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
	// As Ruby's find_variable, whose raise_on_not_found is true: with strict
	// variables, a name that resolves to nothing is an UndefinedVariable
	obj := context.FindVariable(ToString(name, nil), true)

	for i, lookup := range vl.lookups {
		// Each step reads the liquid form of the object it reached, as Ruby
		// calls to_liquid on every result of a lookup
		obj = ToLiquid(obj)
		key := context.Evaluate(lookup)
		key = ToLiquidValue(key)

		// Ruby logic: Try to access as hash/array first
		// If object is a hash- or array-like object we look for the presence of the key
		// Fast path: Check for map[string]interface{} first (most common case)
		if m, ok := obj.(map[string]interface{}); ok {
			if k, ok := key.(string); ok {
				if value, exists := m[k]; exists {
					obj = context.mapValue(m, k, value) // a func is called, as Ruby's lookup_and_evaluate
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

		// Array index access. An Integer key reads the array as Ruby's
		// Array#[]: a negative index counts from the end, and an index out
		// of range gives nil, which strict variables do not raise. Any other
		// key, such as '1' or 1.5, is no index, as in Ruby.
		if arr, ok := obj.([]interface{}); ok {
			if idx, ok := vl.indexKey(i, key); ok {
				obj = indexOf(len(arr), idx, func(i int) interface{} { return context.itemOf(arr, i) })
				continue
			}
		} else if obj != nil {
			// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
			// This matches Ruby's duck-typing behavior: arrays respond to []
			v := reflect.ValueOf(obj)
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				if idx, ok := vl.indexKey(i, key); ok {
					obj = indexOf(v.Len(), idx, func(i int) interface{} { return itemAt(v, i) })
					continue
				}
			}

			// Map access with integer keys (e.g., map[int]interface{})
			// This handles maps where keys are not strings
			if v.Kind() == reflect.Map {
				// Try integer key first
				if idx, err := ToInteger(key); err == nil {
					intKey := reflect.ValueOf(idx)
					if intKey.Type().ConvertibleTo(v.Type().Key()) {
						mapValue := v.MapIndex(intKey.Convert(v.Type().Key()))
						if mapValue.IsValid() {
							obj = mapValue.Interface()
							continue
						}
					}
				}
				// Try key as-is for other map key types; a nil key finds
				// nothing, as in Ruby
				keyValue := reflect.ValueOf(key)
				// a key Go cannot hash (a map, a list) finds nothing
				if keyValue.IsValid() && keyValue.Comparable() && keyValue.Type().ConvertibleTo(v.Type().Key()) {
					mapValue := v.MapIndex(keyValue.Convert(v.Type().Key()))
					if mapValue.IsValid() {
						obj = mapValue.Interface()
						continue
					}
				}
			}
		}

		// Ruby logic: Some special cases. If the part wasn't in square brackets and
		// no key with the same name was found we interpret following calls
		// as commands and call them on the current object
		// (This is lines 67-71 in Ruby)
		if keyStr, ok := key.(string); ok {
			// Check for "to_liquid" property
			if keyStr == "to_liquid" {
				obj = ToLiquid(obj)
				continue
			}

			// Check if it's a command method AND the object responds to it
			if vl.LookupCommand(i) {
				// Command methods for arrays/slices (handle []interface{} first for fast path)
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
					case "empty":
						obj = len(arr) == 0
						continue
					}
				}
				// Use reflection to handle other slice types ([]int, []string, etc.)
				// This matches Ruby's behavior where arrays respond to .size, .first, .last, .empty
				if obj != nil {
					v := reflect.ValueOf(obj)
					if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
						switch keyStr {
						case "size":
							obj = v.Len()
							continue
						case "first":
							if v.Len() > 0 {
								obj = itemAt(v, 0)
								continue
							}
						case "last":
							if v.Len() > 0 {
								obj = itemAt(v, v.Len()-1)
								continue
							}
						case "empty":
							obj = v.Len() == 0
							continue
						}
					}
				}
				// size on a map without a "size" key is its number of keys (Ruby Hash#size)
				if keyStr == "size" && obj != nil && reflect.ValueOf(obj).Kind() == reflect.Map {
					obj = reflect.ValueOf(obj).Len()
					continue
				}
				// Command methods for strings
				if str, ok := obj.(string); ok {
					switch keyStr {
					case "size":
						obj = utf8.RuneCountInString(str)
						continue
					case "empty":
						obj = str == ""
						continue
					case "first": // its first character, as Shopify Liquid 5.14
						obj = firstChar(str)
						continue
					case "last":
						obj = lastChar(str)
						continue
					}
				}
			}

			// A drop gives its keys (forloop.last); a plain Go struct
			// exposes nothing, as a Ruby object that is not a Drop
			if g, ok := asDrop(obj); ok {
				obj = dropGet(context, g, keyStr)
				continue
			}
		} else if g, ok := asDrop(obj); ok {
			// A key that is not a string is a key the drop does not
			// declare, as Ruby's invoke_drop(0)
			obj = dropGet(context, g, ToString(key, nil))
			continue
		}

		// Not found
		if context.StrictVariables() {
			panic(NewUndefinedVariable("undefined variable " + ToString(key, nil)))
		}
		return nil
	}

	// A nil pointer is nil, in a condition as in the output, as Ruby's nil
	if isNilPointer(obj) {
		return nil
	}
	return ToLiquid(obj)
}

// ToString converts a value to string.
func ToString(v interface{}, ctx *Context) string {
	return ToS(v, nil)
}

// integerKey returns key as an index when it is an integer, as Ruby's
// key.is_a?(Integer) in a lookup of an array: a Go integer of any kind,
// named or not, or a float with no fraction, which is how a JSON number comes
// to Go. An index beyond the range of int is out of range of any list.
func integerKey(key interface{}) (int, bool) {
	switch k := key.(type) {
	case int:
		return k, true
	case float64:
		return integralFloat(k)
	}
	v := reflect.ValueOf(key)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return saturateInt64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return saturateUint64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return integralFloat(v.Float())
	}
	return 0, false
}

// integralFloat returns f as an index when it has no fraction.
func integralFloat(f float64) (int, bool) {
	if f != math.Trunc(f) || math.Abs(f) >= 1<<53 {
		return 0, false
	}
	return saturateInt64(int64(f)), true
}

// indexOf returns item idx of a list of n items, as Ruby's Array#[]: a
// negative index counts from the end, and one out of range gives nil.
func indexOf(n, idx int, item func(int) interface{}) interface{} {
	if idx < 0 {
		idx += n
	}
	if idx < 0 || idx >= n {
		return nil
	}
	return item(idx)
}

// indexKey returns key, the lookup at i, as an index of a list: an integer
// (integerKey), or digits written after a dot (indexFlags).
func (vl *VariableLookup) indexKey(i int, key interface{}) (int, bool) {
	if idx, ok := integerKey(key); ok {
		return idx, true
	}
	if s, ok := key.(string); ok && i < 64 && vl.indexFlags&(1<<i) != 0 {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return math.MaxInt, true // more digits than an int64: out of range
		}
		return saturateInt64(n), true
	}
	return 0, false
}

// isDigits reports whether s is made of ASCII digits only.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
