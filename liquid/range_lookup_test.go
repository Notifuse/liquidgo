package liquid

import (
	"testing"
)

func TestRangeLookupParse(t *testing.T) {
	tests := []struct {
		name         string
		startMarkup  string
		endMarkup    string
		check        func(interface{}) bool
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

