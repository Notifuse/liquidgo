# Liquid Go Implementation - Implementation Order

This document provides a phased implementation order that allows incremental testing. Each phase builds on the previous one and includes test milestones.

## Implementation Phases

### Phase 0: Foundation Setup

**Goal**: Set up basic infrastructure with no dependencies.

**Files to implement**:

1. `liquid/version.go` - Version constant
2. `liquid/const.go` - Constants and regex patterns
3. `liquid/errors.go` - Error types
4. `liquid/deprecations.go` - Deprecation warnings

**Tests to write**:

- `liquid/version_test.go` - Version constant test
- `liquid/const_test.go` - Constants test
- `liquid/errors_test.go` - Error types test

**Test milestone**: ✅ Basic types and constants work

**Rationale**: These are the foundation - everything else depends on them.

---

### Phase 1: Basic Utilities

**Goal**: Implement utility functions and basic infrastructure.

**Files to implement**:

1. `liquid/utils.go` - Utility functions
2. `liquid/interrupts.go` - Interrupt handling
3. `liquid/extensions.go` - Extension points
4. `liquid/i18n.go` - Internationalization
5. `liquid/file_system.go` - File system abstraction

**Tests to write**:

- `liquid/utils_test.go`
- `liquid/interrupts_test.go`
- `liquid/extensions_test.go`
- `liquid/i18n_test.go`
- `liquid/file_system_test.go`

**Test milestone**: ✅ Utilities and file system work

**Rationale**: These provide basic functionality needed by parsing and rendering.

---

### Phase 2: Parsing Infrastructure

**Goal**: Implement core parsing components.

**Files to implement** (in order):

1. `liquid/lexer.go` - Lexical analyzer
2. `liquid/tokenizer.go` - Tokenizer
3. `liquid/parser_switching.go` - Parser mode switching
4. `liquid/parser.go` - Template parser

**Tests to write**:

- `liquid/lexer_test.go`
- `liquid/tokenizer_test.go`
- `liquid/parser_switching_test.go`
- `liquid/parser_test.go`

**Test milestone**: ✅ Can tokenize and parse basic templates

**Rationale**: Parsing is fundamental - everything else needs it.

---

### Phase 3: Expression System

**Goal**: Implement expression parsing and evaluation.

**Files to implement** (in order):

1. `liquid/expression.go` - Expression evaluation
2. `liquid/variable_lookup.go` - Variable lookup logic
3. `liquid/range_lookup.go` - Range lookup logic

**Tests to write**:

- `liquid/expression_test.go`
- `liquid/variable_lookup_test.go`
- `liquid/range_lookup_test.go`
- `liquid/expression_integration_test.go`

**Test milestone**: ✅ Can parse and evaluate expressions

**Rationale**: Expressions are needed for variables and conditions.

---

### Phase 4: Variable and Condition Systems

**Goal**: Implement variable handling and conditional logic.

**Files to implement**:

1. `liquid/variable.go` - Variable class
2. `liquid/condition.go` - Condition evaluation

**Tests to write**:

- `liquid/variable_test.go`
- `liquid/condition_test.go`
- `liquid/variable_integration_test.go`

**Test milestone**: ✅ Can evaluate variables and conditions

**Rationale**: Variables and conditions are core to template logic.

---

### Phase 5: Environment and Filters

**Goal**: Implement environment configuration and filter system.

**Files to implement** (in order):

1. `liquid/standardfilters.go` - Standard filter implementations
2. `liquid/strainer_template.go` - Filter strainer
3. `liquid/environment.go` - Environment configuration (partial - without tags dependency)

**Tests to write**:

- `liquid/standardfilters_test.go`
- `liquid/strainer_template_test.go`
- `liquid/environment_test.go` (partial)
- `liquid/standard_filter_integration_test.go`

**Test milestone**: ✅ Filters work, environment can be configured

**Rationale**: Filters are needed for variable rendering. Environment is central configuration.

---

### Phase 6: Tag Infrastructure

**Goal**: Implement base classes for tags.

**Files to implement** (in order):

1. `liquid/tag/disableable.go` - Disableable tag mixin
2. `liquid/tag/disabler.go` - Tag disabler
3. `liquid/tag.go` - Tag base class
4. `liquid/block_body.go` - Block body handling
5. `liquid/block.go` - Block tag base class

**Tests to write**:

- `liquid/tag/disableable_test.go`
- `liquid/tag/disabler_test.go`
- `liquid/tag_test.go`
- `liquid/block_body_test.go`
- `liquid/block_test.go`

**Test milestone**: ✅ Tag base classes work

**Rationale**: All tags depend on these base classes.

---

### Phase 7: Supporting Infrastructure

**Goal**: Implement supporting classes needed by context and template.

**Files to implement**:

1. `liquid/registers.go` - Template registers
2. `liquid/resource_limits.go` - Resource limiting
3. `liquid/partial_cache.go` - Partial template caching
4. `liquid/template_factory.go` - Template factory
5. `liquid/usage.go` - Usage tracking
6. `liquid/parse_tree_visitor.go` - AST visitor pattern

**Tests to write**:

- `liquid/registers_test.go`
- `liquid/resource_limits_test.go`
- `liquid/partial_cache_test.go`
- `liquid/template_factory_test.go`
- `liquid/usage_test.go`
- `liquid/parse_tree_visitor_test.go`

**Test milestone**: ✅ Supporting infrastructure works

**Rationale**: These are needed by context and template.

---

### Phase 8: Parse Context

**Goal**: Implement parse-time context.

**Files to implement**:

1. `liquid/parse_context.go` - Parse-time context

**Tests to write**:

- `liquid/parse_context_test.go`

**Test milestone**: ✅ Parse context works

**Rationale**: Needed for parsing templates.

---

### Phase 9: Context

**Goal**: Implement rendering context.

**Files to implement**:

1. `liquid/context.go` - Rendering context

**Tests to write**:

- `liquid/context_test.go`
- `liquid/context_integration_test.go`

**Test milestone**: ✅ Context can resolve variables and evaluate expressions

**Rationale**: Context is central to rendering - all tags need it.

---

### Phase 10: Drops

**Goal**: Implement drop classes.

**Files to implement** (in order):

1. `liquid/drop.go` - Drop base class
2. `liquid/forloop_drop.go` - For loop drop
3. `liquid/tablerowloop_drop.go` - Table row loop drop
4. `liquid/snippet_drop.go` - Snippet drop

**Tests to write**:

- `liquid/drop_test.go`
- `liquid/forloop_drop_test.go`
- `liquid/tablerowloop_drop_test.go`
- `liquid/snippet_drop_test.go`
- `liquid/drop_integration_test.go`

**Test milestone**: ✅ Drops work

**Rationale**: Some tags need drops (for, table_row, snippet).

---

### Phase 11: Document

**Goal**: Implement document root node.

**Files to implement**:

1. `liquid/document.go` - Document root node

**Tests to write**:

- `liquid/document_test.go`
- `liquid/document_integration_test.go`

**Test milestone**: ✅ Can parse and render simple documents

**Rationale**: Document is the root of the parse tree.

---

### Phase 12: Simple Tags

**Goal**: Implement simple tags (non-block tags).

**Files to implement** (order doesn't matter much):

1. `liquid/tags/assign.go`
2. `liquid/tags/break.go`
3. `liquid/tags/continue.go`
4. `liquid/tags/cycle.go`
5. `liquid/tags/decrement.go`
6. `liquid/tags/echo.go`
7. `liquid/tags/increment.go`
8. `liquid/tags/inline_comment.go`
9. `liquid/tags/raw.go`

**Tests to write**:

- `liquid/tags/assign_test.go`
- `liquid/tags/break_test.go`
- `liquid/tags/continue_test.go`
- `liquid/tags/cycle_test.go`
- `liquid/tags/decrement_test.go`
- `liquid/tags/echo_test.go`
- `liquid/tags/increment_test.go`
- `liquid/tags/inline_comment_test.go`
- `liquid/tags/raw_test.go`
- `liquid/tags/assign_integration_test.go`
- `liquid/tags/break_integration_test.go`
- `liquid/tags/continue_integration_test.go`
- `liquid/tags/cycle_integration_test.go`
- `liquid/tags/echo_integration_test.go`
- `liquid/tags/increment_integration_test.go`
- `liquid/tags/inline_comment_integration_test.go`
- `liquid/tags/raw_integration_test.go`

**Test milestone**: ✅ Simple tags work

**Rationale**: These are simpler and don't have nested content.

---

### Phase 13: Block Tags (Part 1)

**Goal**: Implement basic block tags.

**Files to implement** (in order):

1. `liquid/tags/comment.go`
2. `liquid/tags/doc.go`
3. `liquid/tags/capture.go`
4. `liquid/tags/if.go`
5. `liquid/tags/unless.go` (depends on if)

**Tests to write**:

- `liquid/tags/comment_test.go`
- `liquid/tags/doc_test.go`
- `liquid/tags/capture_test.go`
- `liquid/tags/if_test.go`
- `liquid/tags/unless_test.go`
- `liquid/tags/if_else_integration_test.go`
- `liquid/tags/unless_else_integration_test.go`
- `liquid/tags/capture_integration_test.go`

**Test milestone**: ✅ Basic block tags work

**Rationale**: These are fundamental block tags.

---

### Phase 14: Block Tags (Part 2)

**Goal**: Implement iteration and conditional block tags.

**Files to implement** (in order):

1. `liquid/tags/for.go` (needs forloop_drop)
2. `liquid/tags/ifchanged.go`
3. `liquid/tags/case.go`
4. `liquid/tags/table_row.go` (needs tablerowloop_drop)

**Tests to write**:

- `liquid/tags/for_test.go`
- `liquid/tags/ifchanged_test.go`
- `liquid/tags/case_test.go`
- `liquid/tags/table_row_test.go`
- `liquid/tags/for_integration_test.go`
- `liquid/tags/table_row_integration_test.go`

**Test milestone**: ✅ Iteration and advanced conditionals work

**Rationale**: These are more complex block tags.

---

### Phase 15: Include and Render Tags

**Goal**: Implement template inclusion tags.

**Files to implement**:

1. `liquid/tags/include.go`
2. `liquid/tags/render.go`
3. `liquid/tags/snippet.go` (needs snippet_drop)

**Tests to write**:

- `liquid/tags/include_test.go`
- `liquid/tags/render_test.go`
- `liquid/tags/snippet_test.go`
- `liquid/tags/include_integration_test.go`
- `liquid/tags/render_integration_test.go`
- `liquid/tags/snippet_integration_test.go`

**Test milestone**: ✅ Template inclusion works

**Rationale**: These need file system and template factory.

---

### Phase 16: Tags Registry

**Goal**: Register all tags.

**Files to implement**:

1. `liquid/tags.go` - Tag registry and standard tags

**Tests to write**:

- `liquid/tags_test.go`
- `liquid/tag_integration_test.go` (verify all tags registered)

**Test milestone**: ✅ All tags are registered and accessible

**Rationale**: Completes tag system.

---

### Phase 17: Template

**Goal**: Implement main template class.

**Files to implement**:

1. `liquid/template.go` - Template class

**Tests to write**:

- `liquid/template_test.go`
- `liquid/template_integration_test.go`

**Test milestone**: ✅ Can parse and render full templates

**Rationale**: Template brings everything together.

---

### Phase 18: Profiler

**Goal**: Implement performance profiling.

**Files to implement** (in order):

1. `liquid/profiler/hooks.go` - Profiler hooks
2. `liquid/profiler.go` - Performance profiler

**Tests to write**:

- `liquid/profiler/hooks_test.go`
- `liquid/profiler_test.go`
- `liquid/profiler_integration_test.go`

**Test milestone**: ✅ Profiling works

**Rationale**: Optional feature, can be implemented last.

---

## Quick Reference: File Implementation Order

```
Phase 0:  version, const, errors, deprecations
Phase 1:  utils, interrupts, extensions, i18n, file_system
Phase 2:  lexer → tokenizer → parser_switching → parser
Phase 3:  expression → variable_lookup → range_lookup
Phase 4:  variable, condition
Phase 5:  standardfilters → strainer_template → environment
Phase 6:  tag/disableable → tag/disabler → tag → block_body → block
Phase 7:  registers, resource_limits, partial_cache, template_factory, usage, parse_tree_visitor
Phase 8:  parse_context
Phase 9:  context
Phase 10: drop → forloop_drop, tablerowloop_drop, snippet_drop
Phase 11: document
Phase 12: tags/assign, tags/break, tags/continue, tags/cycle, tags/decrement,
          tags/echo, tags/increment, tags/inline_comment, tags/raw
Phase 13: tags/comment, tags/doc, tags/capture, tags/if → tags/unless
Phase 14: tags/for, tags/ifchanged, tags/case, tags/table_row
Phase 15: tags/include, tags/render, tags/snippet
Phase 16: tags.go
Phase 17: template
Phase 18: profiler/hooks → profiler
```

## Testing Strategy

### After Each Phase

1. **Run unit tests** for the phase
2. **Run integration tests** if applicable
3. **Fix any failing tests** before proceeding
4. **Commit** working code

### Test Coverage Goals

- **Unit tests**: Test individual components in isolation
- **Integration tests**: Test components working together
- **Coverage**: Aim for same coverage as Ruby implementation

### Test File Naming

- Unit tests: `*_test.go` (in same directory as source files)
- Integration tests: `*_integration_test.go` (will be moved to `integration/` folder in the future)

## Implementation Tips

1. **Start small**: Implement minimal functionality first, then expand
2. **Test frequently**: Run tests after each file or small group
3. **Reference Ruby code**: Always check Ruby implementation for behavior
4. **Use Go idioms**: But maintain same logical structure as Ruby
5. **Document deviations**: Note any differences from Ruby implementation

## Milestone Checklist

- [x] Phase 0: Foundation complete
- [x] Phase 1: Utilities complete
- [x] Phase 2: Parsing works
- [x] Phase 3: Expressions work
- [x] Phase 4: Variables and conditions work
- [x] Phase 5: Filters work
- [x] Phase 6: Tag infrastructure works
- [x] Phase 7: Supporting infrastructure works
- [x] Phase 8: Parse context works
- [x] Phase 9: Context works
- [x] Phase 10: Drops work
- [x] Phase 11: Document works
- [x] Phase 12: Simple tags work
- [x] Phase 13: Basic block tags work
- [x] Phase 14: Advanced block tags work
- [x] Phase 15: Include/render work
- [x] Phase 16: Tags registered
- [x] Phase 17: Template works
- [x] Phase 18: Profiler works

## Notes

- **Environment ↔ Tags circular dependency**: Implement environment first with placeholder for tags, then complete after tags are implemented
- **Some phases can be parallelized**: Simple tags (Phase 12) can be implemented in parallel
- **Test as you go**: Don't wait until the end to test - test after each phase
- **Reference tests**: Use Ruby tests as reference for expected behavior
