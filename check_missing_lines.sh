#!/bin/bash

for file in "liquid/tags/comment.go" "liquid/tags/doc.go" "liquid/tags/inline_comment.go" "liquid/usage.go" "liquid/tags/unless.go" "liquid/partial_cache.go" "liquid/string_scanner.go" "liquid/variable.go" "liquid/drop.go"; do
  echo "=== $file ==="
  go tool cover -func=coverage_liquid.out | grep "$file" | grep -v "100.0%" | head -10
  echo ""
done
