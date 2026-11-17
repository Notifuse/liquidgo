// Command libcmp compares Go Liquid libraries with Shopify Liquid 5.14.0.
//
//	go run . render LIB OUT.json    renders the corpus with LIB, writes each result
//	go run . report OUT.json...     compares the results with Shopify's outputs
//	go run . bench LIB              times the render of the benchmark template
//	go run . hostile LIB N          renders hostile template N (bound it: see hostile.sh)
//	go run . export DIR             writes the benchmark and hostile templates for ruby/
//
// It runs from tools/libcmp in the liquidgo repository, whose corpora hold the
// outputs of Shopify Liquid 5.14.0. See docs/library-comparison.md.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const root = "../.."

// result is what a library gave for a case.
type result struct {
	Output string `json:"output"`
	// Failure is "parse", "render", "panic" or "timeout", or empty.
	Failure string `json:"failure,omitempty"`
	Error   string `json:"error,omitempty"`
}

func main() {
	// Shopify's outputs were written in UTC
	time.Local = time.UTC
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: libcmp render LIB OUT | report OUT... | bench LIB")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "render":
		err = render(find(os.Args[2]), os.Args[3])
	case "report":
		err = report(os.Args[2:])
	case "bench":
		err = bench(find(os.Args[2]))
	case "hostile":
		err = hostile(find(os.Args[2]), os.Args[3])
	case "export":
		err = export(os.Args[2])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func find(name string) library {
	for _, l := range libraries {
		if l.name == name {
			return l
		}
	}
	fmt.Fprintf(os.Stderr, "unknown library %q\n", name)
	os.Exit(2)
	return library{}
}

// timeout is the longest a parse and a render of one case may take.
const timeout = 10 * time.Second

func render(lib library, out string) error {
	cases, err := loadCorpus(root)
	if err != nil {
		return err
	}
	results := make(map[string]result, len(cases))
	for _, c := range cases {
		results[c.ID] = renderCase(lib, c)
	}
	b, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(out, b, 0o644)
}

func renderCase(lib library, c Case) result {
	done := make(chan result, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- result{Failure: "panic", Error: fmt.Sprint(r)}
			}
		}()
		fn, err := lib.parse(c.Template)
		if err != nil {
			done <- result{Failure: "parse", Error: err.Error()}
			return
		}
		output, err := fn(c.Data())
		if err != nil {
			done <- result{Output: output, Failure: "render", Error: err.Error()}
			return
		}
		done <- result{Output: output}
	}()
	select {
	case r := <-done:
		return r
	case <-time.After(timeout):
		return result{Failure: "timeout"}
	}
}
