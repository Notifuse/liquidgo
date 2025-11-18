# Liquid Go Performance Suite

This directory contains benchmarking and profiling tools for the Liquid Go template engine. The suite simulates real-world Shopify template rendering to provide realistic performance measurements.

## Overview

The performance suite includes:

- **Theme-based benchmarks**: Renders actual Shopify theme templates with realistic data
- **Unit benchmarks**: Tests individual components (lexer, expression parser)
- **CPU profiling**: Identifies performance bottlenecks
- **Memory profiling**: Tracks memory allocations and usage

## Structure

```
performance/
├── theme_runner.go          # Core infrastructure for loading and rendering templates
├── benchmark_test.go        # Main benchmark suite (tokenize, parse, render)
├── unit_test.go            # Unit benchmarks (expression, lexer)
├── profile.go              # CPU/memory profiling tool
├── memory_profile.go       # Detailed memory profiling tool
├── README.md               # This file
└── shopify/                # Shopify-specific implementations
    ├── liquid.go           # Tag and filter registration
    ├── database.go         # Test data loader
    ├── comment_form.go     # CommentForm tag
    ├── paginate.go         # Paginate tag
    ├── json_filter.go      # JSON filter
    ├── money_filter.go     # Money formatting filters
    ├── shop_filter.go      # Shop-related filters
    ├── tag_filter.go       # Tag management filters
    └── weight_filter.go    # Weight conversion filters
```

## Running Benchmarks

### Quick Start - Run and Save Results

The easiest way to run benchmarks and save results for comparison:

```bash
cd performance
./run_benchmark.sh
```

This automatically saves results with timestamp and commit hash to `benchmark_results/`.

### Manual Benchmark Commands

#### All Benchmarks

Run all benchmarks with default settings:

```bash
go test -bench=.
```

### Specific Benchmark Phases

Use the `PHASE` environment variable to run specific phases:

```bash
# Tokenization only
PHASE=tokenize go test -bench=.

# Parsing only
PHASE=parse go test -bench=.

# Rendering only (using pre-compiled templates)
PHASE=render go test -bench=.

# Parse and render together
PHASE=run go test -bench=.
```

### Benchmark Options

Control benchmark behavior with standard Go testing flags:

```bash
# Run for longer duration (more accurate results)
go test -bench=. -benchtime=10s

# Run with memory allocation stats
go test -bench=. -benchmem

# Run specific benchmarks
go test -bench=BenchmarkRender
go test -bench=BenchmarkExpression
```

## Profiling

### CPU Profiling

Generate and analyze CPU profiles:

```bash
# Generate CPU profile
go run profile.go -cpuprofile=cpu.prof -iterations=200

# Analyze with pprof (interactive)
go tool pprof cpu.prof

# Generate visual graph (requires graphviz)
go tool pprof -http=:8080 cpu.prof
```

Common pprof commands:

- `top` - Show top CPU consumers
- `list <function>` - Show annotated source for a function
- `web` - Generate visual call graph (requires graphviz)

### Memory Profiling

Generate and analyze memory profiles:

```bash
# Generate memory profile
go run profile.go -memprofile=mem.prof -iterations=200

# Analyze with pprof
go tool pprof mem.prof

# Quick memory stats
go run memory_profile.go
```

### Combined Profiling

Profile both CPU and memory in one run:

```bash
go run profile.go -cpuprofile=cpu.prof -memprofile=mem.prof -iterations=500
```

## Unit Benchmarks

The suite includes focused benchmarks for individual components:

### Expression Parser Benchmarks

Test the expression parser with different input types:

```bash
# All expression benchmarks
go test -bench=BenchmarkExpression

# Specific types
go test -bench=BenchmarkExpressionParseString
go test -bench=BenchmarkExpressionParseVariable
go test -bench=BenchmarkExpressionParseNumber
go test -bench=BenchmarkExpressionParseRange
```

### Lexer Benchmarks

Test the lexer/tokenizer:

```bash
go test -bench=BenchmarkLexer
```

## Test Data

The benchmarks use realistic e-commerce data from `reference-liquid/performance/`:

- **Templates**: Real Shopify theme templates from `tests/` directory

  - dropify/ - Minimalist theme
  - ripen/ - Standard theme
  - tribble/ - Complex theme with search
  - vogue/ - Fashion-focused theme

- **Database**: `shopify/vision.database.yml` contains:
  - Products with variants, pricing, inventory
  - Collections and categories
  - Blog posts and articles
  - Navigation links and menus
  - Shopping cart data

## Custom Tags and Filters

The suite includes Shopify-specific implementations:

### Tags

- `paginate` - Pagination for collections
- `form` (comment_form) - Comment form generation

### Filters

- `json` - JSON encoding
- `money`, `money_with_currency` - Price formatting
- `weight`, `weight_with_unit` - Weight conversion
- `asset_url`, `global_asset_url`, `shopify_asset_url` - Asset URL generation
- `product_img_url` - Product image URL with sizing
- `link_to`, `link_to_vendor`, `link_to_type` - Link generation
- `link_to_tag`, `link_to_add_tag`, `link_to_remove_tag` - Tag links
- `default_pagination` - Pagination HTML
- `pluralize` - Singular/plural word selection

## Interpreting Results

### Benchmark Output

Go benchmark results show:

- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

Example:

```
BenchmarkRender-8    100    12345678 ns/op    1234567 B/op    12345 allocs/op
```

### Comparing with Ruby

The Ruby implementation uses `benchmark/ips` (iterations per second), while Go uses nanoseconds per operation:

```
Ruby:  1000 i/s  = 1ms per iteration
Go:    1000000 ns/op = 1ms per iteration
```

To convert:

- `ns/op` to `i/s`: `1000000000 / ns_per_op`
- `i/s` to `ns/op`: `1000000000 / iterations_per_sec`

## Comparing Results Over Time

### Using benchstat (Recommended)

Install benchstat for statistical comparison:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Compare two benchmark runs:

```bash
./compare_benchmarks.sh baseline_results.txt benchmark_results/bench_LATEST.txt
```

Example output:

```
name                           old time/op    new time/op    delta
Tokenize-10                      541µs ± 0%     530µs ± 0%   -2.03%
Parse-10                        6.58ms ± 0%    6.45ms ± 0%   -1.97%
Render-10                       10.5ms ± 0%    10.2ms ± 0%   -2.86%
```

### Tracking Performance Over Time

1. **Baseline**: `baseline_results.txt` contains initial results
2. **History**: See `BENCHMARK_HISTORY.md` for notable performance changes
3. **Scripts**:
   - `run_benchmark.sh` - Run and save benchmarks
   - `compare_benchmarks.sh` - Compare two result files

### Example Workflow

```bash
# Before making changes
./run_benchmark.sh
# Note the filename: benchmark_results/bench_20251118_143022_abc123.txt

# Make your code changes
# ...

# After changes
./run_benchmark.sh

# Compare
./compare_benchmarks.sh \
  benchmark_results/bench_20251118_143022_abc123.txt \
  benchmark_results/bench_20251118_150433_def456.txt
```

See `BENCHMARKING.md` for detailed guide on tracking and analyzing performance.

## Troubleshooting

### Missing Templates

If you get errors about missing templates, ensure you have the reference repository:

```bash
# From liquidgo root
git clone https://github.com/Shopify/liquid reference-liquid
```

### Import Errors

The benchmarks require the liquid package to be importable. Run from the performance directory:

```bash
cd performance
go mod init github.com/pierre/liquidgo/performance  # if needed
go test -bench=.
```

### Memory Issues

For large profiling runs, you may need to increase available memory:

```bash
GOGC=50 go run profile.go -iterations=1000
```

## Contributing

When adding new benchmarks:

1. **Mirror Ruby tests**: Check `reference-liquid/performance/` for equivalent Ruby benchmarks
2. **Use realistic data**: Leverage the existing test templates and database
3. **Document clearly**: Add comments explaining what's being measured
4. **Verify results**: Compare with Ruby implementation where possible

## Performance Goals

Target performance characteristics (compared to Ruby implementation):

- **Parse**: Should be 2-5x faster than Ruby
- **Render**: Should be 3-10x faster than Ruby
- **Memory**: Should use 50-70% less memory than Ruby

Run comparative benchmarks regularly to track progress toward these goals.
