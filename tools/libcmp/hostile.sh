#!/bin/sh
# Renders each hostile template with each library in a Docker container
# bounded to 1 GiB of memory and 20 seconds, and prints the library, the
# template, the exit code (124: out of time, 137: out of memory) and the
# outcome.
set -e
cd "$(dirname "$0")"
out=$(mktemp -d)
CGO_ENABLED=0 GOOS=linux go build -o "$out/libcmp" .
for lib in liquidgo osteele acstech tobi goliquid nchapman; do
  for i in 0 1 2 3; do
    set +e
    docker run --rm --memory=1g --memory-swap=1g -v "$out:/b:ro" alpine:3.20 timeout 20 /b/libcmp hostile "$lib" "$i" >"$out/log" 2>&1
    code=$?
    set -e
    printf '%s\t%s\t%s\t%s\n' "$lib" "$i" "$code" "$(tail -1 "$out/log" | cut -c1-100)"
  done
done
