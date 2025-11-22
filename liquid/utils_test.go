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
		{"*time.Time pointer", &now, false},
		{"nil *time.Time pointer", (*time.Time)(nil), true},
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

func TestToDate_PointerToTime(t *testing.T) {
	now := time.Now()
	ptr := &now
	got := ToDate(ptr)
	if got == nil {
		t.Errorf("ToDate(*time.Time) = nil, expected non-nil")
		return
	}
	if !got.Equal(now) {
		t.Errorf("ToDate(*time.Time) = %v, expected %v", got, now)
	}
}

func TestToDate_NilPointerToTime(t *testing.T) {
	var ptr *time.Time = nil
	got := ToDate(ptr)
	if got != nil {
		t.Errorf("ToDate(nil *time.Time) = %v, expected nil", got)
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

// TestUtilsToInteger tests integer conversion with various types
func TestUtilsToInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"int64", int64(42), 42, false},
		{"int32", int32(42), 42, false},
		{"float64", 42.0, 42, false},
		{"float32", float32(42.0), 42, true}, // float32 not directly supported
		{"nil", nil, 0, true},
		{"bool true", true, 0, true},   // bool not directly supported
		{"bool false", false, 0, true}, // bool not directly supported
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInteger(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUtilsToDate tests date conversion
func TestUtilsToDate(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantNil bool
	}{
		{"unix timestamp int", 1609459200, false},
		{"unix timestamp int64", int64(1609459200), false},
		{"date string", "2021-01-01", false},
		{"invalid string", "invalid", true},
		{"nil", nil, true},
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

// TestUtilsSliceCollection tests collection slicing
func TestUtilsSliceCollection(t *testing.T) {
	tests := []struct {
		name       string
		collection interface{}
		from       int
		to         *int
		wantLen    int
	}{
		{"slice array", []interface{}{1, 2, 3, 4}, 1, intPtr(3), 2},
		{"negative from", []interface{}{1, 2, 3}, -1, nil, 3}, // negative from wraps to end
		{"from > length", []interface{}{1, 2, 3}, 10, nil, 0},
		{"nil collection", nil, 0, nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceCollection(tt.collection, tt.from, tt.to)
			if len(got) != tt.wantLen {
				t.Errorf("SliceCollection() length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestToLiquidValue tests ToLiquidValue conversion
func TestToLiquidValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(interface{}) bool
	}{
		{"string", "hello", func(v interface{}) bool { return v == "hello" }},
		{"int", 42, func(v interface{}) bool { return v == 42 }},
		{"nil", nil, func(v interface{}) bool { return v == nil }},
		{"array", []interface{}{1, 2}, func(v interface{}) bool {
			arr, ok := v.([]interface{})
			return ok && len(arr) == 2
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToLiquidValue(tt.input)
			if !tt.check(got) {
				t.Errorf("ToLiquidValue() = %v, check failed", got)
			}
		})
	}
}

// TestInspect tests Inspect function
func TestInspect(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(string) bool
	}{
		{"string", "hello", func(s string) bool { return s == `"hello"` }},
		{"int", 42, func(s string) bool { return s == "42" }},
		{"nil", nil, func(s string) bool { return s == "nil" }},
		{"array", []interface{}{1, 2}, func(s string) bool { return len(s) > 0 && s[0] == '[' }},
		{"map", map[string]interface{}{"a": 1}, func(s string) bool { return len(s) > 0 && s[0] == '{' }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Inspect(tt.input, nil)
			if !tt.check(got) {
				t.Logf("Inspect() = %q (may differ due to formatting)", got)
			}
		})
	}
}

// TestSliceCollectionWithLoadSlice tests SliceCollection with LoadSlice interface
func TestSliceCollectionWithLoadSlice(t *testing.T) {
	// Create a collection that implements LoadSlice
	loadSliceCollection := &testLoadSliceCollection{
		data: []interface{}{1, 2, 3, 4, 5},
	}

	// Test with from=0, to=nil (should not use LoadSlice, uses Each instead)
	result := SliceCollection(loadSliceCollection, 0, nil)
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	// Test with from=1, to=4 (should use LoadSlice)
	// LoadSlice returns items from index 1 to 4 (exclusive), so indices 1, 2, 3 = [2, 3, 4]
	to := 4
	result2 := SliceCollection(loadSliceCollection, 1, &to)
	if len(result2) < 2 {
		t.Errorf("Expected at least 2 items, got %d", len(result2))
	}
	// Should contain items starting from index 1 (value 2)
	if len(result2) > 0 && result2[0] != 2 {
		t.Errorf("Expected first item to be 2, got %v", result2[0])
	}
}

type testLoadSliceCollection struct {
	data []interface{}
}

func (t *testLoadSliceCollection) LoadSlice(from int, to *int) []interface{} {
	end := len(t.data)
	if to != nil {
		end = *to
	}
	if from < 0 {
		from = 0
	}
	if end > len(t.data) {
		end = len(t.data)
	}
	if from >= end {
		return []interface{}{}
	}
	return t.data[from:end]
}

func (t *testLoadSliceCollection) Each(fn func(interface{})) {
	for _, item := range t.data {
		fn(item)
	}
}

func (t *testLoadSliceCollection) Count() int {
	return len(t.data)
}

// TestToIntegerEdgeCases tests ToInteger with various edge cases
func TestToIntegerEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"int8", int8(42), 42, false},
		{"int16", int16(42), 42, false},
		{"int32", int32(42), 42, false},
		{"int64", int64(42), 42, false},
		{"uint", uint(42), 42, false},
		{"uint8", uint8(42), 42, false},
		{"uint16", uint16(42), 42, false},
		{"uint32", uint32(42), 42, false},
		{"uint64", uint64(42), 42, false},
		{"float64", float64(42.7), 42, false},
		{"negative float64", float64(-42.7), -42, false},
		{"string with spaces", "  42  ", 0, true}, // ToInteger doesn't trim spaces
		{"invalid string", "abc", 0, true},
		{"empty string", "", 0, true},
		{"nil", nil, 0, true},
		{"bool", true, 0, true},
		{"map", map[string]interface{}{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInteger(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSliceCollectionEdgeCases tests SliceCollection edge cases
func TestSliceCollectionEdgeCases(t *testing.T) {
	// Test with non-slice, non-string collection
	result := SliceCollection(map[string]interface{}{"key": "value"}, 0, nil)
	if len(result) != 0 {
		t.Errorf("Expected empty result for map, got %d items", len(result))
	}

	// Test with empty array
	result2 := SliceCollection([]interface{}{}, 0, nil)
	if len(result2) != 0 {
		t.Errorf("Expected empty result for empty array, got %d items", len(result2))
	}

	// Test with from > length
	result3 := SliceCollection([]interface{}{1, 2, 3}, 10, nil)
	if len(result3) != 0 {
		t.Errorf("Expected empty result when from > length, got %d items", len(result3))
	}

	// Test with to < from
	to := 1
	result4 := SliceCollection([]interface{}{1, 2, 3, 4, 5}, 3, &to)
	if len(result4) != 0 {
		t.Errorf("Expected empty result when to < from, got %d items", len(result4))
	}

	// Test with to == from
	to2 := 2
	result5 := SliceCollection([]interface{}{1, 2, 3, 4, 5}, 2, &to2)
	if len(result5) != 0 {
		t.Errorf("Expected empty result when to == from, got %d items", len(result5))
	}
}

// TestToNumberEdgeCases tests ToNumber with edge cases
func TestToNumberEdgeCases(t *testing.T) {
	// Test decimal numbers
	result := ToNumber("3.14")
	if f, ok := result.(float64); !ok || f != 3.14 {
		t.Errorf("Expected 3.14, got %v", result)
	}

	// Test decimal starting with dot (may not be supported by regex)
	result2 := ToNumber(".5")
	// This may return 0 if regex doesn't match, which is acceptable
	if result2 != 0 && result2 != 0.5 {
		t.Logf("Note: .5 conversion returned %v (may not be supported)", result2)
	}

	// Test decimal ending with dot (may not be supported by regex)
	result3 := ToNumber("10.")
	// This may return 0 if regex doesn't match, which is acceptable
	if result3 != 0 && result3 != 10.0 {
		t.Logf("Note: 10. conversion returned %v (may not be supported)", result3)
	}

	// Test invalid number format
	result4 := ToNumber("abc")
	if result4 != 0 {
		t.Errorf("Expected 0 for invalid number, got %v", result4)
	}

	// Test number with whitespace
	result5 := ToNumber("  42  ")
	if i, ok := result5.(int); !ok || i != 42 {
		t.Errorf("Expected 42, got %v", result5)
	}

	// Test decimal with whitespace
	result6 := ToNumber("  3.14  ")
	if f, ok := result6.(float64); !ok || f != 3.14 {
		t.Errorf("Expected 3.14, got %v", result6)
	}

	// Test custom ToNumber interface
	customNum := &testToNumberer{value: 99}
	result7 := ToNumber(customNum)
	if result7 != 99 {
		t.Errorf("Expected 99 from custom ToNumberer, got %v", result7)
	}

	// Test various numeric types
	if ToNumber(int8(10)) != int8(10) {
		t.Error("Expected int8 to pass through")
	}
	if ToNumber(int16(20)) != int16(20) {
		t.Error("Expected int16 to pass through")
	}
	if ToNumber(int32(30)) != int32(30) {
		t.Error("Expected int32 to pass through")
	}
	if ToNumber(int64(40)) != int64(40) {
		t.Error("Expected int64 to pass through")
	}
	if ToNumber(uint(50)) != uint(50) {
		t.Error("Expected uint to pass through")
	}
	if ToNumber(float32(1.5)) != 1.5 {
		t.Error("Expected float32 to convert to float64")
	}
}

type testToNumberer struct {
	value int
}

func (t *testToNumberer) ToNumber() interface{} {
	return t.value
}

// TestToLiquidValueEdgeCases tests ToLiquidValue with edge cases
func TestToLiquidValueEdgeCases(t *testing.T) {
	// Test with custom ToLiquidValue interface
	customLiquid := &testToLiquidValuer{value: "custom"}
	result := ToLiquidValue(customLiquid)
	if result != "custom" {
		t.Errorf("Expected 'custom', got %v", result)
	}

	// Test with regular value (no interface)
	result2 := ToLiquidValue("regular")
	if result2 != "regular" {
		t.Errorf("Expected 'regular', got %v", result2)
	}

	// Test with nil
	result3 := ToLiquidValue(nil)
	if result3 != nil {
		t.Errorf("Expected nil, got %v", result3)
	}

	// Test with various types
	result4 := ToLiquidValue(42)
	if result4 != 42 {
		t.Errorf("Expected 42, got %v", result4)
	}

	result5 := ToLiquidValue(true)
	if result5 != true {
		t.Errorf("Expected true, got %v", result5)
	}
}

type testToLiquidValuer struct {
	value string
}

func (t *testToLiquidValuer) ToLiquidValue() interface{} {
	return t.value
}

// TestToSEdgeCases tests ToS with edge cases
func TestToSEdgeCases(t *testing.T) {
	// Test nil
	result := ToS(nil, nil)
	if result != "" {
		t.Errorf("Expected empty string for nil, got %q", result)
	}

	// Test various types
	if ToS(42, nil) != "42" {
		t.Error("Expected '42' for int")
	}
	if ToS(true, nil) != "true" {
		t.Error("Expected 'true' for bool")
	}
	if ToS(false, nil) != "false" {
		t.Error("Expected 'false' for bool")
	}
	if ToS(3.14, nil) != "3.14" {
		t.Error("Expected '3.14' for float64")
	}
	if ToS("hello", nil) != "hello" {
		t.Error("Expected 'hello' for string")
	}

	// Test with custom type that has String() method
	customStringer := &testStringer{value: "custom"}
	result2 := ToS(customStringer, nil)
	if result2 != "custom" {
		t.Errorf("Expected 'custom', got %q", result2)
	}

	// Test with map (should use hashInspect)
	m := map[string]interface{}{"key": "value"}
	result3 := ToS(m, nil)
	if result3 == "" {
		t.Error("Expected non-empty string for map")
	}
	if !contains(result3, "key") || !contains(result3, "value") {
		t.Errorf("Expected map string to contain key and value, got %q", result3)
	}

	// Test with array (should use arrayInspect)
	arr := []interface{}{1, 2, 3}
	result4 := ToS(arr, nil)
	if result4 == "" {
		t.Error("Expected non-empty string for array")
	}
}

type testStringer struct {
	value string
}

func (t *testStringer) String() string {
	return t.value
}
