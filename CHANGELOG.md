# Changelog

## [Unreleased]

## [5.14.0-port.6]

### Fixed (no lax quirk is left: malformed markup renders as in Ruby)
- **In lax mode, a variable tries the strict parse first, as Ruby**: an expression error of that
  parse was swallowed, leaving the variable empty (`{{ h.k. }}`); it now falls back to the lax
  parse, which reads `h.k`.
- **A bracket that is never closed is skipped** in a lookup, as Ruby's `VariableParser`
  (`{{ arr[ }}` reads `arr`).
- **A number with several dots is Ruby's `to_f`** of it (`{{ 1..3 }}` is the float 1, written `1`
  as every whole float).
- **Lax filters split as Ruby's `FilterParser` and `FilterArgsRegex`**: a comma separates, and a
  quoted argument ends at its closing quote (`split:"t""` takes `"t"`).

### Breaking
- `liquid.SafeParse` and `ParseContext.SafeParseCompleteExpression` are gone: the first swallowed
  the error Ruby raises, and the second was called by nothing. `SafeParseExpression` raises in every
  error mode, as Ruby's `safe_parse_expression`.

### Tests
- `TestStrict2VariableNames`: `{{ self['(x[y'] }}` in lax reads nothing, as Ruby, whose lookup
  split ignores quotes (differential case `tags-27`); it expected `hello`, which Ruby never gave.
- Removed with the functions they tested: `TestExpressionSafeParse` and the three tests of
  `SafeParseCompleteExpression`.
- `TestHostileTemplatesAreBounded`: a lookup nested 101 deep in a condition is refused, as in an
  output. An output now meets the depth bound of the strict parse first, so the bound of a lookup
  needed a case of its own.

## [5.14.0-port.5]

### Breaking
- **Error modes are typed**: `liquid.ErrorMode`, with `ErrorModeLax`, `ErrorModeStrict`,
  `ErrorModeStrict2`, `ErrorModeRigid` and `ErrorModeWarn`. `EnvironmentConfig.ErrorMode`,
  `Environment.ErrorMode()` and `ParseContextInterface.ErrorMode()` use it; a constant string such
  as `ErrorMode: "lax"` still compiles.
- **`ParserSwitching` is gone**: its `ParseWithSelectedParser` and `RigidMode` were called by
  nothing, and its value could not be built outside the package. The fallback parse it held,
  Ruby's `strict_parse_with_error_mode_fallback`, which Shopify Liquid's `Variable` still uses, is
  now internal, with no change of behaviour.

## [5.14.0-port.4]

### Fixed (every divergence listed "minor" now follows Shopify Liquid)
- **`case` renders each `when` that matches**, as Ruby, not only the first; `else` renders when
  none before it did.
- **`for` and `tablerow` over a hash** iterate its `[key, value]` pairs, sorted by key (Ruby keeps
  the order the keys were added, which a Go map does not).
- **A decimal divided by zero** is `Infinity`, `-Infinity` or `NaN`, written as Ruby writes them;
  only an integer divided by an integer zero raises.
- **Integers beyond 64 bits stay exact** through the arithmetic filters, comparisons and output,
  as Ruby's `Integer` (`{{ 9223372036854775807 | plus: 1 }}` gives `9223372036854775808`);
  `ceil` and `floor` return them exact too.
- **A filter given fewer arguments than it requires** raises Ruby's `wrong number of arguments
  (given 1, expected 2)` (`{{ x | replace }}`, `{{ x | split }}`, `{{ 1 | plus }}`), which, as any
  filter error, writes nothing and is recorded.
- **An integer argument is read as Ruby's `Integer()`**: `offset` and `limit` of `for`, and
  `slice`, refuse a fraction (`limit: 2.7`) or text (`slice: 'x'`) with `invalid integer`, and
  read `" 2 "`, `"1_000"`, `"0x1A"`. A whole float still counts as an integer, for JSON data.
  `ToInteger` follows the same rules.
- **strict mode** accepts a range in a condition (`{% if (1..3) != a %}`), refuses a filter colon
  with no argument (`{{ x | upcase: }}`), and every mode refuses an `and` or `or` with no
  expression before or after it (`{% if true and %}`).
- **The name of a variable in lax mode** is Ruby's first quoted fragment of the markup, so
  malformed outputs render as Ruby renders them (`{{{ x }}}`, and 16 listed lax quirks).

### Changed
- `StandardFilters.Slice`, `Ceil` and `Floor` return `(interface{}, error)`.

## [5.14.0-port.3]

### Breaking
- **An environment is built once, from a config, and never changes**:
  `liquid.NewEnvironment(liquid.EnvironmentConfig{...}) (*Environment, error)` takes the tags, the
  filters, the error mode, the file system, the resource limits (a typed `ResourceLimitsConfig`),
  the output escape and the bounds of the partial cache, and keeps a copy of them. The setters are
  gone (`SetErrorMode`, `SetFileSystem`, `SetOutputEscape`, `SetDefaultResourceLimits` and its
  untyped map, `SetPartialCacheLimits`, `RegisterTag`, `RegisterFilter`), as are
  `NewEnvironmentWithStandardTags` and `Environment.Tags`. A setter called during renders wrote
  maps the renders read, without a lock.
- **`tags.Standard()`** gives the standard tags, in a map of the caller's own, to which a custom
  tag is added; it replaces `tags.RegisterStandardTags`.
- `NewEnvironment` refuses an unknown error mode, a tag without a name or a constructor, an invalid
  filter and negative partial cache bounds.

### Tests
- `integration/test_array_last.go` is now `array_commands_test.go`: its test never ran, as the file
  did not end in `_test.go`.

## [5.14.0-port.2]

### Changed
- **`false` is not blank, and a string of Unicode whitespace is blank** (a no-break space
  included), as before 5.14.0-port.1: two deliberate differences from Shopify Liquid 5.14, which
  makes `false` blank and counts ASCII whitespace only. Notifuse's display conditions compare with
  `blank`, and 5.14.0-port.1 changed what they show for `false` and for a no-break space. The rest
  of 5.14's `blank` and `empty` stays (`[] == blank` and `{} == blank` are true, `nil == empty` is
  false).
- **`v5.14.0` is retracted** (`retract` in `go.mod`), for that change.

## [5.14.0-port.1]

liquidgo now ports Shopify Liquid 5.14.0 (from 5.11.0): Shopify's test suite and the differential
corpus are those of 5.14.0, and CI compares with the 5.14.0 gem.

### Breaking
- **`Usage` is gone**, with `GlobalUsage` and `IncrementUsage`: nothing called them.
- **`NewIncrementTag` and `NewDecrementTag` return an error** too, for a name strict2 refuses.
- **`blank` and `empty` compare as Shopify Liquid 5.14**: `false == blank`, `[] == blank` and
  `{} == blank` are now true, and `nil == empty` is now false. `false == blank` was a listed
  deliberate difference; it is gone.

### Added (Shopify Liquid 5.12 to 5.14)
- **`self`**: when no variable is named `self`, `self.x` and `self['x']` read the variable `x`
  through the scopes and the data, so a name that is not an identifier can be read
  (`self['(a(b(c)']`); a `SelfDrop` given to a partial reads the variables of its caller.
- **The `squish` filter**: strips the input and turns each run of whitespace into one space.
- **`first` and `last` on a string** give its first and last character.
- **`blank` and `empty` without ActiveSupport**: `nil` and `false` are blank and not empty, a
  string of whitespace is blank and not empty, a number or `true` is neither; `case`/`when` compares
  with `blank` and `empty` as `if` does.
- **A range costs what the loop visits**: `for` and `tablerow` count each item of a range they
  visit in the render score, an empty body included, and never build the range.
- **strict2 is stricter**: the name of an `assign`, a `capture`, an `increment` or a `decrement`
  must be one identifier; a bare bracket (`{{ ['x'] }}`) is refused, `self['x']` replaces it; the
  error of a tag whose body is blank writes its message (lax and strict still write nothing).
- **`IncludeTag.ForLoop`** tells `include ... for` from `include ... with`.
- **`liquid.Raise`** stops the render of a custom tag with an error, as a standard tag raises.

Shopify's `cumulative_render_score_limit` and `cumulative_assign_score_limit` are not ported: the
limits of a liquidgo render already cover its partials, which do not reset them, and each render
has limits of its own (see [differences](docs/differences-from-shopify.md)).

### Changed
- The standard filters are split by family into `string_filters.go`, `encoding_filters.go`,
  `math_filters.go` and `array_filters.go`; `standardfilters.go` keeps the type and the filters shared
  by all (`size`, `first`, `last`, `date`, `default`). No filter changes.

### Fixed
- **strict2 parsed some expressions as lax**: three checks of the parse context knew `rigid` but not
  `strict2`, its other name, so a malformed expression in strict2 was tolerated.
- **The errors of `if`, `unless`, `case`, `for` and `tablerow` are handled as Ruby's**: they are
  recorded, numbered with their line and raised with `RaiseErrors()`; a blank `if` used to skip them.
- **`for` and `tablerow` read a quoted attribute as a string**: `limit: "a"` read the variable `a`.
  An invalid `offset` or `limit` of `for` now raises `invalid integer` as Ruby (it was ignored), and
  `tablerow` reads its attributes with Ruby's `to_i` (`nil` is 0, `"2.9"` is 2).
- **A `for` with a body was taken for a blank tag**, so its errors wrote nothing.
- **`capture` names its variable as Ruby**: the first run of name characters in its markup.
- **A NUL byte is not the end of a template**: the tokenizer and the number parser took a NUL byte
  for the end of their input, so `{{ x\0 }}` raised a syntax error; it is now read as Ruby reads it.
- **Lax parsing follows Ruby's on malformed markup**: an expression is stripped as Ruby's `strip`
  (ASCII whitespace and NUL bytes, not Unicode spaces), a filter is named by its first word, its
  arguments are the captures of Ruby's `FilterArgsRegex` (without the separator before each), and
  `assign` finds its `name = value` anywhere in its markup, as Ruby's unanchored `Syntax`. Five
  listed divergences now match Shopify (`numbers-08`, `lax-quirks-16`, `parsing_quirks_test.rb`
  97, 99 and 103).

### Tests
- **Invalid input** (security category 8): a template of invalid UTF-8 is refused, control
  characters in its text are kept, data of invalid UTF-8 and control characters goes through every
  standard filter, and the differential corpus compares control characters inside tags with Shopify
  (`control-01` to `control-05`).

## [5.11.0-port.12]

A render is faster: the newsletter template of the [library comparison](docs/library-comparison.md)
renders in 53 µs on GitHub's arm64 runner instead of 95 µs, the fastest of the Go engines and 6.9
times faster than Shopify Liquid with YJIT. Tagged `v5.11.11`.

### Changed
- A render allocates half as much: a filter without arguments gets none, a standard filter borrows
  the argument slice of its render (a filter that takes the place of a standard one still gets
  arguments of its own), `join` and the output make their result at once, and the output of a
  render is its buffer, not a copy. A template keeps the length of its last output to size the
  next one.
- A render looks things up once: the operator of a condition when it is made, a filter per call, a
  key of a map or a scope per lookup; the common values (strings, numbers, maps, lists) skip the
  method lookups and the reflection that other values need; `truncate` walks its input once.

## [5.11.0-port.11]

**Breaking**: the end of phase B, without a new major version (see below). Filters are declared with
`liquid.Filter` and drops declare their keys with `LiquidGet`: nothing of a Go value is read by
reflection any more. Renders share the parse of a partial in a bounded cache, and the parse tree
visitor walks expressions as Ruby's. Tagged `v5.11.10`.

### Breaking
- **Drops declare their keys**: a drop is a value with `LiquidGet(ctx *Context, key string)
  (interface{}, bool, error)`, as a `Liquid::Drop` in Ruby, called with the context of the render.
  `liquid.Drop`, `NewDrop`, `InvokeDropOn`, `IsInvokable`, `GetInvokableMethods` and the method
  blacklist are gone: a drop's methods and fields were found by reflection, under any of three
  spellings (`first_name`, `First_name`, `FirstName`). A key is now read by its exact name, as in
  Ruby: `{{ forloop.Index }}` is nil, and `tablerowloop.cols`, which Ruby does not have, is gone.
  A missing key is nil, or an `UndefinedDropMethod` with strict variables (Ruby's
  `liquid_method_missing`); an error `LiquidGet` returns is recorded. A drop keeps no context, so
  renders can share one. A drop without a `String()` method renders as nothing, never its fields;
  `forloop` and `tablerowloop` render as in Ruby (`Liquid::ForloopDrop`).
- **Duck-typed hooks are gone**: a lookup no longer calls `LiquidMethodMissing(string)` on a value,
  nor does a read from a map or an array filter call `SetContext(*Context)`, nor the array filters
  `Get(string)`: a type that relied on one compiles and renders nothing. Declare its keys with
  `LiquidGet`. `Context.AddFilters` returns an error, the `cached_partials` register holds a
  `map[string]*Template`, and `SliceItems` gives the items of a typed slice.
- **A filter error is recorded**: an error a filter returns, standard or not, gives nil and is
  recorded in `Result.Errors` without being written; with strict filters it is raised (so
  `{{ 1 | divided_by: 0 }}` now raises with strict filters). A Go error a filter or a drop returns
  is recorded as a `*InternalError` (`Liquid error: internal`, as Ruby) that unwraps to it, where
  the render ended with a `*PanicError`. A filter needs a function and a name a template can write
  (`RegisterFilter`, `WithFilters`).
- **An argument given as nil is not a missing one**: a standard filter takes Ruby's default only for
  an argument it is not given, so `{{ arr | join: nothing }}` joins with an empty string and
  `{{ s | truncate: 5, nothing }}` adds no ellipsis, as in Ruby.
- **Filters are declared**: `liquid.Filter(name, func(ctx *Context, input interface{}, args
  ...interface{}) (interface{}, error))`, given to `Environment.RegisterFilter(filters ...FilterDef)`
  and `WithFilters(filters ...FilterDef)`. A struct of filter methods is no longer a filter set: the
  methods were found by reflection, and a name resolved by tries (`find_index`, `FindIndex`, any case).
  A filter is now named exactly, as in Ruby: `{{ x | Downcase }}` is an unknown filter. The standard
  filters are a table generated from `StandardFilters` (`scripts/generate_filter_table.py`).
  `StrainerTemplate`, `StrainerTemplateClass`, `Environment.CreateStrainer`, `RegisterFilters`,
  `FilterMethodNames` (now `FilterNames`) and `Context.Strainer` are gone. An error a filter returns
  renders nothing, and is raised with strict filters, as an argument too many.
- **The parse tree visitor walks expressions, as Ruby's**: a node gives its children with
  `ParseTreeChildren() []interface{}` (`liquid.ParseTreeNode`), where the visitor looked up a
  `ParseTreeVisitor` method by reflection and walked only nodelists; `NewParseTreeVisitor` is gone
  (`ForParseTreeVisitor` stays), and a node without a callback gives nil, as Ruby's. Variables, filter
  arguments, conditions, loop collections and limits, cycle values, assigns, echoes and the partials
  of `include` and `render` are visited, so a callback on `*VariableLookup` lists the variables a
  template reads.
- **`TemplateFactory` is gone**, with the `template_factory` register and the `PartialCache` type:
  `LoadPartial` returns the partial as a `*Template`. A partial is named by its name.

### Changed
- **Renders share the parse of a partial**: an environment keeps the partials it parsed, by their
  text, so that renders that read the same text share one parse, whatever file system each render
  reads it from. The file system is read at each render, so a changed partial is never stale. The
  cache keeps at most 1,024 partials and 4 MiB of their text, the least recently used going first;
  a parsed partial holds 10 to 30 times its text, so 4 MiB of text is about 100 MiB of memory. A
  partial larger than an eighth of the bytes is parsed for its render and not kept.
  `Environment.SetPartialCacheLimits(partials, bytes)` sets the bounds, 0 turning the cache off. A
  template `LoadPartial` returns is shared between renders: a caller must not change it.

### Fixed
- A `reflect.Value` given to a template wrote the fields of what it holds; a value with a `String()`
  method now writes the text of `String()`, never through fmt.
- A channel, a func that a template cannot call, or a pointer fmt would write as an address renders as
  nothing (`nil` inside a list or a hash).
- A key a Go map cannot hash (`{{ y[y] }}` with a map, `{{ y[list] }}`) finds nothing, where the
  render ended with a `*PanicError`.
- A nil pointer is nil in a condition (`{% if np %}` is false, `np == nil` is true) as in the output.

## [5.11.0-port.10]

**Breaking**: the render API, without a new major version (see below). `Template.Render(ctx, data,
opts...)` is the one way to render, a tag is a `liquid.Node`, and the exception renderer, the global
filter and the profiler are gone. Strict variables follow Ruby, and the security checks of the
library run in CI. Tagged `v5.11.9`.

### Breaking

The API breaks without a new major version: a major would change the module path to `/v6`, which
would read as Shopify Liquid 6, and liquidgo's versions follow Shopify's. Templates render as before,
except for the fixes below and for three outcomes, which now end the render with the returned `error`
and `Result.Output` holding what was written before it stopped:

| Outcome | `Render(data, nil)` in v5.11.8 | `Render(ctx, data)` now |
|---|---|---|
| A Go panic in a tag, filter or drop | `Liquid error: internal` in the output, the render goes on, an `InternalError` in `Errors()` | `err` is a `*PanicError` with the stack |
| A resource limit | `Liquid error: Memory limits exceeded` as the output, the error in `Errors()` | `err` is the `*MemoryError` |
| `ctx` done | `Liquid error: render stopped: ...` as the output, the error in `Errors()` | `err` is the `*CanceledError` |

An application that read only `Errors()` must now check `err`: on a limit, `Result.Output` is a
partial body.

- **`Template.Render(ctx, data, opts...) (Result, error)`** is the one way to render. It is the former
  `RenderContext`, renamed, and it returns its `Result` by value (one allocation less per render). It
  replaces `Render(assigns, *RenderOptions) string`, `RenderBang` and `RenderContext`. The data is a map: a drop goes under a key, where `Render` also took a drop or a
  `*Context` as the data.
- **`RenderOptions` becomes render options**:

  | `RenderOptions` | Now |
  |---|---|
  | `Context` | the `ctx` argument |
  | `Registers` | `WithRegisters` |
  | `Filters` | `WithFilters` |
  | `StrictVariables` | `WithStrictVariables` |
  | `StrictFilters` | `WithStrictFilters` |
  | `RenderBang` | `RaiseErrors` |

  `Output`, `GlobalFilter` and `ExceptionRenderer` are gone: the output is `Result.Output`, output
  escaping replaces a global filter, and `Result.Errors` replaces an exception renderer.
- **The template keeps no render state.**
  - Gone: `PersistInstanceAssigns`, `SetPersistInstanceAssigns`, `InstanceAssigns`, `Errors`,
    `ResourceLimits`, `SetResourceLimits` and `Profiler`.
  - A render's errors and scores are in its `Result`, with the new `Result.RenderScore` and
    `Result.AssignScore`.
  - Limits come from the environment's defaults, and `WithLimits` sets them for one render.
- **`TemplateOptions`** loses the fields nothing read: `GlobalFilter`, `ExceptionRenderer`, `Registers`,
  `Filters`, `StrictVariables` and `StrictFilters` (strict variables and filters are render options,
  `WithStrictVariables` and `WithStrictFilters`).
- **Added**: `Template.Environment()`.
- **A template renders through `Render`**: `Template.RenderToOutputBuffer(context, *string)`, the
  render with a `Context` built by hand, is gone. `Template.RenderToOutput(context, output)` stays, for
  a tag that renders a template as a partial of the render it runs in (`render`, `include`).
- **The exception renderer and the global filter are gone**: `Environment.ExceptionRenderer`,
  `Environment.SetExceptionRenderer`, `Context.ExceptionRenderer`, `Context.SetExceptionRenderer`,
  `Context.GlobalFilter`, `Context.SetGlobalFilter`, `Context.ApplyGlobalFilter`, and
  `TagContext.ApplyGlobalFilter`. An error is written as its message, or raised with `RaiseErrors`
  (`ContextConfig.RethrowErrors` for a `Context`); `Result.Errors` lists them; output escaping
  replaces a global filter.
- **The instance-assign leftovers of `Context` are gone**: `ContextConfig.OuterScope`,
  `Context.ClearInstanceAssigns`, `Context.Reset` (for a pool that does not exist) and
  `Context.LookupAndEvaluate`. A template keeps no assigns between renders, so a context's outer scope
  starts empty and holds the render's assigns and captures.
- **`Context.IncrementRenderDepth` and `DecrementRenderDepth` are gone**: the `if` tag counted its
  nesting during the render, which Ruby does not. The depth of blocks is bounded at parse and the
  recursion of partials by the depth of scopes, which stay.
- **The profiler is gone**: `liquid.Profiler`, `liquid.Timing`, `NewProfiler`, `Context.Profiler`,
  `Context.SetProfiler` and `TemplateOptions.Profile`. `Render` no longer used it; Go's own profiler
  (`pprof`) measures a render. `Profile: true` also turned line numbers on: set `LineNumbers`.
- **A tag is a `liquid.Node`**, with one render method:
  `RenderToOutput(context TagContext, output *OutputBuffer)`. liquidgo no longer looks up, by
  reflection, a `Render(context) string` or a `RenderToOutputBuffer(context, *string)` that a tag
  declares; the base `Tag`, `Block`, `BlockBody`, `Document` and the standard tags lose theirs.
  - `Tag` has no render method: a tag that embeds it declares `RenderToOutput`, or it does not compile
    as a `Node`. A tag of another package renames its `Render` or `RenderToOutputBuffer` to
    `RenderToOutput` and writes to the buffer. A `Block` still renders its body by default, so a
    block that keeps the old name compiles and renders only its body.
  - `Environment.RegisterTag` takes a `liquid.TagConstructor`,
    `func(tagName, markup string, pc ParseContextInterface) (Node, error)`, where it took any value
    and called it by reflection; `TagForName` and `Tags` return them. `tags.TagConstructor` is now
    an alias of it. A constructor that returns neither a tag (nil, or a nil pointer) nor an error is a
    syntax error that names the tag, where it gave a generic `Tag`.
  - `tag.Disableable` and `tag.Disabler` lose `RenderToOutputBuffer`, and their `RenderToOutput`
    becomes `Render`, so that a tag that embeds them keeps its own `RenderToOutput`;
    `Disabler.Render` no longer takes the output.
  - Gone: `liquid.RenderToString`. A blank block, or one whose `Render` returned nothing, rendered
    its body in place of its `Render`, as in v5.11.0; with one render method, a block renders
    through it.

### Fixed
- **The end of a comment decides whether the text after it is trimmed**, as Ruby: in
  `{%comment-%}{%endcomment%} `, the `-%}` of the opening tag trimmed the space after `endcomment` too.
- **`LocalFileSystem` reads only under its root**: a symbolic link under the root read a file out of
  it, and the root check took a sibling whose name starts with the root's for the root. The file
  and the root are resolved, and the file must be under the root.
- **`LocalFileSystem` errors name the template, never a path of the host**: an unreadable file gave
  `Failed to read template 'x': read /abs/path/_x.liquid: ...`, written in the render.
- **Strict variables record an undefined name of the first level**, as Ruby's `find_variable`:
  `{{ nope }}` recorded nothing, only `{{ a.nope }}` did. Like Ruby, the error is recorded and not
  written, and it ends its node: `{% if nope %}` renders neither branch.
- **A key that holds nil is not undefined** under strict variables, as in Ruby (`{% if nothing %}` with
  `nothing: nil` renders its else branch).
- **An integer index reads an array as Ruby's `Array#[]`**: a negative index counts from the end
  (`{{ arr[-1] }}` is the last item), and an index out of range is nil, which strict variables do not
  raise. An integer is a Go integer of any kind, named or not, or a float with no fraction (a JSON number);
  an index beyond the range of `int` is out of range, also on 32-bit platforms. Any
  other key is no index, as in Ruby: `{{ arr['1'] }}` and `{{ arr[1.5] }}` render nothing, where they
  read the item at 1. Digits written after a dot stay an index of a list (`{{ order.items.0.sku }}`),
  as Notifuse writes the paths of its display conditions, where Ruby keeps the string `"0"` (lax mode;
  the strict parser rejects a number after a dot, as Ruby's). Three listed divergences from Shopify are gone.
- **A name that holds nil in the data is looked up in the template's assigns next**, as Ruby goes on
  past a nil to the next environment: data `{"x": nil}` with the assign `x = "y"` renders `y`. With
  strict variables, the data decides, nil included, as in Ruby.
- A `Context` built with no environments has one empty environment, as Ruby's `Context.new`.
- **With strict variables, a render given no data finds the template's assigns**, as Ruby's
  `render(nil)` makes them its one environment. Given a map, the map decides, as in Ruby.

### Changed
- **`Result.Errors` lists each error the render recorded**, as Ruby's `template.errors`: the same error
  value raised twice is listed twice, where `RenderContext` listed it once. With `RaiseErrors`, an
  error raised through enclosing blocks and partials is still listed once.
- **The registers of a render reach its partials**: `WithRegisters` writes them in the static
  registers, as Ruby's `registers:` option, which a `{% render %}` partial inherits. They were in the
  change layer, which a `render` partial does not see (as `RenderOptions.Registers` in v5.11.8). A
  `file_system` register now serves nested partials; `WithFileSystem` takes precedence over it.
- **A func of the data is called once per render**, its value kept with the render: the data and the
  template's assigns are never written. Ruby writes a proc's value back into the hash; so did
  liquidgo, so that two renders sharing one data map saw the first one's value, and concurrent renders
  wrote one map. This holds for a func read by a key of the data, of a nested map, or by an index of a
  list, which were rendered as a Go value (`{{ funcs[0] }}` gave an address), as Ruby's
  `lookup_and_evaluate` calls a proc there. A func that lands in a scope of the render (a loop
  variable, an assign) is called there and its value written in the scope, as Ruby: a map a host tag
  pushes (`Context.Push`, `Context.Stack`) is a scope, and belongs to the render.

### Tests
- **Security checks of the categories of the security plan**:
  - each guard is removed in turn by `scripts/guard_mutations` (31 guards), which fails when no test
    fails without it; the Guards workflow runs it on main and before a release;
  - the library imports no network, process or plugin, and `os` only where a host gives files
    (`TestTheLibraryImportsNoSystemAccess`);
  - no error shows a value of the data, but where Ruby names a key or a partial
    (`TestErrorsShowNoValueOfTheData`);
  - with the output escape, a value stays in its HTML destination
    (`TestOutputEscapeKeepsAValueInItsDestination`), and the README lists what the escape does not
    cover.
- **A template keeps the text around its tags**: the text before its first tag starts the
  output and the text after its last one ends it, checked over the differential corpus and in
  `FuzzRender`, which found the comment trim above.
- **The differential corpus renders with strict variables** the templates that ask for it
  (`"strict_variables": true`): 18 templates, checked against Shopify Liquid 5.11.0.
- **The render benchmark renders the same data on every run.** It picked the current collection,
  product and blog, and the cart's items, in the order of a Go map, so its output and its allocations
  changed from one run to the next (30,900 to 38,400 allocations per run of `BenchmarkRender`). It
  takes them in the order of the YAML file, as Ruby's `values.first`.

## [5.11.0-port.9]

`Template.RenderContext`, a render with errors returned that changes nothing in the template. Tagged
`v5.11.8`.

### Added
- **`Template.RenderContext(ctx, data, opts...) (*Result, error)`**, a render with errors returned. It
  renders in a state of its own and changes nothing in the template, so concurrent renders of one
  template never see each other's assigns, errors, limits or profiler. It returns the output and the
  errors the render recorded (`Result`), and an error when the render could not complete: a
  `*MemoryError` for a limit, a `*CanceledError` when `ctx` is done, or the new `*PanicError`, with the
  Go stack, when Go code panics during the render. `Render` keeps writing such a panic as
  `Liquid error: internal`, as Ruby does. Options: `WithFileSystem` (partials of the render, nested ones included),
  `WithLimits` (limits that replace the template's) and `RaiseErrors` (the first recorded error ends
  the render, as Ruby's `render!`). A drop is bound to the render that reads it: give each render its
  own drop instances.

## [5.11.0-port.8]

`RenderBang` no longer changes the template, and a template's limits and profiler can be read while
other renders of it run. Tagged `v5.11.7`.

### Fixed
- **`RenderBang` applies to its own render.** It left the template raising its errors: the next
  `Render` of the same template raised an error out of `Render` instead of recording it. Ruby's
  `render!` also sticks, but there raising is part of the API.
- **`Template.ResourceLimits` and `Template.Profiler` during concurrent renders.** Each render wrote
  its scores into the limits value the template returns, and its profiler into the template,
  without the lock: reading them while another render ran was a data race. A render now publishes a
  new limits value under the lock, and the profiler is written and read under it. Both still give
  the state of the last render, which with concurrent renders may be another one: give a render a
  `Context` to read its own limits.

## [5.11.0-port.7]

A time written without the `date` filter renders as in Ruby. Tagged `v5.11.6`.

### Fixed
- **A time renders as in Ruby.** `{{ t }}` with a `time.Time` wrote Go's form,
  `2026-11-15 10:00:00 +0000 UTC`; it writes Ruby's `Time#to_s`, `2026-11-15 10:00:00 UTC` for a
  time in UTC and `2026-11-15 11:00:00 +0100` otherwise. Inside a list or a hash it writes
  `Time#inspect`, which adds the fraction of a second. The `date` filter is unchanged.

## [5.11.0-port.6]

Fixes from an architecture audit, for templates written by untrusted users. Each render starts
from its own data, a plain Go struct exposes nothing to a template, and local filters work on every
render, also concurrently. Behaviour changes for Go integrations: see Changed. Tagged `v5.11.5`.

### Changed
- **A plain Go struct exposes nothing to a template.** A template read every exported field of
  any struct and called its methods without arguments (`{{ user.password_hash }}`,
  `{{ user.delete }}`), the array filters called its `Get` method, and `{{ user }}` wrote all its
  fields. As in Ruby, only a drop (a type that embeds `liquid.Drop`) now exposes its methods and
  fields. A struct renders as nothing, and as `nil` inside a list or a hash, where Ruby raises an
  internal error for an object without `to_liquid`. An unknown filter no longer reads a property
  of its input (`{{ posts | first | title }}`). To migrate, implement `ToLiquid()` with a value
  receiver and return a map, or embed `*liquid.Drop`. `ToLiquid()` is now applied at each step of
  a lookup, to the items the array filters read and to a value written in the output, as Ruby
  calls `to_liquid`. `time.Time` still renders and takes `date`; `{{ t.year }}` renders nothing,
  as in Ruby.
- **Pointers and hashes render the same on every run.** A pointer renders as what it points to (a
  pointer to a pointer wrote an address) and a nil pointer as nothing (it wrote `<nil>`). A typed
  map renders as a Liquid hash, and every hash with its keys in sorted order: a Go map has no
  order, so `{{ hash }}` could change from one render to the next.
- **Each render starts from its own data.** A `Template` no longer keeps what a render assigns
  (`assign`, `capture`) for its next render: rendered for one recipient who sets a variable, then for
  another who does not, it showed the first recipient's value to the second. Data given as a drop
  (`Render(drop)`) was also kept, and answered every later render of the template, with a map, nil
  or a `Context`; Ruby never kept it. Ruby keeps the assigns;
  `TemplateOptions.PersistInstanceAssigns: true` (or `Template.SetPersistInstanceAssigns(true)`)
  restores that behaviour, and turning it off forgets what earlier renders kept.
- `Template.Errors` takes the template's lock: it raced with concurrent renders. With concurrent
  renders, it returns the errors of the render that ended last.

### Fixed
- **Local filters on every render.** `RenderOptions.Filters` applied only to the render that
  built the environment's strainer cache entry: the next renders with the same filters rendered
  their input unchanged. The cache is also read and written under a lock, so concurrent renders
  with local filters no longer race.
- **Environment filters with local filters.** A render that passed `RenderOptions.Filters` lost the
  filters registered on the environment (`RegisterFilter`): they rendered their input unchanged.
  Both now apply, the local ones first for a name both define, as in Ruby.
- **Filter types with the same name.** The strainer cache told filter types apart by their Go name
  only: two `filters.Filters` from two packages shared one method table.
- **Concurrent renders of partials.** Loading a partial for `render` or `include` wrote to the
  tag's parse state, which every render of the template shares; it no longer does.

## [5.11.0-port.5]

An option to HTML-escape every output, for HTML templates that output untrusted data. Off by
default: nothing changes without it. Tagged `v5.11.4`.

### Added
- **Output escape** (`Environment.SetOutputEscape`), off by default: every `{{ }}` output, `echo`
  and `cycle` is HTML-escaped, except HTML on purpose: the result of the new `raw` filter, of
  `escape`, `escape_once` or `h` (no double escaping), and the text of a `capture`, also through
  `strip`, `newline_to_br`, `truncate` and the like. `assign`, conditions and filters still see the
  data as it is. For HTML templates that output untrusted data, as LiquidJS's `outputEscape`.

## [5.11.0-port.4]

`url_decode` no longer loses a value that holds a lone `%`, and `escape` writes quotes as Ruby does.
Tagged `v5.11.3`.

### Fixed
- **`url_decode`** decodes as Ruby's `CGI.unescape`: a `%` that two hexadecimal digits do not
  follow stays as written (`'%E2%82%AC%2'` gives `€%2`). Before, such a `%` made the filter fail,
  which renders nothing: the whole value was lost. A decoded byte sequence that is not UTF-8 is still
  an error, as in Ruby.
- **`escape`** writes a quote as `&quot;`, as Ruby's `CGI.escapeHTML`; it wrote `&#34;`.

### Tests
- The differential corpus has templates for escaping (`escape`, `escape_once`, `url_encode`,
  `url_decode`, `strip_html`), for markup that could lose content (unclosed tags and outputs, extra
  braces, `{%22` in a URL) and for unusual input (control characters, NUL, BOM, CRLF, U+2028, a
  5,000-character name), compared with Shopify Liquid 5.11.0.
- `TestRecursivePartialsAreBounded`: partials that render themselves stop with "Nesting too deep";
  without the scope depth check, they overflow the Go stack, which ends the process.

## [5.11.0-port.3]

Lookups, parse errors and rendering of lists, ranges and conditions follow Shopify Liquid 5.11.0.
Several outputs change: see the list below before upgrading. Tagged `v5.11.2`, which also ships
5.11.0-port.2.

### Fixed
- **Lookup order**: a variable set by `assign` or `capture` wins over the render data of the same
  name, as in Ruby. Before, the render data was read first.
- **`increment` and `decrement`** counters last the whole render: they no longer restart after a
  `for` loop, no longer start from an `assign` of the same name, and `{{ name }}` reads the counter
  unless an `assign` hides it.
- **An unknown tag inside a block** (`capture`, `for`, `case`, `tablerow`, `ifchanged` and the
  others) is a syntax error in every mode, as in Ruby. Before, it was dropped, or it closed the
  block and the text after it was lost. A `when` without a condition is a syntax error too.
- **A lone quote** (`{{ ' }}`) is an empty string in lax mode, and a condition without a value is
  Ruby's `if` syntax error; both could panic before.
- **Error messages** are Shopify Liquid's: the English locale of 5.11.0 is embedded, so no file in
  the working directory changes a message, and a missing key gives the key instead of panicking.
- **Arrays render as in Ruby**: `{{ list }}`, `echo` and `cycle` write the items one after the other,
  nested lists included and `nil` items left out (`["a", "b"]` renders `ab`); they rendered
  `["a", "b"]`. A filter that takes a string still receives Ruby's inspect of the list, now with
  `nil` for a nil item and quotes for a typed slice such as `[]string`. A list that holds itself
  renders its other items, in an output and in `join` or `reverse`, where it overflowed the stack.
- **Ranges in filters**: `join`, `size`, `first`, `last`, `sort`, `reverse`, `sum`, `uniq` and the
  other array filters treat `(1..3)` as its integers; `first` and `last` give the bounds even of an
  empty range. `join` stops at `render_length_limit` without building the integers, and the filters
  that build them check `assign_score_limit` first. Without these limits, a filter builds at most
  1,048,576 integers of a range and raises `Memory limits exceeded` beyond.
- **Properties in the array filters**: `map`, `where`, `reject`, `has`, `find`, `find_index`,
  `sort`, `sort_natural`, `uniq`, `compact` and `sum` read a property as Ruby's `item[property]`:
  a string item gives the property when it contains it, an integer item is an error (which renders
  nothing), and a float, boolean or nil item makes the filter give nil, or adds 0 to `sum`. A
  pointer to a struct, a drop and a map with string keys such as `map[string]string` give their
  property; they gave nil.
- **Comparisons**: `<`, `>`, `<=` and `>=` with nil, a boolean, a list, a hash or a drop on either
  side are false, as in Ruby, instead of comparing their text or raising; a number against a string
  raises Ruby's message, `comparison of Integer with String failed`. A `*time.Time` compares as
  the time.
- **`case`** keeps the whitespace of its bodies unless every body is blank, as in Ruby.
- **`tablerow` over nil or false** renders nothing; it rendered an empty row.
- **An `elsif` that does not parse** (`{% elsif %}`) is Ruby's `if` syntax error; it was dropped.
- **Strict modes** (`strict`, `strict2` and `rigid`) lex the whole markup of `if`, `unless` and `for`, so an unclosed quote in an
  index (`a['k]`) is a syntax error that names the markup.

### Changed
- **`Sort`, `SortNatural`, `Uniq`, `Compact`, `Where`, `Reject`, `Has`, `Find`, `FindIndex` and
  `Sum`** of `StandardFilters` return a value and an error, as `Map` does; `Has` returns
  `interface{}`: nil when an item is a float, a boolean or nil, as in Ruby.
- **Added**: `liquid.WriteValue` writes a value as an output does, `Range.Len`,
  `Context.Counter` and `Context.SetCounter` for the `increment` and `decrement` counters, and
  `liquid.LocaleOf` and `liquid.NewSyntaxErrorFor` for a tag's syntax error with Ruby's message.

## [5.11.0-port.2]

Rendering under resource limits: the output, loops, filters and parsing stay within the limits a
template is given, a render can be stopped, and an error in one node no longer loses the rest of the
output. Several outputs change: see the list below before upgrading. Not tagged on its own: it ships
in `v5.11.2` with 5.11.0-port.3.

### Added
- **`OutputBuffer`**: the output of a render is written to a buffer that checks
  `render_length_limit` at each write, or counts it in `assign_score_limit` inside a `capture`.
  Blocks, tags, variables, documents and templates have a `RenderToOutput(context, *OutputBuffer)`
  method; every `RenderToOutputBuffer(context, *string)` method stays and gives the same output. A
  custom tag keeps the method it defines: a tag that embeds `*liquid.Tag` or `*liquid.Block` and
  defines `Render` or `RenderToOutputBuffer`, on a pointer or a value receiver, is rendered through
  it, as before. `liquid.RenderToString` adapts one to the other, `NewOutputBufferFor` gives a
  custom tag a buffer under the limits of its context (inside a `capture`, under the assign score
  limit), `liquid.RenderCaptured` renders as `capture` does, and `OutputBuffer.Scratch` gives a
  buffer for output a tag may not keep, which counts only once written.
- **Stopping a render**: `RenderOptions.Context` takes a `context.Context`, and
  `Context.SetCancellation` sets one on a context the caller builds. The render stops at the next
  node or loop iteration once it is done, with a `CanceledError` that wraps the context's error
  (`errors.Is(err, context.DeadlineExceeded)`), and the cause given to `context.WithCancelCause`.
  A `Context` of the caller keeps its own cancellation after a render with `RenderOptions.Context`.
  `Context.CheckCanceled` lets a custom tag check it.

### Fixed
- **Output length**: building the output costs time in proportion to its length, and
  `render_length_limit` stops a render as soon as the output passes it, instead of after the node
  that wrote it.
- **`tablerow` and `ifchanged`** render their body once per item: `increment`, `cycle` and `assign`
  inside them take effect once, as in Ruby, and nesting them no longer costs exponential time.
- **Ranges in `for` and `tablerow`** are iterated without being built, so `(1..2000000000) limit: 1`
  takes no memory. An `offset` plus a `limit` past the integer range saturates instead of dropping
  items. Each iteration counts in `render_score_limit`, even with an empty body.
- **Filter results**: `join`, `replace`, `replace_first`, `replace_last`, `append` and `prepend`
  compute the size of their result and raise `Memory limits exceeded` before building one that
  passes `render_length_limit`; `split` and `concat` do so for a list with more items than
  `assign_score_limit` leaves. The string result of every filter is checked against
  `render_length_limit`, so that a chain of filters that each grow their input stops.
- **`ifchanged`** counts in the limits only the body it writes.
- **An error in a node** writes its message in place of that node and keeps the rest of the output,
  following Ruby's rules: an undefined variable, drop method or filter writes nothing, a tag that
  renders nothing writes nothing, and `Memory limits exceeded` or a stopped render ends the render,
  from any block or partial. Before, the whole output became `Liquid error: internal`. The errors of
  a `render` partial are in `Errors()`, and `include` restores the template name after an error.
- **`cycle`** without values renders nothing.
- **A hash indexed by a nil key** (`{{ h[nothing] }}`) renders nothing, as in Ruby.
- **A drop method that takes arguments** is no longer called by a lookup.
- **Resource limits** given as `int64`, `uint`, `int32` or another integer type, or as a whole
  `float64` such as a number decoded from JSON, are applied; they were ignored. A value past the
  `int` range saturates.
- **Parsing** costs time in proportion to the template: lookups, brackets and parentheses nested
  more than 100 deep (outside quoted strings), and `liquid` tags nested in a `liquid` tag more than
  100 deep, raise `Nesting too deep`, as blocks do. The body of `raw` is built in one pass.
- **`slice` and `truncate`** on a string read only the characters they need.
- **Parse caches**: the process-wide caches of lookups and expressions are gone; each parse keeps
  its own, so memory no longer grows with the number of distinct templates.

### Changed
- **Partials share the limits** of the template that renders them: `render` and `include` no longer
  reset `render_length_limit`, `render_score_limit` and `assign_score_limit`. Ruby resets them, so a
  template can pass the limits through partials there; here it cannot.
- **A lookup nested more than 100 deep** (`a[a[a[...]]]`) is a parse error. Ruby accepts it up to its
  stack size.
- **A custom tag with `RenderToOutputBuffer(context, *string)`** gets a string of its own: it no
  longer sees, or rewrites, what the template wrote before it.
- **With `StrictFilters`, an undefined filter** renders nothing and is reported in `Errors()`, as in Ruby;
  it rendered the error message.

### Tests
- **Hostile templates**: `TestRenderScalesLinearly` renders families of templates at n and 4n and
  fails when 4 times the input takes more than 9 times as long or allocates more than 9 times as
  much; `TestHostileTemplatesAreBounded`
  checks that each resource limit refuses a template before an allocation budget. The cases that
  fail today are listed as known, with the reason, and fail the test once they pass.
- **Differential corpus**: hand-written and generated templates rendered by Shopify Liquid 5.11.0
  in lax and strict mode (`integration/testdata/differential`), with every divergence listed.
- **Fuzzing**: `FuzzRender` checks that no panic leaves the API or becomes an internal error, and
  that two parses of a template render the same output.
- **CI**: the Go version comes from `go.mod`; tests run with the race detector on amd64 and arm64,
  on 386 and on Windows; the differential corpus is checked against
  Shopify Liquid 5.11.0 and `FuzzRender` runs for 2 minutes.

## [5.11.0-port.1]

Conformance fixes, checked by a new suite that replays the Shopify Liquid 5.11.0 test cases.
Several outputs change: see the list below before upgrading.

### Fixed
- **Equality**: `==`, `!=` and `case/when` compare numbers by value across Go numeric types, so a
  JSON number matches an integer literal (`{% if order.total == 100 %}`); comparing two lists or two
  maps no longer renders `Liquid error: internal`; `<`, `>`, `<=`, `>=` accept every numeric type.
  The `where`, `reject`, `has`, `find` and `find_index` filters use the same equality.
- **Quoted `and`/`or`**: `{% if v == "black and white" %}` no longer splits inside the quotes; spaces
  in quoted values are kept; a quoted `|` or parenthesis is part of the value.
- **Multi-line tags**: a tag or variable may span several lines (`{%\nif x\n%}`, `{%- liquid ... -%}`,
  a filter on the next line of an `assign`); an empty line no longer ends a `liquid` tag.
- **UTF-8**: `capitalize`, `truncate`, `truncatewords`, `size`, `slice` and `.size` count characters,
  not bytes: accented and non-Latin text is no longer corrupted.
- **`slice`** accepts a negative offset on strings and lists; **`truncate: 0`** gives the ellipsis;
  **`truncatewords: 0`** keeps one word; **`split`** drops trailing empty strings and splits on runs
  of whitespace with `" "`.
- **Arithmetic** is exact, as Ruby's BigDecimal: `0.1 | plus: 0.2` is `0.3`, `1.005 | round: 2` is
  `1.01`; two integers give an integer, with floor division (`14 | divided_by: 3` is `4`) and floor
  modulo; `"6-3" | plus: 0` is `6`.
- **`date`** follows Ruby's strftime: literal text in the format is kept (`"%d %B %Y, 2pm"`), and
  `%-d`, `%F`, `%T`, `%s`, `%u`, `%k`, `%l`, `%D`, `%R`, `%U`, `%V` and the other directives work. A
  float Unix timestamp (JSON) and more text formats are read.
- **`date`** reads zone names (`PST`, `EDT`, `CET`...) with their offset, whatever the server's
  zone; an unknown zone name (`SGT`) keeps the time as written, read as UTC.
- **`join`** without a separator uses `" "`; **`escape_once`** keeps existing entities;
  **`strip_html`** removes a tag written over several lines; **`sort`** compares numeric strings as
  strings.
- **Filter keyword arguments** reach the filter: `default: "x", allow_false: true` works.
- **`raw`, `comment`, `doc`** end at their first closing tag as in Ruby: `{%- raw -%}...{%- endraw -%}`
  works, and the raw body is never trimmed: `{% raw -%} a{% endraw %}` keeps the space before `a`.
- **Robustness**: a huge `date` width (`%5000d`) is copied as written, a huge `round` precision
  returns at once, and a length beyond the integer range in `truncate`, `truncatewords` or `slice`
  no longer crashes the render; a `for` range ending at the integer maximum stops.

### Changed
- **`contains` on a list** compares elements for equality, as Ruby's `Array#include?`:
  `["vip_gold"] contains "vip"` and `[1, 2] contains "1"` are now false. To search a property of
  list items, map it first: `{% assign titles = posts | map: "title" %}{% if titles contains "x" %}`.
- **`.size` on a map without a `size` key** gives its number of keys, as in Ruby: a template that
  tests an optional `size` field (`{% if item.size %}`) now sees a number when the field is missing.
  Test `item["size"]` instead. A `size` key still wins when it is present.
- **A filter given too many arguments** renders nothing: `{{ x | escape: 1 }}` rendered
  `Liquid error: internal`, and a keyword argument the filter does not take was ignored
  (`{{ "abc" | upcase: locale: "fr" }}` rendered `ABC`); both now render an empty string. Keyword
  arguments fill the filter's next parameter as a map, as in Ruby: `truncate: 5, omission: "!"`
  uses the map as the ellipsis.
- **`truncate` and `truncatewords`** with a length that is not an integer (`"7.0"`) keep the input.
- **`truncatewords`** splits on ASCII whitespace only (a non-breaking space is part of a word), and
  a text ending in whitespace gets the ellipsis, as in Ruby (`"a b " | truncatewords: 2` is `a b...`).
- **`slice` on a number** slices its text (`12345 | slice: 0, 2` is `12`).
- **`date` on a small number** reads it as a Unix timestamp, as Ruby does: `2024 | date: "%Y"` is
  `1970`. `%c` pads the day with a space (`Tue Mar  5`).
- **Module path** is now `github.com/Notifuse/liquidgo/v5`, so that Go accepts the `v5` tags:
  `go get github.com/Notifuse/liquidgo/v5@v5.11.0`. Replace `github.com/Notifuse/liquidgo/` with
  `github.com/Notifuse/liquidgo/v5/` in imports.
- `Version` is `5.11.0` (it said `5.10.0`), with a new `PortRevision` constant.

## [5.11.0]

Compatibility update matching [Shopify Liquid v5.11.0](https://github.com/Shopify/liquid/releases/tag/v5.11.0).

### Added
- `strict2` error mode as alias for `rigid` (backwards compatible)
- `LiquidError` interface for unified error handling
- Integration tests for `strict_filters` mode

### Removed
- **Snippet tag**: `{% snippet %}` tag removed (reverted per Shopify Liquid v5.11.0)
- `SnippetDrop` type removed

### Changed
- **Render tag**: Now only accepts string literals for template names
  ```liquid
  {% render 'template' %}    <!-- Works -->
  {% render variable %}      <!-- Error: invalid syntax -->
  ```
- Simplified `HandleError` using `LiquidError` interface (~100 → ~25 lines)
- Simplified panic recovery in variable.go and template.go

### Fixed
- **Concurrent template rendering** ([#2](https://github.com/Notifuse/liquidgo/issues/2)): Added mutex protection for thread-safe concurrent renders
- **Strict filters**: `strictFilters=true` now correctly raises `UndefinedFilter` errors
- **Error type preservation**: All 16 error types now preserved (7 were incorrectly converted to `InternalError`)
