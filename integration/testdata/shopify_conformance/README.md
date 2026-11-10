# Shopify Liquid conformance data

`cases.json` holds test cases taken from the test suite of
[Shopify Liquid](https://github.com/Shopify/liquid) v5.11.0 (commit `32b50ec`), the Ruby
implementation that liquidgo ports. `integration/shopify_conformance_test.go` replays every case
against liquidgo.

The cases are extracted, not executed, by `scripts/extract_shopify_conformance.py`:

- `assert_template_result(expected, template, assigns, ...)` calls written with literal values
  (`kind: "template"`, compared on the rendered output);
- `assert_equal(expected, @filters.name(...))` calls in `standard_filter_test.rb`
  (`kind: "filter"`, compared on the returned value, the way Ruby's `==` compares it).

The Ruby context of each call is kept: `modes` comes from `with_error_modes` or `error_mode:`
(`null` means the Ruby default: the case runs in `lax`, `strict` and `strict2`, as Ruby runs its
suite), and `raises` names the error a case inside `assert_raises` must produce. Cases inside any
other Ruby block are left out.

Each case also runs a second time with every integer turned into a `float64`, as
`encoding/json` decodes numbers (variant `json`, next to `ruby`).

## known_divergences.json

Every variant where liquidgo differs from Shopify is listed with its reason and liquidgo's actual
result, so that any change, a fix or a regression, fails the test. Reasons start with:

- `bug:` to fix;
- `minor:` a small divergence, to fix when cheap;
- `deliberate:` kept on purpose (see the README of the repository);
- `lax quirk:` malformed markup that Ruby's lax parser tolerates;
- `not comparable:` skipped, the Ruby output cannot be reproduced in Go (`skip: true`).

After a fix, regenerate the file and review the diff:

```sh
LIQUIDGO_CONFORMANCE_WRITE=1 go test ./integration -run TestShopifyConformanceRegenerate
```

## Licence

The test cases come from Shopify Liquid, released under the MIT licence; its notice is in
`SHOPIFY_LIQUID_LICENSE`.
