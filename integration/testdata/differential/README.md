# Differential corpus

`integration/differential_test.go` renders each template of this corpus with liquidgo, in lax
and strict mode, and compares the result with the one Shopify Liquid 5.11.0 gives.

- `templates.json`: templates written by hand, grouped by area (variable lookup, unknown tags,
  ranges, numbers, filters, lax markup, nesting...).
- `data.json`: the data every template is rendered with. Its integers are given to liquidgo as
  `int`, as Ruby has them.
- `expected.json`: the results of Shopify Liquid, written by `scripts/differential_reference.rb`,
  for the templates above and 400 templates it generates from a fixed seed. Never edit it by hand:

  ```sh
  git clone --depth 1 --branch v5.11.0 https://github.com/Shopify/liquid /tmp/shopify-liquid
  gem install strscan bigdecimal
  TZ=UTC ruby -I /tmp/shopify-liquid/lib scripts/differential_reference.rb
  ```

  CI runs the same command with `--check`.
- `differential_divergences.json`: every result of liquidgo that differs from Shopify's, with
  the reason (the prefixes of `../shopify_conformance/README.md`) and liquidgo's actual result,
  so that any change, a fix or a regression, fails the test. After a change, regenerate it and
  review the diff:

  ```sh
  LIQUIDGO_DIFFERENTIAL_WRITE=1 go test ./integration -run TestShopifyDifferential
  ```

  A new divergence is written with the reason `unclassified`, which fails the test until it is
  replaced by the real one.

A result matches when both render the same output, or both fail: error messages are not compared.
