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

	// Test EmptyHash and EmptyArray
	if EmptyHash == nil {
		t.Error("EmptyHash should not be nil")
	}
	if EmptyArray == nil {
		t.Error("EmptyArray should not be nil")
	}

	// Test that EmptyHash is empty
	if len(EmptyHash) != 0 {
		t.Errorf("EmptyHash should be empty, got length %d", len(EmptyHash))
	}

	// Test that EmptyArray is empty
	if len(EmptyArray) != 0 {
		t.Errorf("EmptyArray should be empty, got length %d", len(EmptyArray))
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
