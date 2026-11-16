package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// report prints, for each file of results, how many cases match Shopify's
// output and how many fail to render.
func report(files []string) error {
	cases, err := loadCorpus(root)
	if err != nil {
		return err
	}
	fmt.Printf("%d cases (%d from Shopify's tests, %d from the differential corpus)\n\n", len(cases), count(cases, "shopify:"), count(cases, "differential:"))
	fmt.Println("| Library | Matches Shopify | Shopify's tests | Differential | Failures (parse, render, panic, timeout) |")
	fmt.Println("|---|---|---|---|---|")
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		var results map[string]result
		if err := json.Unmarshal(raw, &results); err != nil {
			return err
		}
		match, shop, diff := 0, 0, 0
		failures := map[string]int{}
		for _, c := range cases {
			r := results[c.ID]
			if r.Failure != "" {
				failures[r.Failure]++
			}
			if r.Failure == "" && r.Output == c.Expected {
				match++
				if strings.HasPrefix(c.ID, "shopify:") {
					shop++
				} else {
					diff++
				}
			}
		}
		total := failures["parse"] + failures["render"] + failures["panic"] + failures["timeout"]
		fmt.Printf("| %s | %d (%.1f%%) | %d / %d | %d / %d | %d (%d, %d, %d, %d) |\n",
			strings.TrimSuffix(filepath.Base(f), ".json"), match, 100*float64(match)/float64(len(cases)),
			shop, count(cases, "shopify:"), diff, count(cases, "differential:"),
			total, failures["parse"], failures["render"], failures["panic"], failures["timeout"])
	}
	return nil
}

func count(cases []Case, prefix string) int {
	n := 0
	for _, c := range cases {
		if strings.HasPrefix(c.ID, prefix) {
			n++
		}
	}
	return n
}
