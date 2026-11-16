package main

// BenchmarkNewsletter times the render of the newsletter template with
// liquidgo, and BenchmarkNewsletterOthers with the engines closest to it in
// time, for profiling: go test -bench Newsletter -cpuprofile cpu.out.

import "testing"

func BenchmarkNewsletter(b *testing.B) {
	fn, err := find("liquidgo").parse(benchTemplate)
	if err != nil {
		b.Fatal(err)
	}
	data := benchData()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := fn(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewsletterOthers(b *testing.B) {
	for _, name := range []string{"goliquid", "nchapman", "acstech"} {
		b.Run(name, func(b *testing.B) {
			fn, err := find(name).parse(benchTemplate)
			if err != nil {
				b.Fatal(err)
			}
			data := benchData()
			b.ReportAllocs()
			for b.Loop() {
				_, _ = fn(data)
			}
		})
	}
}
