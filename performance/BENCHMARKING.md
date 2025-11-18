# Benchmarking Guide

This document explains how to run, store, and compare performance benchmarks for Liquid Go.

## Quick Start

### Run and Save Current Benchmarks

```bash
cd performance
./run_benchmark.sh
```

This will:
- Run all benchmarks with 1 second duration
- Save results to `benchmark_results/bench_<timestamp>_<commit>.txt`
- Display results in the terminal

### Compare Results

```bash
./compare_benchmarks.sh baseline_results.txt benchmark_results/bench_<latest>.txt
```

This uses `benchstat` to show statistical comparison between runs.

## Baseline Results

The file `baseline_results.txt` contains the initial benchmark results from the performance suite implementation (November 18, 2025). Use this as a reference point for future comparisons.

### Initial Results (Apple M1 Pro)

```
BenchmarkTokenize-10                   	    1104	    541380 ns/op	  253453 B/op	    3416 allocs/op
BenchmarkParse-10                      	      87	   6584390 ns/op	 2595259 B/op	   60925 allocs/op
BenchmarkRender-10                     	      55	  10501414 ns/op	18423869 B/op	   34790 allocs/op
BenchmarkParseAndRender-10             	      33	  17325994 ns/op	21531089 B/op	   96146 allocs/op
BenchmarkExpressionParseString-10      	 3953755	       153.7 ns/op	      96 B/op	       6 allocs/op
BenchmarkExpressionParseLiteral-10     	 6831073	        87.77 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseVariable-10    	 5061795	       118.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseNumber-10      	 1773674	       333.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseRange-10       	 2713394	       221.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseAll-10         	  637768	      1225 ns/op	      96 B/op	       6 allocs/op
BenchmarkLexerTokenize-10              	   37323	     16065 ns/op	    8370 B/op	     237 allocs/op
```

## Manual Benchmark Commands

### Run All Benchmarks

```bash
go test -bench=. -benchmem
```

### Run Specific Benchmark

```bash
go test -bench=BenchmarkRender -benchmem
```

### Run with Longer Duration (More Accurate)

```bash
go test -bench=. -benchmem -benchtime=5s
```

### Run Specific Phase (Theme Benchmarks)

```bash
PHASE=tokenize go test -bench=.
PHASE=parse go test -bench=.
PHASE=render go test -bench=.
PHASE=run go test -bench=.
```

## Installing benchstat

The comparison script uses `benchstat` for statistical analysis:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Understanding Results

### Metrics

- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

### Example Output

```
name                           time/op
Tokenize-10                    541µs ± 0%
Parse-10                      6.58ms ± 0%
Render-10                     10.5ms ± 0%
ParseAndRender-10             17.3ms ± 0%
ExpressionParseString-10       154ns ± 0%
```

### benchstat Comparison Output

```
name                           old time/op    new time/op    delta
Tokenize-10                      541µs ± 0%     530µs ± 0%   -2.03%
Parse-10                        6.58ms ± 0%    6.45ms ± 0%   -1.97%
Render-10                       10.5ms ± 0%    10.2ms ± 0%   -2.86%
```

- **Positive delta**: Performance got worse (slower)
- **Negative delta**: Performance got better (faster)

## Best Practices

### Before Making Changes

1. Run and save baseline benchmarks:
   ```bash
   ./run_benchmark.sh
   ```

2. Note the commit hash and timestamp

### After Making Changes

1. Run benchmarks again:
   ```bash
   ./run_benchmark.sh
   ```

2. Compare with previous results:
   ```bash
   ./compare_benchmarks.sh benchmark_results/bench_OLD.txt benchmark_results/bench_NEW.txt
   ```

3. Analyze the differences:
   - Look for regressions (positive delta %)
   - Celebrate improvements (negative delta %)
   - Investigate significant memory changes

### Continuous Monitoring

Add benchmark results to your commit messages for significant changes:

```
git commit -m "Optimize template parsing

Performance impact (benchstat):
- Parse: -15.2% faster
- Memory: -8.3% less allocations
```

## Benchmark Result Files

### Directory Structure

```
performance/
├── baseline_results.txt           # Initial baseline
├── benchmark_results/             # Historical results
│   ├── bench_20251118_123456_abc123.txt
│   ├── bench_20251118_234567_def456.txt
│   └── ...
├── run_benchmark.sh              # Run and save benchmarks
└── compare_benchmarks.sh         # Compare two result files
```

### File Naming Convention

```
bench_<YYYYMMDD>_<HHMMSS>_<commit_hash>.txt
```

Example: `bench_20251118_143022_a1b2c3d.txt`

## Troubleshooting

### Benchmarks Take Too Long

Reduce benchmark time:
```bash
go test -bench=. -benchtime=100ms
```

### Inconsistent Results

1. Close other applications
2. Run multiple times and average
3. Increase benchmark duration: `-benchtime=10s`
4. Use `benchstat` which handles statistical variance

### benchstat Not Found

Install it:
```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Make sure `$GOPATH/bin` is in your `$PATH`.

## Performance Goals

### Target Metrics vs Ruby

Based on the Ruby Liquid implementation:

- **Parse**: 2-5x faster than Ruby
- **Render**: 3-10x faster than Ruby  
- **Memory**: 50-70% less than Ruby

### Acceptable Thresholds

- **Regression**: No more than 5% slower without justification
- **Memory**: No more than 10% increase in allocations
- **Optimization**: Target at least 10% improvement to be meaningful

## Resources

- [Go Benchmarking Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Ruby Liquid Benchmarks](../reference-liquid/performance/)

