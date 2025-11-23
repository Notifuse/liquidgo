# Integration Test Coverage Gap Analysis

Last updated: After fixing filter optional parameters bug (2024)

## Coverage Statistics

**Ruby Liquid (reference implementation):**
- 40 test files
- 774 test methods (`def test_*`)
- 846 `assert_template_result` calls (integration tests)

**liquidgo (current):**
- 12 test files  
- 98 test functions (`func Test*`)
- ~200-300 template assertion tests

**Coverage Gap:** liquidgo has approximately **15-20%** of Ruby Liquid's integration test coverage

---

## Missing Test Categories

### 🔴 High Priority (Core Features - Security & Correctness)

#### 1. **error_handling_test.rb** ❌ MISSING
- Error modes (strict, lax, warn)
- Syntax error handling
- Runtime error handling  
- Error message formatting
- **Risk:** Users may encounter uncaught errors

#### 2. **security_test.rb** ❌ MISSING
- FileSystem access restrictions
- Template injection prevention
- Resource limits enforcement
- Malicious template handling
- **Risk:** Security vulnerabilities

#### 3. **filter_kwarg_test.rb** ❌ MISSING
- Keyword arguments in filters
- Example: `{{ x | default: "val", allow_false: true }}`
- **Status:** Keywords parsed but not passed to filters
- **Risk:** Missing Ruby Liquid feature

#### 4. **variable_test.rb** ❌ MISSING
- Variable parsing edge cases
- Filter chain tests (`{{ x | a | b | c }}`)
- Complex expressions
- Variable lookup precedence
- **Risk:** Subtle bugs in variable resolution

#### 5. **expression_test.rb** ❌ MISSING
- Arithmetic expressions (`{{ 5 + 3 * 2 }}`)
- Comparison operators (`==`, `!=`, `<`, `>`, etc.)
- Logical operators (`and`, `or`, `not`)
- Operator precedence rules
- **Risk:** Wrong calculation results

#### 6. **output_test.rb** ❌ MISSING
- Output tag behavior
- Whitespace handling
- Escaping behavior in different contexts
- **Risk:** Incorrect template output

#### 7. **parsing_quirks_test.rb** ❌ MISSING
- Edge cases in template parsing
- Malformed template handling
- Recovery from parse errors
- **Risk:** Crashes on edge case templates

#### 8. **hash_ordering_test.rb** ❌ MISSING
- Hash/map iteration order stability
- Predictable output for dictionaries
- **Risk:** Non-deterministic output

#### 9. **hash_rendering_test.rb** ❌ MISSING
- Rendering hash/map values
- Nested hashes
- Hash with filters
- **Risk:** Incorrect hash handling

#### 10. **drop_test.rb** ❌ MISSING
- Drop objects (lazy loading pattern)
- Custom drop implementations
- Context-aware drops
- **Risk:** Advanced feature not tested

---

### 🟡 Medium Priority (Tag Coverage)

#### 11. **tags/standard_tag_test.rb** ❌ MISSING
- Comprehensive tests for all standard tags
- Edge cases for each tag

#### 12. **tags/if_else_tag_test.rb** ⚠️ PARTIAL
- **Exists:** `comprehensive_test.go` has basic if/else
- **Missing:** Complex conditionals, nested if/else, all operators
- **Need:** Dedicated test file with 50+ test cases

#### 13. **tags/for_tag_test.rb** ⚠️ PARTIAL  
- **Exists:** `detailed_forloop_test.go` covers some cases
- **Missing:** All forloop object properties, break/continue edge cases
- **Need:** More comprehensive coverage

#### 14. **tags/include_tag_test.rb** ❌ MISSING
- Include tag with variables
- Include with for loops
- Nested includes

#### 15. **tags/render_tag_test.rb** ❌ MISSING
- Render tag (different from include)
- Variable scoping in render
- Render with for loops

#### 16. **tags/cycle_tag_test.rb** ❌ MISSING
- Cycle tag basic usage
- Named cycle groups
- Cycle in loops

#### 17. **tags/increment_tag_test.rb** ❌ MISSING
- Increment/decrement counters
- Persistence across renders

#### 18. **tags/liquid_tag_test.rb** ❌ MISSING
- Liquid tag behavior
- Edge cases

#### 19. **tags/snippet_test.rb** ❌ MISSING
- Snippet tag tests

#### 20. **tags/inline_comment_test.rb** ❌ MISSING
- Comment syntax variations
- Multi-line comments

---

### 🟡 Medium Priority (Filter Coverage)

#### 21. **standard_filter_test.rb** ⚠️ PARTIAL (~30% coverage)

**String Filters Missing Template Tests:**
- `slice` with negative indices, edge cases
- `truncate` with various lengths and ellipsis
- `truncatewords` edge cases
- `split` with regex patterns
- `replace_first`, `replace_last`, `remove_first`, `remove_last`
- `append`, `prepend` edge cases
- `newline_to_br` with various newline types
- `strip_newlines`
- `escape_once` with already escaped content
- `url_encode`, `url_decode` with special characters
- `base64_encode`, `base64_decode` with unicode
- `base64_url_safe_encode`, `base64_url_safe_decode`

**Array Filters Missing Template Tests:**
- `concat` with various array types
- `map` with nested properties, edge cases
- `sum` with/without property parameter
- `reverse` with various types
- `uniq` by property (more test cases needed)
- `compact` by property (more test cases needed)

**Math Filters Missing Template Tests:**
- `abs` with negative numbers
- `ceil`, `floor` with edge cases
- `round` with precision parameter
- `plus`, `minus`, `times`, `divided_by`, `modulo` edge cases
- `at_least`, `at_most` with various number types

**Date Filters Missing Template Tests:**
- `date` with various format strings
- Date parsing edge cases
- Timezone handling

---

### 🟢 Low Priority (Less Common)

#### 22. **profiler_test.rb** ❌ MISSING
- Performance profiling
- Render time tracking

#### 23. **context_test.rb** ⚠️ MINIMAL
- Context manipulation
- Variable scoping rules
- Register access patterns

#### 24. **document_test.rb** ❌ MISSING
- Document object tests
- Template metadata

#### 25. **blank_test.rb** ⚠️ PARTIAL
- Comprehensive blank/empty value tests
- **Exists:** Some coverage in `comprehensive_test.go`

#### 26. **filter_test.rb** ⚠️ PARTIAL
- Custom filter registration
- Filter precedence rules
- **Exists:** Basic coverage in `helper_test.go`

---

## Test Organization Recommendations

### Create New Test Files

```
liquidgo/integration/
├── error_handling_test.go          ❌ NEW - Critical
├── security_test.go                 ❌ NEW - Critical
├── variable_test.go                 ❌ NEW - Important
├── expression_test.go               ❌ NEW - Important
├── output_test.go                   ❌ NEW - Important
├── drop_test.go                     ❌ NEW - Advanced features
├── context_test.go                  ❌ NEW - Core behavior
├── filter_edge_cases_test.go        ❌ NEW - Robustness
├── hash_test.go                     ❌ NEW - Data structure handling
├── parsing_test.go                  ❌ NEW - Parser edge cases
│
└── tags/                            ❌ NEW DIRECTORY
    ├── if_else_comprehensive_test.go
    ├── for_comprehensive_test.go
    ├── include_test.go
    ├── render_test.go
    ├── cycle_test.go
    ├── increment_test.go
    ├── liquid_tag_test.go
    ├── snippet_test.go
    └── comment_test.go
```

### Expand Existing Test Files

- `filter_optional_params_test.go` → Add more edge cases
- `comprehensive_test.go` → Add more tag combination tests
- `trim_mode_test.go` → More whitespace scenarios
- `detailed_forloop_test.go` → Complete forloop coverage

---

## Implementation Priority

### Phase 1: Critical Gaps (Weeks 1-2)
**Goal:** Catch security issues and critical bugs

1. ✅ **Filter optional parameters** - COMPLETED!
2. 🔴 **Error handling tests** (~100 tests)
   - Prevents crashes and improves error messages
3. 🔴 **Security tests** (~50 tests)
   - Prevents template injection, DoS, etc.
4. 🔴 **Variable/Expression tests** (~80 tests)
   - Core template evaluation correctness

**Estimated:** ~230 new tests

### Phase 2: Feature Completeness (Weeks 3-4)
**Goal:** Achieve Ruby Liquid parity

5. 🟡 **Filter keyword arguments** (~30 tests)
   - Implementation + tests for kwargs
6. 🟡 **All standard tag tests** (~150 tests)
   - Comprehensive tag behavior
7. 🟡 **Drop object tests** (~40 tests)
   - Advanced feature support
8. 🟡 **Hash/Output tests** (~60 tests)
   - Data structure handling

**Estimated:** ~280 new tests

### Phase 3: Edge Cases & Polish (Week 5+)
**Goal:** Production-ready robustness

9. 🟢 **Filter edge cases** (~100 tests)
   - All filters tested in templates
10. 🟢 **Parsing quirks** (~40 tests)
    - Malformed template handling
11. 🟢 **Context/Profiler tests** (~50 tests)
    - Advanced features

**Estimated:** ~190 new tests

---

## Overall Estimates

**To achieve Ruby Liquid test parity:**
- **New tests needed:** ~500-600 integration tests
- **New test files:** ~25-30 files
- **Time estimate:** 4-6 weeks of focused work
- **Current coverage:** 15-20%
- **Target coverage:** 90%+ (match Ruby Liquid)

---

## Why This Matters

### Bugs Caught by Integration Tests (Not Unit Tests)

1. **Filter optional parameters bug** (just fixed!)
   - Unit tests passed ✅
   - Integration tests would have caught it ❌
   - Impact: Multiple filters were broken

2. **Potential filter keyword argument bug**
   - Keywords are parsed but not invoked
   - No integration test to catch it
   - Would break: `{{ x | default: "val", allow_false: true }}`

3. **Future bugs prevented:**
   - Variable resolution changes
   - Tag behavior modifications
   - Filter invocation changes
   - Parser modifications

### Best Practice

> **Golden Rule:** If a user would type it in a template, there should be an integration test for it.

---

## Next Steps

1. **Read Ruby Liquid tests** - Port test cases systematically
2. **Prioritize by risk** - Security and correctness first
3. **Test coverage metrics** - Track progress toward 90%
4. **CI/CD integration** - Run all tests on every commit
5. **Documentation** - Link tests to features in docs

---

## References

- Ruby Liquid integration tests: `reference-liquid/test/integration/`
- Ruby Liquid documentation: https://shopify.github.io/liquid/
- Shopify Liquid docs: https://shopify.dev/docs/api/liquid
- liquidgo testing guide: `integration/TESTING_GUIDE.md`

---

**Status:** This analysis completed after fixing the filter optional parameters bug. We've improved from ~13% to ~20% coverage with the new `filter_optional_params_test.go` file.

