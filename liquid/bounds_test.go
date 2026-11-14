package liquid

import (
	"math"
	"math/big"
	"strings"
	"testing"
	"time"
)

// A single filter argument must not make a render take unbounded time or memory.
func TestFilterArgumentsAreBounded(t *testing.T) {
	sf := &StandardFilters{}
	when := time.Date(2024, 3, 5, 14, 30, 7, 0, time.UTC)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, format := range []string{"%300000d", "%99999999999999999999d", "%500000000A", "%99999999999999999999A", "%99999999N"} {
			if got := strftime(&when, format); len(got) > 2000 {
				t.Errorf("strftime(%q) produced %d bytes", format, len(got))
			}
		}
		for _, digits := range []interface{}{30000000, 200000000, -200000000, int64(math.MaxInt64)} {
			_, _ = sf.Round(1.5, digits)
			_, _ = sf.Round(5, digits)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("a huge width or round precision did not return quickly")
	}
	if got, _ := sf.Round(1.5, 400); got != 1.5 {
		t.Errorf("Round(1.5, 400) = %v, want 1.5", got)
	}
	if got, _ := sf.Round(1.5, -400); got != 0.0 {
		t.Errorf("Round(1.5, -400) = %v, want 0", got)
	}
	if got, _ := sf.Round(5, -400); got != 0 {
		t.Errorf("Round(5, -400) = %v, want 0", got)
	}
}

func TestStrftimeHugeWidthIsCopied(t *testing.T) {
	when := time.Date(2024, 3, 5, 14, 30, 7, 0, time.UTC)
	if got := strftime(&when, "%5000d|%d"); got != "%5000d|05" {
		t.Errorf("got %q, want the oversized directive copied as written", got)
	}
	if got := strftime(&when, "%1024d"); len(got) != 1024 || !strings.HasSuffix(got, "05") {
		t.Errorf("a width of 1024 should still pad, got %d bytes", len(got))
	}
}

// Lengths beyond int64 saturate instead of overflowing, as Ruby's 1 << 63 cases.
func TestOutOfRangeLengths(t *testing.T) {
	sf := &StandardFilters{}
	huge, negHuge := 1e20, -1e20
	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"truncate huge", sf.Truncate("foobar", huge, nil), "foobar"},
		{"truncate max int64", sf.Truncate("foobar", int64(math.MaxInt64), nil), "foobar"},
		{"truncate min int64", sf.Truncate("foobar", int64(math.MinInt64), nil), "..."},
		{"truncate -huge", sf.Truncate("foobar", negHuge, nil), "..."},
		{"truncate unparsable", sf.Truncate("foobar", "7.0", nil), "foobar"},
		{"truncatewords huge", sf.TruncateWords("hello world", huge, nil), "hello world"},
		{"slice huge length", sf.Slice("hello", 0, huge), "hello"},
		{"slice huge offset", sf.Slice("hello", huge, 1), ""},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %#v, want %#v", c.name, c.got, c.want)
		}
	}
	if got, _ := ToInteger(1e20); got != math.MaxInt {
		t.Errorf("ToInteger(1e20) = %d, want MaxInt", got)
	}
	if got, _ := ToInteger(-1e20); got != math.MinInt {
		t.Errorf("ToInteger(-1e20) = %d, want MinInt", got)
	}
	if got, _ := ToInteger(uint64(math.MaxUint64)); got != math.MaxInt {
		t.Errorf("ToInteger(MaxUint64) = %d, want MaxInt", got)
	}
}

func TestEqualityReviewCorners(t *testing.T) {
	a := map[string]interface{}{"a": 1}
	b := map[interface{}]interface{}{"a": 1}
	if liquidEqual(a, b) != liquidEqual(b, a) {
		t.Errorf("map equality is not symmetric: %v vs %v", liquidEqual(a, b), liquidEqual(b, a))
	}
	if !liquidEqual(a, b) {
		t.Error("maps with the same keys and values should be equal whatever their key type")
	}
	if !liquidEqual(float32(0.1), 0.1) {
		t.Error("float32(0.1) should equal 0.1, as arithmetic reads it")
	}
}

func TestToDateZones(t *testing.T) {
	for in, offset := range map[string]int{
		"2024-03-05 14:30:00 PST":   -8 * 3600,
		"2024-03-05 14:30:00 EDT":   -4 * 3600,
		"2024-03-05 14:30:00 UTC":   0,
		"2024-03-05 14:30:00 GMT":   0,
		"2024-03-05T14:30:00+0100":  3600,
		"2024-03-05T14:30:00+01:00": 3600,
	} {
		got := ToDate(in)
		if got == nil {
			t.Errorf("ToDate(%q) = nil", in)
			continue
		}
		if _, off := got.Zone(); off != offset {
			t.Errorf("ToDate(%q) offset = %d, want %d", in, off, offset)
		}
	}
	for _, in := range []string{"2024-03-05 14:30:00 XYZ", "Tue, 05 Mar 2024 14:30:00 SGT"} {
		got := ToDate(in)
		if got == nil || got.Format("2006-01-02 15:04 -0700") != "2024-03-05 14:30 +0000" {
			t.Errorf("ToDate(%q) = %v, want the time as written, in UTC", in, got)
		}
	}
}

// The int64 fast path must agree with big.Int arithmetic, overflow included.
func TestInt64ArithmeticFastPath(t *testing.T) {
	values := []int64{0, 1, -1, 2, -2, 3, -3, 7, -7, math.MaxInt64, math.MinInt64, math.MaxInt64 - 1, math.MinInt64 + 1, 1 << 32, -(1 << 32)}
	for _, x := range values {
		for _, y := range values {
			for _, op := range []byte{'+', '-', '*', '/', '%'} {
				if (op == '/' || op == '%') && y == 0 {
					continue
				}
				fast, ok := int64Arithmetic(x, y, op)
				want := bigIntArithmetic(x, y, op)
				if want.IsInt64() != ok || (ok && fast != want.Int64()) {
					t.Errorf("%d %c %d: fast = %d (ok %v), want %s", x, op, y, fast, ok, want)
				}
			}
		}
	}
}

func bigIntArithmetic(x, y int64, op byte) *big.Int {
	a, b := big.NewInt(x), big.NewInt(y)
	switch op {
	case '+':
		return a.Add(a, b)
	case '-':
		return a.Sub(a, b)
	case '*':
		return a.Mul(a, b)
	}
	// floor division and modulo through exact rationals
	q := floorRat(new(big.Rat).SetFrac(a, b))
	if op == '/' {
		return q
	}
	return a.Sub(a, new(big.Int).Mul(b, q))
}

// truncatewords splits like Ruby's split(" ", words + 1), trailing whitespace included.
func TestTruncateWordsRubySplit(t *testing.T) {
	sf := &StandardFilters{}
	tests := []struct {
		in    string
		words interface{}
		want  string
	}{
		{"hello world", 2, "hello world"},
		{"hello world ", 2, "hello world..."},
		{"  hello   world  foo", 2, "hello world..."},
		{"hello\tworld\nfoo", 1, "hello..."},
		{"hello", 1, "hello"},
		{"   ", 1, "   "},
		{"", 1, ""},
		{"one two three", math.MaxInt32, "one two three"},
		{"one two", "x", "one two"},
	}
	for _, tt := range tests {
		if got := sf.TruncateWords(tt.in, tt.words, nil); got != tt.want {
			t.Errorf("TruncateWords(%q, %v) = %q, want %q", tt.in, tt.words, got, tt.want)
		}
	}
	for _, tt := range []struct {
		in    string
		limit int
		want  []string
	}{
		{"a b ", 3, []string{"a", "b", ""}},
		{"a  b c", 2, []string{"a", "b c"}},
		{" a", 2, []string{"a"}},
		{"   ", 2, []string{""}},
		{"", 2, nil},
	} {
		got := splitASCIIWhitespaceLimit(tt.in, tt.limit)
		if strings.Join(got, "|") != strings.Join(tt.want, "|") || len(got) != len(tt.want) {
			t.Errorf("split(%q, %d) = %q, want %q", tt.in, tt.limit, got, tt.want)
		}
	}
}

func TestStrftimeLongFlagRun(t *testing.T) {
	when := time.Date(2024, 3, 5, 14, 30, 7, 0, time.UTC)
	format := "%" + strings.Repeat("-_0^#", 100000) + "d"
	start := time.Now()
	if got := strftime(&when, format); got != "5" {
		t.Errorf("got %q, want 5", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("a long run of flags took %v", d)
	}
}

// A range ending at the int maximum stops instead of wrapping around.
func TestRangeEndingAtMaxIntStops(t *testing.T) {
	got := SliceCollection(&Range{Start: math.MaxInt - 2, End: math.MaxInt}, 0, nil)
	if len(got) != 3 {
		t.Errorf("got %d elements, want 3", len(got))
	}
}
