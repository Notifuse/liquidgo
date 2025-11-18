# File Mapping Reference: Ruby → Go

This document provides a complete mapping between Ruby files in `reference-liquid/lib/liquid/` and their Go equivalents in `liquidgo/liquid/`.

## Quick Reference Rules

1. **Base filename matches exactly** (only `.rb` → `.go` extension changes)
2. **Directory structure mirrors Ruby** (`lib/liquid/` → `liquid/`)
3. **Test files**: `*_test.rb` → `*_test.go` (with `_integration` or `_unit` suffix as needed)

## Core Library Files

| Ruby Path                    | Go Path                  | Notes                                      |
| ---------------------------- | ------------------------ | ------------------------------------------ |
| `lib/liquid.rb`              | `liquid/liquid.go`       | Main package entry point                   |
| `lib/liquid/version.rb`      | `liquid/version.go`      | Version constant - MUST match Ruby version |
| `lib/liquid/const.rb`        | `liquid/const.go`        | Constants and regex patterns               |
| `lib/liquid/deprecations.rb` | `liquid/deprecations.go` | Deprecation warning system                 |

### Template & Parsing

| Ruby Path                          | Go Path                        | Notes                             |
| ---------------------------------- | ------------------------------ | --------------------------------- |
| `lib/liquid/template.rb`           | `liquid/template.go`           | Main Template class - primary API |
| `lib/liquid/template_factory.rb`   | `liquid/template_factory.go`   | Template factory pattern          |
| `lib/liquid/parser.rb`             | `liquid/parser.go`             | Template parser                   |
| `lib/liquid/parser_switching.rb`   | `liquid/parser_switching.go`   | Parser mode switching             |
| `lib/liquid/lexer.rb`              | `liquid/lexer.go`              | Lexical analyzer                  |
| `lib/liquid/tokenizer.rb`          | `liquid/tokenizer.go`          | Tokenizer                         |
| `lib/liquid/parse_context.rb`      | `liquid/parse_context.go`      | Parse-time context                |
| `lib/liquid/parse_tree_visitor.rb` | `liquid/parse_tree_visitor.go` | AST visitor pattern               |

### Context & Rendering

| Ruby Path                       | Go Path                     | Notes                     |
| ------------------------------- | --------------------------- | ------------------------- |
| `lib/liquid/context.rb`         | `liquid/context.go`         | Rendering context         |
| `lib/liquid/environment.rb`     | `liquid/environment.go`     | Environment configuration |
| `lib/liquid/registers.rb`       | `liquid/registers.go`       | Template registers        |
| `lib/liquid/resource_limits.rb` | `liquid/resource_limits.go` | Resource limiting         |

### Variables & Expressions

| Ruby Path                       | Go Path                     | Notes                 |
| ------------------------------- | --------------------------- | --------------------- |
| `lib/liquid/variable.rb`        | `liquid/variable.go`        | Variable handling     |
| `lib/liquid/variable_lookup.rb` | `liquid/variable_lookup.go` | Variable lookup logic |
| `lib/liquid/range_lookup.rb`    | `liquid/range_lookup.go`    | Range lookup          |
| `lib/liquid/expression.rb`      | `liquid/expression.go`      | Expression evaluation |
| `lib/liquid/condition.rb`       | `liquid/condition.go`       | Conditional logic     |

### Tags & Blocks

| Ruby Path                  | Go Path                | Notes                          |
| -------------------------- | ---------------------- | ------------------------------ |
| `lib/liquid/tag.rb`        | `liquid/tag.go`        | Tag base class                 |
| `lib/liquid/block.rb`      | `liquid/block.go`      | Block tag base class           |
| `lib/liquid/block_body.rb` | `liquid/block_body.go` | Block body handling            |
| `lib/liquid/document.rb`   | `liquid/document.go`   | Document root node             |
| `lib/liquid/tags.rb`       | `liquid/tags.go`       | Tag registry and standard tags |

### Drops

| Ruby Path                         | Go Path                       | Notes               |
| --------------------------------- | ----------------------------- | ------------------- |
| `lib/liquid/drop.rb`              | `liquid/drop.go`              | Drop base class     |
| `lib/liquid/forloop_drop.rb`      | `liquid/forloop_drop.go`      | For loop drop       |
| `lib/liquid/tablerowloop_drop.rb` | `liquid/tablerowloop_drop.go` | Table row loop drop |
| `lib/liquid/snippet_drop.rb`      | `liquid/snippet_drop.go`      | Snippet drop        |

### Filters

| Ruby Path                         | Go Path                       | Notes                           |
| --------------------------------- | ----------------------------- | ------------------------------- |
| `lib/liquid/standardfilters.rb`   | `liquid/standardfilters.go`   | Standard filter implementations |
| `lib/liquid/strainer_template.rb` | `liquid/strainer_template.go` | Filter strainer                 |

### File System & I18n

| Ruby Path                   | Go Path                 | Notes                         |
| --------------------------- | ----------------------- | ----------------------------- |
| `lib/liquid/file_system.rb` | `liquid/file_system.go` | File system abstraction       |
| `lib/liquid/i18n.rb`        | `liquid/i18n.go`        | Internationalization          |
| `lib/liquid/locales/en.yml` | `liquid/locales/en.yml` | English locale (keep as YAML) |

### Profiling & Utilities

| Ruby Path                      | Go Path                    | Notes                |
| ------------------------------ | -------------------------- | -------------------- |
| `lib/liquid/profiler.rb`       | `liquid/profiler.go`       | Performance profiler |
| `lib/liquid/profiler/hooks.rb` | `liquid/profiler/hooks.go` | Profiler hooks       |
| `lib/liquid/usage.rb`          | `liquid/usage.go`          | Usage tracking       |
| `lib/liquid/utils.rb`          | `liquid/utils.go`          | Utility functions    |

### Error Handling & Extensions

| Ruby Path                     | Go Path                   | Notes                      |
| ----------------------------- | ------------------------- | -------------------------- |
| `lib/liquid/errors.rb`        | `liquid/errors.go`        | Error types and exceptions |
| `lib/liquid/extensions.rb`    | `liquid/extensions.go`    | Extension points           |
| `lib/liquid/interrupts.rb`    | `liquid/interrupts.go`    | Interrupt handling         |
| `lib/liquid/partial_cache.rb` | `liquid/partial_cache.go` | Partial template caching   |

## Tag Implementations

All tags are in `lib/liquid/tags/` → `liquid/tags/`

| Ruby Path                           | Go Path                         | Tag Name             |
| ----------------------------------- | ------------------------------- | -------------------- |
| `lib/liquid/tags/assign.rb`         | `liquid/tags/assign.go`         | `assign`             |
| `lib/liquid/tags/break.rb`          | `liquid/tags/break.go`          | `break`              |
| `lib/liquid/tags/capture.rb`        | `liquid/tags/capture.go`        | `capture`            |
| `lib/liquid/tags/case.rb`           | `liquid/tags/case.go`           | `case`               |
| `lib/liquid/tags/comment.rb`        | `liquid/tags/comment.go`        | `comment`            |
| `lib/liquid/tags/continue.rb`       | `liquid/tags/continue.go`       | `continue`           |
| `lib/liquid/tags/cycle.rb`          | `liquid/tags/cycle.go`          | `cycle`              |
| `lib/liquid/tags/decrement.rb`      | `liquid/tags/decrement.go`      | `decrement`          |
| `lib/liquid/tags/doc.rb`            | `liquid/tags/doc.go`            | `doc`                |
| `lib/liquid/tags/echo.rb`           | `liquid/tags/echo.go`           | `echo`               |
| `lib/liquid/tags/for.rb`            | `liquid/tags/for.go`            | `for`                |
| `lib/liquid/tags/if.rb`             | `liquid/tags/if.go`             | `if`                 |
| `lib/liquid/tags/ifchanged.rb`      | `liquid/tags/ifchanged.go`      | `ifchanged`          |
| `lib/liquid/tags/include.rb`        | `liquid/tags/include.go`        | `include`            |
| `lib/liquid/tags/increment.rb`      | `liquid/tags/increment.go`      | `increment`          |
| `lib/liquid/tags/inline_comment.rb` | `liquid/tags/inline_comment.go` | `#` (inline comment) |
| `lib/liquid/tags/raw.rb`            | `liquid/tags/raw.go`            | `raw`                |
| `lib/liquid/tags/render.rb`         | `liquid/tags/render.go`         | `render`             |
| `lib/liquid/tags/snippet.rb`        | `liquid/tags/snippet.go`        | `snippet`            |
| `lib/liquid/tags/table_row.rb`      | `liquid/tags/table_row.go`      | `tablerow`           |
| `lib/liquid/tags/unless.rb`         | `liquid/tags/unless.go`         | `unless`             |

## Tag Base Classes

| Ruby Path                       | Go Path                     | Notes                 |
| ------------------------------- | --------------------------- | --------------------- |
| `lib/liquid/tag/disableable.rb` | `liquid/tag/disableable.go` | Disableable tag mixin |
| `lib/liquid/tag/disabler.rb`    | `liquid/tag/disabler.go`    | Tag disabler          |

## Test Files

### Integration Tests

Ruby: `test/integration/*_test.rb` → Go: `liquid/*_integration_test.go`

| Ruby Path                                  | Go Path                                      |
| ------------------------------------------ | -------------------------------------------- |
| `test/integration/assign_test.rb`          | `liquid/assign_integration_test.go`          |
| `test/integration/blank_test.rb`           | `liquid/blank_integration_test.go`           |
| `test/integration/block_test.rb`           | `liquid/block_integration_test.go`           |
| `test/integration/capture_test.rb`         | `liquid/capture_integration_test.go`         |
| `test/integration/context_test.rb`         | `liquid/context_integration_test.go`         |
| `test/integration/document_test.rb`        | `liquid/document_integration_test.go`        |
| `test/integration/drop_test.rb`            | `liquid/drop_integration_test.go`            |
| `test/integration/error_handling_test.rb`  | `liquid/error_handling_integration_test.go`  |
| `test/integration/expression_test.rb`      | `liquid/expression_integration_test.go`      |
| `test/integration/filter_test.rb`          | `liquid/filter_integration_test.go`          |
| `test/integration/filter_kwarg_test.rb`    | `liquid/filter_kwarg_integration_test.go`    |
| `test/integration/hash_ordering_test.rb`   | `liquid/hash_ordering_integration_test.go`   |
| `test/integration/hash_rendering_test.rb`  | `liquid/hash_rendering_integration_test.go`  |
| `test/integration/output_test.rb`          | `liquid/output_integration_test.go`          |
| `test/integration/parsing_quirks_test.rb`  | `liquid/parsing_quirks_integration_test.go`  |
| `test/integration/profiler_test.rb`        | `liquid/profiler_integration_test.go`        |
| `test/integration/security_test.rb`        | `liquid/security_integration_test.go`        |
| `test/integration/standard_filter_test.rb` | `liquid/standard_filter_integration_test.go` |
| `test/integration/tag_test.rb`             | `liquid/tag_integration_test.go`             |
| `test/integration/template_test.rb`        | `liquid/template_integration_test.go`        |
| `test/integration/trim_mode_test.rb`       | `liquid/trim_mode_integration_test.go`       |
| `test/integration/variable_test.rb`        | `liquid/variable_integration_test.go`        |

### Tag Integration Tests

Ruby: `test/integration/tags/*_test.rb` → Go: `liquid/tags/*_integration_test.go`

| Ruby Path                                       | Go Path                                          |
| ----------------------------------------------- | ------------------------------------------------ |
| `test/integration/tags/break_tag_test.rb`       | `liquid/tags/break_integration_test.go`          |
| `test/integration/tags/continue_tag_test.rb`    | `liquid/tags/continue_integration_test.go`       |
| `test/integration/tags/cycle_tag_test.rb`       | `liquid/tags/cycle_integration_test.go`          |
| `test/integration/tags/echo_test.rb`            | `liquid/tags/echo_integration_test.go`           |
| `test/integration/tags/for_tag_test.rb`         | `liquid/tags/for_integration_test.go`            |
| `test/integration/tags/if_else_tag_test.rb`     | `liquid/tags/if_else_integration_test.go`        |
| `test/integration/tags/include_tag_test.rb`     | `liquid/tags/include_integration_test.go`        |
| `test/integration/tags/increment_tag_test.rb`   | `liquid/tags/increment_integration_test.go`      |
| `test/integration/tags/inline_comment_test.rb`  | `liquid/tags/inline_comment_integration_test.go` |
| `test/integration/tags/liquid_tag_test.rb`      | `liquid/tags/liquid_integration_test.go`         |
| `test/integration/tags/raw_tag_test.rb`         | `liquid/tags/raw_integration_test.go`            |
| `test/integration/tags/render_tag_test.rb`      | `liquid/tags/render_integration_test.go`         |
| `test/integration/tags/snippet_test.rb`         | `liquid/tags/snippet_integration_test.go`        |
| `test/integration/tags/standard_tag_test.rb`    | `liquid/tags/standard_integration_test.go`       |
| `test/integration/tags/statements_test.rb`      | `liquid/tags/statements_integration_test.go`     |
| `test/integration/tags/table_row_test.rb`       | `liquid/tags/table_row_integration_test.go`      |
| `test/integration/tags/unless_else_tag_test.rb` | `liquid/tags/unless_else_integration_test.go`    |

### Unit Tests

Ruby: `test/unit/*_unit_test.rb` → Go: `liquid/*_unit_test.go`

| Ruby Path                                  | Go Path                                  |
| ------------------------------------------ | ---------------------------------------- |
| `test/unit/block_unit_test.rb`             | `liquid/block_unit_test.go`              |
| `test/unit/condition_unit_test.rb`         | `liquid/condition_unit_test.go`          |
| `test/unit/environment_filter_test.rb`     | `liquid/environment_filter_unit_test.go` |
| `test/unit/environment_test.rb`            | `liquid/environment_unit_test.go`        |
| `test/unit/file_system_unit_test.rb`       | `liquid/file_system_unit_test.go`        |
| `test/unit/i18n_unit_test.rb`              | `liquid/i18n_unit_test.go`               |
| `test/unit/lexer_unit_test.rb`             | `liquid/lexer_unit_test.go`              |
| `test/unit/parse_context_unit_test.rb`     | `liquid/parse_context_unit_test.go`      |
| `test/unit/parse_tree_visitor_test.rb`     | `liquid/parse_tree_visitor_unit_test.go` |
| `test/unit/parser_unit_test.rb`            | `liquid/parser_unit_test.go`             |
| `test/unit/partial_cache_unit_test.rb`     | `liquid/partial_cache_unit_test.go`      |
| `test/unit/regexp_unit_test.rb`            | `liquid/regexp_unit_test.go`             |
| `test/unit/registers_unit_test.rb`         | `liquid/registers_unit_test.go`          |
| `test/unit/strainer_template_unit_test.rb` | `liquid/strainer_template_unit_test.go`  |
| `test/unit/tag_unit_test.rb`               | `liquid/tag_unit_test.go`                |
| `test/unit/template_factory_unit_test.rb`  | `liquid/template_factory_unit_test.go`   |
| `test/unit/template_unit_test.rb`          | `liquid/template_unit_test.go`           |
| `test/unit/tokenizer_unit_test.rb`         | `liquid/tokenizer_unit_test.go`          |
| `test/unit/variable_unit_test.rb`          | `liquid/variable_unit_test.go`           |

### Tag Unit Tests

Ruby: `test/unit/tags/*_unit_test.rb` → Go: `liquid/tags/*_unit_test.go`

| Ruby Path                                 | Go Path                            |
| ----------------------------------------- | ---------------------------------- |
| `test/unit/tags/case_tag_unit_test.rb`    | `liquid/tags/case_unit_test.go`    |
| `test/unit/tags/comment_tag_unit_test.rb` | `liquid/tags/comment_unit_test.go` |
| `test/unit/tags/doc_tag_unit_test.rb`     | `liquid/tags/doc_unit_test.go`     |
| `test/unit/tags/for_tag_unit_test.rb`     | `liquid/tags/for_unit_test.go`     |
| `test/unit/tags/if_tag_unit_test.rb`      | `liquid/tags/if_unit_test.go`      |

### Tag Disableable Tests

| Ruby Path                                  | Go Path                                      |
| ------------------------------------------ | -------------------------------------------- |
| `test/integration/tag/disableable_test.rb` | `liquid/tag/disableable_integration_test.go` |

## Test Fixtures

| Ruby Path                     | Go Path                              | Notes        |
| ----------------------------- | ------------------------------------ | ------------ |
| `test/fixtures/en_locale.yml` | `liquid/test_fixtures/en_locale.yml` | Keep as YAML |

## Usage Examples

### Finding Go File from Ruby File

**Example 1**: Ruby file `lib/liquid/tags/if.rb`

- Remove `lib/` prefix: `liquid/tags/if.rb`
- Change extension: `liquid/tags/if.go`
- Result: `liquid/tags/if.go`

**Example 2**: Ruby test `test/integration/template_test.rb`

- Remove `test/integration/` prefix: `template_test.rb`
- Change extension and add suffix: `template_integration_test.go`
- Result: `liquid/template_integration_test.go`

**Example 3**: Ruby file `lib/liquid/profiler/hooks.rb`

- Remove `lib/` prefix: `liquid/profiler/hooks.rb`
- Change extension: `liquid/profiler/hooks.go`
- Result: `liquid/profiler/hooks.go`

### Finding Ruby File from Go File

**Example 1**: Go file `liquid/tags/if.go`

- Add `lib/` prefix: `lib/liquid/tags/if.rb`
- Result: `reference-liquid/lib/liquid/tags/if.rb`

**Example 2**: Go test `liquid/template_integration_test.go`

- Remove `_integration` suffix: `template_test.go`
- Change extension: `template_test.rb`
- Add `test/integration/` prefix: `test/integration/template_test.rb`
- Result: `reference-liquid/test/integration/template_test.rb`

## Version Mapping

| Ruby Version File       | Go Version File     |
| ----------------------- | ------------------- |
| `lib/liquid/version.rb` | `liquid/version.go` |

**CRITICAL**: Version numbers MUST match exactly. When Ruby releases a new version, update Go version to match.

## Notes

- All mappings preserve the base filename (without extension)
- Directory structure mirrors Ruby exactly (minus `lib/` prefix)
- Test files use Go naming conventions (`_test.go`) but maintain base name
- Integration vs unit test distinction preserved via suffix
- YAML files (like locales) remain as YAML
