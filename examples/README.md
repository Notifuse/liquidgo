# Liquid Go Examples

This directory contains example programs demonstrating various Liquid Go features.

## Running Examples

Each example is a standalone Go program. To run an example:

```bash
cd examples/basic
go run main.go
```

## Examples

### Basic (`basic/`)

Simple hello world example showing basic template parsing and rendering.

```bash
cd basic && go run main.go
```

**Demonstrates:**
- Template parsing
- Variable substitution
- Basic rendering

### Filters (`filters/`)

Using built-in and custom filters to transform data.

```bash
cd filters && go run main.go
```

**Demonstrates:**
- Standard filters (upcase, size, etc.)
- Custom filter creation
- Filter chaining

### Loops (`loops/`)

Iterating over collections with control flow.

```bash
cd loops && go run main.go
```

**Demonstrates:**
- For loops
- Conditional rendering
- Array filters

## More Examples

For more advanced examples, see:

- **Integration tests**: [`../integration/`](../integration/) - Real-world usage patterns
- **Performance tests**: [`../performance/`](../performance/) - Complex template rendering
- **Unit tests**: [`../liquid/`](../liquid/) - Detailed feature demonstrations

## Creating Your Own

1. Create a new directory under `examples/`
2. Add a `main.go` file
3. Import `github.com/liquidgo/liquidgo/liquid`
4. Write your example code
5. Document what it demonstrates

Example template:

```go
package main

import (
    "fmt"
    "github.com/liquidgo/liquidgo/liquid"
)

func main() {
    // Your example code here
}
```

