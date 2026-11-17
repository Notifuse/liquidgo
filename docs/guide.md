# liquidgo guide

How to render templates with liquidgo, declare filters, drops and tags, and use its options. The
[README](../README.md) gives the overview; [the differences from Shopify Liquid](differences-from-shopify.md)
list what liquidgo does otherwise, and why.

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
    env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard()})
    if err != nil {
        panic(err)
    }

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
in Ruby, only a drop gives keys; Ruby raises an internal error for an object without `to_liquid`,
where liquidgo renders nothing. A pointer renders as what it points to, and a nil
pointer as nothing. A hash renders with its keys in sorted order. To give a struct to a template:

- implement `ToLiquid() interface{}` (with a value receiver) and return a map of what the template
  may read:

  ```go
  func (u User) ToLiquid() interface{} {
      return map[string]interface{}{"name": u.Name}
  }
  ```

- or make it a drop, as a `Liquid::Drop` in Ruby, by declaring the keys a template reads with
  `LiquidGet`, which gets the context of the render. It returns the value and whether the key
  exists: a missing key is nil, or an `UndefinedDropMethod` with strict variables. An error it
  returns is recorded and the render goes on; a Go error is recorded as a `*liquid.InternalError`,
  written `Liquid error: internal` as Ruby writes it, which unwraps to your error. A key is read when a template reads it, so a
  drop can load what it gives lazily; nothing else of the struct is readable, and a drop without
  a `String()` method renders as nothing (`nil` inside a list or a hash, where Ruby writes its class
  name). The items of a `[]User` are drops when `*User` is: a template reads the items of the slice
  themselves, so what a key loads stays for the next key. A nil `*User` is nil.

  ```go
  func (u *User) LiquidGet(ctx *liquid.Context, key string) (interface{}, bool, error) {
      switch key {
      case "first_name":
          return u.FirstName, true, nil
      case "orders":
          orders, err := u.loadOrders()
          return orders, true, err
      }
      return nil, false, nil
  }
  ```

`ToLiquid()` is applied at each step of a lookup, to the items the array filters read and to a
value written in the output. A type with a `String()` or `Error()` method renders that text: a
generated type whose `String()` writes all its fields (protobuf messages, for example) shows them
all, so give a template a map or a drop instead.

### Custom Filters

A filter is declared with its name and a function, which gets the context of the render, the
value it is applied to and its arguments (keyword arguments come as a last `map[string]interface{}`).
A filter of the environment takes the place of a standard one of the same name, and a local filter of
a render (`WithFilters`) takes the place of both; its name is written exactly, as in Ruby. An error
the function returns renders nothing and is recorded in `Result.Errors` (a Go error as a
`*liquid.InternalError` that unwraps to it); with strict filters it is raised, as Ruby raises it.

```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/Notifuse/liquidgo/v5/liquid"
)

func reverse(_ *liquid.Context, input interface{}, _ ...interface{}) (interface{}, error) {
    runes := []rune(fmt.Sprint(input))
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes), nil
}

func shout(_ *liquid.Context, input interface{}, _ ...interface{}) (interface{}, error) {
    return strings.ToUpper(fmt.Sprint(input)) + "!!!", nil
}

func main() {
    // Create an environment that declares the filters
    env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{
        Filters: []liquid.FilterDef{liquid.Filter("reverse", reverse), liquid.Filter("shout", shout)},
    })
    if err != nil {
        panic(err)
    }

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
    "github.com/Notifuse/liquidgo/v5/liquid/tags"
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
    // The standard tags and a custom one
    envTags := tags.Standard()
    envTags["greet"] = func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
        tag, err := NewGreetingTag(tagName, markup, parseContext)
        if err != nil {
            return nil, err
        }
        return tag, nil
    }
    env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: envTags})
    if err != nil {
        panic(err)
    }

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
    userEnv, err := liquid.NewEnvironment(liquid.EnvironmentConfig{
        Tags:    tags.Standard(),
        Filters: []liquid.FilterDef{liquid.Filter("money", money)},
    })
    if err != nil {
        panic(err)
    }

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
- **Safe to share** - An environment never changes once built: `NewEnvironment` keeps a copy of
  its config, so renders on any number of goroutines can share it

## Error Handling

The error mode of an environment decides how its templates parse:

```go
// liquid.ErrorModeLax (default): markup that does not parse is tolerated, as Ruby's lax parser does
// liquid.ErrorModeWarn: a syntax error is a warning, and the markup is parsed as lax
// liquid.ErrorModeStrict: a syntax error fails the parse
// liquid.ErrorModeStrict2 (or ErrorModeRigid): stricter still, as Shopify Liquid 5.14's strict2
env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: liquid.ErrorModeStrict})
```

### Rendering and errors

`Template.Render(ctx, data, opts...)` renders in a state of its own and changes nothing in the template or
in the data: a template is parsed once and rendered from any number of goroutines, and renders never see
each other's assigns, errors or limits. It reads the template's assigns and registers. The data is a map;
a drop goes under a key (`{"page": drop}`). A drop gets the context of the render at each key it gives and
keeps none, so renders can share a drop that is safe for concurrent use.

```go
res, err := tmpl.Render(ctx, data,
    liquid.WithFileSystem(partials),         // partials of this render, nested ones included
    liquid.WithLimits(limits),               // limits set here replace the template's
    liquid.WithRegisters(registers),         // read by tags and filters
    liquid.WithFilters(tenantFilters...),    // local filters, next to the environment's
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

1. **Parse once, render many**: a parsed template renders from any number of goroutines.
2. **Share an environment**: it keeps the tags, the filters and the parse of the partials.
3. **Profile with Go**: `pprof` shows where a render spends its time.

The [README](../README.md#performance) gives the figures, and [`performance/`](../performance/) the
benchmarks.

## Advanced Features

### Resource Limits

Without limits, a template can loop or build strings as long as it likes, as in Ruby. Give limits
to every render of untrusted templates, for the environment or for one render, and a deadline with
the context:

```go
written, score, assigned := 1<<20, 100000, 1<<20
env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{
    Tags: tags.Standard(),
    Limits: liquid.ResourceLimitsConfig{
        RenderLengthLimit: &written,  // bytes written
        RenderScoreLimit:  &score,    // nodes rendered, a loop counting its body and each range item
        AssignScoreLimit:  &assigned, // bytes assigned and captured
    },
})

length := 1 << 20
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
res, err := tmpl.Render(ctx, data, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderLengthLimit: &length}))
// err is a *liquid.MemoryError past a limit, a *liquid.CanceledError past the deadline
```

### Output Escape

For HTML templates that output untrusted data (contact fields, form input, feeds), turn on the
output escape, as LiquidJS's `outputEscape` option:

```go
env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: true})
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

A value stays inside the text of an element, an attribute in double or single quotes, a comment and
an Outlook conditional comment. The escape does not cover:

- an attribute without quotes, which a space in the value ends: quote every attribute;
- the scheme of a URL: `<a href="{{ url }}">` keeps a `javascript:` value; check URLs before they
  reach the template;
- CSS and scripts, as above.

### Listing the Variables of a Template

`liquid.ForParseTreeVisitor` walks a parsed template, its tags and their expressions, as Ruby's
`Liquid::ParseTreeVisitor`. A callback on `*liquid.VariableLookup` lists the variables a template
reads:

```go
package main

import (
    "fmt"

    "github.com/Notifuse/liquidgo/v5/liquid"
    "github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func main() {
    env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard()})
    if err != nil {
        panic(err)
    }
    tmpl, _ := liquid.ParseTemplate(
        "{% if contact.vip %}{{ contact.first_name | default: fallback }}{% endif %}",
        &liquid.TemplateOptions{Environment: env},
    )

    var names []interface{}
    liquid.ForParseTreeVisitor(tmpl.Root(), nil).
        AddCallbackFor([]interface{}{(*liquid.VariableLookup)(nil)}, func(node, ctx interface{}) (interface{}, interface{}) {
            names = append(names, node.(*liquid.VariableLookup).Name())
            return nil, ctx
        }).
        Visit(nil)

    fmt.Println(names) // Output: [contact contact fallback]
}
```

Each item of `Visit` pairs what the callback returned for a node with the visit of its children, so a
callback can also rebuild the tree. A tag of your own gives its children with `ParseTreeChildren()
[]interface{}` (`liquid.ParseTreeNode`); a block gives its body by default. As in Ruby, the keyword
arguments of a filter are not visited (`{{ a | default: b, allow_false: c }}` lists `a` and `b`), nor
is the group name of a `cycle`.

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
