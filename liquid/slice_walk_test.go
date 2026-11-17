package liquid

import (
	"math"
	"testing"
	"unicode/utf8"
)

// sliceByRunes and truncateByRunes are slice and truncate on strings as they
// were before they walked the string: through a copy to []rune.
func sliceByRunes(s string, offset, length int) string {
	runes := []rune(s)
	start, end, ok := sliceBounds(len(runes), offset, length)
	if !ok {
		return ""
	}
	return string(runes[start:end])
}

func truncateByRunes(s string, length int, ellipsis string) string {
	runes := []rune(s)
	if len(runes) <= length {
		return s
	}
	l := 0
	if e := utf8.RuneCountInString(ellipsis); length > e {
		l = length - e
	}
	return string(runes[:l]) + ellipsis
}

// TestSliceAndTruncateWalkAsRunes checks that slice and truncate give, on
// valid and invalid UTF-8, what the copy to []rune gave.
func TestSliceAndTruncateWalkAsRunes(t *testing.T) {
	sf := &StandardFilters{}
	inputs := []string{"", "a", "héllo", "日本語テキスト", "a\xffb\xfe\xfdc", "\xe6\x97", "é\xe6\x97x", "abc\x80"}
	numbers := []int{math.MinInt, -10, -4, -3, -2, -1, 0, 1, 2, 3, 4, 7, 10, math.MaxInt}
	for _, in := range inputs {
		for _, offset := range numbers {
			for _, length := range numbers {
				if got, want := sliceOf(sf, in, offset, length), sliceByRunes(in, offset, length); got != want {
					t.Errorf("slice(%q, %d, %d) = %q, want %q", in, offset, length, got, want)
				}
			}
			for _, ellipsis := range []string{"...", "", "é…"} {
				if got, want := sf.Truncate(in, offset, ellipsis), truncateByRunes(in, offset, ellipsis); got != want {
					t.Errorf("truncate(%q, %d, %q) = %q, want %q", in, offset, ellipsis, got, want)
				}
			}
		}
	}
}
