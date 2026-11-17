package integration

import (
	"net"
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestAnElsifThatDoesNotParseIsASyntaxError checks the results of Shopify
// Liquid 5.11.0 for an elsif whose condition does not parse: a syntax error,
// in lax and strict mode, as an if.
func TestAnElsifThatDoesNotParseIsASyntaxError(t *testing.T) {
	for _, c := range []struct{ mode, source, want string }{
		{"lax", "{% if x == 10 %}y{% elsif ' %}z{% endif %}", "Liquid syntax error: Syntax Error in tag 'if' - Valid syntax: if [expression]"},
		{"strict", "{% if x == 10 %}y{% elsif a[' %}z{% endif %}", `Liquid syntax error: Unexpected character ' in "a['"`},
	} {
		r := renderWith(newEnv(liquid.ErrorMode(c.mode), limits{}), c.source, nil)
		if r.parseErr == nil || r.parseErr.Error() != c.want {
			t.Errorf("%s: %q gives %q (%v), want %q", c.mode, c.source, r.output, r.parseErr, c.want)
		}
	}
}

// TestAConditionIsLexedWholeInStrict2 checks that strict2 and rigid lex the
// markup of a condition as strict mode does, as Ruby's strict2 parse.
func TestAConditionIsLexedWholeInStrict2(t *testing.T) {
	for _, mode := range []liquid.ErrorMode{"strict2", "rigid"} {
		r := renderWith(newEnv(mode, limits{}), "{% if a['k] == 1 %}b{% endif %}", nil)
		want := `Liquid syntax error: Unexpected character ' in "a['k] == 1"`
		if r.parseErr == nil || r.parseErr.Error() != want {
			t.Errorf("%s gives %q (%v), want %q", mode, r.output, r.parseErr, want)
		}
	}
}

type stringerBytes [4]byte

func (s stringerBytes) String() string { return "id-1234" }

type post struct {
	Title string
	Count uint
}

// TestAStringerOrATypedSliceRendersAsBefore checks that a value with a
// String method renders it, even when it is a slice or an array, and that a
// typed slice given to a string filter writes its items without Go types. A
// post is a plain Go struct, not a drop: it writes nothing of its fields,
// and nil in a list.
func TestAStringerOrATypedSliceRendersAsBefore(t *testing.T) {
	data := map[string]interface{}{
		"id":    stringerBytes{1, 2, 3, 4},
		"ip":    net.ParseIP("10.0.0.1"),
		"uints": []uint{5, 6},
		"posts": []post{{"b", 2}},
	}
	for _, c := range []struct{ source, want string }{
		{"{{ id }}|{{ ip }}", "id-1234|10.0.0.1"},
		{`{{ id | append: "" }}|{{ ip | append: "" }}`, "id-1234|10.0.0.1"},
		{`{{ uints | append: "" }}`, "[5, 6]"},
		{`{{ posts | append: "" }}`, "[nil]"},
		{"{{ uints }}", "56"},
	} {
		r := renderWith(newEnv("lax", limits{}), c.source, data)
		if r.output != c.want {
			t.Errorf("%q gives %q, want %q", c.source, r.output, c.want)
		}
	}
}

type titled struct{ Title string }

// TestTheArrayFiltersReadAPropertyOfAPointer checks that a drop has a
// property for the array filters, and that a pointer to a plain Go struct,
// which is not a drop, has none, as in Ruby.
func TestTheArrayFiltersReadAPropertyOfAPointer(t *testing.T) {
	data := map[string]interface{}{
		"pposts": []interface{}{&titled{"a"}, &titled{"b"}, &titled{"a"}},
		"drops":  []interface{}{NewObjectWrapperDrop(2), NewObjectWrapperDrop(1), NewObjectWrapperDrop(2)},
	}
	for _, c := range []struct{ source, want string }{
		{`{{ pposts | uniq: "title" | size }}`, "0"},
		{`{{ pposts | where: "title", "a" | size }}`, "0"},
		{`{{ pposts | sort: "title" | map: "title" | join: "," }}`, ""},
		{`{{ pposts | compact: "title" | size }}`, "0"},
		{`{{ drops | uniq: "value" | size }}`, "2"},
		{`{{ drops | sort: "value" | map: "value" | join: "," }}`, "1,2,2"},
		{`{{ drops | where: "value", 2 | size }}`, "2"},
	} {
		r := renderWith(newEnv("lax", limits{}), c.source, data)
		if r.output != c.want {
			t.Errorf("%q gives %q, want %q", c.source, r.output, c.want)
		}
	}
}

// TestASelfContainingArrayInAFilterDoesNotRecurse checks that join and
// reverse over an array that holds itself leave the inner array out instead
// of recursing without end.
func TestASelfContainingArrayInAFilterDoesNotRecurse(t *testing.T) {
	arr := []interface{}{1, nil, 2}
	arr[1] = arr
	for _, c := range []struct{ source, want string }{
		{`{{ self | join: "," }}`, "1,2"},
		{`{{ self | reverse | join: "," }}`, "2,1"},
	} {
		r := renderWith(newEnv("lax", limits{}), c.source, map[string]interface{}{"self": arr})
		if r.output != c.want {
			t.Errorf("%q gives %q, want %q", c.source, r.output, c.want)
		}
	}
}

// TestAPointerToATimeIsOrdered checks that a *time.Time compares as the time.
func TestAPointerToATimeIsOrdered(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Hour)
	r := renderWith(newEnv("lax", limits{}), "{% if pt < t2 %}y{% endif %}{% if t2 > pt %}z{% endif %}", map[string]interface{}{"pt": &t1, "t2": t2})
	if r.output != "yz" {
		t.Errorf("a *time.Time comparison gives %q, want %q", r.output, "yz")
	}
}

// TestATablerowOverFalseRendersNothing checks the result of Shopify Liquid
// 5.11.0 for a tablerow over false: nothing, as over nil.
func TestATablerowOverFalseRendersNothing(t *testing.T) {
	r := renderWith(newEnv("lax", limits{}), "{% tablerow i in f %}x{% endtablerow %}", map[string]interface{}{"f": false})
	if r.output != "" {
		t.Errorf("a tablerow over false gives %q", r.output)
	}
}

// TestAnIntegerItemRaisesThePropertyError checks that the array filters
// return Ruby's error for an integer item, which the render shows empty.
func TestAnIntegerItemRaisesThePropertyError(t *testing.T) {
	sf := &liquid.StandardFilters{}
	ints := []interface{}{3, 1}
	for name, call := range map[string]func() (interface{}, error){
		"map":          func() (interface{}, error) { return sf.Map(ints, "a") },
		"where":        func() (interface{}, error) { return sf.Where(ints, "a", nil) },
		"reject":       func() (interface{}, error) { return sf.Reject(ints, "a", nil) },
		"sort_natural": func() (interface{}, error) { return sf.SortNatural(ints, "a") },
		"has":          func() (interface{}, error) { return sf.Has(ints, "a", nil) },
		"find":         func() (interface{}, error) { return sf.Find(ints, "a", nil) },
		"find_index":   func() (interface{}, error) { return sf.FindIndex(ints, "a", nil) },
		"sort":         func() (interface{}, error) { return sf.Sort(ints, "a") },
		"uniq":         func() (interface{}, error) { return sf.Uniq(ints, "a") },
		"compact":      func() (interface{}, error) { return sf.Compact(ints, "a") },
		"sum":          func() (interface{}, error) { return sf.Sum(ints, "a") },
	} {
		v, err := call()
		if err == nil || err.Error() != "Liquid error: cannot select the property 'a'" || v != nil {
			t.Errorf("%s over integers gives %v, %v", name, v, err)
		}
	}
	if v, err := sf.Has([]interface{}{1.5}, "a", nil); err != nil || v != nil {
		t.Errorf("has over a float gives %v, %v, want nil", v, err)
	}
}
