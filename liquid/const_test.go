package liquid

import (
	"regexp"
	"testing"
)

func TestConstants(t *testing.T) {
	// Test that constants are not nil
	if FilterSeparator == nil {
		t.Error("FilterSeparator should not be nil")
	}
	if TagStart == nil {
		t.Error("TagStart should not be nil")
	}
	if TagEnd == nil {
		t.Error("TagEnd should not be nil")
	}
	if VariableStart == nil {
		t.Error("VariableStart should not be nil")
	}
	if VariableEnd == nil {
		t.Error("VariableEnd should not be nil")
	}

	// Test EMPTY_HASH and EMPTY_ARRAY
	if EMPTY_HASH == nil {
		t.Error("EMPTY_HASH should not be nil")
	}
	if EMPTY_ARRAY == nil {
		t.Error("EMPTY_ARRAY should not be nil")
	}

	// Test that EMPTY_HASH is empty
	if len(EMPTY_HASH) != 0 {
		t.Errorf("EMPTY_HASH should be empty, got length %d", len(EMPTY_HASH))
	}

	// Test that EMPTY_ARRAY is empty
	if len(EMPTY_ARRAY) != 0 {
		t.Errorf("EMPTY_ARRAY should be empty, got length %d", len(EMPTY_ARRAY))
	}
}

func TestRegexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern *regexp.Regexp
		text    string
		want    bool
	}{
		{"TagStart", TagStart, "{%", true},
		{"TagStart", TagStart, "{% if", true},
		{"TagStart", TagStart, "{{", false},
		{"TagEnd", TagEnd, "%}", true},
		{"TagEnd", TagEnd, "endif %}", true},
		{"VariableStart", VariableStart, "{{", true},
		{"VariableStart", VariableStart, "{{ name", true},
		{"VariableStart", VariableStart, "{%", false},
		{"VariableEnd", VariableEnd, "}}", true},
		{"VariableEnd", VariableEnd, "name }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pattern.MatchString(tt.text)
			if got != tt.want {
				t.Errorf("Pattern %s.MatchString(%q) = %v, want %v", tt.name, tt.text, got, tt.want)
			}
		})
	}
}
