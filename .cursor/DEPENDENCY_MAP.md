# Liquid Go Implementation - Dependency Map

This document maps the dependency hierarchy of Liquid files, showing what depends on what. Files are organized by dependency level, with Level 0 having no dependencies.

## Dependency Levels

### Level 0: Foundation (No Dependencies)

These files have no internal dependencies and can be implemented first:

- `liquid/version.go` - Version constant
- `liquid/const.go` - Constants and regex patterns
- `liquid/errors.go` - Error types (extends stdlib errors)
- `liquid/deprecations.go` - Deprecation warning system

**Dependencies**: None (only stdlib)

---

### Level 1: Basic Utilities

These depend only on Level 0:

- `liquid/utils.go` - Utility functions
- `liquid/interrupts.go` - Interrupt handling
- `liquid/extensions.go` - Extension points (uses stdlib time/date)
- `liquid/i18n.go` - Internationalization (uses YAML)
- `liquid/file_system.go` - File system abstraction

**Dependencies**: Level 0

---

### Level 2: Parsing Infrastructure

Core parsing components:

- `liquid/lexer.go` - Lexical analyzer (tokenizes input)
- `liquid/tokenizer.go` - Tokenizer (uses StringScanner)
- `liquid/parser_switching.go` - Parser mode switching
- `liquid/parser.go` - Template parser (depends on lexer)

**Dependencies**: Level 0, Level 1

**Order within level**: lexer → tokenizer → parser_switching → parser

---

### Level 3: Expression System

Expression parsing and evaluation:

- `liquid/expression.go` - Expression evaluation
- `liquid/variable_lookup.go` - Variable lookup logic (depends on expression)
- `liquid/range_lookup.go` - Range lookup logic

**Dependencies**: Level 0, Level 1, Level 2

**Order within level**: expression → variable_lookup → range_lookup

---

### Level 4: Variable System

Variable handling:

- `liquid/variable.go` - Variable class (depends on expression, variable_lookup)

**Dependencies**: Level 0-3

---

### Level 5: Condition System

Conditional logic:

- `liquid/condition.go` - Condition evaluation (depends on expression)

**Dependencies**: Level 0-3

---

### Level 6: Environment and Filters

Environment configuration and filter system:

- `liquid/standardfilters.go` - Standard filter implementations
- `liquid/strainer_template.go` - Filter strainer (uses Set)
- `liquid/environment.go` - Environment configuration (depends on tags, strainer_template, standardfilters)

**Dependencies**: Level 0-1

**Note**: `environment.go` has a circular dependency with tags (tags depend on environment, environment depends on tags). This is resolved at runtime via lazy initialization.

---

### Level 7: Tag Infrastructure

Base classes for tags:

- `liquid/tag/disableable.go` - Disableable tag mixin
- `liquid/tag/disabler.go` - Tag disabler
- `liquid/tag.go` - Tag base class (depends on tag/disableable, tag/disabler, parser_switching)
- `liquid/block_body.go` - Block body handling
- `liquid/block.go` - Block tag base class (extends tag, depends on block_body)

**Dependencies**: Level 0-2, Level 6 (for parser_switching)

**Order within level**: tag/disableable → tag/disabler → tag → block_body → block

---

### Level 8: Supporting Infrastructure

Supporting classes needed by context and template:

- `liquid/registers.go` - Template registers
- `liquid/resource_limits.go` - Resource limiting
- `liquid/partial_cache.go` - Partial template caching
- `liquid/template_factory.go` - Template factory
- `liquid/usage.go` - Usage tracking
- `liquid/parse_tree_visitor.go` - AST visitor pattern

**Dependencies**: Level 0-1

---

### Level 9: Parse Context

Parse-time context:

- `liquid/parse_context.go` - Parse-time context (depends on environment, parser, tokenizer, block_body, i18n)

**Dependencies**: Level 0-2, Level 6, Level 7, Level 8

---

### Level 10: Context

Rendering context:

- `liquid/context.go` - Rendering context (depends on environment, variable, expression, registers, resource_limits, variable_lookup, strainer_template)

**Dependencies**: Level 0-5, Level 6, Level 8

---

### Level 11: Drops

Drop classes for special objects:

- `liquid/drop.go` - Drop base class
- `liquid/forloop_drop.go` - For loop drop (depends on drop)
- `liquid/tablerowloop_drop.go` - Table row loop drop (depends on drop)
- `liquid/snippet_drop.go` - Snippet drop (depends on drop)

**Dependencies**: Level 0-1, Level 10 (for context)

**Order within level**: drop → forloop_drop, tablerowloop_drop, snippet_drop

---

### Level 12: Document

Document root node:

- `liquid/document.go` - Document root node (depends on block_body, parse_context)

**Dependencies**: Level 0-2, Level 7, Level 9

---

### Level 13: Tag Implementations

All tag implementations (depend on tag/block base classes):

**Simple Tags** (extend Tag):
- `liquid/tags/assign.go`
- `liquid/tags/break.go`
- `liquid/tags/continue.go`
- `liquid/tags/cycle.go`
- `liquid/tags/decrement.go`
- `liquid/tags/echo.go`
- `liquid/tags/increment.go`
- `liquid/tags/inline_comment.go`
- `liquid/tags/raw.go`

**Block Tags** (extend Block):
- `liquid/tags/capture.go`
- `liquid/tags/case.go`
- `liquid/tags/comment.go`
- `liquid/tags/doc.go`
- `liquid/tags/for.go` (depends on forloop_drop)
- `liquid/tags/if.go`
- `liquid/tags/ifchanged.go`
- `liquid/tags/include.go`
- `liquid/tags/render.go`
- `liquid/tags/snippet.go` (depends on snippet_drop)
- `liquid/tags/table_row.go` (depends on tablerowloop_drop)
- `liquid/tags/unless.go` (depends on if)

**Dependencies**: Level 0-7, Level 10-11

**Order within level**: Simple tags first, then block tags. `unless` depends on `if`.

---

### Level 14: Tags Registry

Tag registry:

- `liquid/tags.go` - Tag registry and standard tags (depends on all tag implementations)

**Dependencies**: Level 13

---

### Level 15: Template

Main template class:

- `liquid/template.go` - Template class (depends on document, parse_context, context, environment, profiler)

**Dependencies**: Level 0-14

---

### Level 16: Profiler

Performance profiling:

- `liquid/profiler/hooks.go` - Profiler hooks
- `liquid/profiler.go` - Performance profiler (depends on profiler/hooks)

**Dependencies**: Level 0-1

**Note**: Profiler can be implemented earlier but is typically used by Template.

---

## Dependency Graph Summary

```
Level 0: version, const, errors, deprecations
    ↓
Level 1: utils, interrupts, extensions, i18n, file_system
    ↓
Level 2: lexer → tokenizer → parser_switching → parser
    ↓
Level 3: expression → variable_lookup → range_lookup
    ↓
Level 4: variable
    ↓
Level 5: condition
    ↓
Level 6: standardfilters, strainer_template → environment
    ↓
Level 7: tag/disableable → tag/disabler → tag → block_body → block
    ↓
Level 8: registers, resource_limits, partial_cache, template_factory, usage, parse_tree_visitor
    ↓
Level 9: parse_context
    ↓
Level 10: context
    ↓
Level 11: drop → forloop_drop, tablerowloop_drop, snippet_drop
    ↓
Level 12: document
    ↓
Level 13: tags/* (all tag implementations)
    ↓
Level 14: tags.go (registry)
    ↓
Level 15: template
```

## Circular Dependencies

### Environment ↔ Tags

- `environment.go` references `Tags::STANDARD_TAGS` (from tags.go)
- `tags.go` may reference environment for configuration

**Resolution**: Use lazy initialization or forward declarations. In Go, this can be handled by:
- Defining tags as a map/constant that environment references
- Tags can be registered after environment initialization

## External Dependencies

### Standard Library

- `strscan` (Ruby) → Go equivalent: `strings` or custom scanner
- `set` (Ruby) → Go: `map[T]bool` or use a Set library
- `yaml` (Ruby) → Go: `gopkg.in/yaml.v3` or similar
- `cgi`, `base64`, `bigdecimal` (Ruby) → Go stdlib equivalents
- `time`, `date` (Ruby) → Go: `time` package

## Notes

1. **ParserSwitching**: Used by Tag and Variable classes for different parsing modes
2. **BlockBody**: Used by Block, Document, and many block tags
3. **Expression**: Core to Variable, VariableLookup, and Condition
4. **Context**: Central to rendering - used by all tags, variables, and drops
5. **Environment**: Configuration container - referenced by many components
6. **ParseContext**: Used during parsing - separate from rendering Context

