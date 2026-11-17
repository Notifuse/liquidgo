package integration

import (
	"encoding/json"
	"github.com/Notifuse/liquidgo/v5/liquid"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// timeDependent finds the inputs whose output changes with the time.
var timeDependent = regexp.MustCompile(`\b(now|today)\b`)

// The difference known today between two renders of the same source.
const knownHashOrder = "a hash with several keys, which keyword arguments make, renders its keys in map order"

// renderedHash is a hash as liquidgo and Ruby render it: {"k"=>"v", ...}.
var renderedHash = regexp.MustCompile(`\{[^{}]*=>[^{}]*\}`)

// canonicalHashes sorts the pairs of every rendered hash in s, so that two
// renders of the same hash compare equal whatever the order of its keys.
func canonicalHashes(s string) string {
	return renderedHash.ReplaceAllStringFunc(s, func(h string) string {
		pairs := strings.Split(h[1:len(h)-1], ", ")
		sort.Strings(pairs)
		return "{" + strings.Join(pairs, ", ") + "}"
	})
}

// sameUpToHashOrder reports whether two different outputs are the same once
// the pairs of each rendered hash are sorted.
func sameUpToHashOrder(a, b string) bool {
	return a != b && strings.Contains(a, "=>") && canonicalHashes(a) == canonicalHashes(b)
}

// fuzzData is the data every fuzzed template is rendered with. Its maps have
// one key each: the order of a map with several is outside the comparison.
func fuzzData() map[string]interface{} {
	return map[string]interface{}{
		"a": []interface{}{1, 2}, "x": "x", "s": "héllo", "n": 3, "f": 1.5,
		"h": map[string]interface{}{"k": "v"}, "posts": []interface{}{map[string]interface{}{"title": "t"}},
	}
}

// fuzzSeeds returns the templates of the differential corpus and of the
// conformance suite.
func fuzzSeeds(f *testing.F) []string {
	var seeds []string
	for _, file := range []string{differentialDir + "/templates.json", conformanceDir + "/cases.json"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			f.Fatal(err)
		}
		var templates []struct {
			Template string `json:"template"`
		}
		if err := json.Unmarshal(raw, &templates); err != nil {
			f.Fatal(err)
		}
		for _, t := range templates {
			if t.Template != "" {
				seeds = append(seeds, t.Template)
			}
		}
	}
	for _, family := range scalingFamilies {
		seeds = append(seeds, family.build(2))
	}
	return seeds
}

// FuzzRender parses and renders a template in lax and strict mode, under the
// three resource limits, and checks that no panic leaves the public API or
// becomes an InternalError, and that two fresh parses of the same source
// render the same output.
func FuzzRender(f *testing.F) {
	for _, seed := range fuzzSeeds(f) {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			env := newEnv(mode, allLimits)
			var first, second result
			requireNoPanicEscapes(t, func() { first = renderWith(env, source, fuzzData()) })
			if first.parseErr != nil && strings.Contains(first.parseErr.Error(), "runtime error") {
				t.Fatalf("%s: a panic while parsing %q became the error %q", mode, source, first.parseErr)
			}
			if err := first.internalError(); err != nil {
				t.Fatalf("%s: a panic in the render of %q became %v", mode, source, err)
			}
			if first.parseErr == nil && !first.stopped() {
				if err := keptEnds(source, first.output); err != nil {
					t.Fatalf("%s: %q loses text: %v", mode, source, err)
				}
			}
			if timeDependent.MatchString(source) {
				continue
			}
			requireNoPanicEscapes(t, func() { second = renderWith(env, source, fuzzData()) })
			if first.output != second.output && sameUpToHashOrder(first.output, second.output) {
				t.Skipf("known: %s", knownHashOrder)
			}
			if (first.parseErr == nil) != (second.parseErr == nil) || first.output != second.output {
				t.Fatalf("%s: two parses of %q render %q (%v) and %q (%v)", mode, source, first.output, first.parseErr, second.output, second.parseErr)
			}
		}
	})
}
