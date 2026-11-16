# Liquid Go - Liquid Template Engine for Go

> **💌 Built by [Notifuse](https://www.notifuse.com/)** - The modern open-source emailing platform to send newsletters, transactional emails & write blogs.
> Notifuse uses Liquid templating to personalize emails and blog templates with variables like `{{ contact.first_name }}`. Self-hosted, free forever, and a modern alternative to Mailchimp, Resend etc...
> [Try the live demo →](https://www.notifuse.com/)

[![Tests](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml)
[![Benchmarks](https://github.com/Notifuse/liquidgo/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/benchmarks.yml)
[![codecov](https://codecov.io/github/notifuse/liquidgo/graph/badge.svg?token=3ATZBZSIYR)](https://codecov.io/github/notifuse/liquidgo)
[![Go Reference](https://pkg.go.dev/badge/github.com/Notifuse/liquidgo/v5.svg)](https://pkg.go.dev/github.com/Notifuse/liquidgo/v5)
[![Go Report Card](https://goreportcard.com/badge/github.com/Notifuse/liquidgo)](https://goreportcard.com/report/github.com/Notifuse/liquidgo)

A full-featured Go implementation of [Shopify's Liquid template engine](https://github.com/Shopify/liquid), maintaining feature parity with the Ruby version.

- [Documentation](#documentation)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Contributing](IMPLEMENTATION.md)

## Introduction

Liquid is a template engine written with specific requirements:

- **Beautiful and simple markup** - Clean, readable template syntax
- **Secure and non-evaluating** - Safe for user-generated templates without code execution
- **Stateless** - Separate parse and render phases for optimal performance

## Why Liquid Go?

- ✅ **Full feature parity** with Ruby Liquid 5.11.0
- ⚡ **High performance** - 3-10x faster than Ruby implementation
- 🔒 **Secure** - Safe for user-generated templates
- 📦 **Zero dependencies** - Pure Go implementation
- 🧪 **Superior test coverage** - ~90% code coverage with comprehensive edge case testing, exceeding the official Ruby library's test suite
- 🎯 **Production ready** - Used in real-world applications

## Installation

```bash
go get github.com/Notifuse/liquidgo/v5
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/Notifuse/liquidgo/v5/liquid"
)

func main() {
    // Parse template
    tmpl, err := liquid.ParseTemplate("Hello {{ name }}!", nil)
    if err != nil {
        panic(err)
    }

    // Render with data
    res, err := tmpl.Render(context.Background(), map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(res.Output) // Output: Hello World!
}
```

### Using Tags and Filters

```go
package main

import (
    "context"
    "fmt"
    "github.com/Notifuse/liquidgo/v5/liquid"
    "github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func main() {
    // Create environment with standard tags
    env := liquid.NewEnvironment()
    tags.RegisterStandardTags(env)

    // Parse template with conditionals and loops
    source := `
    {% if user %}
        <h1>Hello {{ user.name | capitalize }}!</h1>
        <ul>
        {% for item in user.items %}
            <li>{{ item }}</li>
        {% endfor %}
        </ul>
    {% else %}
        <p>Please log in.</p>
    {% endif %}
    `

    tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{
        Environment: env,
    })
    if err != nil {
        panic(err)
    }

    // Render with nested data
    res, err := tmpl.Render(context.Background(), map[string]interface{}{
        "user": map[string]interface{}{
            "name": "john doe",
            "items": []string{"apple", "banana", "cherry"},
        },
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(res.Output)
}
```

### What a Template Reads

A template reads maps (of any key and value type), slices and arrays, strings, numbers, booleans
and `time.Time`, which renders and takes the `date` filter. A Go struct exposes nothing to a
template, neither fields nor methods, and renders as nothing (`nil` inside a list or a hash). As
in Ruby, only a drop has its methods invoked; Ruby raises an internal error for an object without
`to_liquid`, where liquidgo renders nothing. A pointer renders as what it points to, and a nil
pointer as nothing. A hash renders with its keys in sorted order. To give a struct to a template:

- implement `ToLiquid() interface{}` (with a value receiver) and return a map of what the template
  may read:

  ```go
  func (u User) ToLiquid() interface{} {
      return map[string]interface{}{"name": u.Name}
  }
  ```

- or make it a drop by embedding `*liquid.Drop`, initialised with `liquid.NewDrop()`: its exported
  methods without arguments and its exported fields become readable (`{{ user.first_name }}` calls
  `FirstName()`), the methods it gets from embedded types included. Give a drop only the methods a
  template may call.

`ToLiquid()` is applied at each step of a lookup, to the items the array filters read and to a
value written in the output. A type with a `String()` or `Error()` method renders that text: a
generated type whose `String()` writes all its fields (protobuf messages, for example) shows them
all, so give a template a map or a drop instead.

### Custom Filters

```go
package main

import (
    "context"
    "fmt"
    "strings"
    "github.com/Notifuse/liquidgo/v5/liquid"
)

// Define custom filter
type MyFilters struct{}

func (f *MyFilters) Reverse(input string) string {
    runes := []rune(input)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

func (f *MyFilters) Shout(input string) string {
    return strings.ToUpper(input) + "!!!"
}

func main() {
    // Create environment and register filters
    env := liquid.NewEnvironment()
    env.RegisterFilter(&MyFilters{})

    tmpl, _ := liquid.ParseTemplate(
        "{{ 'hello' | reverse | shout }}",
        &liquid.TemplateOptions{Environment: env},
    )

    res, _ := tmpl.Render(context.Background(), nil)
    fmt.Println(res.Output) // Output: OLLEH!!!
}
```

### Custom Tags

A tag is a `liquid.Node`: it has one render method, `RenderToOutput`, which
writes to the render's output buffer. `*liquid.Tag` has no render method, so
a tag that embeds it declares `RenderToOutput`, or it is not a `Node`; a tag
that embeds `*liquid.Block` renders its body unless it declares its own. A
tag constructor returns `(liquid.Node, error)`: an error, or no tag, is a
syntax error of the template.

```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/Notifuse/liquidgo/v5/liquid"
)

// Custom tag implementation
type GreetingTag struct {
    *liquid.Tag
    name string
}

func NewGreetingTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*GreetingTag, error) {
    return &GreetingTag{
        Tag:  liquid.NewTag(tagName, markup, parseContext),
        name: strings.TrimSpace(markup), // the markup keeps the space before %}
    }, nil
}

func (g *GreetingTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
    ctx := context.Context().(*liquid.Context)
    name := ctx.FindVariable(g.name, false)
    output.WriteString(fmt.Sprintf("Greetings, %v!", name))
}

func main() {
    // Register custom tag
    env := liquid.NewEnvironment()
    env.RegisterTag("greet", func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
        tag, err := NewGreetingTag(tagName, markup, parseContext)
        if err != nil {
            return nil, err
        }
        return tag, nil
    })

    tmpl, _ := liquid.ParseTemplate(
        "{% greet user_name %}",
        &liquid.TemplateOptions{Environment: env},
    )

    res, _ := tmpl.Render(context.Background(), map[string]interface{}{
        "user_name": "Alice",
    })

    fmt.Println(res.Output) // Output: Greetings, Alice!
}
```

## Template Syntax

### Variables

```liquid
{{ variable }}
{{ object.property }}
{{ array[0] }}
```

### Filters

```liquid
{{ "hello" | capitalize }}
{{ product.price | money }}
{{ "now" | date: "%Y-%m-%d" }}
```

Filters can be chained:

```liquid
{{ "HELLO world" | downcase | capitalize }}
```

### Tags

#### Control Flow

```liquid
{% if user.age >= 18 %}
    Adult content
{% elsif user.age >= 13 %}
    Teen content
{% else %}
    Child content
{% endif %}

{% unless user.subscribed %}
    Subscribe now!
{% endunless %}

{% case product.type %}
{% when "shirt" %}
    Clothing item
{% when "book" %}
    Reading material
{% else %}
    Other product
{% endcase %}
```

#### Loops

```liquid
{% for item in array %}
    {{ forloop.index }}: {{ item }}
{% endfor %}

{% for i in (1..10) %}
    Number {{ i }}
{% endfor %}

{% tablerow product in collection.products %}
    {{ product.title }}
{% endtablerow %}
```

#### Variable Assignment

```liquid
{% assign name = "John" %}
{% capture greeting %}
    Hello {{ name }}!
{% endcapture %}
```

#### Comments

```liquid
{% comment %}
    This won't be rendered
{% endcomment %}

{% # This is an inline comment %}
```

## Environments

Use environments to encapsulate custom tags, filters, and configurations:

```go
package main

import (
    "github.com/Notifuse/liquidgo/v5/liquid"
    "github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func main() {
    // Create isolated environment
    userEnv := liquid.NewEnvironment()
    tags.RegisterStandardTags(userEnv)
    userEnv.RegisterFilter(&MyCustomFilters{})

    // Use environment in template
    tmpl, _ := liquid.ParseTemplate(source, &liquid.TemplateOptions{
        Environment: userEnv,
    })
}
```

Benefits of environments:

- **Encapsulation** - Keep different contexts separate
- **Security** - Limit available tags/filters per context
- **Maintainability** - Clearer scope of customizations
- **No conflicts** - Avoid global state issues

## Error Handling

Liquid supports three error modes:

```go
env := liquid.NewEnvironment()

// Lax mode (default) - render errors inline
env.SetErrorMode("lax")

// Warn mode - collect warnings
env.SetErrorMode("warn")

// Strict mode - return errors immediately
env.SetErrorMode("strict")
```

### Rendering and errors

`Template.Render(ctx, data, opts...)` renders in a state of its own and changes nothing in the template or
in the data: a template is parsed once and rendered from any number of goroutines, and renders never see
each other's assigns, errors or limits. It reads the template's assigns and registers. The data is a map;
a drop goes under a key (`{"page": drop}`), and is bound to the render that reads it: give each render its
own drop instances.

```go
res, err := tmpl.Render(ctx, data,
    liquid.WithFileSystem(partials),         // partials of this render, nested ones included
    liquid.WithLimits(limits),               // limits set here replace the template's
    liquid.WithRegisters(registers),         // read by tags and filters
    liquid.WithFilters(&TenantFilters{}),    // local filters, next to the environment's
    liquid.WithStrictVariables(),            // an undefined variable is an error
)
switch {
case err != nil:
    // *liquid.MemoryError (a limit), *liquid.CanceledError (ctx is done),
    // *liquid.PanicError (Go code panicked, with its Stack); res.Output is
    // what was written before the render stopped
case len(res.Errors) > 0:
    // errors the render recorded and went on, as Ruby: most wrote
    // "Liquid error: ..." in res.Output; an undefined variable, filter or
    // drop method writes nothing
}
fmt.Println(res.Output, res.RenderScore, res.AssignScore)
```

`liquid.RaiseErrors()` makes the first recorded error end the render, as Ruby's `render!`. A Go panic in a
tag, a filter or a drop ends the render with a `*liquid.PanicError`, where Ruby writes
`Liquid error: internal` and goes on; a Liquid error they raise is recorded as in Ruby.

## Performance

Liquid Go is optimized for performance:

### Benchmarks (Apple M1 Pro)

```
BenchmarkTokenize       1981 ops    540 µs/op     253 KB/op
BenchmarkParse           180 ops   6.66 ms/op    2.59 MB/op
BenchmarkRender          100 ops  10.43 ms/op   20.90 MB/op
BenchmarkFull             67 ops  18.49 ms/op   24.09 MB/op
```

### Performance Tips

1. **Parse once, render many** - Templates are compiled once and reused
2. **Use environments** - Pre-register filters and tags
3. **Profile with Go** - `pprof` shows where a render spends its time
4. **Cache templates** - Store compiled templates in memory

See [`performance/`](performance/) directory for detailed benchmarks.

## Advanced Features

### Resource Limits

```go
env := liquid.NewEnvironment()
env.SetDefaultResourceLimits(map[string]interface{}{
    "render_length_limit": 1000000,  // 1MB output limit
    "render_score_limit":  100000,   // Complexity limit
})
```

### Output Escape

For HTML templates that output untrusted data (contact fields, form input, feeds), turn on the
output escape, as LiquidJS's `outputEscape` option:

```go
env := liquid.NewEnvironment()
env.SetOutputEscape(true)
```

Every `{{ }}` output, `echo` and `cycle` is then HTML-escaped, so a value cannot add markup to the
page. An output is written as it is when its value is HTML on purpose:

- the result of `raw` (`{{ feed.html | raw }}`), or of `escape`, `escape_once` or `h`, which have
  escaped it already;
- the text of a `capture` (its outputs were escaped when it was captured), also after an `assign`
  copies it, and in the partials of the render;
- either of these through `strip`, `lstrip`, `rstrip`, `strip_newlines`, `newline_to_br`,
  `truncate` or `truncatewords` (`{{ bio | escape | newline_to_br }}`), as long as the ellipsis
  of `truncate` and `truncatewords` is written in the template.

A capture is recognised by its text: an output equal to the text of a capture of the render is
written as it is, so data can only repeat markup the template itself captured.

Any other filter after them makes the result ordinary text again, escaped unless `raw` follows.
`assign`, conditions and filters see the data as it is: only what an output writes is escaped. The
escape is for HTML: a value written into CSS (`<style>`) or a script needs `raw`, and only for data
you trust. The option is off by default, as in Shopify Liquid, which has no `raw` filter.

### File System

```go
type MyFileSystem struct{}

func (fs *MyFileSystem) ReadTemplateFile(path string) (string, error) {
    // Load template from database, S3, etc.
    return loadTemplate(path)
}

tmpl := liquid.NewTemplate(&liquid.TemplateOptions{
    Environment: env,
})
tmpl.Registers()["file_system"] = &MyFileSystem{}
```

### Partial Templates

```liquid
{% render "header", title: page.title %}
{% include "sidebar" %}
```

## Standard Filters

Liquid Go includes all standard filters:

**String**: `capitalize`, `downcase`, `upcase`, `strip`, `lstrip`, `rstrip`, `strip_html`, `strip_newlines`, `newline_to_br`, `escape`, `escape_once`, `url_encode`, `url_decode`, `slice`, `truncate`, `truncatewords`, `split`, `replace`, `replace_first`, `remove`, `remove_first`, `append`, `prepend`

**Array**: `join`, `first`, `last`, `concat`, `map`, `reverse`, `sort`, `sort_natural`, `uniq`, `where`, `group_by`, `compact`, `size`

**Math**: `abs`, `ceil`, `floor`, `round`, `plus`, `minus`, `times`, `divided_by`, `modulo`, `at_least`, `at_most`

**Date**: `date`

**Default**: `default`

See [documentation](https://shopify.dev/docs/api/liquid/filters) for details.

## Standard Tags

All standard Liquid tags are supported:

- Control flow: `if`, `elsif`, `else`, `endif`, `unless`, `case`, `when`
- Loops: `for`, `break`, `continue`, `tablerow`
- Variables: `assign`, `capture`, `increment`, `decrement`
- Templates: `include`, `render`
- Other: `comment`, `raw`, `echo`, `liquid`

## Testing

Liquid Go maintains **superior test coverage** compared to the official Ruby library:

- **~90% code coverage** across the entire codebase
- **Comprehensive edge case testing** - Tests cover error paths, boundary conditions, and unusual scenarios
- **Integration tests** - Full template rendering scenarios matching Ruby test suite
- **Performance tests** - Benchmark suite ensuring optimal performance

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run integration tests
go test ./integration/...

# Run benchmarks
cd performance && go test -bench=.
```

### Shopify conformance suite

`integration/shopify_conformance_test.go` replays the test cases of the Shopify Liquid Ruby test
suite (extracted by `scripts/extract_shopify_conformance.py`) in each error mode, with Ruby
numbers and with JSON-decoded numbers. Every remaining difference is listed, with its reason, in
`integration/testdata/shopify_conformance/known_divergences.json`; see
[its README](integration/testdata/shopify_conformance/README.md).

## Project Structure

```
liquidgo/
├── liquid/              # Core library
│   ├── tags/           # Standard tag implementations
│   ├── tag/            # Tag base types
│   └── locales/        # i18n support
├── integration/        # Integration tests
├── performance/        # Benchmarking suite
└── reference-liquid/   # Ruby reference implementation
```

## Version Compatibility

Current version: **5.11.0**, port revision 9 (`liquid.Version`, `liquid.PortRevision`). Port
revision N is tagged `v5.11.(N-1)`: revision 9 is `v5.11.8`. Revision 2 has no tag of its own: it ships
in `v5.11.2` with revision 3.

Liquid Go maintains version parity with [Shopify Liquid](https://github.com/Shopify/liquid). This ensures compatibility with templates written for the Ruby version.

### Deliberate differences from Shopify Liquid

A few behaviours differ on purpose, because Go applications usually pass data decoded from JSON,
where every number is a `float64`, and render text meant for people (emails, pages):

- **An empty string is falsy**: `{% if "" %}` is false, so `{% if contact.first_name %}` skips an
  empty name. In Ruby, only `nil` and `false` are falsy. `where`, `reject`, `has`, `find` and
  `find_index` without a value also treat an empty list or hash as falsy, and with a value match a
  string to the number it writes (`where: "id", "10"` keeps an `id` of `10`).
- **Whole floats render without `.0`**: `{{ 2.0 }}` renders `2`, and `{{ 100 | plus: 1 }}` with a
  JSON number renders `101`, not `101.0`. A JSON integer stays a float in arithmetic, so
  `{{ 1999 | divided_by: 100 }}` with JSON data gives `19.99` (an integer literal gives `19`).
- **A filter error renders nothing**: `{{ 1 | divided_by: 0 }}`, an invalid `base64_decode`, `map`
  over integers and a filter given too many arguments render an empty string instead of
  `Liquid error: ...`; a `truncate` or `truncatewords` length that is not an integer keeps the
  input.
- **Template-controlled sizes are bounded**: a `date` directive wider than 1024 (`%5000d`) is
  copied as written, a `round` precision beyond ±400 digits is not computed, and a filter builds
  at most 1,048,576 integers of a range (`join` writes at most as many), raising `Memory limits
  exceeded` beyond, where Ruby would build them.
- **A Go struct renders as nothing**, unless it is a drop or has `ToLiquid()`, `String()` or
  `Error()`; Ruby raises an internal error for an object without `to_liquid`. See
  [What a Template Reads](#what-a-template-reads).
- **Each render starts from its own data**: what a render assigns (`assign`, `capture`) is not
  kept in the template for the next render, so a value assigned while rendering for one recipient
  never reaches the next. Ruby keeps it.
- **A render takes a map**, where Ruby also takes a drop as the root of the variables: give the drop
  under a key.
- **A float with no fraction is an array index**, as a JSON number is a `float64` in Go:
  `{{ arr[1.0] }}` and an index computed with `divided_by: 2.0` read an item, where Ruby takes only an
  Integer as an index.
- **`RaiseErrors` raises the errors of its own render only**; Ruby's `render!` leaves the template
  raising them in later renders.
- **A Go panic, a resource limit and a cancellation end the render** with the returned error and the
  output written so far; Ruby writes `Liquid error: internal` for an exception and goes on, and
  writes the message of a limit as the output.
- **`blank`** compares as before (`false == blank` is false).
- **`date`** accepts a float Unix timestamp, and reads a date without a zone as UTC.
- **`increment` and `decrement`** counters start at 0, never at a render data value of the same
  name.
- **Partials share the resource limits** of the template that renders them: `render` and `include`
  do not reset them, as Ruby does.
- **Nesting is bounded at 100** for lookups and `liquid` tags too, as for blocks: `a[a[a[...]]]`
  nested deeper is a parse error.
- **A filter result longer than `render_length_limit`** raises `Memory limits exceeded`, even when
  it is assigned rather than written; the filters that can multiply their input check before they
  build it.
- Some malformed markup that Ruby's lax parser tolerates (`{{ 'a' | split$$$:' ' }}`,
  `{{ false - }}`) renders differently.

## Documentation

- [Implementation Guide](IMPLEMENTATION.md) - Architecture and implementation details
- [Performance Guide](performance/README.md) - Benchmarking and optimization
- [Ruby Liquid Docs](https://shopify.dev/docs/api/liquid) - Template syntax reference
- [Liquid Wiki](https://github.com/Shopify/liquid/wiki) - Additional resources

## Contributing

Contributions are welcome! When implementing new features:

1. Reference the Ruby implementation in `reference-liquid/`
2. Maintain file naming conventions (see [IMPLEMENTATION.md](IMPLEMENTATION.md))
3. Add tests matching Ruby test coverage
4. Run benchmarks to check performance impact
5. Update documentation

## License

Liquid Go is released under the MIT License. See LICENSE file for details.

The Ruby reference implementation is © Shopify Inc., also under MIT License.

## Author

Liquid Go is created and maintained by **Pierre Bazoge** ([@pierre-b](https://github.com/pierre-b)).

## Credits

Liquid Go is a Go port of [Shopify's Liquid](https://github.com/Shopify/liquid) template engine, maintaining full compatibility with the Ruby implementation.

Original Liquid created by Tobias Lütke (@tobi).
