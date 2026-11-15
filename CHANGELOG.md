# Changelog

## [Unreleased]

### Added
- **Output escape** (`Environment.SetOutputEscape`), off by default: every `{{ }}` output and
  `echo` is HTML-escaped, except after the new `raw` filter, after `escape`, `escape_once` or `h`
  (no double escaping), and the text of a `capture`. `assign`, conditions and filters still see
  the data as it is. For HTML templates that output untrusted data, as LiquidJS's `outputEscape`.

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
