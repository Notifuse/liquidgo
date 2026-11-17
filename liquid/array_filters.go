package liquid

import (
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Join joins the items of the input with a separator, as Ruby's
// InputIterator#join: nested arrays are flattened, a range gives its
// integers, and any other input is a single item.
func (sf *StandardFilters) Join(input interface{}, separator interface{}) string {
	sep := " "
	if separator != nil {
		sep = ToS(separator, nil)
	}
	if r, ok := input.(*Range); ok {
		return sf.joinRange(r, sep)
	}
	if arr, ok := input.([]interface{}); ok && !hasNestedList(arr) {
		return sf.joinItems(arr, sep)
	}
	iter := NewInputIterator(input, sf.context)
	parts := make([]string, 0, len(iter.input))
	iter.Each(func(item interface{}) {
		parts = append(parts, ToS(item, nil))
	})
	sf.checkLength(joinedLength(parts, sep))
	return strings.Join(parts, sep)
}

// joinRange joins the integers of a range without building them first, and
// stops at the render length limit, or past maxRangeItems integers.
// joinItems joins the items of a list without a nested list, as Join does,
// writing each item once, within the render length limit.
func (sf *StandardFilters) joinItems(arr []interface{}, sep string) string {
	var b strings.Builder
	if n, ok := joinedStrings(arr, sep); ok {
		sf.checkLength(n)
		b.Grow(n)
	}
	for i, item := range arr {
		s := ToS(ToLiquidValue(ToLiquid(item)), nil)
		n := len(s)
		if i > 0 {
			n += len(sep)
		}
		sf.checkLength(b.Len() + n)
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(s)
	}
	return b.String()
}

// joinedStrings returns the length of arr joined by sep when every item is a
// string, so that the result is made at once, saturated.
func joinedStrings(arr []interface{}, sep string) (int, bool) {
	total := 0
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return 0, false
		}
		if total > math.MaxInt-len(s) {
			return math.MaxInt, true
		}
		total += len(s)
	}
	if len(arr) > 1 && len(sep) > 0 {
		gaps := len(arr) - 1
		if gaps > (math.MaxInt-total)/len(sep) {
			return math.MaxInt, true
		}
		total += gaps * len(sep)
	}
	return total, true
}

// hasNestedList reports whether arr holds a list, which Join flattens.
func hasNestedList(arr []interface{}) bool {
	for _, item := range arr {
		if _, ok := item.([]interface{}); ok {
			return true
		}
	}
	return false
}

func (sf *StandardFilters) joinRange(r *Range, sep string) string {
	n := r.Len()
	if n > maxRangeItems {
		sf.checkRangeItems(n)
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(strconv.Itoa(r.Start + i))
		sf.checkLength(b.Len())
	}
	return b.String()
}

// Reverse reverses the order of items in an array.
// Mirrors Ruby's reverse from standardfilters.rb:523
func (sf *StandardFilters) Reverse(input interface{}) []interface{} {
	iter := NewInputIterator(input, sf.context)
	return iter.Reverse()
}

// Sort sorts items in an array.
// Mirrors Ruby's sort from standardfilters.rb:378
func (sf *StandardFilters) Sort(input interface{}, property interface{}) (interface{}, error) {
	return sf.sortBy(input, property, nilSafeCompare)
}

// SortNatural sorts items case-insensitively.
// Mirrors Ruby's sort_natural from standardfilters.rb:407
func (sf *StandardFilters) SortNatural(input interface{}, property interface{}) (interface{}, error) {
	return sf.sortBy(input, property, nilSafeCasecmp)
}

// sortBy sorts the items of input with compare, by a property when one is
// given. As Ruby, sorting by a property gives nil unless every item has a
// [] method, and an integer item raises the property error once two items
// are compared.
func (sf *StandardFilters) sortBy(input interface{}, property interface{}, compare func(a, b interface{}) int) (interface{}, error) {
	arr := NewInputIterator(input, sf.context).ToArray()
	if len(arr) == 0 {
		return []interface{}{}, nil
	}
	sorted := make([]interface{}, len(arr))
	copy(sorted, arr)
	if property == nil {
		sort.SliceStable(sorted, func(i, j int) bool {
			return compare(sorted[i], sorted[j]) < 0
		})
		return sorted, nil
	}

	keys := make(map[int]interface{}, len(arr))
	failed := false
	for i, item := range arr {
		key, kind := indexItem(sf.context, item, property)
		if kind == noIndex {
			return nil, nil
		}
		failed = failed || kind == indexError
		keys[i] = key
	}
	if failed && len(arr) > 1 {
		return nil, raisePropertyError(property)
	}
	order := make([]int, len(arr))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return compare(keys[order[i]], keys[order[j]]) < 0
	})
	for i, k := range order {
		sorted[i] = arr[k]
	}
	return sorted, nil
}

// Uniq removes duplicate items from an array.
// Mirrors Ruby's uniq from standardfilters.rb:497
func (sf *StandardFilters) Uniq(input interface{}, property interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}, nil
	}

	if property == nil {
		return iter.Uniq(nil), nil
	}

	// Uniq by property: as Ruby, an item without [] gives nil, and a single
	// item is kept without reading its property, as Array#uniq
	items := iter.ToArray()
	if len(items) == 1 {
		return items, nil
	}
	seen := make(map[string]bool)
	result := make([]interface{}, 0)
	for _, item := range items {
		key, kind := indexItem(sf.context, ToLiquidValue(item), property)
		switch kind {
		case indexError:
			return nil, raisePropertyError(property)
		case noIndex:
			return nil, nil
		}
		keyStr := ""
		if key != nil {
			keyStr = ToS(key, nil)
		}
		if !seen[keyStr] {
			seen[keyStr] = true
			result = append(result, item)
		}
	}
	return result, nil
}

// Compact removes nil items from an array.
// Mirrors Ruby's compact from standardfilters.rb:557
func (sf *StandardFilters) Compact(input interface{}, property interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return []interface{}{}, nil
	}

	if property == nil {
		return iter.Compact(), nil
	}

	// Compact by property - remove items where property is nil; as Ruby, an
	// item without [] gives nil
	result := make([]interface{}, 0)
	for _, item := range iter.ToArray() {
		val, kind := indexItem(sf.context, item, property)
		switch kind {
		case indexError:
			return nil, raisePropertyError(property)
		case noIndex:
			return nil, nil
		}
		if val != nil {
			result = append(result, item)
		}
	}

	return result, nil
}

// Map extracts property values from array items.
// Mirrors Ruby's map from standardfilters.rb:535
func (sf *StandardFilters) Map(input interface{}, property interface{}) (interface{}, error) {
	if property == nil {
		return nil, raisePropertyError(property)
	}

	iter := NewInputIterator(input, sf.context)
	propStr := ToS(property, nil)
	result := make([]interface{}, 0)

	for _, item := range iter.ToArray() {
		// Special case for "to_liquid"
		if propStr == "to_liquid" {
			result = append(result, item)
			continue
		}
		val, kind := indexItem(sf.context, item, property)
		if kind == indexError {
			return nil, raisePropertyError(property)
		}
		result = append(result, val)
	}

	return result, nil
}

// Where filters array items by property value.
// Mirrors Ruby's where from standardfilters.rb:434
func (sf *StandardFilters) Where(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	result := make([]interface{}, 0)
	return sf.filterArray(input, property, targetValue, []interface{}{}, func(i int, item interface{}, match bool) bool {
		if match {
			result = append(result, item)
		}
		return false
	}, func() interface{} { return result })
}

// Reject filters out array items by property value.
// Mirrors Ruby's reject from standardfilters.rb:447
func (sf *StandardFilters) Reject(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	result := make([]interface{}, 0)
	return sf.filterArray(input, property, targetValue, []interface{}{}, func(i int, item interface{}, match bool) bool {
		if !match {
			result = append(result, item)
		}
		return false
	}, func() interface{} { return result })
}

// Has tests if any array item has a property value.
// Mirrors Ruby's has from standardfilters.rb:460
func (sf *StandardFilters) Has(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	found := false
	return sf.filterArray(input, property, targetValue, false, func(i int, item interface{}, match bool) bool {
		found = match
		return match
	}, func() interface{} { return found })
}

// Find returns the first array item with a property value.
// Mirrors Ruby's find from standardfilters.rb:473
func (sf *StandardFilters) Find(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	var found interface{}
	return sf.filterArray(input, property, targetValue, nil, func(i int, item interface{}, match bool) bool {
		if match {
			found = item
		}
		return match
	}, func() interface{} { return found })
}

// FindIndex returns the index of the first array item with a property value.
// Mirrors Ruby's find_index from standardfilters.rb:486
func (sf *StandardFilters) FindIndex(input interface{}, property interface{}, targetValue interface{}) (interface{}, error) {
	var found interface{}
	return sf.filterArray(input, property, targetValue, nil, func(i int, item interface{}, match bool) bool {
		if match {
			found = i
		}
		return match
	}, func() interface{} { return found })
}

// filterArray is Ruby's filter_array: it gives each item of input to visit,
// in order, with whether its property is truthy, or equal to targetValue
// when one is given, until visit returns true, and then returns result. An
// empty input gives empty. As Ruby, an integer item raises the property
// error, and an item without [] makes the filter give nil.
func (sf *StandardFilters) filterArray(input, property, targetValue, empty interface{}, visit func(i int, item interface{}, match bool) bool, result func() interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)
	if iter.Empty() {
		return empty, nil
	}
	if property == nil {
		return empty, nil
	}
	for i, item := range iter.ToArray() {
		val, kind := indexItem(sf.context, item, property)
		switch kind {
		case indexError:
			return nil, raisePropertyError(property)
		case noIndex:
			return nil, nil
		}
		// Truthiness and equality stay liquidgo's: an empty string or
		// list is falsy, and a string matches the number it writes
		var match bool
		if targetValue == nil {
			match = isTruthy(val)
		} else {
			match = valuesEqual(val, targetValue)
		}
		if visit(i, item, match) {
			break
		}
	}
	return result(), nil
}

// Concat concatenates two arrays.
// Mirrors Ruby's concat from standardfilters.rb:682
func (sf *StandardFilters) Concat(input interface{}, array interface{}) (interface{}, error) {
	// Validate that second argument is an array
	isArray := false
	if _, ok := array.([]interface{}); ok {
		isArray = true
	} else if array != nil {
		// Reflection fallback for typed slices ([]BlogPost, []string, []int, etc.)
		v := reflect.ValueOf(array)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			isArray = true
		}
	}

	if !isArray {
		return nil, NewArgumentError("concat filter requires an array argument")
	}

	iter := NewInputIterator(input, sf.context)
	sf.checkItems(len(iter.input) + reflectLen(array))
	return iter.Concat(array), nil
}

// Sum returns the sum of array elements.
// Mirrors Ruby's sum from standardfilters.rb:953
func (sf *StandardFilters) Sum(input interface{}, property interface{}) (interface{}, error) {
	iter := NewInputIterator(input, sf.context)

	if iter.Empty() {
		return 0, nil
	}

	sum := integerOf(0)
	for _, item := range iter.ToArray() {
		if property != nil {
			// As Ruby, an item without [] adds 0
			val, kind := indexItem(sf.context, item, property)
			if kind == indexError {
				return nil, raisePropertyError(property)
			}
			item = val
		}
		sum = arithmetic(sum, toLiquidNumber(item), '+')
	}

	return sum.value(), nil
}

// Helper functions for array filters

// indexKind says how Ruby's item[property] went for an item.
type indexKind int

const (
	// indexed: the item has a [] method that gave a value.
	indexed indexKind = iota
	// noIndex: the item has no [] method, as a float, a boolean or nil.
	noIndex
	// indexError: the item's [] raised a TypeError, as an integer's for a
	// string property.
	indexError
)

// indexItem reads property on item as Ruby's item[property] in the array
// filters. A hash or a drop gives its property. A string gives the property
// when it contains it, as String#[], or its character at an integer
// property. An integer gives its bit at an integer property and raises a
// TypeError for any other property, as Integer#[].
func indexItem(ctx *Context, item, property interface{}) (interface{}, indexKind) {
	// Ruby's array filters read each item's to_liquid
	item = ToLiquid(item)
	switch v := item.(type) {
	case nil, bool, float32, float64:
		return nil, noIndex
	case string:
		if sub, ok := property.(string); ok {
			if strings.Contains(v, sub) {
				return sub, indexed
			}
			return nil, indexed
		}
		if n, ok := integerItem(property); ok {
			runes := []rune(v)
			if n < 0 {
				n += int64(len(runes))
			}
			if n < 0 || n >= int64(len(runes)) {
				return nil, indexed
			}
			return string(runes[n]), indexed
		}
		return nil, indexError
	}
	if n, ok := integerItem(item); ok {
		bit, ok := integerItem(property)
		if !ok {
			return nil, indexError
		}
		switch {
		case bit >= 64:
			if n < 0 {
				return 1, indexed
			}
			return 0, indexed
		case bit < 0:
			return 0, indexed
		}
		return int((n >> uint(bit)) & 1), indexed
	}
	if supportsIndexing(item) {
		return getProperty(ctx, item, ToS(property, nil)), indexed
	}
	if rv := reflect.ValueOf(item); rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		// Array#[] with a property that is not an integer
		return nil, indexError
	}
	return nil, noIndex
}

// integerItem returns the value of an item of an integer kind.
func integerItem(item interface{}) (int64, bool) {
	rv := reflect.ValueOf(item)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(rv.Uint()), true
	}
	return 0, false
}

// supportsIndexing checks if an item supports property access.
func supportsIndexing(item interface{}) bool {
	if item == nil {
		return false
	}

	switch item.(type) {
	case map[string]interface{}:
		return true
	default:
		// A drop gives its properties
		if isDrop(item) {
			return true
		}
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		if v.Kind() == reflect.Map {
			return true
		}
	}

	return false
}

// getProperty returns the property of item, a map or a drop, for the render
// of ctx.
func getProperty(ctx *Context, item interface{}, property string) interface{} {
	switch v := item.(type) {
	case nil:
		return nil
	case map[string]interface{}:
		return v[property]
	}
	if g, ok := asDrop(item); ok {
		return dropGet(ctx, g, property)
	}
	rv := reflect.ValueOf(item)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		if val := rv.MapIndex(reflect.ValueOf(property).Convert(rv.Type().Key())); val.IsValid() {
			return val.Interface()
		}
	}
	return nil
}

// isTruthy checks if a value is truthy in Liquid context.
func isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}

	if b, ok := val.(bool); ok {
		return b
	}

	// Empty strings and arrays are falsy
	switch v := val.(type) {
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	}

	// Reflection fallback for typed slices/arrays/maps
	// This matches Ruby's duck-typing behavior: empty collections are falsy
	if val != nil {
		v := reflect.ValueOf(val)
		kind := v.Kind()
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return v.Len() > 0
		}
	}

	return true
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Liquid equality: numbers by value, lists and maps element by element
	if liquidEqual(a, b) {
		return true
	}

	// String comparison
	aStr := ToS(a, nil)
	bStr := ToS(b, nil)

	return aStr == bStr
}
