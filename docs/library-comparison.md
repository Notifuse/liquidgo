# Go Liquid libraries compared

Measured on 2026-11-17, with Shopify Liquid 5.14.0 as the reference; liquidgo v5.14.3.

## Libraries

| Library | Version | Error mode used | Resource limits |
|---|---|---|---|
| [Notifuse/liquidgo](https://github.com/Notifuse/liquidgo) (this one) | v5.14.3 | lax | yes, and a deadline (`context`) |
| [tobi/liquid-go](https://github.com/tobi/liquid-go) | commit of 2026-07-10 (no tag) | lax | no |
| [nchapman/go-liquid](https://github.com/nchapman/go-liquid) | commit of 2026-05-11 (no tag) | lax | yes |
| [go-liquid/liquid](https://github.com/go-liquid/liquid) | v0.2.0 | lax | no |
| [osteele/liquid](https://github.com/osteele/liquid) | v1.9.2 | none (no modes) | no |
| [acstech/liquid](https://github.com/acstech/liquid) | v1.1.0 | none (no modes) | no |
| [Shopify Liquid](https://github.com/Shopify/liquid), the reference | 5.14.0 on Ruby 3.3 with YJIT | lax | yes |

**Left out:** [karlseguin/liquid](https://github.com/karlseguin/liquid), unchanged since 2015 and
without a `go.mod`, which acstech/liquid continues; [joaqu1m/liquid](https://github.com/joaqu1m/liquid),
a fork of osteele/liquid that keeps its module path.

## Cases

962 templates, each with the output Shopify Liquid 5.14.0 renders for it in lax mode:

- the 360 template cases of Shopify Liquid's own test suite that render without partials and
  without raising (`integration/testdata/shopify_conformance/cases.json`, extracted from the Ruby
  tests);
- the 602 templates of liquidgo's differential corpus (`integration/testdata/differential`),
  rendered by Ruby with the data of `data.json`, the strict variables runs left out.

Shopify's tests are independent of liquidgo. The differential corpus is liquidgo's own test data:
it was written while porting, so it favours liquidgo, and its column is given apart.

Left out of both: the outputs where Ruby writes a hash with more than one key, which Ruby writes in
the order its keys were added and a Go map does not keep (20 cases).

## Method

1. `tools/libcmp` parses and renders each case with each library, in a process of its own per
   library, with the data decoded as Ruby holds it (an integer is an `int`, a float a `float64`)
   and the local time zone set to UTC, as the reference was. A parse error, a render error, a
   panic or a case that takes more than 10 s counts as a template the library cannot render, but
   for a template the library refuses on purpose and lists as a deliberate difference, which counts
   as a difference.
2. An output matches when it is the same string as Shopify's.
3. Hostile templates: each of four templates written to exhaust the process (a billion
   iterations, a string doubled 40 times, a billion bytes of output, 100,000 nested blocks) runs in
   a Docker container bounded to 1 GiB of memory and 20 s (`tools/libcmp/hostile.sh`). A library
   contains the template when it stops it with an error, refuses it, or renders it within the
   bounds. The libraries that take resource limits get the same: 1 MiB of output, a render score of
   100,000 and 1 MiB assigned.
4. Time: the median of 15 rounds of at least 200 ms of renders of a newsletter template (a loop
   over 50 products with conditions and filters, `tools/libcmp/bench.go`), parsed once, on GitHub's
   Linux arm64 runner (`.github/workflows/library-comparison.yml`): each library three times in
   turn, the middle run given; the three runs differ by less than 2%. Shopify Liquid renders the
   same template and data with Ruby 3.3 and YJIT (`tools/libcmp/ruby/bench.rb`). acstech/liquid has
   no `round` filter, so the template uses none.
5. Shopify Liquid also renders the hostile templates, with the same resource limits, in the same
   bounded container (`tools/libcmp/ruby.sh`).

To run it again, from the root of the repository:

```bash
cd tools/libcmp
mkdir -p /tmp/libcmp
for lib in liquidgo tobi nchapman goliquid osteele acstech; do go run . render $lib /tmp/libcmp/$lib.json; done
go run . report /tmp/libcmp/*.json
for lib in liquidgo tobi nchapman goliquid osteele acstech; do go run . bench $lib; done
./hostile.sh
./ruby.sh path/to/a/checkout/of/shopify/liquid/v5.14.0
```

The times of the table come from the workflow "Library comparison", started by hand.

## Results

| Library | Matches Shopify | Shopify's tests | Differential corpus | Templates it cannot render | Hostile templates contained | Newsletter template |
|---|---|---|---|---|---|---|
| Notifuse/liquidgo | **903 (93.9%)** | 355 / 360 | **548 / 602** | **0** | **4 / 4** | **53 µs** |
| tobi/liquid-go | 841 (87.4%) | **357 / 360** | 484 / 602 | 0 | 0 / 4 | 317 µs |
| nchapman/go-liquid | 760 (79.0%) | 321 / 360 | 439 / 602 | 71 | 3 / 4 | 88 µs |
| go-liquid/liquid | 618 (64.2%) | 271 / 360 | 347 / 602 | 194 | 1 / 4 | 76 µs |
| osteele/liquid | 502 (52.2%) | 252 / 360 | 250 / 602 | 336 | 1 / 4 | 304 µs |
| acstech/liquid | 323 (33.6%) | 193 / 360 | 130 / 602 | 488 | 1 / 4 | 54 µs |
| Shopify Liquid (Ruby 3.3, YJIT) | the reference | | | | **4 / 4** | 368 µs |

- **liquidgo**: each of its 59 differences is listed in its divergence files with its reason. 50
  are deliberate (counters start at 0, an empty string is false, a filter error writes nothing,
  whole floats without `.0`, `false` is not blank: see [the differences](differences-from-shopify.md)),
  4 are malformed markup Ruby's lax parser tolerates and 5 compare Ruby objects with no Go
  equivalent; none is left to fix. It refuses one template on purpose, a lookup nested more than 100 deep, which would
  exhaust the stack of a deeper one: a deliberate difference, not a template it cannot render.
- **tobi/liquid-go** comes closest on Shopify's tests. It compares an `int` from Go data as a
  string (`{% if product.discount > 0 %}` gives `Liquid error: comparison of String with 0
  failed`), and has no bound: each hostile template runs out of time or memory.
- **nchapman/go-liquid** stops three hostile templates with its limits; 100,000 nested blocks run
  out of time while parsing. It lacks `{% liquid %}` blocks and part of the lax syntax.
- **go-liquid/liquid** has no `{% liquid %}`, `{% echo %}` or `{% doc %}` tag.
- **osteele/liquid** has no `increment` or `decrement` tag, a `cycle` that takes no group, and stops
  the render on an error, where Shopify writes it and goes on.
- **acstech/liquid** has no `cycle`, `increment`, `decrement` or `tablerow` tag, no `round` filter,
  and fails on whitespace control (`{{-`). It renders the newsletter template in 54 µs, the closest
  to liquidgo's 53 µs: the three runs of each do not overlap (53.1 to 53.3 µs against 54.1 to
  54.8 µs).
- Hostile templates: 100,000 nested blocks render with osteele/liquid, acstech/liquid and
  go-liquid/liquid; the other three templates exhaust them. liquidgo refuses the nesting while
  parsing and stops the other three with its limits.
- **Shopify Liquid** refuses the nesting and stops the other three with its limits: since 5.14 its
  `for` counts the items of a range as it visits them, where 5.11 built the whole range first and
  ran out of memory on the billion iterations. It renders the newsletter template in 368 µs, 6.9
  times the time of liquidgo.
