# Testing Guide for liquidgo

This guide explains the testing philosophy and best practices for liquidgo development.

## Why We Need Both Unit and Integration Tests

### The Problem We Discovered

During development, we discovered a critical bug where filters with optional parameters (like `default`, `sort`, `where`) failed to invoke correctly from templates. Interestingly, **all unit tests passed**, but the filters were broken in actual use.

**Why did this happen?**

- **Unit tests** directly called filter methods: `sf.Default(nil, "bar", nil)` ✅ Passed
- **Actual usage** went through template rendering: `{{ x | default: "bar" }}` ❌ Failed
- The bug was in the **filter invocation system**, not the filter logic itself

This revealed a gap in our testing strategy.

## Testing Philosophy

liquidgo follows a **two-layer testing approach**:

```
┌─────────────────────────────────────────────────────┐
│  Integration Tests (Template Rendering)             │
│  Tests the entire pipeline: parse → render → output │
│  ✓ Tests user-facing behavior                       │
│  ✓ Catches invocation/infrastructure bugs           │
│  Example: filter_optional_params_test.go            │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│  Unit Tests (Direct Method Calls)                   │
│  Tests individual components in isolation            │
│  ✓ Tests logic and edge cases                       │
│  ✓ Fast and focused                                 │
│  Example: standardfilters_test.go                   │
└─────────────────────────────────────────────────────┘
```

## Unit Tests vs Integration Tests

### Unit Tests (liquid/*_test.go)

**Purpose:** Test individual methods/functions in isolation

**Example:**
```go
func TestDefault(t *testing.T) {
    sf := &StandardFilters{}
    
    // Direct method call
    result := sf.Default(nil, "bar", nil)
    if result != "bar" {
        t.Errorf("Expected 'bar', got %v", result)
    }
}
```

**What they test:**
- Filter logic is correct
- Edge cases are handled
- Return values are as expected

**What they DON'T test:**
- Whether the filter can be invoked from a template
- Argument parsing and conversion
- The filter registration and lookup system

### Integration Tests (integration/*_test.go)

**Purpose:** Test the full template rendering pipeline

**Example:**
```go
func TestFiltersWithOptionalParameters(t *testing.T) {
    template := `{{ x | default: "fallback" }}`
    data := map[string]interface{}{"x": nil}
    
    // Full template rendering
    result := renderTemplate(template, data)
    assertEqual(t, "fallback", result)
}
```

**What they test:**
- Templates parse correctly
- Filters are invoked with correct arguments
- The entire render pipeline works end-to-end
- User-facing behavior matches expectations

## When to Write Each Type of Test

### Write Unit Tests When:
1. Testing complex filter logic with many edge cases
2. Testing internal utility functions
3. Testing data structures and their methods
4. Fast iteration on algorithm correctness

### Write Integration Tests When:
1. Adding a new filter (test it works in templates!)
2. Adding a new tag
3. Testing filter/tag combinations
4. Testing optional parameters or special syntax
5. Reproducing user-reported bugs
6. Ensuring Ruby Liquid compatibility

**Golden Rule:** If a user would type it in a template, write an integration test for it.

## Integration Test Best Practices

### 1. Use assertTemplateResult Helper

```go
func TestMyFilter(t *testing.T) {
    assertTemplateResult(t, 
        "expected output",           // Expected
        `{{ input | my_filter }}`,   // Template
        map[string]interface{}{      // Data
            "input": "test",
        },
    )
}
```

### 2. Test With and Without Optional Parameters

```go
{
    name:     "sort without property",
    template: `{{ arr | sort }}`,
    // ...
},
{
    name:     "sort with property",
    template: `{{ arr | sort: "field" }}`,
    // ...
},
```

### 3. Test Backward Compatibility

When fixing bugs, ensure existing working code still works:

```go
func TestBackwardCompatibility(t *testing.T) {
    // Test that filters with all args still work
    tests := []struct{
        name     string
        template string
        // ...
    }{
        {
            name:     "filter with all arguments",
            template: `{{ x | filter: arg1, arg2, arg3 }}`,
            // ...
        },
    }
}
```

### 4. Test Edge Cases in Templates, Not Just Unit Tests

```go
{
    name:     "empty array",
    template: `{{ arr | where: "active" | size }}`,
    data:     map[string]interface{}{"arr": []interface{}{}},
    expected: "0",
},
```

## Ruby Liquid Test Parity

liquidgo aims for **feature parity** with Ruby Liquid. When implementing filters:

1. **Check the Ruby tests:** `reference-liquid/test/integration/standard_filter_test.rb`
2. **Port relevant tests** to liquidgo integration tests
3. **Look for `assert_template_result`** - these are integration tests

**Example from Ruby:**
```ruby
def test_default
  assert_equal("bar", @filters.default(nil, "bar"))  # Unit test
  assert_template_result('bar', "{{ false | default: 'bar' }}")  # Integration test ✓
end
```

**Port to liquidgo:**
```go
func TestFiltersWithOptionalParameters(t *testing.T) {
    tests := []struct{
        name     string
        template string
        data     map[string]interface{}
        expected string
    }{
        {
            name:     "default filter with false",
            template: `{{ false | default: "bar" }}`,  // Same as Ruby!
            data:     map[string]interface{}{"false": false},
            expected: "bar",
        },
    }
}
```

## Common Pitfalls

### ❌ Only Testing Direct Method Calls

```go
// This only tests the method logic
func TestMyFilter(t *testing.T) {
    sf := &StandardFilters{}
    result := sf.MyFilter("input")
    // ...
}
```

### ✅ Also Test Through Templates

```go
// This tests the full pipeline
func TestMyFilterInTemplate(t *testing.T) {
    assertTemplateResult(t, "expected", `{{ input | my_filter }}`, ...)
}
```

### ❌ Assuming Optional Parameters "Just Work"

```go
func (sf *StandardFilters) MyFilter(input interface{}, optional interface{}) {
    // Unit test calls: sf.MyFilter("test", nil)  ✓ Works
    // Template calls: {{ x | my_filter }}         ❓ Does it work?
}
```

### ✅ Explicitly Test Optional Parameter Scenarios

```go
{
    name:     "my_filter without optional param",
    template: `{{ x | my_filter }}`,  // Test this!
    // ...
},
{
    name:     "my_filter with optional param",
    template: `{{ x | my_filter: "arg" }}`,  // And this!
    // ...
},
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run only integration tests
go test ./integration

# Run only unit tests for liquid package
go test ./liquid

# Run specific test
go test ./integration -run TestFiltersWithOptionalParameters

# Run with verbose output
go test -v ./integration
```

## Test File Organization

```
liquidgo/
├── liquid/                    # Core package
│   ├── standardfilters.go    # Filter implementations
│   ├── standardfilters_test.go   # Unit tests ← Test filter logic
│   ├── strainer_template.go  # Filter invocation system
│   └── strainer_template_test.go  # Unit tests
│
└── integration/               # Integration tests
    ├── helper_test.go        # Test helpers (assertTemplateResult)
    ├── filter_optional_params_test.go  # ← Test filter invocation from templates
    ├── comprehensive_test.go # Full feature tests
    └── TESTING_GUIDE.md      # This file
```

## Contributing

When adding new features:

1. ✅ Write unit tests for the logic
2. ✅ Write integration tests for user-facing behavior
3. ✅ Check Ruby Liquid tests for compatibility
4. ✅ Run full test suite to catch regressions

**Remember:** If users interact with it through templates, write an integration test!

## References

- Ruby Liquid tests: `reference-liquid/test/integration/`
- Ruby Liquid docs: https://shopify.github.io/liquid/
- Shopify Liquid docs: https://shopify.dev/docs/api/liquid

---

**Key Takeaway:** Unit tests verify that code works in isolation. Integration tests verify that code works for users. We need both!

