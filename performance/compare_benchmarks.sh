#!/bin/bash
# Compare benchmark results using benchstat

if ! command -v benchstat &> /dev/null; then
    echo "benchstat not found. Installing..."
    go install golang.org/x/perf/cmd/benchstat@latest
fi

if [ $# -eq 0 ]; then
    echo "Usage: $0 <old_results> <new_results>"
    echo ""
    echo "Available result files:"
    ls -1t benchmark_results/*.txt baseline_results.txt 2>/dev/null | head -5
    exit 1
fi

OLD_FILE=$1
NEW_FILE=$2

if [ ! -f "$OLD_FILE" ]; then
    echo "Error: File not found: $OLD_FILE"
    exit 1
fi

if [ ! -f "$NEW_FILE" ]; then
    echo "Error: File not found: $NEW_FILE"
    exit 1
fi

echo "Comparing benchmarks:"
echo "  Old: $OLD_FILE"
echo "  New: $NEW_FILE"
echo ""

benchstat "$OLD_FILE" "$NEW_FILE"

