#!/bin/bash
# Run benchmarks and save results with timestamp

set -e

RESULTS_DIR="benchmark_results"
mkdir -p "$RESULTS_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
RESULTS_FILE="$RESULTS_DIR/bench_${TIMESTAMP}_${COMMIT_HASH}.txt"

echo "Running benchmarks..."
echo "Timestamp: $(date)" | tee "$RESULTS_FILE"
echo "Commit: $COMMIT_HASH" | tee -a "$RESULTS_FILE"
echo "Go Version: $(go version)" | tee -a "$RESULTS_FILE"
echo "CPU: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || uname -m)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Run benchmarks
go test -bench=. -benchmem -benchtime=1s | tee -a "$RESULTS_FILE"

echo ""
echo "Results saved to: $RESULTS_FILE"
echo ""
echo "To compare with baseline:"
echo "  benchstat baseline_results.txt $RESULTS_FILE"

