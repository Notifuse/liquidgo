# Liquid Go Implementation - Detailed Guide

## File Mapping Reference

This document provides a comprehensive mapping between Ruby files and their Go equivalents.

### Core Library Files

| Ruby File                          | Go File                        | Description                     |
| ---------------------------------- | ------------------------------ | ------------------------------- |
| `lib/liquid.rb`                    | `liquid/liquid.go`             | Main package entry point        |
| `lib/liquid/version.rb`            | `liquid/version.go`            | Version constant                |
| `lib/liquid/template.rb`           | `liquid/template.go`           | Template class - main API       |
| `lib/liquid/environment.rb`        | `liquid/environment.go`        | Environment configuration       |
| `lib/liquid/parser.rb`             | `liquid/parser.go`             | Template parser                 |
| `lib/liquid/lexer.rb`              | `liquid/lexer.go`              | Lexical analyzer                |
| `lib/liquid/tokenizer.rb`          | `liquid/tokenizer.go`          | Tokenizer                       |
| `lib/liquid/context.rb`            | `liquid/context.go`            | Rendering context               |
| `lib/liquid/variable.rb`           | `liquid/variable.go`           | Variable handling               |
| `lib/liquid/variable_lookup.rb`    | `liquid/variable_lookup.go`    | Variable lookup logic           |
| `lib/liquid/expression.rb`         | `liquid/expression.go`         | Expression evaluation           |
| `lib/liquid/condition.rb`          | `liquid/condition.go`          | Conditional logic               |
| `lib/liquid/block.rb`              | `liquid/block.go`              | Block tag base class            |
| `lib/liquid/block_body.rb`         | `liquid/block_body.go`         | Block body handling             |
| `lib/liquid/document.rb`           | `liquid/document.go`           | Document root node              |
| `lib/liquid/tag.rb`                | `liquid/tag.go`                | Tag base class                  |
| `lib/liquid/drop.rb`               | `liquid/drop.go`               | Drop base class                 |
| `lib/liquid/errors.rb`             | `liquid/errors.go`             | Error types                     |
| `lib/liquid/file_system.rb`        | `liquid/file_system.go`        | File system abstraction         |
| `lib/liquid/standardfilters.rb`    | `liquid/standardfilters.go`    | Standard filter implementations |
| `lib/liquid/i18n.rb`               | `liquid/i18n.go`               | Internationalization            |
| `lib/liquid/registers.rb`          | `liquid/registers.go`          | Template registers              |
| `lib/liquid/resource_limits.rb`    | `liquid/resource_limits.go`    | Resource limiting               |
| `lib/liquid/partial_cache.rb`      | `liquid/partial_cache.go`      | Partial template caching        |
| `lib/liquid/parse_context.rb`      | `liquid/parse_context.go`      | Parse-time context              |
| `lib/liquid/parse_tree_visitor.rb` | `liquid/parse_tree_visitor.go` | AST visitor pattern             |
| `lib/liquid/parser_switching.rb`   | `liquid/parser_switching.go`   | Parser mode switching           |
| `lib/liquid/profiler.rb`           | `liquid/profiler.go`           | Performance profiling           |
| `lib/liquid/profiler/hooks.rb`     | `liquid/profiler/hooks.go`     | Profiler hooks                  |
| `lib/liquid/range_lookup.rb`       | `liquid/range_lookup.go`       | Range lookup logic              |
| `lib/liquid/snippet_drop.rb`       | `liquid/snippet_drop.go`       | Snippet drop                    |
| `lib/liquid/forloop_drop.rb`       | `liquid/forloop_drop.go`       | For loop drop                   |
| `lib/liquid/tablerowloop_drop.rb`  | `liquid/tablerowloop_drop.go`  | Table row loop drop             |
| `lib/liquid/strainer_template.rb`  | `liquid/strainer_template.go`  | Filter strainer                 |
| `lib/liquid/tags.rb`               | `liquid/tags.go`               | Tag registry                    |
| `lib/liquid/template_factory.rb`   | `liquid/template_factory.go`   | Template factory                |
| `lib/liquid/usage.rb`              | `liquid/usage.go`              | Usage tracking                  |
| `lib/liquid/utils.rb`              | `liquid/utils.go`              | Utility functions               |
| `lib/liquid/const.rb`              | `liquid/const.go`              | Constants                       |
| `lib/liquid/deprecations.rb`       | `liquid/deprecations.go`       | Deprecation warnings            |
| `lib/liquid/extensions.rb`         | `liquid/extensions.go`         | Extension points                |
| `lib/liquid/interrupts.rb`         | `liquid/interrupts.go`         | Interrupt handling              |

### Tag Files

| Ruby File                           | Go File                         | Description        |
| ----------------------------------- | ------------------------------- | ------------------ |
| `lib/liquid/tags/assign.rb`         | `liquid/tags/assign.go`         | Assign tag         |
| `lib/liquid/tags/break.rb`          | `liquid/tags/break.go`          | Break tag          |
| `lib/liquid/tags/capture.rb`        | `liquid/tags/capture.go`        | Capture tag        |
| `lib/liquid/tags/case.rb`           | `liquid/tags/case.go`           | Case tag           |
| `lib/liquid/tags/comment.rb`        | `liquid/tags/comment.go`        | Comment tag        |
| `lib/liquid/tags/continue.rb`       | `liquid/tags/continue.go`       | Continue tag       |
| `lib/liquid/tags/cycle.rb`          | `liquid/tags/cycle.go`          | Cycle tag          |
| `lib/liquid/tags/decrement.rb`      | `liquid/tags/decrement.go`      | Decrement tag      |
| `lib/liquid/tags/doc.rb`            | `liquid/tags/doc.go`            | Doc tag            |
| `lib/liquid/tags/echo.rb`           | `liquid/tags/echo.go`           | Echo tag           |
| `lib/liquid/tags/for.rb`            | `liquid/tags/for.go`            | For tag            |
| `lib/liquid/tags/if.rb`             | `liquid/tags/if.go`             | If tag             |
| `lib/liquid/tags/ifchanged.rb`      | `liquid/tags/ifchanged.go`      | Ifchanged tag      |
| `lib/liquid/tags/include.rb`        | `liquid/tags/include.go`        | Include tag        |
| `lib/liquid/tags/increment.rb`      | `liquid/tags/increment.go`      | Increment tag      |
| `lib/liquid/tags/inline_comment.rb` | `liquid/tags/inline_comment.go` | Inline comment tag |
| `lib/liquid/tags/raw.rb`            | `liquid/tags/raw.go`            | Raw tag            |
| `lib/liquid/tags/render.rb`         | `liquid/tags/render.go`         | Render tag         |
| `lib/liquid/tags/snippet.rb`        | `liquid/tags/snippet.go`        | Snippet tag        |
| `lib/liquid/tags/table_row.rb`      | `liquid/tags/table_row.go`      | Table row tag      |
| `lib/liquid/tags/unless.rb`         | `liquid/tags/unless.go`         | Unless tag         |

### Tag Base Classes

| Ruby File                       | Go File                     | Description           |
| ------------------------------- | --------------------------- | --------------------- |
| `lib/liquid/tag/disableable.rb` | `liquid/tag/disableable.go` | Disableable tag mixin |
| `lib/liquid/tag/disabler.rb`    | `liquid/tag/disabler.go`    | Tag disabler          |

### Test Files

| Ruby Test File                             | Go Test File                                 | Description                   |
| ------------------------------------------ | -------------------------------------------- | ----------------------------- |
| `test/integration/template_test.rb`        | `liquid/template_integration_test.go`        | Template integration tests    |
| `test/integration/assign_test.rb`          | `liquid/assign_integration_test.go`          | Assign tag integration tests  |
| `test/integration/block_test.rb`           | `liquid/block_integration_test.go`           | Block integration tests       |
| `test/integration/capture_test.rb`         | `liquid/capture_integration_test.go`         | Capture tag integration tests |
| `test/integration/context_test.rb`         | `liquid/context_integration_test.go`         | Context integration tests     |
| `test/integration/document_test.rb`        | `liquid/document_integration_test.go`        | Document integration tests    |
| `test/integration/drop_test.rb`            | `liquid/drop_integration_test.go`            | Drop integration tests        |
| `test/integration/error_handling_test.rb`  | `liquid/error_handling_integration_test.go`  | Error handling tests          |
| `test/integration/expression_test.rb`      | `liquid/expression_integration_test.go`      | Expression integration tests  |
| `test/integration/filter_test.rb`          | `liquid/filter_integration_test.go`          | Filter integration tests      |
| `test/integration/filter_kwarg_test.rb`    | `liquid/filter_kwarg_integration_test.go`    | Filter keyword arg tests      |
| `test/integration/standard_filter_test.rb` | `liquid/standard_filter_integration_test.go` | Standard filter tests         |
| `test/integration/tag_test.rb`             | `liquid/tag_integration_test.go`             | Tag integration tests         |
| `test/integration/variable_test.rb`        | `liquid/variable_integration_test.go`        | Variable integration tests    |
| `test/unit/template_unit_test.rb`          | `liquid/template_unit_test.go`               | Template unit tests           |
| `test/unit/block_unit_test.rb`             | `liquid/block_unit_test.go`                  | Block unit tests              |
| `test/unit/condition_unit_test.rb`         | `liquid/condition_unit_test.go`              | Condition unit tests          |
| `test/unit/environment_test.rb`            | `liquid/environment_unit_test.go`            | Environment unit tests        |
| `test/unit/lexer_unit_test.rb`             | `liquid/lexer_unit_test.go`                  | Lexer unit tests              |
| `test/unit/parser_unit_test.rb`            | `liquid/parser_unit_test.go`                 | Parser unit tests             |
| `test/unit/tokenizer_unit_test.rb`         | `liquid/tokenizer_unit_test.go`              | Tokenizer unit tests          |
| `test/unit/variable_unit_test.rb`          | `liquid/variable_unit_test.go`               | Variable unit tests           |

## Implementation Patterns

### Ruby to Go Translation Patterns

#### Classes → Structs

```ruby
# Ruby
class Template
  attr_accessor :root, :name
  def initialize
    @root = nil
  end
end
```

```go
// Go
type Template struct {
    Root *Document
    Name string
}

func NewTemplate() *Template {
    return &Template{}
}
```

#### Modules → Interfaces or Packages

```ruby
# Ruby
module Liquid
  module StandardFilters
    def upcase(input)
      input.to_s.upcase
    end
  end
end
```

```go
// Go
package liquid

type StandardFilters struct{}

func (f *StandardFilters) Upcase(input interface{}) interface{} {
    // implementation
}
```

#### Methods → Methods or Functions

```ruby
# Ruby
class Template
  def parse(source)
    # ...
  end
end
```

```go
// Go
func (t *Template) Parse(source string) error {
    // implementation
}
```

#### Error Handling

```ruby
# Ruby
raise Liquid::SyntaxError, "Invalid syntax"
```

```go
// Go
return nil, &SyntaxError{Message: "Invalid syntax"}
```

#### Constants

```ruby
# Ruby
module Liquid
  VERSION = "5.10.0"
  TagStart = /\{\%/
end
```

```go
// Go
const Version = "5.10.0"
var TagStart = regexp.MustCompile(`\{\%`)
```

## Testing Strategy

### Test Structure

- Use Go's standard `testing` package
- Mirror test cases from Ruby implementation
- Use table-driven tests for multiple cases
- Integration tests verify end-to-end behavior
- Unit tests verify individual component behavior

### Test Helper Functions

Create helper functions similar to Ruby's `assert_template_result`:

```go
func assertTemplateResult(t *testing.T, expected, template string, assigns map[string]interface{}) {
    // implementation
}
```

### Test Coverage

- Aim for 100% parity with Ruby test coverage
- Each Ruby test should have a corresponding Go test
- Test edge cases and error conditions

## Applying Changelog Updates

### Process

1. **Read History.md**: Check `reference-liquid/History.md` for new releases
2. **Identify Changes**: Note which files were modified in the changelog
3. **Check Git History**: Use `git log` or `git diff` in `reference-liquid/` to see exact changes
4. **Map Files**: Use the file mapping table to find Go equivalents
5. **Implement**: Apply changes maintaining same behavior
6. **Update Version**: Update `liquid/version.go` to match release version
7. **Update Tests**: Add/modify tests as needed
8. **Verify**: Run tests and compare behavior with Ruby version

### Example: Applying Version 5.10.0 Changes

From History.md:

```
## 5.10.0
* Introduce support for Inline Snippets
```

Steps:

1. Check git diff for version 5.10.0 in `reference-liquid/`
2. Identify modified files (likely `lib/liquid/tags/snippet.rb`)
3. Map to `liquid/tags/snippet.go`
4. Implement inline snippet support
5. Update version to "5.10.0"
6. Add tests for inline snippets

## Architecture Decisions

### Package Structure

- Single `liquid` package for core functionality
- Sub-packages for tags: `liquid/tags`
- Sub-packages for tag base classes: `liquid/tag`
- Sub-packages for profiler: `liquid/profiler`

### Error Handling

- Use Go error interface
- Create custom error types matching Ruby exceptions
- Return errors from functions, don't panic (unless appropriate)

### API Design

- Maintain similar public API to Ruby version
- Use Go naming conventions (exported names start with capital)
- Provide both methods and functions where appropriate

### Performance Considerations

- Use Go's strengths (goroutines, channels) where beneficial
- Maintain performance characteristics similar to Ruby version
- Profile and optimize as needed

## Version Tracking

Version is stored in `liquid/version.go`:

```go
const Version = "5.10.0"
```

When updating:

1. Check `reference-liquid/lib/liquid/version.rb` for new version
2. Update `liquid/version.go` to match exactly
3. Document changes in commit message

## Reference Files

Always refer to these Ruby files when implementing:

- `reference-liquid/lib/liquid/` - Core implementation
- `reference-liquid/test/` - Test cases
- `reference-liquid/History.md` - Changelog
- `reference-liquid/README.md` - Documentation
