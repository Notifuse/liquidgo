package performance

import (
	"os"
	"testing"
)

var themeRunner *ThemeRunner

// Initialize theme runner once for all benchmarks
func TestMain(m *testing.M) {
	var err error
	themeRunner, err = NewThemeRunner()
	if err != nil {
		panic("Failed to initialize ThemeRunner: " + err.Error())
	}
	os.Exit(m.Run())
}

// BenchmarkTokenize benchmarks just the tokenization phase
func BenchmarkTokenize(b *testing.B) {
	phase := os.Getenv("PHASE")
	if phase != "" && phase != "all" && phase != "tokenize" {
		b.Skip("Skipping tokenize benchmark (PHASE=" + phase + ")")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := themeRunner.Tokenize(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParse benchmarks just the parse/compilation phase
func BenchmarkParse(b *testing.B) {
	phase := os.Getenv("PHASE")
	if phase != "" && phase != "all" && phase != "parse" {
		b.Skip("Skipping parse benchmark (PHASE=" + phase + ")")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := themeRunner.Compile(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRender benchmarks just the render phase (using pre-compiled templates)
func BenchmarkRender(b *testing.B) {
	phase := os.Getenv("PHASE")
	if phase != "" && phase != "all" && phase != "render" {
		b.Skip("Skipping render benchmark (PHASE=" + phase + ")")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := themeRunner.Render(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseAndRender benchmarks parse and render together
func BenchmarkParseAndRender(b *testing.B) {
	phase := os.Getenv("PHASE")
	if phase != "" && phase != "all" && phase != "run" {
		b.Skip("Skipping parse & render benchmark (PHASE=" + phase + ")")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := themeRunner.Run(); err != nil {
			b.Fatal(err)
		}
	}
}
