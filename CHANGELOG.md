# Changelog

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
