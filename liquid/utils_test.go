package liquid

import (
	"testing"
	"time"
)

func TestSliceCollection(t *testing.T) {
	tests := []struct {
		name       string
		collection interface{}
		from       int
		to         *int
		want       int
	}{
		{"slice array", []interface{}{1, 2, 3, 4, 5}, 1, intPtr(4), 3},
		{"slice to end", []interface{}{1, 2, 3}, 1, nil, 2},
		{"empty string", "", 0, nil, 0},
		{"non-empty string", "hello", 0, nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceCollection(tt.collection, tt.from, tt.to)
			if len(got) != tt.want {
				t.Errorf("SliceCollection() length = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestToInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"int", 42, 42, false},
		{"string int", "42", 42, false},
		{"invalid string", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInteger(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantType string
	}{
		{"int", 42, "int"},
		{"float", 3.14, "float64"},
		{"string int", "42", "int"},
		{"string float", "3.14", "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToNumber(tt.input)
			gotType := getTypeName(got)
			if gotType != tt.wantType {
				t.Errorf("ToNumber() type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestToDate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		input   interface{}
		wantNil bool
	}{
		{"time.Time", now, false},
		{"now string", "now", false},
		{"today string", "today", false},
		{"unix timestamp string", "1609459200", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToDate(tt.input)
			if (got == nil) != tt.wantNil {
				t.Errorf("ToDate() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

func TestToS(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		wantContains string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"array", []interface{}{1, 2, 3}, "["},
		{"map", map[string]interface{}{"key": "value"}, "{"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToS(tt.input, nil)
			if !contains(got, tt.wantContains) {
				t.Errorf("ToS() = %v, want to contain %v", got, tt.wantContains)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func getTypeName(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case float64:
		return "float64"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
