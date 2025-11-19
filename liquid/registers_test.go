package liquid

import (
	"testing"
)

func TestRegistersBasic(t *testing.T) {
	r := NewRegisters(nil)
	if r == nil {
		t.Fatal("Expected Registers, got nil")
	}
}

func TestRegistersSetGet(t *testing.T) {
	r := NewRegisters(nil)
	r.Set("key", "value")

	val := r.Get("key")
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestRegistersStatic(t *testing.T) {
	static := map[string]interface{}{
		"static_key": "static_value",
	}
	r := NewRegisters(static)

	val := r.Get("static_key")
	if val != "static_value" {
		t.Errorf("Expected 'static_value', got %v", val)
	}
}

func TestRegistersChangesOverrideStatic(t *testing.T) {
	static := map[string]interface{}{
		"key": "static_value",
	}
	r := NewRegisters(static)
	r.Set("key", "changed_value")

	val := r.Get("key")
	if val != "changed_value" {
		t.Errorf("Expected 'changed_value', got %v", val)
	}
}

func TestRegistersDelete(t *testing.T) {
	r := NewRegisters(nil)
	r.Set("key", "value")
	r.Delete("key")

	val := r.Get("key")
	if val != nil {
		t.Errorf("Expected nil after delete, got %v", val)
	}
}

func TestRegistersFetch(t *testing.T) {
	r := NewRegisters(nil)

	// Test with default value
	val := r.Fetch("nonexistent", "default")
	if val != "default" {
		t.Errorf("Expected 'default', got %v", val)
	}

	// Test with existing value
	r.Set("key", "value")
	val = r.Fetch("key", "default")
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestRegistersHasKey(t *testing.T) {
	static := map[string]interface{}{
		"static_key": "value",
	}
	r := NewRegisters(static)

	if !r.HasKey("static_key") {
		t.Error("Expected HasKey to return true for static key")
	}

	r.Set("dynamic_key", "value")
	if !r.HasKey("dynamic_key") {
		t.Error("Expected HasKey to return true for dynamic key")
	}

	if r.HasKey("nonexistent") {
		t.Error("Expected HasKey to return false for nonexistent key")
	}
}

func TestRegistersFromRegisters(t *testing.T) {
	r1 := NewRegisters(map[string]interface{}{"key": "value"})
	r2 := NewRegisters(r1)

	val := r2.Get("key")
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestRegistersStaticMap(t *testing.T) {
	static := map[string]interface{}{
		"static_key": "static_value",
	}
	r := NewRegisters(static)

	staticMap := r.Static()
	if staticMap == nil {
		t.Fatal("Expected non-nil static map")
	}
	if staticMap["static_key"] != "static_value" {
		t.Errorf("Expected 'static_value', got %v", staticMap["static_key"])
	}
}

func TestRegistersChangesMap(t *testing.T) {
	r := NewRegisters(nil)
	r.Set("key", "value")

	changes := r.Changes()
	if changes == nil {
		t.Fatal("Expected non-nil changes map")
	}
	if changes["key"] != "value" {
		t.Errorf("Expected 'value', got %v", changes["key"])
	}
}
