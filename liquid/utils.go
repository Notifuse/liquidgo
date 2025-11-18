package liquid

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// DECIMAL_REGEX matches decimal numbers
	DECIMAL_REGEX = regexp.MustCompile(`^-?\d+\.\d+$`)

	// UNIX_TIMESTAMP_REGEX matches unix timestamps
	UNIX_TIMESTAMP_REGEX = regexp.MustCompile(`^\d+$`)
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
func ToInteger(num interface{}) (int, error) {
	switch v := num.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
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

// ToNumber converts a value to a number (int or float64).
func ToNumber(obj interface{}) interface{} {
	switch v := obj.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v
	case string:
		trimmed := strings.TrimSpace(v)
		if DECIMAL_REGEX.MatchString(trimmed) {
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return 0
			}
			return f
		}
		i, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0
		}
		return i
	default:
		if toNumberer, ok := obj.(interface {
			ToNumber() interface{}
		}); ok {
			return toNumberer.ToNumber()
		}
		return 0
	}
}

// ToDate converts a value to a time.Time.
func ToDate(obj interface{}) *time.Time {
	// If it already has a Strftime method (time.Time), return it
	if t, ok := obj.(time.Time); ok {
		return &t
	}

	switch v := obj.(type) {
	case string:
		if v == "" {
			return nil
		}
		lower := strings.ToLower(v)
		if lower == "now" || lower == "today" {
			now := time.Now()
			return &now
		}
		if UNIX_TIMESTAMP_REGEX.MatchString(v) {
			ts, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil
			}
			t := time.Unix(ts, 0)
			return &t
		}
		// Try parsing as RFC3339 or common formats
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Try other common formats
			formats := []string{
				time.RFC1123,
				time.RFC1123Z,
				"2006-01-02",
				"2006-01-02 15:04:05",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, v); err == nil {
					return &t
				}
			}
			return nil
		}
		return &t
	case int, int64:
		var ts int64
		switch vv := v.(type) {
		case int:
			ts = int64(vv)
		case int64:
			ts = vv
		}
		t := time.Unix(ts, 0)
		return &t
	default:
		return nil
	}
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
func ToS(obj interface{}, seen map[uintptr]bool) string {
	// Handle nil - in Liquid, nil renders as empty string (like Ruby's nil.to_s)
	if obj == nil {
		return ""
	}

	if seen == nil {
		seen = make(map[uintptr]bool)
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		return hashInspect(v, seen)
	case []interface{}:
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

