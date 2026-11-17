package liquid

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// DecimalRegex matches decimal numbers
	DecimalRegex = regexp.MustCompile(`^-?\d+\.\d+$`)

	// UnixTimestampRegex matches unix timestamps
	UnixTimestampRegex = regexp.MustCompile(`^\d+$`)
)

// SliceCollection slices a collection from index `from` to index `to` (exclusive).
// If `to` is nil, slices to the end.
func SliceCollection(collection interface{}, from int, to *int) []interface{} {
	// Check if collection has a LoadSlice method (for custom collections)
	if loadSlicer, ok := collection.(interface {
		LoadSlice(from int, to *int) []interface{}
	}); ok {
		if from != 0 || to != nil {
			return loadSlicer.LoadSlice(from, to)
		}
	}

	return sliceCollectionUsingEach(collection, from, to)
}

func sliceCollectionUsingEach(collection interface{}, from int, to *int) []interface{} {
	var segments []interface{}

	// Handle strings specially
	if str, ok := collection.(string); ok {
		if str == "" {
			return []interface{}{}
		}
		return []interface{}{str}
	}

	// Handle Range objects specially (convert to array)
	if r, ok := collection.(*Range); ok {
		for i := r.Start; i <= r.End && i >= r.Start; i++ { // i >= r.Start stops at an int overflow
			if to != nil && *to <= (i-r.Start) {
				break
			}
			if from <= (i - r.Start) {
				segments = append(segments, i)
			}
		}
		return segments
	}

	// Check if collection implements Each method
	if eacher, ok := collection.(interface {
		Each(func(interface{}))
	}); ok {
		index := 0
		eacher.Each(func(item interface{}) {
			if to != nil && *to <= index {
				return
			}
			if from <= index {
				segments = append(segments, item)
			}
			index++
		})
		return segments
	}

	// Use reflection to iterate
	v := reflect.ValueOf(collection)
	if v.Kind() == reflect.Map {
		return slicePairs(hashPairs(v), from, to)
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return []interface{}{}
	}

	index := 0
	for i := 0; i < v.Len(); i++ {
		if to != nil && *to <= index {
			break
		}

		if from <= index {
			segments = append(segments, itemAt(v, i))
		}

		index++
	}

	return segments
}

// hashPairs returns the pairs of a hash as Ruby's Hash#each gives them, each
// a [key, value] list, in the order of their keys as written: a Go map keeps
// no order, and the hash output (hashInspect) sorts the same way.
func hashPairs(v reflect.Value) [][2]interface{} {
	pairs := make([][2]interface{}, 0, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		pairs = append(pairs, [2]interface{}{iter.Key().Interface(), iter.Value().Interface()})
	}
	sort.Slice(pairs, func(i, j int) bool { return ToS(pairs[i][0], nil) < ToS(pairs[j][0], nil) })
	return pairs
}

// slicePairs returns the pairs from from to to, each as a list.
func slicePairs(pairs [][2]interface{}, from int, to *int) []interface{} {
	hi := len(pairs)
	if to != nil {
		hi = min(hi, *to)
	}
	lo := max(from, 0)
	if hi <= lo {
		return []interface{}{}
	}
	items := make([]interface{}, 0, hi-lo)
	for _, p := range pairs[lo:hi] {
		items = append(items, []interface{}{p[0], p[1]})
	}
	return items
}

// wholeFloat returns a float with no fraction as an integer, saturated to the
// int range: a number decoded from JSON is a float64 (a deliberate difference
// from Ruby, which refuses any Float). A fraction, NaN and the infinities are
// an invalid integer, as Ruby's Utils.to_integer.
func wholeFloat(f float64) (int, error) {
	if f != math.Trunc(f) || math.IsInf(f, 0) {
		return 0, NewArgumentError("invalid integer")
	}
	return saturateFloat(f), nil
}

// ToInteger converts a value to an integer.
// Optimization: Fast path for int type (most common case).
func ToInteger(num interface{}) (int, error) {
	// Fast path for most common case
	if v, ok := num.(int); ok {
		return v, nil
	}

	// Handle other numeric types; values beyond the int range saturate
	switch v := num.(type) {
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return saturateInt64(v), nil
	case uint:
		return saturateUint64(uint64(v)), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return saturateUint64(uint64(v)), nil
	case uint64:
		return saturateUint64(v), nil
	case *big.Int:
		if v == nil {
			return 0, NewArgumentError("invalid integer")
		}
		if v.IsInt64() {
			return saturateInt64(v.Int64()), nil
		}
		if v.Sign() < 0 {
			return math.MinInt, nil
		}
		return math.MaxInt, nil
	case float32:
		return wholeFloat(float64(v))
	case float64:
		return wholeFloat(v)
	case string:
		// as Ruby's Integer(): around spaces, a sign, underscores and the
		// 0x, 0o, 0b and 0 prefixes; nothing else
		i, err := strconv.ParseInt(strings.TrimSpace(v), 0, 0)
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				return saturateInt64(i), nil
			}
			return 0, NewArgumentError("invalid integer")
		}
		return int(i), nil
	default:
		return 0, NewArgumentError("invalid integer")
	}
}

func saturateInt64(v int64) int {
	switch {
	case v > int64(math.MaxInt):
		return math.MaxInt
	case v < int64(math.MinInt):
		return math.MinInt
	}
	return int(v)
}

func saturateUint64(v uint64) int {
	if v > uint64(math.MaxInt) {
		return math.MaxInt
	}
	return int(v)
}

// saturateFloat truncates toward zero, as Ruby's Float#to_i, and saturates at the int
// range (Go's conversion of an out-of-range float is undefined). NaN gives 0.
func saturateFloat(v float64) int {
	switch {
	case math.IsNaN(v):
		return 0
	case v >= float64(math.MaxInt):
		return math.MaxInt
	case v <= float64(math.MinInt):
		return math.MinInt
	}
	return int(v)
}

// ToNumber converts a value to a number (int, int64, or float64).
func ToNumber(obj interface{}) (interface{}, bool) {
	switch v := obj.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		trimmed := strings.TrimSpace(v)
		if DecimalRegex.MatchString(trimmed) {
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return 0, false
			}
			return f, true
		}
		i, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, false
		}
		return float64(i), true
	default:
		if toNumberer, ok := obj.(interface {
			ToNumber() interface{}
		}); ok {
			return toNumberer.ToNumber(), true
		}
		return 0, false
	}
}

// toDateLayouts are the text formats ToDate reads, standing in for Ruby's Time.parse.
// A format without a zone is read as UTC.
var toDateLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z0700",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 15:04:05 MST",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"2006/01/02 15:04:05",
	"2006/01/02",
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC850,
	time.RFC822,
	time.RFC822Z,
	"January 2, 2006 15:04",
	"January 2, 2006",
	"Jan 2, 2006",
	"2 January 2006 15:04",
	"2 January 2006",
	"2 Jan 2006",
}

// zoneAbbreviations are the offsets, in hours, of the zone names ToDate reads, as in
// Ruby's Date zone table. Go's time.Parse reads an abbreviation with the server's local
// zone, or with a zero offset when it does not know it, so ToDate looks it up here.
var zoneAbbreviations = map[string]float64{
	"UTC": 0, "UT": 0, "GMT": 0, "Z": 0, "WET": 0,
	"EST": -5, "EDT": -4, "CST": -6, "CDT": -5, "MST": -7, "MDT": -6, "PST": -8, "PDT": -7,
	"AKST": -9, "AKDT": -8, "HST": -10,
	"BST": 1, "WEST": 1, "CET": 1, "CEST": 2, "EET": 2, "EEST": 3, "MSK": 3,
	"IST": 5.5, "AWST": 8, "JST": 9, "KST": 9, "ACST": 9.5, "AEST": 10, "AEDT": 11,
	"NZST": 12, "NZDT": 13,
}

// withZoneAbbreviation gives t the offset of its zone abbreviation. An unknown
// abbreviation (SGT, HKT...) keeps the time as written, read as UTC like a date without
// a zone (Ruby reads it in the server's zone).
func withZoneAbbreviation(t time.Time) *time.Time {
	name, _ := t.Zone()
	hours, ok := zoneAbbreviations[name]
	if !ok {
		hours = 0
	}
	offset := int(hours * 3600)
	var loc *time.Location
	if offset == 0 {
		loc = time.UTC
	} else {
		loc = time.FixedZone(name, offset)
	}
	zoned := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	return &zoned
}

// ToDate converts a value to a time.Time, as Ruby Liquid's Utils.to_date: a time, a
// Unix timestamp (number or digits), "now"/"today", or a date in a common text format.
// Unlike Ruby, a float timestamp is accepted too, since decoded JSON numbers are floats.
func ToDate(obj interface{}) *time.Time {
	switch v := obj.(type) {
	case time.Time:
		return &v
	case *time.Time:
		return v
	case nil, bool:
		return nil
	case string:
		if v == "" {
			return nil
		}
		trimmed := strings.TrimSpace(v)
		lower := strings.ToLower(trimmed)
		if lower == "now" || lower == "today" {
			now := time.Now()
			return &now
		}
		if UnixTimestampRegex.MatchString(trimmed) {
			ts, err := strconv.ParseInt(trimmed, 10, 64)
			if err != nil {
				return nil
			}
			t := time.Unix(ts, 0)
			return &t
		}
		for _, layout := range toDateLayouts {
			// in UTC, which defines no abbreviation but UTC, so that the server's own zone
			// never decides what "EDT" means
			if t, err := time.ParseInLocation(layout, trimmed, time.UTC); err == nil {
				if strings.Contains(layout, "MST") {
					return withZoneAbbreviation(t)
				}
				return &t
			}
		}
		return nil
	}
	if n, ok := asNumber(obj); ok {
		var t time.Time
		switch n.kind {
		case signedNumber:
			t = time.Unix(n.i, 0)
		case unsignedNumber:
			if n.u > math.MaxInt64 {
				return nil
			}
			t = time.Unix(int64(n.u), 0)
		default:
			if math.IsNaN(n.f) || math.IsInf(n.f, 0) || math.Abs(n.f) > 1e15 {
				return nil
			}
			sec := math.Floor(n.f)
			t = time.Unix(int64(sec), int64((n.f-sec)*1e9))
		}
		return &t
	}
	return nil
}

// ToLiquidValue converts an object to its liquid representation.
func ToLiquidValue(obj interface{}) interface{} {
	if _, ok := obj.(string); ok {
		return obj // the most common value, inlined
	}
	return toLiquidValueMethod(obj)
}

// toLiquidValueMethod is ToLiquidValue for the values that may have the
// method, apart so that ToLiquidValue is inlined.
func toLiquidValueMethod(obj interface{}) interface{} {
	switch obj.(type) {
	case nil, int, float64, bool, map[string]interface{}, []interface{}:
		return obj // the common values have no method: no method lookup
	}
	if v, ok := obj.(interface {
		ToLiquidValue() interface{}
	}); ok {
		return v.ToLiquidValue()
	}
	return obj
}

// ToS converts an object to a string representation.
// Optimization: Fast path for common types to enable compiler inlining.
func ToS(obj interface{}, seen map[uintptr]bool) string {
	// Handle nil - in Liquid, nil renders as empty string (like Ruby's nil.to_s)
	if obj == nil {
		return ""
	}
	// A value writes its liquid form, also as an item of a list or a hash
	obj = ToLiquid(obj)
	switch v := obj.(type) {
	case string:
		return v // before indirect, which reads the kind of obj by reflection
	case int:
		return strconv.Itoa(v)
	}
	if target, ok := indirect(obj); ok {
		return ToS(target, seen)
	}

	// Fast path for common types (helps with inlining)
	switch v := obj.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return floatToS(v)
	case time.Time:
		return rubyTime(v, false)
	case *time.Time:
		return rubyTime(*v, false)
	case map[string]interface{}:
		if seen == nil {
			seen = make(map[uintptr]bool)
		}
		return hashInspect(v, seen)
	case []interface{}:
		if seen == nil {
			seen = make(map[uintptr]bool)
		}
		return arrayInspect(v, seen)
	default:
		// String, not fmt: fmt writes what a reflect.Value holds
		if s, ok := obj.(fmt.Stringer); ok {
			return s.String()
		}
		if items, ok := SliceItems(obj); ok {
			if seen == nil {
				seen = make(map[uintptr]bool)
			}
			return arrayInspect(items, seen)
		}
		if isPlainStruct(obj) {
			return ""
		}
		if rv := reflect.ValueOf(obj); rv.Kind() == reflect.Map {
			if seen == nil {
				seen = make(map[uintptr]bool)
			}
			return mapInspect(rv, seen)
		}
		if hasNoText(obj) {
			return ""
		}
		return fmt.Sprintf("%v", obj)
	}
}

// hasNoText reports whether obj is a value fmt would write as an address: a
// channel, a func, a pointer or an unsafe pointer, which render as nothing.
// An error writes its text.
func hasNoText(obj interface{}) bool {
	if _, ok := obj.(error); ok {
		return false
	}
	switch reflect.ValueOf(obj).Kind() {
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.UnsafePointer:
		return true
	}
	return false
}

// rubyTime writes a time as Ruby's Time#to_s, or Time#inspect, which adds
// the fraction of a second: "2026-11-15 10:00:00 UTC" for a time in UTC,
// "2026-11-15 11:00:00 +0100" otherwise.
func rubyTime(t time.Time, inspect bool) string {
	layout := "2006-01-02 15:04:05"
	if inspect {
		layout += ".999999999"
	}
	if t.Location() == time.UTC {
		return t.Format(layout) + " UTC"
	}
	return t.Format(layout + " -0700")
}

// indirect returns what a pointer that renders as its target points to:
// nil for a nil pointer, the end of a chain of pointers otherwise. A pointer
// whose type has a String or an Error method renders itself; indirect then
// reports false, as for a value that is not a pointer. A drop renders as
// what it points to: never through fmt, which writes the fields it holds.
func indirect(obj interface{}) (interface{}, bool) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr {
		return nil, false
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, true
		}
		switch v := rv.Interface(); v.(type) {
		case fmt.Stringer, error:
			if rv.Interface() == obj {
				return nil, false
			}
			return v, true
		}
		rv = rv.Elem()
	}
	return rv.Interface(), true
}

// isPlainStruct reports whether obj is a Go struct, or a pointer to one,
// without a String or an Error method, a drop or not. It renders as nothing,
// so that fmt never writes its fields.
func isPlainStruct(obj interface{}) bool {
	switch obj.(type) {
	case fmt.Stringer, error:
		return false
	}
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Struct
}

// mapInspect writes a map of any type as hashInspect does, so that the
// values it holds are written as Liquid writes them.
func mapInspect(rv reflect.Value, seen map[uintptr]bool) string {
	ptr := rv.Pointer()
	if seen[ptr] {
		return "{...}"
	}
	seen[ptr] = true
	defer delete(seen, ptr)

	keys := make([]string, 0, rv.Len())
	values := make(map[string]interface{}, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		k := Inspect(iter.Key().Interface(), seen)
		keys = append(keys, k)
		values[k] = iter.Value().Interface()
	}
	return writeInspectedHash(keys, values, seen)
}

// writeInspectedHash writes a hash from its inspected keys, in sorted order:
// a Go map has no order of its own, and the output must not change from one
// render to the next.
func writeInspectedHash(keys []string, values map[string]interface{}, seen map[uintptr]bool) string {
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteString("=>")
		b.WriteString(Inspect(values[k], seen))
	}
	b.WriteString("}")
	return b.String()
}

// SliceItems returns the items of a typed slice or array, which Ruby has as
// an Array; an item whose pointer is a drop is given by its address.
func SliceItems(obj interface{}) ([]interface{}, bool) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	items := make([]interface{}, rv.Len())
	for i := range items {
		items[i] = itemAt(rv, i)
	}
	return items, true
}

// Inspect returns a detailed string representation of an object.
func Inspect(obj interface{}, seen map[uintptr]bool) string {
	if seen == nil {
		seen = make(map[uintptr]bool)
	}
	obj = ToLiquid(obj)
	if target, ok := indirect(obj); ok {
		return Inspect(target, seen)
	}

	switch v := obj.(type) {
	case nil:
		return "nil"
	case map[string]interface{}:
		return hashInspect(v, seen)
	case []interface{}:
		return arrayInspect(v, seen)
	case string:
		return strconv.Quote(v)
	case time.Time:
		return rubyTime(v, true)
	case *time.Time:
		return rubyTime(*v, true)
	default:
		// A Go value is written without its Go type: a value with a String
		// method as a string, a number or a boolean as it renders, a typed
		// slice as an array, and anything else as fmt's %v
		if s, ok := obj.(fmt.Stringer); ok {
			return strconv.Quote(s.String())
		}
		if _, ok := asNumber(obj); ok {
			return ToS(obj, nil)
		}
		if b, ok := obj.(bool); ok {
			return strconv.FormatBool(b)
		}
		if items, ok := SliceItems(obj); ok {
			return arrayInspect(items, seen)
		}
		if rv := reflect.ValueOf(obj); rv.Kind() == reflect.String {
			return strconv.Quote(rv.String())
		}
		if isPlainStruct(obj) {
			return "nil"
		}
		if rv := reflect.ValueOf(obj); rv.Kind() == reflect.Map {
			return mapInspect(rv, seen)
		}
		if hasNoText(obj) {
			return "nil"
		}
		return fmt.Sprintf("%v", obj)
	}
}

func arrayInspect(arr []interface{}, seen map[uintptr]bool) string {
	ptr := reflect.ValueOf(arr).Pointer()
	if seen[ptr] {
		return "[...]"
	}

	seen[ptr] = true
	defer delete(seen, ptr)

	var b strings.Builder
	b.WriteString("[")
	for i, item := range arr {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(Inspect(item, seen))
	}
	b.WriteString("]")
	return b.String()
}

func hashInspect(hash map[string]interface{}, seen map[uintptr]bool) string {
	ptr := reflect.ValueOf(hash).Pointer()
	if seen[ptr] {
		return "{...}"
	}

	seen[ptr] = true
	defer delete(seen, ptr)

	keys := make([]string, 0, len(hash))
	values := make(map[string]interface{}, len(hash))
	for key, value := range hash {
		k := Inspect(key, seen)
		keys = append(keys, k)
		values[k] = value
	}
	return writeInspectedHash(keys, values, seen)
}

// floatToS writes a float as Ruby's Float#to_s for the special values
// (Infinity, -Infinity, NaN), and without exponent otherwise.
func floatToS(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case math.IsNaN(f):
		return "NaN"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}
