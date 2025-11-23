# Implementation Gaps

## Not Implemented

1. **Filter Keyword Arguments** - `{{ x | filter: key: value }}` syntax not supported
2. **Warning Collection** - Warnings in `warn` mode not collected properly
3. **Error Handling in Drops** - Errors from drop methods not caught/handled correctly
4. **Arithmetic Expressions** - `{{ 5 + 3 }}` not working in output tags
5. **Logical Operators** - `and`, `or`, `not` causing panics in some contexts
6. **Error Message Formatting** - Line numbers and template names not always included
7. **Strict Mode Parsing** - Some syntax errors not caught in strict mode
8. **Exception Renderer** - Custom exception renderers not invoked properly

## Test Results Summary

- **~26 tests failing** - Mostly error handling and drop-related
- **~3 tests skipped** - Warning collection features
- **~10 tests passing** - Basic expressions, comparisons, parsing

## Notes

- Core variable resolution works
- Comparison operators work
- Security boundaries enforced
- Most parsing edge cases handled
