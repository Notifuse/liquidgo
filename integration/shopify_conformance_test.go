package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// The conformance suite replays cases taken from the Shopify Liquid Ruby test suite
// (see testdata/shopify_conformance/README.md). Every case runs in each error mode it
// declares and with two typings of its numbers: as Ruby has them ("ruby": integers stay
// integers) and as encoding/json decodes them ("json": every number is a float64), which
// is how applications such as Notifuse hand data to the engine.
//
// A variant that does not match Shopify must be listed in known_divergences.json with
// liquidgo's actual result, so that any change of behaviour, fix or regression, fails.

const conformanceDir = "testdata/shopify_conformance"

type conformanceCase struct {
	ID           string                 `json:"id"`
	Source       string                 `json:"source"`
	Test         string                 `json:"test"`
	Kind         string                 `json:"kind"`
	Template     string                 `json:"template"`
	Assigns      map[string]interface{} `json:"assigns"`
	Partials     map[string]string      `json:"partials"`
	RenderErrors bool                   `json:"render_errors"`
	Modes        []string               `json:"modes"`
	Raises       string                 `json:"raises"`
	Expected     interface{}            `json:"expected"`
}

type knownDivergence struct {
	Reason string            `json:"reason"`
	Skip   bool              `json:"skip,omitempty"`
	Actual map[string]string `json:"actual,omitempty"`
}

// conformanceModes returns the error modes a case runs in: the modes it declares, or
// lax, strict and strict2, the three modes Ruby runs its whole suite in.
func conformanceModes(c conformanceCase) []string {
	if c.Modes == nil {
		return []string{"lax", "strict", "strict2"}
	}
	var modes []string
	for _, m := range c.Modes {
		if m == "lax" || m == "strict" || m == "strict2" || m == "warn" {
			modes = append(modes, m)
		}
	}
	return modes
}

// conformanceValue converts a value from cases.json. Floats are written {"__f": x};
// integers become int ("ruby") or float64 ("json").
func conformanceValue(v interface{}, typing string) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		if f, ok := x["__f"]; ok && len(x) == 1 {
			return f.(float64)
		}
		out := make(map[string]interface{}, len(x))
		for k, e := range x {
			out[k] = conformanceValue(e, typing)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, e := range x {
			out[i] = conformanceValue(e, typing)
		}
		return out
	case float64:
		if typing == "json" {
			return x
		}
		return int(x)
	}
	return v
}

func hasInteger(v interface{}) bool {
	switch x := v.(type) {
	case map[string]interface{}:
		if _, ok := x["__f"]; ok && len(x) == 1 {
			return false
		}
		for _, e := range x {
			if hasInteger(e) {
				return true
			}
		}
	case []interface{}:
		for _, e := range x {
			if hasInteger(e) {
				return true
			}
		}
	case float64:
		return true
	}
	return false
}

// runConformance renders one variant. For a "filter" case it returns the value assigned
// to __result instead of the output.
func runConformance(c conformanceCase, mode, typing string) (out string, value interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			// a Liquid error is what a raises case expects; any other panic is a bug
			if le, ok := r.(liquid.LiquidError); ok {
				err = le
			} else {
				err = goPanic{r}
			}
		}
	}()
	fs := NewStubFileSystem(c.Partials)
	env := liquid.NewEnvironment()
	env.SetFileSystem(fs)
	env.SetErrorMode(mode)
	tags.RegisterStandardTags(env)

	tmpl, perr := liquid.ParseTemplate(c.Template, &liquid.TemplateOptions{Environment: env, LineNumbers: true})
	if perr != nil {
		return "", nil, fmt.Errorf("parse: %v", perr)
	}
	assigns, _ := conformanceValue(c.Assigns, typing).(map[string]interface{})
	ctx := liquid.BuildContext(liquid.ContextConfig{
		Environment:        env,
		StaticEnvironments: []map[string]interface{}{assigns},
		Registers:          liquid.NewRegisters(map[string]interface{}{"file_system": fs}),
		RethrowErrors:      !c.RenderErrors,
	})
	out = tmpl.Render(ctx, &liquid.RenderOptions{})
	if c.Kind == "filter" {
		value = ctx.FindVariable("__result", false)
	}
	return out, value, nil
}

// goPanic is a panic that is not a Liquid error, such as an index out of range.
type goPanic struct{ value interface{} }

func (p goPanic) Error() string { return fmt.Sprintf("panic: %v", p.value) }

// rubyEqual compares a liquidgo value with a Ruby literal the way Ruby's == would:
// numbers by value whatever their type, arrays and hashes element by element.
func rubyEqual(want, got interface{}) bool {
	if f, ok := toFloat(want); ok {
		g, ok := toFloat(got)
		return ok && (f == g || (math.IsNaN(f) && math.IsNaN(g)))
	}
	switch w := want.(type) {
	case nil:
		return got == nil
	case []interface{}:
		rv := reflect.ValueOf(got)
		if !rv.IsValid() || (rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array) || rv.Len() != len(w) {
			return false
		}
		for i := range w {
			if !rubyEqual(w[i], rv.Index(i).Interface()) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		g, ok := got.(map[string]interface{})
		if !ok || len(g) != len(w) {
			return false
		}
		for k, v := range w {
			gv, ok := g[k]
			if !ok || !rubyEqual(v, gv) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(want, got)
}

func toFloat(v interface{}) (float64, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 0, false
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	}
	return 0, false
}

// conformanceOutcome runs a variant and reports whether it matches Shopify, with a
// one-line description of liquidgo's result to pin in known_divergences.json.
func conformanceOutcome(c conformanceCase, mode, typing string) (ok bool, actual string) {
	out, value, err := runConformance(c, mode, typing)
	if c.Raises != "" {
		if _, crashed := err.(goPanic); crashed {
			return false, "error: " + err.Error()
		}
		if err != nil {
			return true, "error"
		}
		return false, "output: " + out
	}
	if err != nil {
		return false, "error: " + err.Error()
	}
	if c.Kind == "filter" {
		want := conformanceValue(c.Expected, "ruby")
		if rubyEqual(want, value) {
			return true, ""
		}
		return false, fmt.Sprintf("value: %#v", value)
	}
	want, _ := c.Expected.(string)
	if out == want {
		return true, ""
	}
	return false, "output: " + out
}

func loadConformance(t *testing.T) ([]conformanceCase, map[string]knownDivergence) {
	t.Helper()
	raw, err := os.ReadFile(conformanceDir + "/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []conformanceCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(conformanceDir + "/known_divergences.json")
	if err != nil {
		t.Fatal(err)
	}
	var known map[string]knownDivergence
	if err := json.Unmarshal(raw, &known); err != nil {
		t.Fatal(err)
	}
	return cases, known
}

func conformanceVariants(c conformanceCase) []string {
	var variants []string
	typings := []string{"ruby"}
	if hasInteger(c.Assigns) {
		typings = append(typings, "json")
	}
	for _, mode := range conformanceModes(c) {
		for _, typing := range typings {
			variants = append(variants, mode+"/"+typing)
		}
	}
	return variants
}

func containsString(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

func TestShopifyConformance(t *testing.T) {
	cases, known := loadConformance(t)
	ids := map[string]bool{}
	for _, c := range cases {
		ids[c.ID] = true
		pinned := known[c.ID]
		if pinned.Skip {
			continue
		}
		variants := conformanceVariants(c)
		if len(variants) == 0 {
			t.Errorf("%s runs in no variant: check its modes", c.ID)
		}
		for variant := range pinned.Actual {
			if !containsString(variants, variant) {
				t.Errorf("%s: pinned variant %s is not run", c.ID, variant)
			}
		}
		for _, variant := range variants {
			parts := strings.SplitN(variant, "/", 2)
			ok, actual := conformanceOutcome(c, parts[0], parts[1])
			want, isPinned := pinned.Actual[variant]
			switch {
			case ok && isPinned:
				t.Errorf("%s [%s] now matches Shopify: remove it from known_divergences.json\n  template: %q", c.ID, variant, c.Template)
			case !ok && !isPinned:
				t.Errorf("%s [%s] differs from Shopify\n  template: %q\n  assigns:  %v\n  want:     %#v\n  got:      %s", c.ID, variant, c.Template, c.Assigns, c.Expected, actual)
			case !ok && actual != want:
				t.Errorf("%s [%s] changed\n  template: %q\n  pinned:   %s\n  got:      %s", c.ID, variant, c.Template, want, actual)
			}
		}
	}
	var stale []string
	for id, d := range known {
		if !ids[id] {
			stale = append(stale, id)
		}
		if d.Reason == "" {
			t.Errorf("%s: known divergence without a reason", id)
		}
	}
	sort.Strings(stale)
	for _, id := range stale {
		t.Errorf("%s: in known_divergences.json but not in cases.json", id)
	}
}

// TestShopifyConformanceRegenerate writes known_divergences.json from the current
// behaviour, keeping existing reasons. Run it only by hand, after a fix:
//
//	LIQUIDGO_CONFORMANCE_WRITE=1 go test ./integration -run TestShopifyConformanceRegenerate
//
// then review the diff: every new entry needs a reason.
func TestShopifyConformanceRegenerate(t *testing.T) {
	if os.Getenv("LIQUIDGO_CONFORMANCE_WRITE") == "" {
		t.Skip("set LIQUIDGO_CONFORMANCE_WRITE=1 to rewrite known_divergences.json")
	}
	cases, known := loadConformance(t)
	next := map[string]knownDivergence{}
	for _, c := range cases {
		if known[c.ID].Skip {
			next[c.ID] = known[c.ID]
			continue
		}
		for _, variant := range conformanceVariants(c) {
			parts := strings.SplitN(variant, "/", 2)
			if ok, actual := conformanceOutcome(c, parts[0], parts[1]); !ok {
				d, seen := next[c.ID]
				if !seen {
					d = knownDivergence{Reason: known[c.ID].Reason, Actual: map[string]string{}}
				}
				d.Actual[variant] = actual
				next[c.ID] = d
			}
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(next); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(conformanceDir+"/known_divergences.json", buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
