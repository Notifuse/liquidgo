# Differences from Shopify Liquid

A few behaviours differ on purpose, because Go applications usually pass data decoded from JSON,
where every number is a `float64`, and render text meant for people (emails, pages):

- **An empty string is falsy**: `{% if "" %}` is false, so `{% if contact.first_name %}` skips an
  empty name. In Ruby, only `nil` and `false` are falsy. `where`, `reject`, `has`, `find` and
  `find_index` without a value also treat an empty list or hash as falsy, and with a value match a
  string to the number it writes (`where: "id", "10"` keeps an `id` of `10`).
- **Whole floats render without `.0`**: `{{ 2.0 }}` renders `2`, and `{{ 100 | plus: 1 }}` with a
  JSON number renders `101`, not `101.0`. A JSON integer stays a float in arithmetic, so
  `{{ 1999 | divided_by: 100 }}` with JSON data gives `19.99` (an integer literal gives `19`). A
  large whole float keeps its digits (`13835058055282164000`), where Ruby writes
  `1.3835058055282164e+19`.
- **A hash iterates in the order of its keys**: `{% for pair in hash %}` gives each `[key, value]`
  as Ruby does, but sorted by key, as a whole hash is written: a Go map keeps no order, where Ruby
  keeps the order the keys were added.
- **A filter error renders nothing**: `{{ 1 | divided_by: 0 }}`, an invalid `base64_decode`, `map`
  over integers and a filter given too many or too few arguments render an empty string instead of
  `Liquid error: ...`, and the error is recorded in `Result.Errors` (with strict filters it is
  raised, as in Ruby); a `truncate` or `truncatewords` length that is not an integer keeps the
  input.
- **Template-controlled sizes are bounded**: a `date` directive wider than 1024 (`%5000d`) is
  copied as written, a `round` precision beyond ±400 digits is not computed, and a filter builds
  at most 1,048,576 integers of a range (`join` writes at most as many), raising `Memory limits
  exceeded` beyond, where Ruby would build them.
- **A Go struct renders as nothing**, unless it is a drop or has `ToLiquid()`, `String()` or
  `Error()`; Ruby raises an internal error for an object without `to_liquid`. See
  [What a Template Reads](guide.md#what-a-template-reads).
- **Each render starts from its own data**: what a render assigns (`assign`, `capture`) is not
  kept in the template for the next render, so a value assigned while rendering for one recipient
  never reaches the next. Ruby keeps it.
- **A render takes a map**, where Ruby also takes a drop as the root of the variables: give the drop
  under a key.
- **Digits after a dot are an array index**: `{{ order.items.0.sku }}` reads the first item, as the
  paths of Notifuse's display conditions are written; Ruby takes `0` as the string key `"0"` and
  gives nil. A map still reads `h.0` as its key `"0"`. This is lax mode: the strict parser rejects a
  number after a dot, as Ruby's.
- **A float with no fraction is an array index**, as a JSON number is a `float64` in Go:
  `{{ arr[1.0] }}` and an index computed with `divided_by: 2.0` read an item, where Ruby takes only an
  Integer as an index.
- **`RaiseErrors` raises the errors of its own render only**; Ruby's `render!` leaves the template
  raising them in later renders.
- **A Go panic, a resource limit and a cancellation end the render** with the returned error and the
  output written so far; Ruby writes `Liquid error: internal` for an exception and goes on, and
  writes the message of a limit as the output.
- **`blank`**: `false` is not blank, and a string of Unicode whitespace (a no-break space included)
  is blank, as the display conditions of Notifuse expect; Shopify Liquid 5.14 makes `false` blank
  and counts ASCII whitespace only.
- **`date`** accepts a float Unix timestamp, and reads a date without a zone as UTC.
- **`increment` and `decrement`** counters start at 0, never at a render data value of the same
  name.
- **Partials share the resource limits** of the template that renders them: `render` and `include`
  do not reset them, as Ruby does. Ruby's cumulative limits (`cumulative_render_score_limit`,
  `cumulative_assign_score_limit`), which bound a render with its partials, are therefore not
  needed and not ported; each render has limits of its own, where Ruby's cumulative scores go on
  across the renders of a template.
- **Nesting is bounded at 100** for lookups and `liquid` tags too, as for blocks: `a[a[a[...]]]`
  nested deeper is a parse error.
- **A filter result longer than `render_length_limit`** raises `Memory limits exceeded`, even when
  it is assigned rather than written; the filters that can multiply their input check before they
  build it.

## Other differences

Every other output where liquidgo differs from Shopify Liquid 5.14.0 is listed with its reason, so
that a change fails a test:

- [`integration/testdata/shopify_conformance/known_divergences.json`](../integration/testdata/shopify_conformance/known_divergences.json),
  for the cases of Shopify's test suite;
- [`integration/testdata/differential/differential_divergences.json`](../integration/testdata/differential/differential_divergences.json),
  for the differential corpus.

A reason starts with `deliberate:` (above), `minor:` (to fix when cheap), `lax quirk:` (malformed
markup Ruby's lax parser tolerates) or `not comparable:` (Ruby objects with no Go equivalent).
