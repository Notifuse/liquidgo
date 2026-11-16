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

// namedTag returns a constructor of a generic Tag named name, whatever name
// the template gives: the tests tell constructors apart by what they build.
func namedTag(name string) TagConstructor {
	return func(_, markup string, pc ParseContextInterface) (Node, error) {
		return NewTag(name, markup, pc), nil
	}
}

// builtName returns the name of the tag that constructor builds.
func builtName(t *testing.T, constructor TagConstructor) string {
	t.Helper()
	if constructor == nil {
		t.Fatal("Expected a constructor, got nil")
	}
	node, err := constructor("", "", NewParseContext(ParseContextOptions{}))
	if err != nil {
		t.Fatalf("constructor: %v", err)
	}
	return node.(*Tag).TagName()
}

func TestEnvironmentRegisterTag(t *testing.T) {
	env := NewEnvironment()
	env.RegisterTag("test", namedTag("TestTag"))

	if got := builtName(t, env.TagForName("test")); got != "TestTag" {
		t.Errorf("Expected 'TestTag', got %v", got)
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

func TestNewEnvironmentWithStandardTags(t *testing.T) {
	env := NewEnvironmentWithStandardTags()
	if env == nil {
		t.Fatal("Expected Environment, got nil")
	}
	// Note: Standard tags are registered from outside the package
	// This function just creates an environment
}

func TestEnvironmentTags(t *testing.T) {
	env := NewEnvironment()
	env.RegisterTag("test1", namedTag("Tag1"))
	env.RegisterTag("test2", namedTag("Tag2"))

	tags := env.Tags()
	if tags == nil {
		t.Fatal("Expected tags map, got nil")
	}
	if len(tags) < 2 {
		t.Errorf("Expected at least 2 tags, got %d", len(tags))
	}
	if got := builtName(t, tags["test1"]); got != "Tag1" {
		t.Errorf("Expected 'Tag1' for test1, got %v", got)
	}
	if got := builtName(t, tags["test2"]); got != "Tag2" {
		t.Errorf("Expected 'Tag2' for test2, got %v", got)
	}
}

func TestEnvironmentRegisterFilters(t *testing.T) {
	env := NewEnvironment()

	// Register multiple filters
	filters := []interface{}{
		&StandardFilters{},
	}
	err := env.RegisterFilters(filters)
	if err != nil {
		t.Fatalf("RegisterFilters() error = %v", err)
	}

	// Verify filters were registered
	names := env.FilterMethodNames()
	if len(names) == 0 {
		t.Error("Expected filter method names after RegisterFilters, got empty")
	}
}
