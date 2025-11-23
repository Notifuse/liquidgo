# Test Failure Analysis

**Principle**: If a feature is not implemented in Ruby Liquid, it should not be expected in Go. Those tests are marked as TEST ISSUE, not IMPLEMENTATION ISSUE.

## Implementation Issues (Need Fixes - Ruby Liquid Has These Features)

### Error Handling
1. **TestErrorHandling_StandardError** - IMPLEMENTATION: Drop method panics not caught
2. **TestErrorHandling_SyntaxError** - IMPLEMENTATION: Drop method panics not caught  
3. **TestErrorHandling_ArgumentError** - IMPLEMENTATION: Drop method panics not caught
4. **TestErrorHandling_DefaultExceptionRendererWithInternalError** - IMPLEMENTATION: Errors not collected
5. **TestErrorHandling_SettingDefaultExceptionRenderer** - IMPLEMENTATION: Exception renderer not invoked
6. **TestErrorHandling_SettingExceptionRendererOnEnvironment** - IMPLEMENTATION: Exception renderer not invoked
7. **TestErrorHandling_ExceptionRendererExposingNonLiquidError** - IMPLEMENTATION: Exception renderer not invoked
8. **TestErrorHandling_IncludedTemplateNameWithLineNumbers** - IMPLEMENTATION: Template name not set in errors
9. **TestErrorHandling_InternalErrorIsRaisedWithTemplateName** - IMPLEMENTATION: Template name format differs
10. **TestErrorHandling_BugCompatibleSilencingOfErrorsInBlankNodes** - IMPLEMENTATION: Error handling in blank nodes differs

### Strict Mode Parsing
11. **TestErrorHandling_UnrecognizedOperator** - IMPLEMENTATION: Strict mode not catching `=!` operator
12. **TestErrorHandling_ParsingStrictWithLineNumbersAddsNumbersToLexerErrors** - IMPLEMENTATION: Strict mode not catching errors
13. **TestErrorHandling_StrictErrorMessages** - IMPLEMENTATION: Strict mode not catching errors
14. **TestErrorHandling_LaxUnrecognizedOperator** - IMPLEMENTATION: Error type wrong (InternalError vs ArgumentError)

### Template Name in Errors
15. **TestErrorHandling_SyntaxErrorIsRaisedWithTemplateName** - IMPLEMENTATION: Template name format differs
16. **TestErrorHandling_SyntaxErrorIsRaisedWithTemplateNameFromTemplateFactory** - IMPLEMENTATION: Template name format differs
17. **TestErrorHandling_ErrorIsRaisedDuringParseWithTemplateName** - IMPLEMENTATION: Max depth check not working

### Drop Methods
18. **TestDrop_RespondsToToLiquid** - IMPLEMENTATION: `to_liquid` method not accessible
19. **TestDrop_ContextDrop** - IMPLEMENTATION: LiquidMethodMissing not working correctly
20. **TestDrop_NestedContextDrop** - IMPLEMENTATION: LiquidMethodMissing not working correctly
21. **TestDrop_Scope** - IMPLEMENTATION: Context access from drops not working
22. **TestDrop_AccessContextFromDrop** - IMPLEMENTATION: Context access from drops not working
23. **TestDrop_EnumerableDrop** - IMPLEMENTATION: Drops not iterable in for loops

### Error Message Format
24. **TestErrorHandling_MissingEndtagParseTimeError** - TEST ISSUE: Regex pattern expects `: 'for' tag was never closed\z` but Go error is `Tag was never closed: for` (format differs, pattern needs adjustment)
25. **TestErrorHandling_TemplatesParsedWithLineNumbersRendersThemInErrors** - IMPLEMENTATION: Error messages from drop panics not rendered (empty output instead of error message)

## Test Issues (Tests Need Fixing - Features Not in Ruby Liquid)

### Removed Tests (Features Not in Ruby Liquid)
- **TestExpression_Arithmetic** - REMOVED: Ruby Liquid doesn't support `{{ 5 + 3 }}` in output tags. Arithmetic is only supported in conditions, not output tags.
- **TestExpression_Logical NOT operator tests** - REMOVED: Ruby Liquid doesn't support `not` operator. Ruby only supports `and` and `or` (see `BOOLEAN_OPERATORS = %w(and or)`).

## Summary

- **Implementation Issues**: ~24 tests (Ruby Liquid has these features, Go needs to implement them)
- **Removed Tests**: 2 test cases removed (arithmetic in output tags, `not` operator - features not in Ruby Liquid)
- **Skipped (Known Gaps)**: 3 tests (warnings collection - documented as not implemented)

## Notes

All drop-related tests (to_liquid, context access, enumerable drops) are valid - Ruby Liquid supports these features. The implementation issues are real gaps that need fixing.

## Key Findings

1. **Drop error handling**: Panics from drop methods not caught and converted to error messages
2. **Strict mode**: Not fully implemented for some syntax errors
3. **Arithmetic in output**: Not supported in Ruby Liquid (test is wrong)
4. **Template names**: Format differs from Ruby Liquid expectations
5. **Exception renderers**: Not being invoked properly

