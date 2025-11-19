# Generate Code Coverage Report Per File

## Overview

Create a markdown document (`COVERAGE.md`) that lists code coverage percentages for every file in the `liquid/` package, organized by directory structure. This will serve as a prioritized list for improving test coverage.

## Current Coverage Status

### Overall Coverage

- `liquid/` package: 83.8%
- `liquid/tag/` package: 92.0%
- `liquid/tags/` package: 89.2%
- **Total: 85.4%**

### Files Needing Test Improvement (Sorted by Priority - Lowest Coverage First)

#### Critical Priority (<70% coverage)

1. `liquid/parse_tree_visitor.go` - 57.1% (AddCallbackFor: 0.0%)
2. `liquid/parser_switching.go` - 60.0% (MarkupContext: 0.0%)
3. `liquid/resource_limits.go` - 61.1% (IncrementWriteScore: 28.6%)
4. `liquid/utils.go` - 64.7% (SliceCollection: 50.0%, ToInteger: 61.1%)
5. `liquid/condition.go` - 66.7% (checkMethodLiteral: 66.7%, compareValues: 53.3%, toNumber: 57.1%)
6. `liquid/context.go` - 66.7% (HandleError: 66.7%, Pop: 66.7%, Set: 66.7%, SetLast: 66.7%, tryVariableFindInEnvironments: 54.5%, checkOverflow: 50.0%)
7. `liquid/drop.go` - 66.7% (InvokeDropOld: 40.9%, InvokeDropOn: 46.2%, stringsTitle: 66.7%)
8. `liquid/document.go` - 66.7% (ParseDocument: 56.2%, parseBody: 54.5%, Parse: 58.3%)
9. `liquid/tokenizer.go` - 66.7% (nextVariableToken: 54.5%, tokenize: 57.1%, nextTagTokenWithStart: 0.0%)
10. `liquid/variable.go` - 70.0% (NewVariable: 70.0%)
11. `liquid/tags/render.go` - 73.5% (RenderToOutputBuffer: 73.5%)
12. `liquid/tags/if.go` - 75.0% (RenderToOutputBuffer: 75.0%, parseBodyForBlock: 78.6%)
13. `liquid/tags/table_row.go` - 75.4% (RenderToOutputBuffer: 75.4%)
14. `liquid/tags/include.go` - 85.7% (RenderToOutputBuffer: 85.7%)
15. `liquid/block_body.go` - 87.1% (parseForDocument: 87.1%, Parse: 75.0%, parseForLiquidTag: 79.2%, renderNodeOptimized: 79.3%)
16. `liquid/tags/case.go` - 86.4% (parseBodyForBlock: 86.4%, RenderToOutputBuffer: 83.3%)
17. `liquid/tags/for.go` - 90.9% (parseBody: 90.9%, Parse: 83.3%)
18. `liquid/tags/if.go` - 78.6% (parseBodyForBlock: 78.6%)

#### Medium Priority (70-85% coverage)

- `liquid/parser.go` - 48.9% (Expression: 48.9%)
- `liquid/parse_context.go` - 33.3% (computePartialOptions: 33.3%)
- `liquid/file_system.go` - 55.6% (ReadTemplateFile: 55.6%, FullPath: 82.4%)
- `liquid/standardfilters.go` - Various filters with 66-72% coverage
- `liquid/tags/assign.go` - 88.9% (NewAssignTag: 88.9%)
- `liquid/tags/capture.go` - 93.3% (NewCaptureTag: 93.3%)
- `liquid/tags/doc.go` - 88.9% (NewDocTag: 88.9%)
- `liquid/tags/inline_comment.go` - 87.5% (NewInlineCommentTag: 87.5%)
- `liquid/tags/raw.go` - 90.0% (NewRawTag: 90.0%)
- `liquid/tags/snippet.go` - 80.0% (NewSnippetTag: 80.0%)
- `liquid/tags/unless.go` - 87.5% (NewUnlessTag: 87.5%)

#### High Coverage (≥90% - maintain)

- Most accessor methods and simple functions are at 100%
- `liquid/tags/for.go` - 100% for most methods after recent improvements
- `liquid/tags/if.go` - 100% for parseIfCondition
- `liquid/tag/disableable.go` - 100%
- `liquid/tag/disabler.go` - 100%

## Implementation Steps

### 1. Generate Coverage Data

- Run `go test ./liquid/... -coverprofile=coverage.out` to generate coverage data
- Use `go tool cover -func=coverage.out` to get function-level coverage

### 2. Process Coverage Data

- Parse coverage output to aggregate function-level coverage into file-level coverage
- Calculate average coverage per file
- Group files by directory structure:
- `liquid/` (root package)
- `liquid/tag/` (tag base types)
- `liquid/tags/` (tag implementations)

### 3. Create Coverage Report Document

- Create `COVERAGE.md` in the repository root
- Structure the document with:
- Header with overall coverage summary
- Prioritized list of files needing improvement (sorted by coverage percentage)
- Sections for each directory with detailed breakdown
- Table format showing:
- File name
- Coverage percentage
- Key functions/methods with low coverage
- Status indicator (color coding: green ≥90%, yellow 80-89%, red <80%)
- Summary statistics at the end

### 4. Format and Organize

- Sort files by coverage percentage (lowest first) for prioritization
- Include total files count and average coverage per directory
- Add timestamp of when coverage was generated
- Include instructions for regenerating the report

## Output Format

The `COVERAGE.md` file will contain:

- Overall package coverage summary
- Prioritized list of files needing test improvements
- Per-file coverage breakdown organized by directory
- Key uncovered functions/methods per file
- Visual indicators for coverage levels
- Summary statistics

## Files to Create/Modify

- `COVERAGE.md` - New coverage report document with prioritized improvement list
