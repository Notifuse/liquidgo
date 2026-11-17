#!/bin/sh
# Times the benchmark template with each Go library and with Shopify Liquid
# (Ruby 3.3 with YJIT), each in a Docker container on the same machine, and
# renders the hostile templates with Shopify Liquid and its resource limits,
# bounded as hostile.sh bounds the Go libraries.
# ./ruby.sh SHOPIFY_LIQUID_5_14_CHECKOUT
set -e
cd "$(dirname "$0")"
liquid=$(cd "$1" && pwd)
out=$(mktemp -d)
CGO_ENABLED=0 GOOS=linux go build -o "$out/libcmp" .
go run . export "$out"
cp ruby/*.rb "$out/"
for lib in liquidgo tobi nchapman goliquid osteele acstech; do
  docker run --rm -v "$out:/b:ro" alpine:3.20 /b/libcmp bench "$lib"
done
ruby="docker run --rm -e RUBY_YJIT_ENABLE=1 -v $out:/b:ro -v $liquid:/liquid:ro ruby:3.3"
$ruby sh -c "gem install --no-document strscan:'>= 3.1.1' bigdecimal >/dev/null 2>&1 && ruby -I /liquid/lib /b/bench.rb /b/templates.json"
for i in 0 1 2 3; do
  set +e
  docker run --rm --memory=1g --memory-swap=1g -v "$out:/b:ro" -v "$liquid:/liquid:ro" ruby:3.3 \
    sh -c "gem install --no-document strscan:'>= 3.1.1' bigdecimal >/dev/null 2>&1 && timeout 20 ruby -I /liquid/lib /b/hostile.rb /b/templates.json $i" >"$out/log" 2>&1
  code=$?
  set -e
  printf 'ruby\t%s\t%s\t%s\n' "$i" "$code" "$(tail -1 "$out/log" | cut -c1-100)"
done
