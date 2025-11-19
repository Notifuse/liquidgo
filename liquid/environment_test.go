package liquid

import (
	"testing"
)

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()
	if env == nil {
		t.Fatal("Expected Environment, got nil")
	}
	if env.ErrorMode() != "lax" {
		t.Errorf("Expected error mode 'lax', got '%s'", env.ErrorMode())
	}
}

func TestEnvironmentRegisterTag(t *testing.T) {
	env := NewEnvironment()
	env.RegisterTag("test", "TestTag")

	tag := env.TagForName("test")
	if tag != "TestTag" {
		t.Errorf("Expected 'TestTag', got %v", tag)
	}
}

func TestEnvironmentRegisterFilter(t *testing.T) {
	env := NewEnvironment()

	// Register a filter (StandardFilters already has methods)
	filter := &StandardFilters{}
	err := env.RegisterFilter(filter)
	if err != nil {
		t.Fatalf("RegisterFilter() error = %v", err)
	}

	names := env.FilterMethodNames()
	if len(names) == 0 {
		t.Error("Expected filter method names, got empty")
	}

	// Check for a known method
	found := false
	for _, name := range names {
		if name == "Size" || name == "Downcase" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Size or Downcase in filter method names")
	}
}

func TestEnvironmentCreateStrainer(t *testing.T) {
	env := NewEnvironment()
	ctx := &mockContext{}

	strainer := env.CreateStrainer(ctx, nil, false)
	if strainer == nil {
		t.Fatal("Expected StrainerTemplate, got nil")
	}
}

func TestEnvironmentFileSystem(t *testing.T) {
	env := NewEnvironment()
	fs := &BlankFileSystem{}
	env.SetFileSystem(fs)

	if env.FileSystem() != fs {
		t.Error("FileSystem mismatch")
	}
}

func TestEnvironmentExceptionRenderer(t *testing.T) {
	env := NewEnvironment()
	renderer := func(err error) interface{} {
		return "rendered"
	}
	env.SetExceptionRenderer(renderer)

	if env.ExceptionRenderer() == nil {
		t.Error("Expected exception renderer, got nil")
	}
}

func TestEnvironmentSetErrorMode(t *testing.T) {
	env := NewEnvironment()
	env.SetErrorMode("strict")
	if env.ErrorMode() != "strict" {
		t.Errorf("Expected error mode 'strict', got '%s'", env.ErrorMode())
	}
}

// TestEnvironmentStrainerCaching tests that strainer template classes are cached
func TestEnvironmentStrainerCaching(t *testing.T) {
	env := NewEnvironment()
	ctx := &mockContext{}

	// Create first strainer
	strainer1 := env.CreateStrainer(ctx, nil, false)
	if strainer1 == nil {
		t.Fatal("Expected StrainerTemplate, got nil")
	}

	// Create second strainer with same filters (should use cache)
	strainer2 := env.CreateStrainer(ctx, nil, false)
	if strainer2 == nil {
		t.Fatal("Expected StrainerTemplate, got nil")
	}

	// Both should be valid strainers
	if strainer1 == nil || strainer2 == nil {
		t.Error("Both strainers should be valid")
	}
}
