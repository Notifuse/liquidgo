package liquid

import (
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
)

// The filters that can multiply the size of their input check the size of
// their result before they build it: a string longer than the render
// length limit, or an array with more items than the assign score limit
// leaves, raises a MemoryError, as a write to the output would.

// limits returns the resource limits of the render the filters belong to.
func (sf *StandardFilters) limits() *ResourceLimits {
	if sf == nil || sf.context == nil {
		return nil
	}
	return sf.context.resourceLimits
}

// checkLength raises a MemoryError if a string of n bytes passes the render
// length limit.
func (sf *StandardFilters) checkLength(n int) {
	if rl := sf.limits(); rl != nil && rl.renderLengthLimit != nil && n > *rl.renderLengthLimit {
		rl.raiseLimitsReached()
	}
}

// countItems counts the n items of an array a filter is about to build in
// the assign score.
func (sf *StandardFilters) countItems(n int) {
	if rl := sf.limits(); rl != nil && n > 0 {
		rl.IncrementAssignScore(n)
	}
}

// replacedLength returns the length of s once count occurrences of search
// are replaced by replacement, saturated at math.MaxInt.
func replacedLength(s, search, replacement string, count int) int {
	grow := len(replacement) - len(search)
	if grow <= 0 || count == 0 {
		return len(s) + count*grow
	}
	if count > (math.MaxInt-len(s))/grow {
		return math.MaxInt
	}
	return len(s) + count*grow
}

// joinedLength returns the length of parts joined by sep, saturated.
func joinedLength(parts []string, sep string) int {
	total := 0
	for _, p := range parts {
		if total > math.MaxInt-len(p) {
			return math.MaxInt
		}
		total += len(p)
	}
	if len(parts) > 1 && len(sep) > 0 {
		gaps := len(parts) - 1
		if gaps > (math.MaxInt-total)/len(sep) {
			return math.MaxInt
		}
		total += gaps * len(sep)
	}
	return total
}

// splitCount returns the number of parts strings.Split(s, sep) gives.
func splitCount(s, sep string) int {
	if sep == "" {
		return utf8.RuneCountInString(s)
	}
	return strings.Count(s, sep) + 1
}

// reflectLen returns the length of a slice or an array, or 0.
func reflectLen(v interface{}) int {
	if arr, ok := v.([]interface{}); ok {
		return len(arr)
	}
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return rv.Len()
	}
	return 0
}
