package integration

import (
	"errors"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestTheArrayFiltersReadAPropertyAsRuby checks the results of Shopify Liquid
// 5.11.0 for the filters that read a property of each item, over items that
// are not hashes. They read it as Ruby's item[property]: a string gives the
// property when it contains it, an integer raises "cannot select the
// property", and an item without [] (a float, a boolean, nil) makes the
// filter give nil, or 0 for sum. A filter error renders empty in liquidgo
// (deliberate: no error text in emails), where Ruby writes
// "Liquid error: cannot select the property 'a'", so the rows over integers
// cannot tell the error from a nil result: TestAnIntegerItemRaisesThePropertyError
// checks the error each filter returns.
func TestTheArrayFiltersReadAPropertyAsRuby(t *testing.T) {
	data := func() map[string]interface{} {
		return map[string]interface{}{
			"ints": []interface{}{3, 1},
			"flts": []interface{}{1.5, 2.5},
			"ss":   []interface{}{"ab", "cd"},
			"bs":   []interface{}{true, false},
			"hf":   []interface{}{map[string]interface{}{"a": 1}, 1.5},
			"fh":   []interface{}{1.5, map[string]interface{}{"a": 1}},
			"nh":   []interface{}{nil, map[string]interface{}{"a": 1}},
			"one":  []interface{}{5},
			"hi":   []interface{}{map[string]interface{}{"a": 1}, 2},
			"hs": []interface{}{
				map[string]interface{}{"a": ""}, map[string]interface{}{"a": []interface{}{}},
				map[string]interface{}{"a": 0}, map[string]interface{}{"a": false},
				map[string]interface{}{"a": nil}, map[string]interface{}{"a": map[string]interface{}{}},
				map[string]interface{}{"a": "x"},
			},
		}
	}
	for _, c := range []struct{ source, want string }{
		// Deliberate: an empty string, list or hash is falsy, and a string
		// matches the number it writes. Ruby gives 5, 2, 0, 1, 1 and 0.
		{"{{ hs | where: 'a' | size }}", "2"},
		{"{{ hs | reject: 'a' | size }}", "5"},
		{"{{ hs | find_index: 'a' }}", "2"},
		{"{{ hs | where: 'a', '' | size }}", "1"},
		{"{{ hs | where: 'a', 0 | size }}", "1"},
		{"{{ hs | where: 'a', '0' | size }}", "1"},
		{"{{ ints | map: 'a' | join: ',' }}", ""},
		{"{{ ints | where: 'a' | join: ',' }}", ""},
		{"{{ ints | where: 'a', 1 | join: ',' }}", ""},
		{"{{ ints | reject: 'a' | join: ',' }}", ""},
		{"{{ ints | has: 'a' | join: ',' }}", ""},
		{"{{ ints | has: 'a', 1 | join: ',' }}", ""},
		{"{{ ints | find: 'a' | join: ',' }}", ""},
		{"{{ ints | find_index: 'a' | join: ',' }}", ""},
		{"{{ ints | sort: 'a' | join: ',' }}", ""},
		{"{{ ints | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ ints | uniq: 'a' | join: ',' }}", ""},
		{"{{ ints | compact: 'a' | join: ',' }}", ""},
		{"{{ ints | sum: 'a' | join: ',' }}", ""},
		{"{{ flts | map: 'a' | join: ',' }}", ","},
		{"{{ flts | where: 'a' | join: ',' }}", ""},
		{"{{ flts | where: 'a', 1 | join: ',' }}", ""},
		{"{{ flts | reject: 'a' | join: ',' }}", ""},
		{"{{ flts | has: 'a' | join: ',' }}", ""},
		{"{{ flts | has: 'a', 1 | join: ',' }}", ""},
		{"{{ flts | find: 'a' | join: ',' }}", ""},
		{"{{ flts | find_index: 'a' | join: ',' }}", ""},
		{"{{ flts | sort: 'a' | join: ',' }}", ""},
		{"{{ flts | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ flts | uniq: 'a' | join: ',' }}", ""},
		{"{{ flts | compact: 'a' | join: ',' }}", ""},
		{"{{ flts | sum: 'a' | join: ',' }}", "0"},
		{"{{ ss | map: 'a' | join: ',' }}", "a,"},
		{"{{ ss | where: 'a' | join: ',' }}", "ab"},
		{"{{ ss | where: 'a', 1 | join: ',' }}", ""},
		{"{{ ss | reject: 'a' | join: ',' }}", "cd"},
		{"{{ ss | has: 'a' | join: ',' }}", "true"},
		{"{{ ss | has: 'a', 1 | join: ',' }}", "false"},
		{"{{ ss | find: 'a' | join: ',' }}", "ab"},
		{"{{ ss | find_index: 'a' | join: ',' }}", "0"},
		{"{{ ss | sort: 'a' | join: ',' }}", "ab,cd"},
		{"{{ ss | sort_natural: 'a' | join: ',' }}", "ab,cd"},
		{"{{ ss | uniq: 'a' | join: ',' }}", "ab,cd"},
		{"{{ ss | compact: 'a' | join: ',' }}", "ab"},
		{"{{ ss | sum: 'a' | join: ',' }}", "0"},
		{"{{ bs | map: 'a' | join: ',' }}", ","},
		{"{{ bs | where: 'a' | join: ',' }}", ""},
		{"{{ bs | where: 'a', 1 | join: ',' }}", ""},
		{"{{ bs | reject: 'a' | join: ',' }}", ""},
		{"{{ bs | has: 'a' | join: ',' }}", ""},
		{"{{ bs | has: 'a', 1 | join: ',' }}", ""},
		{"{{ bs | find: 'a' | join: ',' }}", ""},
		{"{{ bs | find_index: 'a' | join: ',' }}", ""},
		{"{{ bs | sort: 'a' | join: ',' }}", ""},
		{"{{ bs | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ bs | uniq: 'a' | join: ',' }}", ""},
		{"{{ bs | compact: 'a' | join: ',' }}", ""},
		{"{{ bs | sum: 'a' | join: ',' }}", "0"},
		{"{{ hf | map: 'a' | join: ',' }}", "1,"},
		{"{{ hf | where: 'a' | join: ',' }}", ""},
		{"{{ hf | where: 'a', 1 | join: ',' }}", ""},
		{"{{ hf | reject: 'a' | join: ',' }}", ""},
		{"{{ hf | has: 'a' | join: ',' }}", "true"},
		{"{{ hf | has: 'a', 1 | join: ',' }}", "true"},
		{"{{ hf | find: 'a' | join: ',' }}", "{\"a\"=>1}"},
		{"{{ hf | find_index: 'a' | join: ',' }}", "0"},
		{"{{ hf | sort: 'a' | join: ',' }}", ""},
		{"{{ hf | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ hf | uniq: 'a' | join: ',' }}", ""},
		{"{{ hf | compact: 'a' | join: ',' }}", ""},
		{"{{ hf | sum: 'a' | join: ',' }}", "1"},
		{"{{ fh | map: 'a' | join: ',' }}", ",1"},
		{"{{ fh | where: 'a' | join: ',' }}", ""},
		{"{{ fh | where: 'a', 1 | join: ',' }}", ""},
		{"{{ fh | reject: 'a' | join: ',' }}", ""},
		{"{{ fh | has: 'a' | join: ',' }}", ""},
		{"{{ fh | has: 'a', 1 | join: ',' }}", ""},
		{"{{ fh | find: 'a' | join: ',' }}", ""},
		{"{{ fh | find_index: 'a' | join: ',' }}", ""},
		{"{{ fh | sort: 'a' | join: ',' }}", ""},
		{"{{ fh | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ fh | uniq: 'a' | join: ',' }}", ""},
		{"{{ fh | compact: 'a' | join: ',' }}", ""},
		{"{{ fh | sum: 'a' | join: ',' }}", "1"},
		{"{{ nh | map: 'a' | join: ',' }}", ",1"},
		{"{{ nh | where: 'a' | join: ',' }}", ""},
		{"{{ nh | where: 'a', 1 | join: ',' }}", ""},
		{"{{ nh | reject: 'a' | join: ',' }}", ""},
		{"{{ nh | has: 'a' | join: ',' }}", ""},
		{"{{ nh | has: 'a', 1 | join: ',' }}", ""},
		{"{{ nh | find: 'a' | join: ',' }}", ""},
		{"{{ nh | find_index: 'a' | join: ',' }}", ""},
		{"{{ nh | sort: 'a' | join: ',' }}", ""},
		{"{{ nh | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ nh | uniq: 'a' | join: ',' }}", ""},
		{"{{ nh | compact: 'a' | join: ',' }}", ""},
		{"{{ nh | sum: 'a' | join: ',' }}", "1"},
		{"{{ one | map: 'a' | join: ',' }}", ""},
		{"{{ one | where: 'a' | join: ',' }}", ""},
		{"{{ one | where: 'a', 1 | join: ',' }}", ""},
		{"{{ one | reject: 'a' | join: ',' }}", ""},
		{"{{ one | has: 'a' | join: ',' }}", ""},
		{"{{ one | has: 'a', 1 | join: ',' }}", ""},
		{"{{ one | find: 'a' | join: ',' }}", ""},
		{"{{ one | find_index: 'a' | join: ',' }}", ""},
		{"{{ one | sort: 'a' | join: ',' }}", "5"},
		{"{{ one | sort_natural: 'a' | join: ',' }}", "5"},
		{"{{ one | uniq: 'a' | join: ',' }}", "5"},
		{"{{ one | compact: 'a' | join: ',' }}", ""},
		{"{{ one | sum: 'a' | join: ',' }}", ""},
		{"{{ hi | map: 'a' | join: ',' }}", ""},
		{"{{ hi | where: 'a' | join: ',' }}", ""},
		{"{{ hi | where: 'a', 1 | join: ',' }}", ""},
		{"{{ hi | reject: 'a' | join: ',' }}", ""},
		{"{{ hi | has: 'a' | join: ',' }}", "true"},
		{"{{ hi | has: 'a', 1 | join: ',' }}", "true"},
		{"{{ hi | find: 'a' | join: ',' }}", "{\"a\"=>1}"},
		{"{{ hi | find_index: 'a' | join: ',' }}", "0"},
		{"{{ hi | sort: 'a' | join: ',' }}", ""},
		{"{{ hi | sort_natural: 'a' | join: ',' }}", ""},
		{"{{ hi | uniq: 'a' | join: ',' }}", ""},
		{"{{ hi | compact: 'a' | join: ',' }}", ""},
		{"{{ hi | sum: 'a' | join: ',' }}", ""},
	} {
		r := renderWith(newEnv("lax", limits{}), c.source, data())
		if r.parseErr != nil || r.output != c.want || !onlyFilterArgumentErrors(r.errors) {
			t.Errorf("%q gives %q (%v, %v), want %q", c.source, r.output, r.parseErr, r.errors, c.want)
		}
	}
}

// onlyFilterArgumentErrors reports whether errs holds only the ArgumentErrors
// a filter that cannot select a property records.
func onlyFilterArgumentErrors(errs []error) bool {
	for _, e := range errs {
		var argument *liquid.ArgumentError
		if !errors.As(e, &argument) || !strings.Contains(e.Error(), "cannot select the property") {
			return false
		}
	}
	return true
}
