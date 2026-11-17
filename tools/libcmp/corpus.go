package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Case is a template with its data and the output of Shopify Liquid 5.14.0
// in lax mode.
type Case struct {
	ID       string
	Template string
	data     []byte // JSON, decoded anew for each render
	Expected string
}

// Data decodes the data of the case as Ruby holds it: an integer is an int,
// a float a float64.
func (c Case) Data() map[string]any {
	if len(c.data) == 0 {
		return map[string]any{}
	}
	dec := json.NewDecoder(bytes.NewReader(c.data))
	dec.UseNumber()
	var v map[string]any
	if err := dec.Decode(&v); err != nil {
		panic(err)
	}
	return numbers(v).(map[string]any)
}

func numbers(v any) any {
	switch x := v.(type) {
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return int(i)
		}
		f, _ := x.Float64()
		return f
	case map[string]any:
		// Shopify's cases write a float {"__f": x}, so that 100.0 stays a float
		if f, ok := x["__f"].(json.Number); ok && len(x) == 1 {
			v, _ := f.Float64()
			return v
		}
		for k, e := range x {
			x[k] = numbers(e)
		}
	case []any:
		for i, e := range x {
			x[i] = numbers(e)
		}
	}
	return v
}

// loadCorpus reads the two corpora of liquidgo whose outputs come from
// Shopify Liquid 5.14.0: the template cases of Shopify's test suite that
// render in lax mode without partials and without raising, and the lax
// results of the differential corpus.
func loadCorpus(root string) ([]Case, error) {
	var cases []Case
	raw, err := os.ReadFile(filepath.Join(root, "integration/testdata/shopify_conformance/cases.json"))
	if err != nil {
		return nil, err
	}
	var shopify []struct {
		ID, Kind, Template, Raises string
		Expected                   any
		Modes                      []string
		Assigns                    json.RawMessage
		Partials                   map[string]string
	}
	if err := json.Unmarshal(raw, &shopify); err != nil {
		return nil, err
	}
	for _, c := range shopify {
		expected, ok := c.Expected.(string)
		if c.Kind != "template" || !ok || c.Raises != "" || len(c.Partials) > 0 || !lax(c.Modes) {
			continue
		}
		cases = append(cases, Case{ID: "shopify:" + c.ID, Template: c.Template, data: c.Assigns, Expected: expected})
	}

	dir := filepath.Join(root, "integration/testdata/differential")
	data, err := os.ReadFile(filepath.Join(dir, "data.json"))
	if err != nil {
		return nil, err
	}
	raw, err = os.ReadFile(filepath.Join(dir, "expected.json"))
	if err != nil {
		return nil, err
	}
	var expected struct {
		Results []struct {
			ID, Mode, Template string
			Output             *string
			StrictVariables    bool `json:"strict_variables"`
		}
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		return nil, err
	}
	for _, r := range expected.Results {
		if r.Mode != "lax" || r.Output == nil || r.StrictVariables {
			continue
		}
		cases = append(cases, Case{ID: "differential:" + r.ID, Template: r.Template, data: data, Expected: *r.Output})
	}
	sort.SliceStable(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	return comparable(cases), nil
}

// comparable leaves out the cases no Go library can match: Ruby writes a hash
// with its keys in the order they were added, which a Go map does not keep.
func comparable(cases []Case) []Case {
	kept := cases[:0]
	for _, c := range cases {
		if strings.Count(c.Expected, "=>") > 1 {
			continue
		}
		kept = append(kept, c)
	}
	return kept
}

func lax(modes []string) bool {
	if modes == nil {
		return true
	}
	for _, m := range modes {
		if m == "lax" {
			return true
		}
	}
	return false
}
