package liquid

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

// Text filters count characters, not bytes, as Ruby strings do.

func TestCapitalizeUnicode(t *testing.T) {
	sf := &StandardFilters{}
	for in, want := range map[string]string{
		"élodie":         "Élodie",
		"ÉLODIE":         "Élodie",
		"my great TITLE": "My great title",
		"":               "",
		"日本":             "日本",
		"émile zola":     "Émile zola",
	} {
		if got := sf.Capitalize(in); got != want {
			t.Errorf("Capitalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSizeUnicode(t *testing.T) {
	sf := &StandardFilters{}
	if got := sf.Size("été"); got != 3 {
		t.Errorf(`Size("été") = %v, want 3`, got)
	}
	if got := sf.Size("测试"); got != 2 {
		t.Errorf(`Size("测试") = %v, want 2`, got)
	}
}

func TestSliceStrings(t *testing.T) {
	sf := &StandardFilters{}
	for _, tt := range []struct {
		in     interface{}
		offset interface{}
		length interface{}
		want   string
	}{
		{"été", 0, 2, "ét"},
		{"été", 1, nil, "t"},
		{"Bonjour", -3, 2, "ou"},
		{"foobar", -2, 1000, "ar"},
		{"foobar", -1, nil, "r"},
		{"foobar", -6, 2, "fo"},
		{"foobar", -7, 2, ""},
		{"foobar", 6, 1, ""},
		{"foobar", 7, 1, ""},
		{"foobar", 1, -1, ""},
		{"foobar", 1, 0, ""},
		{12345, 1, 2, "23"},
		{"测试测试", 1, 2, "试测"},
	} {
		if got := sf.Slice(tt.in, tt.offset, tt.length); got != tt.want {
			t.Errorf("Slice(%#v, %v, %v) = %#v, want %q", tt.in, tt.offset, tt.length, got, tt.want)
		}
	}
}

func TestSliceArrays(t *testing.T) {
	sf := &StandardFilters{}
	list := []interface{}{"a", "b", "c", "d"}
	for _, tt := range []struct {
		in     interface{}
		offset interface{}
		length interface{}
		want   []interface{}
	}{
		{list, 1, 2, []interface{}{"b", "c"}},
		{list, -2, 2, []interface{}{"c", "d"}},
		{list, -1, nil, []interface{}{"d"}},
		{list, -5, 2, []interface{}{}},
		{list, 4, 1, []interface{}{}},
		{list, 1, -1, []interface{}{}},
		{[]string{"a", "b", "c"}, -2, 5, []interface{}{"b", "c"}},
		{[2]int{1, 2}, -1, 1, []interface{}{2}},
	} {
		got := sf.Slice(tt.in, tt.offset, tt.length)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Slice(%#v, %v, %v) = %#v, want %#v", tt.in, tt.offset, tt.length, got, tt.want)
		}
	}
}

func TestTruncateUnicode(t *testing.T) {
	sf := &StandardFilters{}
	for _, tt := range []struct {
		in     string
		length interface{}
		ellips interface{}
		want   string
	}{
		{"Événement spécial", 6, nil, "Évé..."},
		{"测试测试测试测试", 5, nil, "测试..."},
		{"测试测试测试测试", 6, "…", "测试测试测…"},
		{"1234567890", 0, nil, "..."},
		{"1234567890", -1, nil, "..."},
		{"1234567890", 7, nil, "1234..."},
		{"1234567890", 20, nil, "1234567890"},
		{"1234567890", 10, nil, "1234567890"},
		{"1234567890", 5, "", "12345"},
		{"été", 3, nil, "été"},
		{"", 0, nil, ""},
	} {
		got := sf.Truncate(tt.in, tt.length, tt.ellips)
		if got != tt.want {
			t.Errorf("Truncate(%q, %v, %v) = %q, want %q", tt.in, tt.length, tt.ellips, got, tt.want)
		}
		if !utf8.ValidString(got) {
			t.Errorf("Truncate(%q, %v) produced invalid UTF-8", tt.in, tt.length)
		}
	}
	long := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if got := sf.Truncate(long, nil, nil); utf8.RuneCountInString(got) != 50 {
		t.Errorf("Truncate without a length should keep 50 characters, got %q", got)
	}
}

func TestTruncateWordsEdges(t *testing.T) {
	sf := &StandardFilters{}
	for _, tt := range []struct {
		in    string
		words interface{}
		want  string
	}{
		{"one two three four", 0, "one..."},
		{"one two three four", -1, "one..."},
		{"one two three four", 2, "one two..."},
		{"one two three four", 4, "one two three four"},
		{"  one   two three", 2, "one two..."},
		{"été hiver printemps", 2, "été hiver..."},
		{"a b c", 1, "a b..."},
	} {
		if got := sf.TruncateWords(tt.in, tt.words, nil); got != tt.want {
			t.Errorf("TruncateWords(%q, %v) = %q, want %q", tt.in, tt.words, got, tt.want)
		}
	}
}

func TestSplitRubySemantics(t *testing.T) {
	sf := &StandardFilters{}
	for _, tt := range []struct {
		in      interface{}
		pattern interface{}
		want    []interface{}
	}{
		{"a,b,c", ",", []interface{}{"a", "b", "c"}},
		{"a,b,,", ",", []interface{}{"a", "b"}},
		{",a", ",", []interface{}{"", "a"}},
		{"", ",", []interface{}{}},
		{nil, " ", []interface{}{}},
		{"  a  b ", " ", []interface{}{"a", "b"}},
		{"a\tb\nc", " ", []interface{}{"a", "b", "c"}},
		{"été", "", []interface{}{"é", "t", "é"}},
		{"abc", nil, []interface{}{"a", "b", "c"}},
	} {
		got := sf.Split(tt.in, tt.pattern)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Split(%#v, %#v) = %#v, want %#v", tt.in, tt.pattern, got, tt.want)
		}
	}
}
