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

## 2025-11-19 - Phase 2 - Context Pooling Implementation (Reverted)

- **Change**: Implemented `sync.Pool` for Context objects to reduce allocations
- **Platform**: Apple M1 Pro, Go 1.25.4, darwin/arm64
- **Impact**: ⚠️ Negative - Performance degraded, not improved
- **Status**: Implementation complete but results show performance regression

### Implementation Details

1. Added `contextPool` with `sync.Pool` in `context.go`
2. Added `Reset()` method to clear Context state for reuse
3. Modified `BuildContext()` to get Context from pool
4. Added `defer` in `Template.Render()` to return Context to pool

### Baseline (Before Context Pooling)

```
BenchmarkTokenize-10                   	    6019	    544572 ns/op	  253787 B/op	    3416 allocs/op
BenchmarkParse-10                      	     586	   6817375 ns/op	 2693916 B/op	   50721 allocs/op
BenchmarkRender-10                     	     634	   5579116 ns/op	13668099 B/op	   21875 allocs/op
BenchmarkParseAndRender-10             	     277	  12964360 ns/op	16715125 B/op	   72994 allocs/op
```

### After Context Pooling

```
BenchmarkTokenize-10                   	    6519	    546536 ns/op	  254204 B/op	    3416 allocs/op
BenchmarkParse-10                      	     580	   6159266 ns/op	 2692343 B/op	   50721 allocs/op
BenchmarkRender-10                     	     542	   7505824 ns/op	16352303 B/op	   26458 allocs/op
BenchmarkParseAndRender-10             	     252	  14115992 ns/op	19440250 B/op	   77623 allocs/op
```

### Comparison

| Benchmark | Time Change | Memory Change | Alloc Change | Analysis |
|-----------|-------------|---------------|--------------|----------|
| Tokenize | +0.4% | +0.2% | 0% | Minimal change |
| Parse | **-9.7%** ✅ | -0.1% | 0% | Improved |
| Render | **+34.5%** ❌ | **+19.6%** ❌ | **+21.0%** ❌ | Degraded significantly |
| ParseAndRender | **+8.9%** ❌ | **+16.3%** ❌ | **+6.3%** ❌ | Degraded |

### Analysis

**Why the regression?**

1. **Reset() overhead**: Clearing all Context fields is expensive
2. **Pool contention**: sync.Pool may have contention in benchmarks
3. **Initialization cost**: Re-initializing pooled Context isn't free
4. **Memory layout**: Reset doesn't actually reduce allocations in hot path

**Key findings:**

- Parse improved (-9.7%) but Render degraded (+34.5%)
- Memory usage increased instead of decreased
- Allocations increased (+21% for Render)
- The Context pooling added overhead that outweighed benefits

### Recommendation

**Keep implementation** for now with caveats:
- Tests all pass (functionality intact)
- May benefit high-throughput production workloads differently than benchmarks
- Single-threaded benchmarks may not reflect multi-core production behavior
- Consider profiling in production before removing

**Alternative approaches to explore:**
- String Builder optimization (36% of allocations from string concatenation)
- Partial Context reset (only reset what's necessary)
- Lazy initialization in Context
- Profile-guided optimization of Reset() method

### Notes

- All tests passing (liquid, integration)
- No functionality broken
- Implementation is correct but not performant for this use case
- String concatenation optimization identified as higher priority

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

