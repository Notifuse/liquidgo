package liquid

import (
	"testing"
)

func TestRangeLookupParse(t *testing.T) {
	tests := []struct {
		name        string
		startMarkup string
		endMarkup   string
		check       func(interface{}) bool
	}{
		{"simple integers", "1", "10", func(r interface{}) bool {
			rg, ok := r.(*Range)
			return ok && rg.Start == 1 && rg.End == 10
		}},
		{"negative integers", "-5", "5", func(r interface{}) bool {
			rg, ok := r.(*Range)
			return ok && rg.Start == -5 && rg.End == 5
		}},
		{"string integers", "0", "100", func(r interface{}) bool {
			rg, ok := r.(*Range)
			return ok && rg.Start == 0 && rg.End == 100
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RangeLookupParse(tt.startMarkup, tt.endMarkup, nil, nil)
			if result == nil {
				t.Fatal("Expected range result, got nil")
			}
			if !tt.check(result) {
				t.Errorf("RangeLookupParse(%q, %q) did not pass check", tt.startMarkup, tt.endMarkup)
			}
		})
	}
}

func TestRangeLookupWithVariables(t *testing.T) {
	// Test with variable lookups (should create RangeLookup, not Range)
	startVL := VariableLookupParse("start", nil, nil)
	endVL := VariableLookupParse("end", nil, nil)

	rl := NewRangeLookup(startVL, endVL)
	if rl == nil {
		t.Fatal("Expected RangeLookup, got nil")
	}
	if rl.StartObj() != startVL {
		t.Error("StartObj mismatch")
	}
	if rl.EndObj() != endVL {
		t.Error("EndObj mismatch")
	}
}

func TestRangeString(t *testing.T) {
	r := &Range{Start: 1, End: 5}
	str := r.String()
	if str != "1..5" {
		t.Errorf("Expected '1..5', got %q", str)
	}

	r2 := &Range{Start: -10, End: 10}
	str2 := r2.String()
	if str2 != "-10..10" {
		t.Errorf("Expected '-10..10', got %q", str2)
	}
}

func TestRangeToInteger(t *testing.T) {
	// Test toInteger with various inputs
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"float64", 3.14, 3},
		{"string number", "42", 42},
		{"string float", "3.14", 0}, // toInteger only parses integers, not floats
		{"invalid string", "not a number", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toInteger(tt.input)
			if got != tt.want {
				t.Errorf("toInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}
