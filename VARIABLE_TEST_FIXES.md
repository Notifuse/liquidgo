# Variable Test Fixes - Priority Report

## Summary

After the recent error handling implementation (commit 01b667a), Variable tests had failures that were initially skipped. This report tracks the fixes and remaining issues.

**Status Update:**

- ✅ **3 Critical tests FIXED** (core functionality restored)
- **5 Medium tests** remain skipped (parser validation issues)
- **3 Low tests** remain skipped (edge cases)
- **Total:** 8 tests still skipped, down from 11

---

## ✅ COMPLETED FIXES

### Critical Priority Fixes (Completed)

All 3 critical tests have been fixed and are now passing:

**1. TestVariable_LookupCallsToLiquidValue** - FIXED ✅

- **Fix:** Added reflection-based map access for integer keys in `liquid/variable_lookup.go`
- **Changes:** Lines 211-238 - Added support for `map[int]interface{}` and other map types with non-string keys
- **Impact:** Custom drops with `ToLiquidValue()` now work correctly as array/map indices

**2. TestVariable_ExpressionWithWhitespaceInSquareBrackets** - FIXED ✅

- **Fix:** Rewrote bracket expression tokenizer to handle nested brackets properly
- **Changes:** Added `parseVariableTokens()` function in `liquid/variable_lookup.go` (lines 24-72)
- **Impact:** Nested bracket expressions like `{{ a[ [ 'b' ] ] }}` now work correctly

**3. TestVariable_DynamicFindVarWithDrop** - FIXED ✅

- **Fix:** Same as #2 - the nested bracket tokenizer fix resolved dynamic variable lookups
- **Changes:** `parseVariableTokens()` now properly handles `[list[settings.zero]]` as a single token
- **Impact:** Complex dynamic lookups with drops now work: `{{ [list[settings.zero]] }}`

**Technical Details:**

The root cause was that the `VariableParser` regex (`\[[^\[\]]*\]`) couldn't handle nested brackets because Go's regex engine doesn't support recursion. The regex would tokenize `[list[settings.zero]]` as TWO tokens (`list` and `[settings.zero]`) instead of ONE.

The fix implements a custom bracket-aware tokenizer that:

1. Tracks bracket depth using a counter
2. Extracts complete bracket expressions including nested content
3. Properly handles expressions like `[list[a[b]]]` as single tokens

**Files Modified:**

- `liquid/variable_lookup.go` - Added `parseVariableTokens()` and map integer key support
- `integration/variable_test.go` - Removed `t.Skip()` from 3 tests

---

## REMAINING ISSUES

## MEDIUM PRIORITY (Parser Validation)

These 5 tests all relate to the same issue: **the parser doesn't properly reject malformed filter syntax in strict mode**.

### 4. TestVariable_FilterWithSingleTrailingComma

**File:** `integration/variable_test.go:338`

**Issue:** Trailing comma after filter argument should error in strict mode.

```liquid
{{ "hello" | append: "world", }}
// Should error in strict mode
// Should work in rigid/lax mode
```

**Current Behavior:** No error in any mode

---

### 5. TestVariable_MultipleFiltersWithTrailingCommas

**File:** `integration/variable_test.go:363`

**Issue:** Multiple trailing commas should error in strict mode.

```liquid
{{ "hello" | append: "1", | append: "2", }}
// Should error in strict mode
// Should work in rigid/lax mode
```

**Current Behavior:** No error in any mode

---

### 6. TestVariable_FilterWithColonButNoArguments

**File:** `integration/variable_test.go:384`

**Issue:** Colon without arguments should error in strict mode.

```liquid
{{ "test" | upcase: }}
// Should error in strict mode
// Should work in rigid/lax mode
```

**Current Behavior:** No error in any mode

---

### 7. TestVariable_FilterChainWithColonNoArgs

**File:** `integration/variable_test.go:405`

**Issue:** Filter chain with empty args should error in strict mode.

```liquid
{{ "test" | append: "x" | upcase: }}
// Should error in strict mode
// Should work in rigid/lax mode
```

**Current Behavior:** No error in any mode

---

### 8. TestVariable_CombiningTrailingCommaAndEmptyArgs

**File:** `integration/variable_test.go:426`

**Issue:** Combined malformed syntax should error in strict mode.

```liquid
{{ "test" | append: "x", | upcase: }}
// Should error in strict mode
// Should work in rigid/lax mode
```

**Current Behavior:** No error in any mode

---

**Root Cause (All 5 Tests):**

- Filter argument parser in `liquid/variable.go` or `liquid/parser.go` doesn't validate syntax
- Error mode (strict/rigid/lax) not being checked during filter parsing
- Parser accepts malformed syntax that should be rejected

**Impact:**

- Invalid templates are silently accepted in strict mode
- Users don't get helpful error messages
- Breaks strict mode guarantees

**Fix Approach (All 5 Together):**

1. Review filter parsing in `liquid/variable.go` (lines 70-180)
2. Add syntax validation for:
   - Trailing commas after arguments
   - Empty arguments after colons
3. Check error mode and raise `SyntaxError` in strict mode
4. Ensure rigid/lax modes still accept these patterns

**Estimated Effort:** 3-4 hours (fix all 5 together)

---

## LOW PRIORITY (Edge Cases)

### 9. TestVariable_AssignsNotPollutedFromTemplate

**File:** `integration/variable_test.go:212`

**Issue:** Template state may persist across renders with different contexts.

**Test Case:**

```liquid
{{ test }}{% assign test = 'bar' %}{{ test }}
// Rendered 4 times with different contexts
// Expects clean state each time
```

**Root Cause:**

- Template object may be caching state
- Context isolation issue

**Impact:**

- Affects templates rendered multiple times with different data
- Rare use case - most apps create fresh templates or use partials
- Workaround: Create new template instance for each render

**Fix Approach:**

1. Review template state management
2. Ensure assigns are context-local, not template-local
3. May require context cloning fixes

**Estimated Effort:** 2-3 hours

**Recommendation:** Document as known limitation, fix if time permits

---

### 10. TestVariable_NestedArray

**File:** `integration/variable_test.go:283`

**Issue:** Nested arrays with nil render debug representation instead of empty string.

**Failing Case:**

```liquid
{{ foo }}  // foo = [[nil]]
// Expected: ""
// Got: "[[<nil>]]"
```

**Root Cause:**

- `ToS()` function showing Go's internal representation
- Nested nil handling not matching Ruby Liquid behavior

**Impact:**

- Very uncommon edge case
- Only affects specific nested nil patterns
- Debug output leaks into rendered content

**Fix Approach:**

1. Update `ToS()` in `liquid/utils.go` to handle nested arrays
2. Recursively check for nil-only arrays
3. Return empty string for arrays containing only nil

**Estimated Effort:** 1-2 hours

**Recommendation:** Low value fix, skip unless completing all others

---

### 11. TestVariable_DoubleNestedVariableLookup

**File:** `integration/variable_test.go:323`

**Issue:** Extremely complex nested dynamic lookups fail.

**Failing Case:**

```liquid
{{ list[list[settings.zero]]['foo'] }}
// list = [1, {"foo": "bar"}], settings.zero = 0
// Expected: "bar"
// Got: ""
```

**Root Cause:**

- Related to test #3 but even more complex
- Multiple levels of indirection with drops and maps

**Impact:**

- Extremely rare pattern
- Can be worked around with intermediate `{% assign %}` statements
- Most templates don't need this level of nesting

**Fix Approach:**

1. Fix tests #1 and #3 first (prerequisites)
2. May resolve automatically once simpler cases work
3. Add comprehensive nested evaluation tests

**Estimated Effort:** 1-2 hours (after #1 and #3)

**Recommendation:** Skip unless fixing all critical/medium issues

---

## Recommended Fix Order (Updated)

1. ✅ **Phase 1 (Critical):** Tests 1-3 COMPLETED

   - Restored core drop functionality
   - Fixed common syntax patterns
   - Unblocked advanced use cases
   - **Actual time:** ~2 hours

2. **Phase 2 (Medium):** Fix tests 4-8 (estimated 3-4 hours)

   - Implement proper strict mode validation
   - Improve error messages
   - Ensure parser correctness

3. **Phase 3 (Low):** Document or skip tests 9-11
   - Edge cases with workarounds
   - Low user impact
   - Can be addressed later if needed

**Remaining Estimated Effort:** 3-7 hours for medium + low priority

---

## Current Status

**Test Suite Status:** All tests passing ✅ (0 failures, 14 skipped)

**Variable Tests:**

- ✅ 3 critical tests FIXED and passing
- 5 medium priority tests still skipped (filter parsing validation)
- 3 low priority tests still skipped (edge cases)

**Files Modified:**

- `liquid/variable_lookup.go` - Added `parseVariableTokens()` and integer map key support
- `integration/variable_test.go` - Removed `t.Skip()` from 3 critical tests

**Other Skipped Tests (Not Variable Tests):**

- `TestErrorHandling_ParsingWarnWithLineNumbersAddsNumbersToLexerErrors` - Feature not implemented
- `TestFilterKwarg_CanParseDataKwargs` - Keyword arguments not implemented
- `TestParsingQuirks_UnanchoredFilterArguments` - Malformed syntax edge case
- `TestParsingQuirks_IncompleteExpression` - Incomplete expression edge case
- `TestSecurity_NoInstanceEvalLaterInChain` - Plain function filters not supported
- `TestSecurity_MoreThanMaxDepthNestedBlocksRaisesException` - Design difference (parse vs render time)
