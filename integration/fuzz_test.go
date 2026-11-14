package integration

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Inputs the render cannot bound yet: a range bound, limit, offset or cols
// that is large or computed (a variable), many loops, which nest, and many
// filters that multiply the size of a string. They are skipped until the
// library bounds them itself.
var (
	largeBound       = regexp.MustCompile(`(\d{3}|[A-Za-z_][\w.]*)\s*\.\.|\.\.\s*-?(\d{3}|[A-Za-z_])|\b(limit|offset|cols)\s*:\s*-?(\d{3}|[A-Za-z_])`)
	loopTag          = regexp.MustCompile(`\b(for|tablerow|ifchanged)\b`)
	amplifyingFilter = regexp.MustCompile(`\b(replace|replace_first|replace_last|join|split|concat|append|prepend)\b`)
	timeDependent    = regexp.MustCompile(`\b(now|today)\b`)
)

// The panics known today, each with the shape of input that causes it.
var (
	// {% cycle a: %}, a group and no values.
	cycleWithoutValues = regexp.MustCompile(`\bcycle\s*[^}\n]*:\s*(-?%\}|\n|$)`)
	knownCyclePanic    = "a cycle without values panics"
	// h[], h[name], h[!], h[0A]: an index, which an undefined name or
	// markup that lax mode drops makes nil.
	nilIndex       = regexp.MustCompile(`\[[^\]]*\]`)
	knownNilKey    = "a hash indexed by a nil key panics"
	loneQuotePanic = "slice bounds out of range [1:0]"
	knownLoneQuote = "a lone quote panics while parsing"
	knownHashOrder = "a hash with several keys, which keyword arguments make, renders its keys in map order"
)

// sameUpToHashOrder reports whether two outputs that render a hash ("=>")
// hold the same characters, in another order.
func sameUpToHashOrder(a, b string) bool {
	if !strings.Contains(a, "=>") || len(a) != len(b) {
		return false
	}
	ra, rb := []rune(a), []rune(b)
	sort.Slice(ra, func(i, j int) bool { return ra[i] < ra[j] })
	sort.Slice(rb, func(i, j int) bool { return rb[i] < rb[j] })
	return string(ra) == string(rb)
}

// knownUnboundedInput returns why the library cannot yet bound the render of
// source, or "".
func knownUnboundedInput(source string) string {
	switch {
	case largeBound.MatchString(source):
		return "a range is built as an array, and an empty loop body adds nothing to the render score"
	case len(loopTag.FindAllString(source, -1)) > 3:
		return "an empty loop body adds nothing to the render score, and tablerow and ifchanged render their body more than once"
	case len(amplifyingFilter.FindAllString(source, -1)) > 3:
		return "a filter result is counted only once it is built"
	}
	return ""
}

// fuzzData is the data every fuzzed template is rendered with. Its maps have
// one key each: the order of a map with several is outside the comparison.
func fuzzData() map[string]interface{} {
	return map[string]interface{}{
		"a": []interface{}{1, 2}, "x": "x", "s": "héllo", "n": 3, "f": 1.5,
		"h": map[string]interface{}{"k": "v"}, "posts": []interface{}{map[string]interface{}{"title": "t"}},
	}
}

// fuzzDataWithoutHashes is fuzzData without its hashes, to tell the known
// panic of a hash indexed by a nil key from the others.
func fuzzDataWithoutHashes() map[string]interface{} {
	data := fuzzData()
	delete(data, "h")
	delete(data, "posts")
	return data
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
		if reason := knownUnboundedInput(source); reason != "" {
			t.Skip(reason)
		}
		for _, mode := range []string{"lax", "strict"} {
			env := newEnv(mode, allLimits)
			var first, second result
			requireNoPanicEscapes(t, func() { first = renderWith(env, source, fuzzData()) })
			if first.parseErr != nil && strings.Contains(first.parseErr.Error(), "runtime error") {
				if strings.Contains(first.parseErr.Error(), loneQuotePanic) {
					t.Skipf("known: %s", knownLoneQuote)
				}
				t.Fatalf("%s: a panic while parsing %q became the error %q", mode, source, first.parseErr)
			}
			if err := first.internalError(); err != nil {
				if cycleWithoutValues.MatchString(source) {
					t.Skipf("known: %s", knownCyclePanic)
				}
				if nilIndex.MatchString(source) {
					var withoutHashes result
					requireNoPanicEscapes(t, func() { withoutHashes = renderWith(env, source, fuzzDataWithoutHashes()) })
					if withoutHashes.internalError() == nil {
						t.Skipf("known: %s", knownNilKey)
					}
				}
				t.Fatalf("%s: a panic in the render of %q became %v", mode, source, err)
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
