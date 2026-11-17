# liquidgo - The best Liquid engine in the world

A Go port of [Shopify Liquid](https://github.com/Shopify/liquid) 5.14: the closest to Shopify and the
fastest of the Go engines, the only Go engine that stops every hostile template, as Shopify's
own does, and 7 times faster than Shopify's. Every difference from Shopify is listed.

It renders the emails and blog pages of [Notifuse](https://www.notifuse.com/), the open-source
emailing platform, where every template is written by its users.

[![Tests](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml)
[![Coverage](https://codecov.io/github/notifuse/liquidgo/graph/badge.svg?token=3ATZBZSIYR)](https://codecov.io/github/notifuse/liquidgo)
[![Go Reference](https://pkg.go.dev/badge/github.com/Notifuse/liquidgo/v5.svg)](https://pkg.go.dev/github.com/Notifuse/liquidgo/v5)
![Liquid](https://img.shields.io/badge/Liquid-5.14-orange)
![License](https://img.shields.io/badge/license-MIT-blue)

## Go Liquid libraries compared

Measured on 962 templates against Shopify Liquid 5.14.0 on 2026-11-17 ([method](docs/library-comparison.md)):

| Library | Same output as Shopify | Templates it cannot render | Hostile templates contained | Newsletter template |
|---|---|---|---|---|
| **liquidgo** | **93.9%** | **0** | **4 / 4** | **53 µs** |
| [tobi/liquid-go](https://github.com/tobi/liquid-go) | 87.4% | 0 | 0 / 4 | 317 µs |
| [nchapman/go-liquid](https://github.com/nchapman/go-liquid) | 79.0% | 71 | 3 / 4 | 88 µs |
| [go-liquid/liquid](https://github.com/go-liquid/liquid) | 64.2% | 194 | 1 / 4 | 76 µs |
| [osteele/liquid](https://github.com/osteele/liquid) | 52.2% | 336 | 1 / 4 | 304 µs |
| [acstech/liquid](https://github.com/acstech/liquid) | 33.6% | 488 | 1 / 4 | 54 µs |
| [Shopify Liquid](https://github.com/Shopify/liquid) (Ruby 3.3, YJIT) | the reference | | 4 / 4 | 368 µs |

The templates are the cases of Shopify Liquid's own test suite and liquidgo's differential corpus,
which favours liquidgo; on Shopify's tests alone, tobi/liquid-go matches 357 of 360 and liquidgo 355,
the others being liquidgo's [deliberate differences](docs/differences-from-shopify.md) and markup
Ruby's lax parser tolerates. A hostile template is written to exhaust the process (a billion
iterations, a string doubled 40 times, a billion bytes of output, 100,000 nested blocks); Shopify
Liquid stops all four with its resource limits since 5.14, which no longer builds a range whole.

## Why liquidgo

- **Same output as Shopify Liquid.** A port of the Ruby gem 5.14.0, file by file. Shopify's test
  suite and a differential corpus run against the gem in CI; every difference is listed with its
  reason.
- **Safe with untrusted templates.** A template cannot crash your program, exhaust its memory or CPU,
  or read anything of your Go values but what you give it. See [Safety](#safety).
- **Fast.** The fastest Go engine on the newsletter template, and 7 times faster than Shopify
  Liquid with YJIT. Parse once, render from any number of
  goroutines; renders share the parse of a partial.
- **Explicit.** Filters are declared with `liquid.Filter`, and drops declare the keys a template
  reads: nothing is found by reflection.
- **Pure Go.** One dependency, `gopkg.in/yaml.v3`, for the translations of error messages.

## Install

```bash
go get github.com/Notifuse/liquidgo/v5
```

```go
package main

import (
    "context"
    "fmt"

    "github.com/Notifuse/liquidgo/v5/liquid"
    "github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func main() {
    env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard()})
    if err != nil {
        panic(err)
    }

    tmpl, err := liquid.ParseTemplate(
        "Hello {{ name | capitalize }}! {% for p in products %}{{ p.title }} {% endfor %}",
        &liquid.TemplateOptions{Environment: env},
    )
    if err != nil {
        panic(err)
    }
    res, err := tmpl.Render(context.Background(), map[string]interface{}{
        "name":     "ann",
        "products": []interface{}{map[string]interface{}{"title": "Pen"}},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(res.Output) // Hello Ann! Pen
}
```

## API

```go
tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
res, err := tmpl.Render(ctx, data, opts...) // res.Output, res.Errors
```

A template is parsed once and rendered from any number of goroutines: a render changes nothing in
the template or in the data. The options of a render:

| Option | Effect |
|---|---|
| `liquid.WithLimits(liquid.ResourceLimitsConfig{...})` | bounds the bytes written, the nodes rendered and the bytes assigned |
| `liquid.WithFilters(liquid.Filter(name, fn), ...)` | filters of this render, before the environment's |
| `liquid.WithFileSystem(fs)` | where `include` and `render` read partials |
| `liquid.WithRegisters(map)` | values a custom tag reads |
| `liquid.WithStrictVariables()`, `liquid.WithStrictFilters()` | an undefined variable or filter is an error |
| `liquid.RaiseErrors()` | the first error ends the render, as Ruby's `render!` |

The context bounds the time: past its deadline, the render stops at its next node.

### Errors

In lax mode, as in Ruby, an error writes its message in place of its tag and the render goes on;
every error is in `res.Errors`. A filter error writes nothing (see the
[differences](docs/differences-from-shopify.md)). `Render` returns an error only when the render
could not complete: a `*liquid.MemoryError` past a limit, a `*liquid.CanceledError` past the
deadline, a `*liquid.PanicError` when Go code panicked (a bug, with its stack), or the first error
with `RaiseErrors`.

### Filters and drops

```go
shout := liquid.Filter("shout", func(ctx *liquid.Context, input interface{}, args ...interface{}) (interface{}, error) {
    return strings.ToUpper(fmt.Sprint(input)) + "!", nil
})
env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard(), Filters: []liquid.FilterDef{shout}})

// A drop declares the keys a template reads: {{ user.first_name }}
func (u *User) LiquidGet(ctx *liquid.Context, key string) (interface{}, bool, error) {
    switch key {
    case "first_name":
        return u.FirstName, true, nil
    }
    return nil, false, nil
}
```

A template reads maps, lists, numbers, strings, booleans, times, what `ToLiquid()` returns and the
keys of drops; a Go struct exposes nothing. The [guide](docs/guide.md) covers custom tags,
environments, the output escape, partials and the parse tree visitor.

## Safety

A template is untrusted input. Each threat is countered and tested:

| Threat | Countered by | Tests |
|---|---|---|
| A loop, a string or an output that grows without end | the render length, render score and assign score limits, checked as the render goes, filters included | `TestHostileTemplatesAreBounded` (25 templates), `TestFilterResultsStayWithinTheRenderLength` |
| A render that runs too long | the deadline of the context, checked at each node | `TestRenderLimitsAndCancellation` |
| Nesting deep enough to exhaust the stack | blocks, lookups and brackets nested more than 100 deep are refused while parsing | `TestSecurity_MoreThanMaxDepthNestedBlocksRaisesException`, `TestHostileTemplatesAreBounded` |
| A partial that renders itself | the depth of partials is bounded | `TestRecursivePartialsAreBounded` |
| Reading the fields or methods of a Go value | nothing is read by reflection: only maps, lists, `ToLiquid()` and the keys a drop declares | `TestAPlainStructExposesNothing`, `TestATemplateCannotCallAMethodOfAPlainStruct` |
| Writing a struct's fields or an address in the output | a struct without `String()`, a channel or a func renders as nothing | `TestADropNeverWritesItsFields`, `TestADropOnASliceNeverWritesTheFieldsItHolds`, `TestAReflectValueWritesNoField`, `TestAChannelOrAFuncWritesNoAddress` |
| Data leaking from one render to the next | a render keeps its assigns, errors and drops to itself | `TestAuditAssignDoesNotLeakBetweenRenders`, `TestAuditDropDoesNotLeakBetweenRenders`, `TestRenderIsolatesConcurrentRenders` |
| Reading a file outside the partials' directory | the local file system resolves links and refuses paths out of its root | `TestLocalFileSystemReadsOnlyUnderItsRoot`, `TestLocalFileSystemALoopingLink` |
| Values of the data in error messages | an error shows no value of the data, but a key or a partial name the template reads from a variable, as Ruby | `TestErrorsShowNoValueOfTheData` |
| A cache that grows with the templates | the parse caches and the shared partial cache are bounded | `TestParseCachesAreBounded`, `TestThePartialCacheIsBounded` |
| A panic that crashes the program | a panic during a render is returned as a `*liquid.PanicError` | `TestRenderGoPanicIsFatal` |
| Concurrent renders corrupting each other | renders share only what they never write | `TestConcurrentRendersOfPartials`, the suite under `-race` |

Limits are off by default, as in Ruby: set them for untrusted templates (see the
[guide](docs/guide.md#resource-limits)).

## Tests

- **Shopify's test suite**: the cases of Shopify Liquid 5.14.0's tests replayed in each error mode,
  with Ruby numbers and JSON numbers ([its README](integration/testdata/shopify_conformance/README.md)).
- **Differential corpus**: 293 templates, each in lax and strict mode, whose expected outputs CI
  renders with the Ruby gem on each pull request; liquidgo matches each or lists why not.
- **Fuzzing**: two minutes on each pull request, for panics and outputs that change from one parse
  to the next.
- **Guard mutations**: 36 checks of the library are removed one at a time, and a test must fail
  each time (`scripts/guard_mutations`), before each release.
- **Platforms**: the suite runs on amd64, arm64, 386 and Windows, and under the race detector;
  coverage stays over 90%.

```bash
go test -race ./...
```

## Performance

The newsletter template of the comparison (a loop over 50 products with conditions and filters)
renders in 53 µs on GitHub's arm64 runner, against 54 µs for the next Go engine and 368 µs for
Shopify Liquid with YJIT. The benchmarks of [`performance/`](performance/) render a Shopify theme; a render
of v5.11.10 allocates half as much as v5.11.9.

## Differences from Shopify Liquid

A few behaviours differ on purpose, for Go data decoded from JSON and text read by people:

- an empty string is false in a condition, so `{% if contact.first_name %}` skips an empty name;
- a whole float renders without `.0` (`{{ 2.0 }}` gives `2`);
- a filter error writes nothing, and is recorded in `res.Errors`;
- a render keeps its assigns to itself, and an `increment` counter starts at 0, never at a value
  of the data of the same name.

Every difference, deliberate or not, is listed with its reason:
[differences from Shopify Liquid](docs/differences-from-shopify.md).

## Version

Current version: **5.14.0**, port revision 6 (`liquid.Version`, `liquid.PortRevision`). The minor
version is Shopify's: port revision N of Shopify Liquid 5.14.0 is tagged `v5.14.(N-1)`, so
revision 6 is `v5.14.5`. `v5.14.0` is retracted: it made `false` blank, which changed the display
conditions of Notifuse. The port of 5.11.0 ended at revision 12, `v5.11.11`.

## Contributing

Each change follows the Ruby implementation in `reference-liquid/`, with tests ported from its
suite; a bug fix comes with a test that fails without it. See [IMPLEMENTATION.md](IMPLEMENTATION.md).

## License

MIT. Shopify Liquid, which liquidgo ports, is © Shopify Inc., also under the MIT License.

Created and maintained by Pierre Bazoge ([@pierre-b](https://github.com/pierre-b)).
