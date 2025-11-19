package liquid

import (
	"testing"
)

func TestVariableLookupParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*VariableLookup) bool
	}{
		{"simple", "var", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) == 0
		}},
		{"with dot", "var.method", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) == 1
		}},
		{"with brackets", "var[0]", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) == 1
		}},
		{"nested brackets", "var[method][0]", func(vl *VariableLookup) bool {
			return vl.name == "var" && len(vl.lookups) >= 2
		}},
		{"command method", "items.size", func(vl *VariableLookup) bool {
			if vl.name != "items" || len(vl.lookups) != 1 {
				return false
			}
			return vl.LookupCommand(0)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VariableLookupParse(tt.input, nil, nil)
			if result == nil {
				t.Fatal("Expected VariableLookup, got nil")
			}
			if !tt.check(result) {
				t.Errorf("VariableLookupParse(%q) did not pass check", tt.input)
			}
		})
	}
}

func TestVariableLookupCommandMethods(t *testing.T) {
	vl := VariableLookupParse("items.size", nil, nil)
	if !vl.LookupCommand(0) {
		t.Error("Expected size to be a command method")
	}

	vl = VariableLookupParse("items.first", nil, nil)
	if !vl.LookupCommand(0) {
		t.Error("Expected first to be a command method")
	}

	vl = VariableLookupParse("items.last", nil, nil)
	if !vl.LookupCommand(0) {
		t.Error("Expected last to be a command method")
	}

	vl = VariableLookupParse("items.name", nil, nil)
	if vl.LookupCommand(0) {
		t.Error("Expected name NOT to be a command method")
	}
}
