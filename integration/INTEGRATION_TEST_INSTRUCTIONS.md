# Integration Test Instructions

## Overview

Integration tests in the `integration/` directory verify end-to-end template rendering behavior and ensure feature parity with the Ruby Liquid implementation. These tests match the Ruby integration tests from `reference-liquid/test/integration/`.

## File Structure and Naming

Integration test files should mirror the Ruby test file structure:

- **Ruby**: `reference-liquid/test/integration/template_test.rb`
- **Go**: `integration/template_test.go`

- **Ruby**: `reference-liquid/test/integration/assign_test.rb`
- **Go**: `integration/assign_test.go`

- **Ruby**: `reference-liquid/test/integration/tags/for_tag_test.rb`
- **Go**: `integration/tags/for_test.go`

### Key Points:

- Base file names match (without extension)
- Directory structure mirrors Ruby structure
- Test function names should match Ruby test method names (converted to Go naming conventions)

## Test Implementation Guidelines

### 1. Match Ruby Test Cases

Each Ruby test method should have a corresponding Go test function:

```ruby
# Ruby
def test_assign_with_hyphen_in_variable_name
  assert_template_result("Print this-thing", template_source)
end
```

```go
// Go
func TestAssignWithHyphenInVariableName(t *testing.T) {
    assertTemplateResult(t, "Print this-thing", templateSource, nil)
}
```

### 2. Use Helper Functions

Use the helper functions from `helper_test.go`:

- `assertTemplateResult()` - Main helper for template rendering tests
- `assertMatchSyntaxError()` - Test syntax error handling
- `assertSyntaxError()` - Simplified syntax error test
- `withCustomTag()` - Test custom tags
- `withErrorModes()` - Test different error modes
- `withGlobalFilter()` - Test global filters

### 3. Test Data Types

Use test drops from `test_drops.go` when needed:

- `ThingWithToLiquid`
- `IntegerDrop`, `BooleanDrop`, `StringDrop`
- `ErrorDrop`
- `TemplateContextDrop`
- `SettingsDrop`

## Debugging Failed Tests

When an integration test fails, follow this process to determine if the issue is in the **test implementation** or the **Go implementation**:

### Step 1: Verify Test Setup

1. **Check tag registration**: Ensure `tags.RegisterStandardTags(env)` is called before parsing
2. **Check environment**: Verify the environment is passed correctly to `ParseTemplate`
3. **Check helper function**: Review `assertTemplateResult` and related helpers for correctness

### Step 2: Compare with Ruby Implementation

1. **Read the Ruby test** (`reference-liquid/test/integration/[test_file].rb`)

   - Understand what the test is verifying
   - Check the expected behavior
   - Note any edge cases or special handling

2. **Read the Ruby implementation** (`reference-liquid/lib/liquid/[component].rb`)

   - Understand how the feature is implemented in Ruby
   - Check for any special logic or edge case handling
   - Note error handling and validation

3. **Read the Go implementation** (`liquid/[component].go`)
   - Compare the logic with Ruby implementation
   - Check if all edge cases are handled
   - Verify error handling matches Ruby behavior

### Step 3: Identify the Root Cause

Ask these questions:

1. **Is the test correctly written?**

   - Does it match the Ruby test structure?
   - Are all parameters correct?
   - Is the expected output correct?

2. **Is the Go implementation missing functionality?**

   - Does the Go code handle the same cases as Ruby?
   - Are there missing features or edge cases?
   - Is error handling equivalent?

3. **Is there a type mismatch or API difference?**
   - Are types being used correctly?
   - Is the API being called correctly?
   - Are there Go-specific considerations?

### Step 4: Fix the Issue

**If it's a test problem:**

- Fix the test to match Ruby behavior
- Correct helper function usage
- Update test data or expectations

**If it's an implementation problem:**

- Update the Go implementation to match Ruby behavior
- Add missing features or edge case handling
- Fix error handling or validation
- **IMPORTANT**: Update corresponding unit tests (see "Updating Unit Tests" below)

## Common Issues and Solutions

### Issue: "unknown_tag" Error

**Symptoms**: Tests fail with "errors.syntax.unknown_tag" panic

**Possible Causes**:

1. Tags not registered before parsing
2. Environment not passed correctly to template
3. Type mismatch in tag constructor (TagConstructor vs function type)

**Solution**:

1. Ensure `tags.RegisterStandardTags(env)` is called
2. Verify environment is set in `TemplateOptions`
3. Check `block_body.go` tag lookup logic matches registered tag types

### Issue: Variables Return `<nil>`

**Symptoms**: Template renders `<nil>` instead of expected value

**Possible Causes**:

1. Context not set up correctly
2. Variable assignment not working
3. Scope issues with assigns

**Solution**:

1. Check how assigns are passed to `Render()`
2. Verify context scopes are set up correctly
3. Compare with Ruby's context setup

### Issue: Syntax Errors Not Caught

**Symptoms**: Tests expecting syntax errors don't fail

**Possible Causes**:

1. Error mode not set correctly
2. Error handling differs from Ruby
3. Parser too permissive

**Solution**:

1. Check error mode in environment (strict vs lax)
2. Compare error handling with Ruby implementation
3. Verify parser error detection logic

## Updating Unit Tests When Implementation Changes

When you fix an implementation bug discovered through integration tests, **always update the corresponding unit tests** to ensure the fix is properly tested at the unit level.

### Finding Corresponding Unit Tests

Unit tests are located alongside the implementation files:

- **Implementation**: `liquid/tags/assign.go`
- **Unit Tests**: `liquid/tags/assign_test.go`

- **Implementation**: `liquid/tags/capture.go`
- **Unit Tests**: `liquid/tags/capture_test.go`

- **Implementation**: `liquid/block_body.go`
- **Unit Tests**: `liquid/block_body_test.go`

### What to Update in Unit Tests

1. **Add test cases for the bug fix**:

   - Create a unit test that reproduces the specific issue
   - Test the component in isolation (not end-to-end)
   - Verify the fix works at the unit level

2. **Update existing tests if needed**:

   - If the fix changes behavior, update affected unit tests
   - Ensure unit tests still pass with the new implementation

3. **Test edge cases**:
   - Add unit tests for edge cases discovered during integration testing
   - Test error conditions and boundary cases

### Example: Updating Unit Tests After Fix

**Scenario**: Integration test `TestCapturesBlockContentInVariable` fails because capture tag doesn't handle quoted variable names.

**Fix Applied**: Updated `liquid/tags/capture.go` to handle quoted strings in variable names.

**Unit Test Update**:

```go
// In liquid/tags/capture_test.go

func TestCaptureTagWithQuotedVariableName(t *testing.T) {
    pc := liquid.NewParseContext(liquid.ParseContextOptions{})

    // Test with single quotes
    tag, err := NewCaptureTag("capture", "'var'", pc)
    if err != nil {
        t.Fatalf("NewCaptureTag() error = %v", err)
    }
    if tag.To() != "var" {
        t.Errorf("Expected To 'var', got %q", tag.To())
    }

    // Test with double quotes
    tag, err = NewCaptureTag("capture", `"var"`, pc)
    if err != nil {
        t.Fatalf("NewCaptureTag() error = %v", err)
    }
    if tag.To() != "var" {
        t.Errorf("Expected To 'var', got %q", tag.To())
    }
}
```

### Benefits of Updating Unit Tests

1. **Faster feedback**: Unit tests run faster than integration tests
2. **Better isolation**: Unit tests verify the specific component works correctly
3. **Regression prevention**: Prevents the bug from reoccurring
4. **Documentation**: Unit tests document expected behavior
5. **Easier debugging**: Unit tests help identify exactly where issues occur

### Checklist

When fixing an implementation bug:

- [ ] Fix the implementation code
- [ ] Verify integration test passes
- [ ] Add/update unit test for the fix
- [ ] Verify unit test passes
- [ ] Run all unit tests: `go test ./liquid/...`
- [ ] Run all integration tests: `go test ./integration/...`
- [ ] Check for related unit tests that might need updates

## Testing Workflow

1. **Write the test** matching Ruby test structure
2. **Run the test**: `go test ./integration/... -v -run TestName`
3. **If it fails**:
   - Read Ruby test to understand intent
   - Read Ruby implementation to understand behavior
   - Read Go implementation to find differences
   - Determine if test or implementation needs fixing
   - Fix and re-test
   - **Update unit tests** if implementation was changed
4. **If it passes**: Move to next test

## Reference Files

- **Ruby Tests**: `reference-liquid/test/integration/`
- **Ruby Unit Tests**: `reference-liquid/test/unit/`
- **Ruby Implementation**: `reference-liquid/lib/liquid/`
- **Go Implementation**: `liquid/`
- **Go Unit Tests**: `liquid/*_test.go` (alongside implementation files)
- **Test Helpers**: `integration/helper_test.go`
- **Test Drops**: `integration/test_drops.go`

## Best Practices

1. **One test per Ruby test method** - Keep tests focused and isolated
2. **Use descriptive test names** - Match Ruby test names when possible
3. **Comment complex tests** - Explain what's being tested
4. **Reference Ruby code** - Add comments referencing Ruby implementation
5. **Test edge cases** - Include all edge cases from Ruby tests
6. **Verify error messages** - Ensure error messages match Ruby (when applicable)
7. **Update unit tests** - When fixing implementation bugs, always add/update corresponding unit tests
8. **Test at multiple levels** - Integration tests verify end-to-end, unit tests verify components in isolation

## Example: Debugging a Failed Test

```go
func TestAssignWithHyphenInVariableName(t *testing.T) {
    templateSource := `{% assign this-thing = 'Print this-thing' -%}
{{ this-thing -}}`
    assertTemplateResult(t, "Print this-thing", templateSource, nil)
}
```

**If this fails:**

1. **Check Ruby test** (`reference-liquid/test/integration/assign_test.rb`):

   - Ruby uses `assert_template_result("Print this-thing", template_source)`
   - No assigns needed, just template

2. **Check Ruby implementation** (`reference-liquid/lib/liquid/tags/assign.rb`):

   - How does assign handle hyphens in variable names?
   - How does variable lookup work?

3. **Check Go implementation** (`liquid/tags/assign.go`):

   - Does Go assign tag handle hyphens correctly?
   - Does variable lookup support hyphens?

4. **Check test helper** (`integration/helper_test.go`):

   - Is `assertTemplateResult` working correctly?
   - Are tags registered?

5. **Determine fix**:
   - If Ruby supports it but Go doesn't → Fix Go implementation
   - If test is wrong → Fix test
   - If helper is wrong → Fix helper

## Notes

- Integration tests verify **end-to-end behavior**, not just unit functionality
- Tests should be **deterministic** and **repeatable**
- When in doubt, **match Ruby behavior exactly**
- Document any intentional deviations from Ruby behavior
