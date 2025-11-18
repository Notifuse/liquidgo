# Benchmark History

This file tracks significant benchmark results and performance improvements over time.

## Format

```
## [Date] - [Commit Hash] - [Description]
- **Change**: Brief description of what changed
- **Impact**: Performance impact summary
- **Details**: Benchmark comparison results
```

---

## 2025-11-18 - Initial Baseline - Performance Suite Implementation

- **Change**: Initial implementation of Go performance suite, ported from Ruby
- **Platform**: Apple M1 Pro, Go 1.25.4, darwin/arm64
- **Impact**: Established baseline for future comparisons

### Results

```
BenchmarkTokenize-10                   	    1981	    540173 ns/op	  252614 B/op	    3415 allocs/op
BenchmarkParse-10                      	     180	   6658812 ns/op	 2598179 B/op	   60926 allocs/op
BenchmarkRender-10                     	     100	  10432303 ns/op	20900473 B/op	   41492 allocs/op
BenchmarkParseAndRender-10             	      67	  18492732 ns/op	24091650 B/op	  102880 allocs/op
BenchmarkExpressionParseString-10      	 7848061	       152.1 ns/op	      96 B/op	       6 allocs/op
BenchmarkExpressionParseLiteral-10     	13739139	        87.49 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseVariable-10    	10297563	       116.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseNumber-10      	 3629349	       329.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseRange-10       	 5510266	       221.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpressionParseAll-10         	 1277396	       944.3 ns/op	      96 B/op	       6 allocs/op
BenchmarkLexerTokenize-10              	   75652	     16027 ns/op	    8370 B/op	     237 allocs/op
```

### Notes

- Used real Shopify theme templates (dropify, ripen, tribble, vogue)
- Realistic e-commerce data from vision.database.yml (~946 lines)
- Fixed nil pointer dereference in liquid engine's `IsInvokable()` function
- All benchmarks passing successfully

### Key Metrics

- **Tokenize**: ~540 µs per operation (parsing template syntax)
- **Parse**: ~6.7 ms per operation (building AST)
- **Render**: ~10.4 ms per operation (executing pre-compiled templates)
- **Full Cycle**: ~18.5 ms per operation (parse + render)

### Memory Usage

- **Parse**: ~2.6 MB allocated, ~61k allocations
- **Render**: ~21 MB allocated, ~41k allocations
- **Full Cycle**: ~24 MB allocated, ~103k allocations

---

## How to Add New Entries

When you make a performance-impacting change:

1. Run benchmarks before and after:
   ```bash
   ./run_benchmark.sh  # Before
   # Make your changes
   ./run_benchmark.sh  # After
   ```

2. Compare results:
   ```bash
   ./compare_benchmarks.sh benchmark_results/bench_OLD.txt benchmark_results/bench_NEW.txt
   ```

3. Add entry to this file with:
   - Date and commit hash
   - Description of changes
   - benchstat comparison output
   - Any relevant notes

Example entry:

```markdown
## 2025-11-19 - abc1234 - Optimize Expression Parser

- **Change**: Cached compiled regex patterns in expression parser
- **Impact**: 15% faster expression parsing, no memory increase
- **Platform**: Apple M1 Pro, Go 1.25.4

### Comparison

\`\`\`
name                           old time/op    new time/op    delta
ExpressionParseString-10         152ns ± 0%     129ns ± 0%  -15.13%
ExpressionParseVariable-10       117ns ± 0%      99ns ± 0%  -15.38%

name                           old alloc/op   new alloc/op   delta
ExpressionParseString-10          96.0B ± 0%     96.0B ± 0%     ~
\`\`\`

### Notes
- Regex compilation was happening on every parse
- Moved to package-level variables
- No breaking changes
```

