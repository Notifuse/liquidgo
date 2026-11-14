package liquid

import (
	"math"
	"testing"
)

func filtersWithLimits(renderLength, assignScore int) *StandardFilters {
	ctx := NewContext()
	ctx.SetResourceLimits(NewResourceLimits(ResourceLimitsConfig{RenderLengthLimit: &renderLength, AssignScoreLimit: &assignScore}))
	return &StandardFilters{context: ctx}
}

func raisesMemoryError(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if _, ok := recover().(*MemoryError); !ok {
			t.Errorf("%s: want a MemoryError", name)
		}
	}()
	fn()
}

// TestFiltersRefuseResultsOverTheLimits checks that each filter that can
// multiply its input refuses a result over the limits before it builds it,
// and gives its result under them.
func TestFiltersRefuseResultsOverTheLimits(t *testing.T) {
	sf := filtersWithLimits(10, 10)
	for name, fn := range map[string]func(){
		"replace":       func() { sf.Replace("aaaa", "a", "aaa") },
		"replace_first": func() { sf.ReplaceFirst("aaaaaaaa", "a", "xxxxx") },
		"replace_last":  func() { sf.ReplaceLast("aaaaaaaa", "a", "xxxxx") },
		"append":        func() { sf.Append("aaaaaa", "bbbbbb") },
		"prepend":       func() { sf.Prepend("aaaaaa", "bbbbbb") },
		"join":          func() { sf.Join([]interface{}{"a", "b", "c"}, "xxxxx") },
		"join typed":    func() { sf.Join([]string{"a", "b", "c"}, "xxxxx") },
		"split":         func() { sf.Split("a,b,c,d,e,f,g,h,i,j,k", ",") },
		"split chars":   func() { sf.Split("abcdefghijk", "") },
		"split spaces":  func() { sf.Split("a b c d e f g h i j k", " ") },
		"concat":        func() { _, _ = sf.Concat([]interface{}{1, 2, 3, 4, 5, 6}, []interface{}{7, 8, 9, 10, 11}) },
	} {
		raisesMemoryError(t, name, fn)
	}

	sf = filtersWithLimits(10, 10)
	if got := sf.Replace("aa", "a", "bb"); got != "bbbb" {
		t.Errorf("replace under the limit gives %q", got)
	}
	if got := sf.Join([]interface{}{"a", "b"}, "-"); got != "a-b" {
		t.Errorf("join under the limit gives %q", got)
	}
	if got := sf.Split("a,b", ","); len(got) != 2 {
		t.Errorf("split under the limit gives %v", got)
	}

	// Without limits, nothing is refused.
	free := &StandardFilters{context: NewContext()}
	if got := free.Replace("aaaa", "a", "aaa"); len(got) != 12 {
		t.Errorf("replace without limits gives %q", got)
	}
}

func TestFilterResultLengths(t *testing.T) {
	if got := replacedLength("abc", "", "xx", 4); got != 11 {
		t.Errorf("replacedLength with an empty search = %d, want 11", got)
	}
	if got := replacedLength("aaa", "a", "", 3); got != 0 {
		t.Errorf("replacedLength that shrinks = %d, want 0", got)
	}
	if got := replacedLength("a", "a", "xxxx", math.MaxInt/2+1); got != math.MaxInt {
		t.Errorf("replacedLength does not saturate: %d", got)
	}
	if got := joinedLength([]string{"a", "b", "c"}, ", "); got != 7 {
		t.Errorf("joinedLength = %d, want 7", got)
	}
	if got := joinedLength([]string{"a", "b"}, string(make([]byte, 1<<20))); got != 1<<20+2 {
		t.Errorf("joinedLength = %d", got)
	}
	if got := splitCount("héllo", ""); got != 5 {
		t.Errorf("splitCount by characters = %d, want 5", got)
	}
	if got := splitCount("a,b,,c", ","); got != 4 {
		t.Errorf("splitCount = %d, want 4", got)
	}
}
