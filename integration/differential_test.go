package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

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
}

// unclassified is the reason a regeneration gives a new divergence; the test
// fails until it is replaced by the real one.
const unclassified = "unclassified"

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
		same := (output == nil) == (r.Output == nil) && (output == nil || *output == *r.Output)
		if same {
			if _, listed := known[r.key()]; listed {
				t.Errorf("%s now matches Shopify (%s): remove it from differential_divergences.json", r.key(), describe(r.Output, r.Error))
			}
			continue
		}
		actual := describe(output, errText)
		reason := unclassified
		if k, listed := known[r.key()]; listed {
			reason = k.Reason
			if k.Actual != actual {
				t.Errorf("%s: template %q\nliquidgo now gives %q\nthe file lists     %q\nShopify gives      %q", r.key(), r.Template, actual, k.Actual, describe(r.Output, r.Error))
			}
		} else {
			t.Errorf("%s: template %q\nliquidgo gives %q\nShopify gives  %q\n(not listed in differential_divergences.json)", r.key(), r.Template, actual, describe(r.Output, r.Error))
		}
		if reason == unclassified {
			t.Errorf("%s: the divergence has no reason", r.key())
		}
		got[r.key()] = differentialDivergence{Reason: reason, Actual: actual}
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
