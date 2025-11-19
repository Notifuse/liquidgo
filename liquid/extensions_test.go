package liquid

import (
	"testing"
)

// Test type that implements ToLiquid
type testLiquidValue struct {
	value string
}

func (t *testLiquidValue) ToLiquid() interface{} {
	return t.value
}

func TestToLiquid(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		want     interface{}
		checkMap bool
	}{
		{"implements ToLiquid", &testLiquidValue{value: "test"}, "test", false},
		{"string", "hello", "hello", false},
		{"int", 42, 42, false},
		{"nil", nil, nil, false},
		{"map", map[string]interface{}{"key": "value"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToLiquid(tt.input)
			if tt.checkMap {
				// For maps, just verify it's a map
				if _, ok := got.(map[string]interface{}); !ok {
					t.Errorf("ToLiquid() = %T, want map[string]interface{}", got)
				}
			} else {
				if got != tt.want {
					t.Errorf("ToLiquid() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestToLiquidWithCustomType(t *testing.T) {
	custom := &testLiquidValue{value: "custom"}
	result := ToLiquid(custom)
	if result != "custom" {
		t.Errorf("Expected 'custom', got %v", result)
	}
}
