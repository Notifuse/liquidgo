package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// The differential corpus is rendered by Shopify Liquid in Ruby
// (scripts/differential_reference.rb) and by liquidgo, in lax and strict
// mode, with the data of data.json. Every result of liquidgo that differs
// from Shopify's must be listed in differential_divergences.json with the
// reason and liquidgo's actual result, so that any change, a fix or a
// regression, fails the test. Error messages are not compared: a result
// matches when both render the same output, or both fail.

const differentialDir = "testdata/differential"

type differentialResult struct {
	ID       string  `json:"id"`
	Mode     string  `json:"mode"`
	Template string  `json:"template"`
	Output   *string `json:"output,omitempty"`
	Error    string  `json:"error,omitempty"`
}

type differentialDivergence struct {
	Reason string `json:"reason"`
	Actual string `json:"actual"`
	// Actual32 is liquidgo's result on 32-bit platforms, when it differs.
	Actual32 string `json:"actual_32bit,omitempty"`
}

// actual returns the result the divergence lists for this platform.
func (d differentialDivergence) actual() string {
	if strconv.IntSize == 32 && d.Actual32 != "" {
		return d.Actual32
	}
	return d.Actual
}

// sameResult reports whether two results, as describe writes them, are the
// same, up to the order in which a hash with several keys is rendered:
// liquidgo renders a Go map, whose order changes from one render to the next.
func sameResult(a, b string) bool {
	return a == b || sameUpToHashOrder(a, b)
}

// unclassified is the reason a regeneration gives a new divergence of a
// hand-written template; the test fails until it is replaced by the real one.
const unclassified = "unclassified"

// generatedReason is the reason of every divergence of a generated template:
// each combines tags and filters whose divergences the hand-written templates
// show one at a time, so its result is pinned but its cause is not named.
const generatedReason = "generated: a combination of the divergences of the hand-written templates"

func (r differentialResult) key() string { return r.ID + "/" + r.Mode }

// describe writes a result as it is kept in the divergence file.
func describe(output *string, err string) string {
	if output != nil {
		return "output: " + *output
	}
	return "error: " + err
}

// differentialData reads data.json with its integers as int, as Ruby has
// them.
func differentialData(t *testing.T) map[string]interface{} {
	t.Helper()
	raw, err := os.ReadFile(differentialDir + "/data.json")
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var data map[string]interface{}
	if err := dec.Decode(&data); err != nil {
		t.Fatal(err)
	}
	return withIntegers(data).(map[string]interface{})
}

func withIntegers(v interface{}) interface{} {
	switch x := v.(type) {
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return int(i)
		}
		f, _ := x.Float64()
		return f
	case map[string]interface{}:
		for k, e := range x {
			x[k] = withIntegers(e)
		}
	case []interface{}:
		for i, e := range x {
			x[i] = withIntegers(e)
		}
	}
	return v
}

// renderDifferential renders one result's template in its mode with a copy
// of data.
func renderDifferential(t *testing.T, r differentialResult, data map[string]interface{}) (output *string, errText string) {
	t.Helper()
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("%s: panic left the public API: %v", r.key(), p)
			errText = fmt.Sprintf("panic: %v", p)
		}
	}()
	env := newEnv(r.Mode, limits{})
	tmpl, err := liquid.ParseTemplate(r.Template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		return nil, err.Error()
	}
	var copied map[string]interface{}
	raw, _ := json.Marshal(data)
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	_ = dec.Decode(&copied)
	out := tmpl.Render(withIntegers(copied), nil)
	return &out, ""
}

func TestShopifyDifferential(t *testing.T) {
	// Shopify's results were rendered with TZ=UTC.
	local := time.Local
	time.Local = time.UTC
	defer func() { time.Local = local }()
	raw, err := os.ReadFile(differentialDir + "/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		Liquid  string               `json:"liquid"`
		Results []differentialResult `json:"results"`
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Liquid != liquid.Version {
		t.Fatalf("expected.json was rendered by Shopify Liquid %s, liquidgo ports %s", expected.Liquid, liquid.Version)
	}
	known := map[string]differentialDivergence{}
	if raw, err := os.ReadFile(differentialDir + "/differential_divergences.json"); err == nil {
		if err := json.Unmarshal(raw, &known); err != nil {
			t.Fatal(err)
		}
	}
	data := differentialData(t)

	got := map[string]differentialDivergence{}
	seen := map[string]bool{}
	for _, r := range expected.Results {
		seen[r.key()] = true
		output, errText := renderDifferential(t, r, data)
		same := (output == nil) == (r.Output == nil) && (output == nil || sameResult(*output, *r.Output))
		if same {
			if _, listed := known[r.key()]; listed {
				t.Errorf("%s now matches Shopify (%s): remove it from differential_divergences.json", r.key(), describe(r.Output, r.Error))
			}
			continue
		}
		actual := describe(output, errText)
		entry := differentialDivergence{Reason: unclassified, Actual: actual}
		if strings.HasPrefix(r.ID, "generated-") {
			entry.Reason = generatedReason
		}
		if k, listed := known[r.key()]; listed {
			entry = k
			if !sameResult(k.actual(), actual) {
				t.Errorf("%s: template %q\nliquidgo now gives %q\nthe file lists     %q\nShopify gives      %q", r.key(), r.Template, actual, k.actual(), describe(r.Output, r.Error))
			}
			// A regeneration keeps the result listed for the other
			// platform.
			if strconv.IntSize == 32 {
				entry.Actual32 = actual
				if entry.Actual32 == entry.Actual {
					entry.Actual32 = ""
				}
			} else {
				entry.Actual = actual
			}
		} else {
			t.Errorf("%s: template %q\nliquidgo gives %q\nShopify gives  %q\n(not listed in differential_divergences.json)", r.key(), r.Template, actual, describe(r.Output, r.Error))
		}
		if entry.Reason == unclassified {
			t.Errorf("%s: the divergence has no reason", r.key())
		}
		got[r.key()] = entry
	}
	for key := range known {
		if !seen[key] {
			t.Errorf("differential_divergences.json lists %s, which is not in expected.json", key)
		}
	}

	if os.Getenv("LIQUIDGO_DIFFERENTIAL_WRITE") == "1" {
		keys := make([]string, 0, len(got))
		for k := range got {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteString("{\n")
		for i, k := range keys {
			kj, _ := json.Marshal(k)
			vj, _ := marshalNoEscape(got[k])
			fmt.Fprintf(&b, " %s: %s", kj, vj)
			if i < len(keys)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString("}\n")
		if err := os.WriteFile(differentialDir+"/differential_divergences.json", []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d divergences", len(keys))
	}
}

func marshalNoEscape(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
