package liquid

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
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
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return []interface{}{}
	}

	index := 0
	for i := 0; i < v.Len(); i++ {
		if to != nil && *to <= index {
			break
		}

		if from <= index {
			segments = append(segments, v.Index(i).Interface())
		}

		index++
	}

	return segments
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
	case float32:
		return saturateFloat(float64(v)), nil
	case float64:
		return saturateFloat(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, NewArgumentError("invalid integer")
		}
		return i, nil
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
	if toLiquid, ok := obj.(interface {
		ToLiquidValue() interface{}
	}); ok {
		return toLiquid.ToLiquidValue()
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
		return strconv.FormatFloat(v, 'f', -1, 64)
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
		return fmt.Sprintf("%v", obj)
	}
}

// Inspect returns a detailed string representation of an object.
func Inspect(obj interface{}, seen map[uintptr]bool) string {
	if seen == nil {
		seen = make(map[uintptr]bool)
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return hashInspect(v, seen)
	case []interface{}:
		return arrayInspect(v, seen)
	default:
		return fmt.Sprintf("%#v", obj)
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

	var b strings.Builder
	b.WriteString("{")
	first := true
	for key, value := range hash {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(Inspect(key, seen))
		b.WriteString("=>")
		b.WriteString(Inspect(value, seen))
	}
	b.WriteString("}")
	return b.String()
}
