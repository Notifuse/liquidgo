package liquid

import (
	"fmt"
	"strconv"
)

// RangeLookup represents a range lookup expression (start..end).
type RangeLookup struct {
	startObj interface{}
	endObj   interface{}
}

// StartObj returns the start object.
func (rl *RangeLookup) StartObj() interface{} {
	return rl.startObj
}

// EndObj returns the end object.
func (rl *RangeLookup) EndObj() interface{} {
	return rl.endObj
}

// Range represents a simple integer range.
type Range struct {
	Start int
	End   int
}

// String returns the string representation of the range (e.g., "1..5").
func (r *Range) String() string {
	return fmt.Sprintf("%d..%d", r.Start, r.End)
}

// RangeLookupParse parses start and end markups into a RangeLookup or range.
func RangeLookupParse(startMarkup, endMarkup string, ss *StringScanner, cache map[string]interface{}) interface{} {
	startObj := Parse(startMarkup, ss, cache)
	endObj := Parse(endMarkup, ss, cache)

	// Check if either is an evaluable expression (has Evaluate method)
	// For now, we'll check if they're VariableLookup instances
	if _, ok := startObj.(*VariableLookup); ok {
		return NewRangeLookup(startObj, endObj)
	}
	if _, ok := endObj.(*VariableLookup); ok {
		return NewRangeLookup(startObj, endObj)
	}

	// Both are primitives, convert to integers and create range
	startInt := toInteger(startObj)
	endInt := toInteger(endObj)

	// Create a simple range representation
	// In Go, we'll use a struct to represent ranges
	return &Range{
		Start: startInt,
		End:   endInt,
	}
}

// NewRangeLookup creates a new RangeLookup.
func NewRangeLookup(startObj, endObj interface{}) *RangeLookup {
	return &RangeLookup{
		startObj: startObj,
		endObj:   endObj,
	}
}

func toInteger(input interface{}) int {
	switch v := input.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		// Try to parse as integer
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
		return 0
	case nil:
		return 0
	default:
		// Try ToInteger utility
		if val, err := ToInteger(input); err == nil {
			return val
		}
		return 0
	}
}

