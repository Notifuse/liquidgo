# Liquid Go - Liquid Template Engine for Go

> **💌 Built by [Notifuse](https://www.notifuse.com/)** - The modern open-source emailing platform to send newsletters, transactional emails & write blogs.
> Notifuse uses Liquid templating to personalize emails and blog templates with variables like `{{ contact.first_name }}`. Self-hosted, free forever, and a modern alternative to Mailchimp, Resend etc...
> [Try the live demo →](https://www.notifuse.com/)

[![Tests](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/test.yml)
[![Benchmarks](https://github.com/Notifuse/liquidgo/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/Notifuse/liquidgo/actions/workflows/benchmarks.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Notifuse/liquidgo.svg)](https://pkg.go.dev/github.com/Notifuse/liquidgo)
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

- ✅ **Full feature parity** with Ruby Liquid 5.10.0
- ⚡ **High performance** - 3-10x faster than Ruby implementation
- 🔒 **Secure** - Safe for user-generated templates
- 📦 **Zero dependencies** - Pure Go implementation
- 🧪 **Well tested** - Comprehensive test suite matching Ruby tests
- 🎯 **Production ready** - Used in real-world applications

## Installation

```bash
go get github.com/Notifuse/liquidgo
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/Notifuse/liquidgo/liquid"
)

func main() {
    // Parse template
    tmpl, err := liquid.ParseTemplate("Hello {{ name }}!", nil)
    if err != nil {
        panic(err)
    }

    // Render with data
    output := tmpl.Render(map[string]interface{}{
        "name": "World",
    }, nil)

    fmt.Println(output) // Output: Hello World!
}
```

### Using Tags and Filters

```go
package main

import (
    "fmt"
    "github.com/Notifuse/liquidgo/liquid"
    "github.com/Notifuse/liquidgo/liquid/tags"
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
    output := tmpl.Render(map[string]interface{}{
        "user": map[string]interface{}{
            "name": "john doe",
            "items": []string{"apple", "banana", "cherry"},
        },
    }, nil)

    fmt.Println(output)
}
```

### Custom Filters

```go
package main

import (
    "fmt"
    "strings"
    "github.com/Notifuse/liquidgo/liquid"
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

    output := tmpl.Render(nil, nil)
    fmt.Println(output) // Output: OLLEH!!!
}
```

### Custom Tags

```go
package main

import (
    "fmt"
    "github.com/Notifuse/liquidgo/liquid"
)

// Custom tag implementation
type GreetingTag struct {
    *liquid.Tag
    name string
}

func NewGreetingTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*GreetingTag, error) {
    return &GreetingTag{
        Tag:  liquid.NewTag(tagName, markup, parseContext),
        name: markup,
    }, nil
}

func (g *GreetingTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
    ctx := context.Context().(*liquid.Context)
    name := ctx.FindVariable(g.name, false)
    *output += fmt.Sprintf("Greetings, %v!", name)
}

func main() {
    // Register custom tag
    env := liquid.NewEnvironment()
    env.RegisterTag("greet", func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
        return NewGreetingTag(tagName, markup, parseContext)
    })

    tmpl, _ := liquid.ParseTemplate(
        "{% greet user_name %}",
        &liquid.TemplateOptions{Environment: env},
    )

    output := tmpl.Render(map[string]interface{}{
        "user_name": "Alice",
    }, nil)

    fmt.Println(output) // Output: Greetings, Alice!
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
    "github.com/Notifuse/liquidgo/liquid"
    "github.com/Notifuse/liquidgo/liquid/tags"
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
3. **Enable profiling** - Use built-in profiler for optimization
4. **Cache templates** - Store compiled templates in memory

See [`performance/`](performance/) directory for detailed benchmarks.

## Advanced Features

### Template Profiling

```go
tmpl, _ := liquid.ParseTemplate(source, &liquid.TemplateOptions{
    Profile: true,
})

output := tmpl.Render(data, nil)

// Access profiling data
profiler := tmpl.Profiler()
fmt.Println(profiler.String())
```

### Resource Limits

```go
env := liquid.NewEnvironment()
env.SetDefaultResourceLimits(map[string]interface{}{
    "render_length_limit": 1000000,  // 1MB output limit
    "render_score_limit":  100000,   // Complexity limit
})
```

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

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./integration/...

# Run benchmarks
cd performance && go test -bench=.
```

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

Current version: **5.10.0**

Liquid Go maintains version parity with [Shopify Liquid](https://github.com/Shopify/liquid). This ensures compatibility with templates written for the Ruby version.

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

## Credits

Liquid Go is a Go port of [Shopify's Liquid](https://github.com/Shopify/liquid) template engine, maintaining full compatibility with the Ruby implementation.

Original Liquid created by Tobias Lütke (@tobi).
