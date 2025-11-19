# Liquid Directory Test Coverage Analysis

## Summary

**Overall Coverage: 89.8%**

This report identifies missing code paths in the `liquid` directory that need test coverage.

## Critical Issues (0% Coverage - Completely Untested)

The following functions have **NO test coverage** and should be prioritized:

| File | Line | Function | Coverage |
|------|------|----------|----------|
| `liquid/tags/comment.go` | 30 | `RenderToOutputBuffer` | 0.0% |
| `liquid/tags/doc.go` | 42 | `RenderToOutputBuffer` | 0.0% |
| `liquid/tags/inline_comment.go` | 36 | `RenderToOutputBuffer` | 0.0% |
| `liquid/usage.go` | 9 | `Increment` | 0.0% |
| `liquid/variable.go` | 378 | `AddWarning` | 0.0% |

## High Priority (< 60% Coverage)

| File | Line | Function | Coverage |
|------|------|----------|----------|
| `liquid/tokenizer.go` | 88 | `tokenize` | 57.1% |

## Medium Priority (60-75% Coverage)

| File | Line | Function | Coverage |
|------|------|----------|----------|
| `liquid/context.go` | 622 | `squashInstanceAssignsWithEnvironments` | 62.5% |
| `liquid/drop.go` | 52 | `InvokeDropOn` | 66.7% |
| `liquid/drop.go` | 276 | `stringsTitle` | 66.7% |
| `liquid/string_scanner.go` | 110 | `Rest` | 66.7% |
| `liquid/string_scanner.go` | 134 | `Byteslice` | 66.7% |
| `liquid/string_scanner.go` | 145 | `runeAt` | 66.7% |
| `liquid/tag.go` | 101 | `RenderToOutputBuffer` | 66.7% |
| `liquid/variable.go` | 44 | `NewVariable` | 70.0% |
| `liquid/range_lookup.go` | 69 | `toInteger` | 72.7% |
| `liquid/tags/render.go` | 94 | `RenderToOutputBuffer` | 73.5% |
| `liquid/strainer_template.go` | 96 | `Invoke` | 74.4% |

## Areas Needing Improvement (75-85% Coverage)

### Core Parsing & Rendering
- `liquid/block_body.go` - Multiple functions with 75-85% coverage
- `liquid/template.go` - Core rendering functions at 78-82%
- `liquid/tokenizer.go` - Token processing functions at 75-77%

### Tag Implementations
- `liquid/tags/render.go` - 73.5% RenderToOutputBuffer
- `liquid/tags/table_row.go` - 75.4% RenderToOutputBuffer
- `liquid/tags/if.go` - 75.0% RenderToOutputBuffer

### Utility Functions
- `liquid/string_scanner.go` - Multiple utility functions at 66-80%
- `liquid/variable_lookup.go` - 77.8% Evaluate
- `liquid/range_lookup.go` - 72.7% toInteger

## Critical Code Paths Missing Tests

### 1. Comment Tag Rendering
All comment-style tags have 0% coverage on their `RenderToOutputBuffer` methods:
- `comment.go`
- `doc.go`
- `inline_comment.go`

These are likely simple (just return empty string or nothing), but should still be tested.

### 2. Variable System
- `variable.go:378` - `AddWarning` (0%)
- `variable.go:44` - `NewVariable` (70%)
- `variable_lookup.go:116` - `Evaluate` (77.8%)

### 3. Drop System
- `drop.go:148` - `InvokeDropOld` (40.9%)
- `drop.go:52` - `InvokeDropOn` (66.7%)

### 4. Tokenizer
- `tokenizer.go:88` - `tokenize` (57.1%)
- This is a critical parsing component that needs better coverage

### 5. String Scanner Utilities
Multiple low-coverage utility functions in `string_scanner.go`:
- `Rest` (66.7%)
- `Byteslice` (66.7%)
- `runeAt` (66.7%)

## Recommendations

### Priority 1 - Zero Coverage Functions
Add basic tests for all 0% coverage functions. These are likely:
- Simple render methods that return nothing (comments)
- Unused utility functions that might be needed in future

### Priority 2 - Core Parsing Components
Focus on:
- `tokenizer.go` - Critical for parsing, currently at 57.1%
- `drop.go` - `InvokeDropOld` at 40.9%
- `context.go` - `squashInstanceAssignsWithEnvironments` at 62.5%

### Priority 3 - Edge Cases
Many functions are at 80-95%, indicating some edge cases are missing:
- Error handling paths
- Nil/empty input handling
- Boundary conditions
- Error recovery scenarios

## How to View Detailed Coverage

An HTML coverage report has been generated:
```bash
open coverage_liquid.html
```

This shows line-by-line coverage with:
- 🟢 Green: Covered lines
- 🔴 Red: Uncovered lines
- ⚪ Gray: Not executable

## Next Steps

1. **Add tests for 0% coverage functions** (5 functions)
2. **Improve tokenizer coverage** (currently 57.1%)
3. **Add tests for drop system** (InvokeDropOld at 40.9%)
4. **Test edge cases** in functions with 75-85% coverage
5. **Review error paths** - many missing branches are likely error handling

Target: **95%+ coverage** for all core parsing and rendering components

---

## Detailed Analysis of Critical Missing Tests

### 1. Comment Tags (0% Coverage - EASY FIX)

These are trivial functions that don't render anything but still need tests:

**`liquid/tags/comment.go:30`** - `RenderToOutputBuffer`
```go
func (c *CommentTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
    // Comments don't render anything
}
```
**Test needed:** Verify that rendering produces no output.

**`liquid/tags/doc.go:42`** - `RenderToOutputBuffer`
```go
func (d *DocTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
    // Docs don't render anything
}
```
**Test needed:** Verify that rendering produces no output.

**`liquid/tags/inline_comment.go:36`** - `RenderToOutputBuffer`
Similar to above - needs test to verify no output.

### 2. Usage Tracking (0% Coverage - TODO)

**`liquid/usage.go:9`** - `Increment`
```go
func (u *Usage) Increment(name string) {
    // TODO: Implement usage tracking
}
```
**Test needed:** Either implement the feature and test it, or add a test that verifies it's a no-op.

### 3. Variable Warnings (0% Coverage)

**`liquid/variable.go:378`** - `AddWarning`
```go
func (p *parseContextWrapper) AddWarning(error) {
    // No-op for wrapper
}
```
**Test needed:** Test that calling this doesn't panic and behaves as a no-op.

### 4. Drop System (40.9% Coverage - COMPLEX)

**`liquid/drop.go:148`** - `InvokeDropOld` (40.9%)
This is a complex reflection-based method invocation system. Missing test paths likely include:
- Methods with different visibility
- Fields vs methods
- Pointer vs value receivers
- Method not found scenarios
- Invalid reflection cases

**Recommendation:** Compare with Ruby implementation to ensure all edge cases are tested.

### 5. Tokenizer (57.1% Coverage - CRITICAL)

**`liquid/tokenizer.go:88`** - `tokenize` (57.1%)
This is core parsing logic. Missing 42.9% of paths suggests:
- Edge cases in token detection
- Error handling paths not covered
- Boundary conditions
- Malformed input handling

**Recommendation:** This needs immediate attention as it's critical infrastructure.

### 6. String Scanner Utilities (66% Coverage)

**`liquid/string_scanner.go`** - Multiple functions at 66-77%
- `Rest()` - Get remaining string
- `Byteslice()` - Get byte slice
- `runeAt()` - Get rune at position

These are utility functions likely missing boundary condition tests:
- Empty string input
- Out of bounds access
- Unicode edge cases

---

## Test Coverage Goals by Component

| Component | Current | Target | Priority |
|-----------|---------|--------|----------|
| **Comment Tags** | 0% | 100% | 🔴 High (Easy) |
| **Usage System** | 0% | 100% | 🟢 Low (TODO) |
| **Tokenizer** | 57% | 95%+ | 🔴 High (Critical) |
| **Drop System** | 41-67% | 90%+ | 🟠 Medium |
| **String Scanner** | 66-80% | 95%+ | 🟠 Medium |
| **Variable System** | 70-85% | 95%+ | 🟠 Medium |
| **Template/Rendering** | 78-82% | 95%+ | 🟠 Medium |

---

## Quick Wins (Easy Tests to Add)

1. **Comment tag rendering** (3 functions, 0% → 100%)
2. **Usage.Increment** (1 function, 0% → 100%)
3. **parseContextWrapper.AddWarning** (1 function, 0% → 100%)

These 5 functions can be fully tested with minimal effort and will improve overall coverage from 89.8% to ~90.2%.

---

## Coverage Files Generated

- `coverage_liquid.out` - Raw coverage data
- `coverage_liquid.html` - Interactive HTML report (open in browser)
- `liquid_coverage_per_file.txt` - Per-function coverage list
- `liquid_missing_coverage.txt` - Functions with < 100% coverage
- `COVERAGE_REPORT.md` - This report

## Commands for Analysis

```bash
# View overall coverage
go tool cover -func=coverage_liquid.out | tail -1

# View HTML report
open coverage_liquid.html

# Re-run coverage
go test -coverprofile=coverage_liquid.out -covermode=atomic ./liquid/...

# Check specific package
go test -coverprofile=coverage.out -covermode=atomic ./liquid/tags/
go tool cover -html=coverage.out
```
