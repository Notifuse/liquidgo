package performance

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

// Test data for expression benchmarks
var (
	stringMarkups = []string{
		`"foo"`,
		`"fooooooooooo"`,
		`"foooooooooooooooooooooooooooooo"`,
		`'foo'`,
		`'fooooooooooo'`,
		`'foooooooooooooooooooooooooooooo'`,
	}

	variableMarkups = []string{
		"article",
		"article.title",
		"article.title.size",
		"very_long_variable_name_2024_11_05",
		"very_long_variable_name_2024_11_05.size",
	}

	numberMarkups = []string{
		"0",
		"35",
		"1241891024912849",
		"3.5",
		"3.51214128409128",
		"12381902839.123819283910283",
		"123.456.789",
		"-123",
		"-12.33",
		"-405.231",
		"-0",
		"0",
		"0.0",
		"0.0000000000000000000000",
		"0.00000000001",
	}

	rangeMarkups = []string{
		"(1..30)",
		"(1...30)",
		"(1..30..5)",
		"(1.0...30.0)",
		"(1.........30)",
		"(1..foo)",
		"(foo..30)",
		"(foo..bar)",
		"(foo...bar...100)",
		"(foo...bar...100.0)",
	}

	literalMarkups = []string{
		"",
		"nil",
		"null",
		"",
		"true",
		"false",
		"blank",
		"empty",
	}

	allExpressionMarkups []string

	// Lexer test expressions
	lexerExpressions = []string{
		"foo[1..2].baz",
		"12.0",
		"foo.bar.based",
		"21 - 62",
		"foo.bar.baz",
		"foo > 12",
		"foo < 12",
		"foo <= 12",
		"foo >= 12",
		"foo <> 12",
		"foo == 12",
		"foo != 12",
		"foo contains 12",
		"foo contains 'bar'",
		"foo != 'bar'",
		"'foo' contains 'bar'",
		"234089",
		"foo | default: -1",
	}
)

func init() {
	// Combine all expression markups for the "all" benchmark
	allExpressionMarkups = append(allExpressionMarkups, stringMarkups...)
	allExpressionMarkups = append(allExpressionMarkups, variableMarkups...)
	allExpressionMarkups = append(allExpressionMarkups, numberMarkups...)
	allExpressionMarkups = append(allExpressionMarkups, rangeMarkups...)
	allExpressionMarkups = append(allExpressionMarkups, literalMarkups...)
}

// BenchmarkExpressionParseString benchmarks Expression.Parse with string literals
func BenchmarkExpressionParseString(b *testing.B) {
	env := liquid.NewEnvironment()
	parseCtx := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, markup := range stringMarkups {
			parseCtx.ParseExpression(markup)
		}
	}
}

// BenchmarkExpressionParseLiteral benchmarks Expression.Parse with literal values
func BenchmarkExpressionParseLiteral(b *testing.B) {
	env := liquid.NewEnvironment()
	parseCtx := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, markup := range literalMarkups {
			parseCtx.ParseExpression(markup)
		}
	}
}

// BenchmarkExpressionParseVariable benchmarks Expression.Parse with variables
func BenchmarkExpressionParseVariable(b *testing.B) {
	env := liquid.NewEnvironment()
	parseCtx := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, markup := range variableMarkups {
			parseCtx.ParseExpression(markup)
		}
	}
}

// BenchmarkExpressionParseNumber benchmarks Expression.Parse with numbers
func BenchmarkExpressionParseNumber(b *testing.B) {
	env := liquid.NewEnvironment()
	parseCtx := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, markup := range numberMarkups {
			parseCtx.ParseExpression(markup)
		}
	}
}

// BenchmarkExpressionParseRange benchmarks Expression.Parse with ranges
func BenchmarkExpressionParseRange(b *testing.B) {
	env := liquid.NewEnvironment()
	parseCtx := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, markup := range rangeMarkups {
			parseCtx.ParseExpression(markup)
		}
	}
}

// BenchmarkExpressionParseAll benchmarks Expression.Parse with all markup types
func BenchmarkExpressionParseAll(b *testing.B) {
	env := liquid.NewEnvironment()
	parseCtx := liquid.NewParseContext(liquid.ParseContextOptions{Environment: env})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, markup := range allExpressionMarkups {
			parseCtx.ParseExpression(markup)
		}
	}
}

// BenchmarkLexerTokenize benchmarks Lexer tokenization
func BenchmarkLexerTokenize(b *testing.B) {
	lexer := &liquid.Lexer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, expr := range lexerExpressions {
			ss := liquid.NewStringScanner(expr)
			_, _ = lexer.Tokenize(ss)
		}
	}
}
