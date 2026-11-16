# Go Liquid libraries compared

Measured on 2026-11-16, with Shopify Liquid 5.11.0 as the reference.

## Libraries

| Library | Version | Error mode used | Resource limits |
|---|---|---|---|
| [Notifuse/liquidgo](https://github.com/Notifuse/liquidgo) (this one) | v5.11.10 | lax | yes, and a deadline (`context`) |
| [tobi/liquid-go](https://github.com/tobi/liquid-go) | commit of 2026-07-10 (no tag) | lax | no |
| [nchapman/go-liquid](https://github.com/nchapman/go-liquid) | commit of 2026-05-11 (no tag) | lax | yes |
| [go-liquid/liquid](https://github.com/go-liquid/liquid) | v0.2.0 | lax | no |
| [osteele/liquid](https://github.com/osteele/liquid) | v1.9.2 | none (no modes) | no |
| [acstech/liquid](https://github.com/acstech/liquid) | v1.1.0 | none (no modes) | no |

**Left out:** [karlseguin/liquid](https://github.com/karlseguin/liquid), unchanged since 2015 and
without a `go.mod`, which acstech/liquid continues; [joaqu1m/liquid](https://github.com/joaqu1m/liquid),
a fork of osteele/liquid that keeps its module path.

## Cases

938 templates, each with the output Shopify Liquid 5.11.0 renders for it in lax mode:

- the 350 template cases of Shopify Liquid's own test suite that render without partials and
  without raising (`integration/testdata/shopify_conformance/cases.json`, extracted from the Ruby
  tests);
- the 588 templates of liquidgo's differential corpus (`integration/testdata/differential`),
  rendered by Ruby with the data of `data.json`, the strict variables runs left out.

Shopify's tests are independent of liquidgo. The differential corpus is liquidgo's own test data:
it was written while porting, so it favours liquidgo, and its column is given apart.

Left out of both: the outputs where Ruby writes a hash with more than one key, which Ruby writes in
the order its keys were added and a Go map does not keep (20 cases).

## Method

1. `tools/libcmp` parses and renders each case with each library, in a process of its own per
   library, with the data decoded as Ruby holds it (an integer is an `int`, a float a `float64`)
   and the local time zone set to UTC, as the reference was. A parse error, a render error, a
   panic or a case that takes more than 10 s counts as a failure.
2. An output matches when it is the same string as Shopify's.
3. Hostile templates: each of four templates written to exhaust the process (a billion
   iterations, a string doubled 40 times, a billion bytes of output, 100,000 nested blocks) runs in
   a Docker container bounded to 1 GiB of memory and 20 s (`tools/libcmp/hostile.sh`). A library
   contains the template when it stops it with an error, refuses it, or renders it within the
   bounds. The libraries that take resource limits get the same: 1 MiB of output, a render score of
   100,000 and 1 MiB assigned.
4. Time: the median of 15 rounds of at least 200 ms of renders of a newsletter template (a loop
   over 50 products with conditions and filters, `tools/libcmp/bench.go`), parsed once, on an
   Apple M-series laptop. acstech/liquid has no `round` filter, so the template uses none.

To run it again, from the root of the repository:

```bash
cd tools/libcmp
mkdir -p /tmp/libcmp
for lib in liquidgo tobi nchapman goliquid osteele acstech; do go run . render $lib /tmp/libcmp/$lib.json; done
go run . report /tmp/libcmp/*.json
for lib in liquidgo tobi nchapman goliquid osteele acstech; do go run . bench $lib; done
./hostile.sh
```

## Results

| Library | Matches Shopify | Shopify's tests | Differential corpus | Render failures | Hostile templates contained | Newsletter template |
|---|---|---|---|---|---|---|
| Notifuse/liquidgo | **848 (90.4%)** | 328 / 350 | **520 / 588** | **1** | **4 / 4** | 46 µs |
| tobi/liquid-go | 821 (87.5%) | **347 / 350** | 474 / 588 | 0 | 0 / 4 | 185 µs |
| nchapman/go-liquid | 738 (78.7%) | 307 / 350 | 431 / 588 | 69 | 3 / 4 | 128 µs |
| go-liquid/liquid | 607 (64.7%) | 262 / 350 | 345 / 588 | 192 | 1 / 4 | 47 µs |
| osteele/liquid | 497 (53.0%) | 247 / 350 | 250 / 588 | 324 | 1 / 4 | 153 µs |
| acstech/liquid | 315 (33.6%) | 186 / 350 | 129 / 588 | 480 | 1 / 4 | 28 µs |

- **liquidgo**: each of its 90 differences is listed in its divergence files with its reason. 40
  are deliberate (counters start at 0, an empty string is false, a filter error writes nothing,
  whole floats without `.0`: see [the differences](differences-from-shopify.md)), 25 are malformed
  markup Ruby's lax parser tolerates, 20 are minor and 5 compare Ruby objects with no Go
  equivalent. Its render failure is a lookup nested more than 100 deep, refused on purpose.
- **tobi/liquid-go** comes closest on Shopify's tests. It compares an `int` from Go data as a
  string (`{% if product.discount > 0 %}` gives `Liquid error: comparison of String with 0
  failed`), and has no bound: each hostile template runs out of time or memory.
- **nchapman/go-liquid** stops three hostile templates with its limits; 100,000 nested blocks run
  out of time while parsing. It lacks `{% liquid %}` blocks and part of the lax syntax.
- **go-liquid/liquid** has no `{% liquid %}`, `{% echo %}` or `{% doc %}` tag.
- **osteele/liquid** has no `increment` or `decrement` tag, a `cycle` that takes no group, and stops
  the render on an error, where Shopify writes it and goes on.
- **acstech/liquid** has no `cycle`, `increment`, `decrement` or `tablerow` tag, no `round` filter,
  and fails on whitespace control (`{{-`).
- Hostile templates: 100,000 nested blocks render with osteele/liquid, acstech/liquid and
  go-liquid/liquid; the other three templates exhaust them. liquidgo refuses the nesting while
  parsing and stops the other three with its limits.
