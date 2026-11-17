package integration

import (
	gocontext "context"
	"errors"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// The tests of Shopify Liquid 5.14 that the conformance extraction leaves
// out, ported with their templates and expected outputs.

func render514(t *testing.T, src string, partials map[string]string, opts ...liquid.RenderOption) (liquid.Result, error) {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: &mapFileSystem{templates: partials}})
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	return tmpl.Render(gocontext.Background(), nil, opts...)
}

// standard_filter_test.rb: test_squish_filter
func TestSquish(t *testing.T) {
	for src, want := range map[string]string{
		"{{ \" foo   bar\n\t   boo   \" | squish }}": "foo bar boo",
		"{{ nil | squish }}":                         "",
		`{{ " " | squish }}`:                         "",
	} {
		if res, err := render514(t, src, nil); err != nil || res.Output != want {
			t.Errorf("%q gives %q (%v), want %q", src, res.Output, err, want)
		}
	}
}

// self_drop_context_test.rb
func TestSelfDropAcrossRenders(t *testing.T) {
	for _, c := range []struct {
		name, src, want string
		partials        map[string]string
	}{
		{"passed as a render param keeps its scope",
			"{%- assign var = 42 -%}\n{%- assign s = self -%}\n{%- render \"snippet1\", other_self: s -%}\n", "42|43",
			map[string]string{"snippet1": "{%- assign var = 43 -%}\n{{- other_self.var }}|{{ self.var -}}\n"}},
		{"in a render, the inner scope",
			"{%- assign var = 42 -%}\n{%- render \"snippet1\" -%}\n", "99",
			map[string]string{"snippet1": "{%- assign var = 99 -%}\n{{- self.var -}}\n"}},
		{"nested renders keep each level",
			"{%- assign a = 1 -%}\n{%- assign s1 = self -%}\n{%- render \"snippet1\", outer: s1 -%}\n", "1|2|3",
			map[string]string{
				"snippet1": "{%- assign a = 2 -%}\n{%- assign s2 = self -%}\n{%- render \"snippet2\", outer: outer, middle: s2 -%}\n",
				"snippet2": "{%- assign a = 3 -%}\n{{- outer.a }}|{{ middle.a }}|{{ self.a -}}\n",
			}},
		{"variables assigned after it",
			"{%- assign s = self -%}\n{%- assign x = 42 %}{{ s.x -}}\n", "42", nil},
	} {
		if res, err := render514(t, c.src, c.partials); err != nil || res.Output != c.want {
			t.Errorf("%s: %q (%v), want %q", c.name, res.Output, err, c.want)
		}
	}
}

// for_tag_test.rb and table_row_test.rb: a range is not built whole, and each
// item it visits counts in the render score, an empty body included.
func TestRangeLoopsChargeEachItem(t *testing.T) {
	score := 3
	for _, src := range []string{"{% for i in (1..1000) %}{% endfor %}", "{% tablerow i in (1..1000) %}{% endtablerow %}"} {
		res, err := render514(t, src, nil, liquid.WithLimits(liquid.ResourceLimitsConfig{RenderScoreLimit: &score}))
		var memory *liquid.MemoryError
		if !errors.As(err, &memory) || res.RenderScore != 4 {
			t.Errorf("%s: score %d (%v), want a MemoryError at 4", src, res.RenderScore, err)
		}
	}
	// the outputs of Shopify's tests, and the render scores Shopify Liquid
	// 5.14 gives: a range item counts 1 besides its body
	for _, c := range []struct {
		src, want string
		score     int
	}{
		{"{% for i in (1..1000) reversed offset:997 limit:2 %}{{ forloop.length }}:{{ i }}{% endfor %}", "2:9992:998", 9},
		{"{% for i in (1..1000) reversed %}{{ forloop.length }}:{{ i }}{% break %}{% endfor %}", "1000:1000", 6},
		{"{% tablerow i in (1..1000) offset:997 limit:2 %}{{ tablerowloop.length }}:{{ tablerowloop.index }}:{{ i }}{% endtablerow %}", "<tr class=\"row1\">\n<td class=\"col1\">2:1:998</td><td class=\"col2\">2:2:999</td></tr>\n", 13},
		{"{% for i in (1..3) %}x{% endfor %}", "xxx", 7},
		{"{% for i in (1..3) %}{% endfor %}", "", 4},
		{"{% tablerow i in (1..3) %}x{% endtablerow %}", "<tr class=\"row1\">\n<td class=\"col1\">x</td><td class=\"col2\">x</td><td class=\"col3\">x</td></tr>\n", 7},
	} {
		if res, err := render514(t, c.src, nil); err != nil || res.Output != c.want || res.RenderScore != c.score {
			t.Errorf("%s: %q score %d (%v), want %q and %d", c.src, res.Output, res.RenderScore, err, c.want, c.score)
		}
	}
}

// include_tag_test.rb: for_loop? tells for from with
func TestIncludeForLoop(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	for src, want := range map[string]bool{
		"{% include 'product' for products %}":             true,
		"{% include 'product' with product %}":             false,
		"{% include 'header' %}":                           false,
		"{% include 'product' for products as item %}":     true,
		"{% include 'product' with products[0] as item %}": false,
	} {
		tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		node, ok := tmpl.Root().Nodelist()[0].(*tags.IncludeTag)
		if !ok || node.ForLoop() != want {
			t.Errorf("%s: ForLoop() = %v, want %v", src, ok && node.ForLoop(), want)
		}
	}
}

func parse514(t *testing.T, src string, mode liquid.ErrorMode) (*liquid.Template, error) {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), ErrorMode: mode})
	return liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env, LineNumbers: true})
}

// blank_body_error_handling_test.rb: the error of a tag whose body is blank
// writes no text in lax and strict, and writes its text in strict2; raised,
// it is never swallowed.
func TestBlankBodyErrorText(t *testing.T) {
	const comparison = "Liquid error (line 1): comparison of Integer with String failed"
	const invalidInteger = "Liquid error (line 1): invalid integer"
	data := map[string]interface{}{"xs": "bad"}
	for _, c := range []struct {
		src, mode, want string
	}{
		{`{% if 5 > "x" %}{% endif %}`, "lax", ""},
		{`{% if 5 > "x" %}{% endif %}`, "strict", ""},
		{`{% unless 5 > "x" %} {% endunless %}`, "lax", ""},
		{`{% unless 5 > "x" %} {% endunless %}`, "strict", ""},
		{`{% for i in (1..3) offset: xs %}{% endfor %}`, "lax", ""},
		{`{% for i in (1..3) offset: xs %}{% endfor %}`, "strict", ""},
		{`{% if 5 > "x" %}{% endif %}`, "strict2", comparison},
		{`{% if 5 > "x" %}   {% endif %}`, "strict2", comparison},
		{`{% if 5 > "x" %}{% assign a = 1 %}{% endif %}`, "strict2", comparison},
		{`{% if 5 > "x" %}{% comment %}c{% endcomment %}{% endif %}`, "strict2", comparison},
		{`{% if 5 > "x" %}{% capture c %}text{% endcapture %}{% endif %}`, "strict2", comparison},
		{`{% unless 5 > "x" %} {% endunless %}`, "strict2", comparison},
		{`{% for i in (1..3) offset: xs %}{% endfor %}`, "strict2", invalidInteger},
		{`{% if 5 > "x" %}{% echo 1 %}{% endif %}`, "lax", comparison},
		{`{% if 5 > "x" %}{{ "" }}{% endif %}`, "strict", comparison},
		{`{% if 5 > "x" %}{% else %}E{% endif %}`, "strict2", comparison},
	} {
		tmpl, err := parse514(t, c.src, liquid.ErrorMode(c.mode))
		if err != nil {
			t.Fatal(err)
		}
		res, err := tmpl.Render(gocontext.Background(), data)
		if err != nil || res.Output != c.want {
			t.Errorf("%s in %s: %q (%v), want %q", c.src, c.mode, res.Output, err, c.want)
		}
		for _, mode := range []liquid.ErrorMode{"lax", "strict", "strict2"} {
			tmpl, _ := parse514(t, c.src, mode)
			if _, err := tmpl.Render(gocontext.Background(), data, liquid.RaiseErrors()); err == nil {
				t.Errorf("%s in %s: a raised error is swallowed", c.src, mode)
			}
		}
	}
}

// parser.rb: strict2 refuses a bare bracket, which self[...] replaces
func TestStrict2RejectsBareBrackets(t *testing.T) {
	data := map[string]interface{}{"foo": "baz", "bar": "foo"}
	for _, src := range []string{`{{ ["foo"] }}`, `{{ [bar] }}`, `{{ a[ [ 'b' ] ] }}`, `{% if ["foo"] %}x{% endif %}`} {
		_, err := parse514(t, src, "strict2")
		if err == nil || !strings.Contains(err.Error(), "Bare bracket access is not allowed. Use self['...'] instead") {
			t.Errorf("%s in strict2: %v, want the bare bracket error", src, err)
		}
	}
	for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
		tmpl, err := parse514(t, `{{ ["foo"] }}|{{ [bar] }}`, mode)
		if err != nil {
			t.Fatal(err)
		}
		if res, err := tmpl.Render(gocontext.Background(), data); err != nil || res.Output != "baz|baz" {
			t.Errorf("%s: %q (%v), want baz|baz", mode, res.Output, err)
		}
	}
	tmpl, err := parse514(t, `{{ self["foo"] }}|{{ self[bar] }}`, "strict2")
	if err != nil {
		t.Fatal(err)
	}
	if res, err := tmpl.Render(gocontext.Background(), data); err != nil || res.Output != "baz|baz" {
		t.Errorf("self[...] in strict2: %q (%v), want baz|baz", res.Output, err)
	}
}

// assign_test.rb, capture_test.rb, increment_tag_test.rb: in strict2, the
// name of an assign, a capture, an increment or a decrement is one
// identifier; lax keeps any name.
func TestStrict2VariableNames(t *testing.T) {
	for _, src := range []string{
		"{% assign (a(b(c) = 1234 %}", "{% assign [x.y] = 'hello' %}", "{% assign a.b = 'hello' %}", "{% assign 1abc = 'hello' %}",
		"{% capture (x[y %}hello{% endcapture %}", "{% capture a.b %}hello{% endcapture %}", "{% capture 1abc %}hello{% endcapture %}",
		"{% increment foo bar %}", "{% increment 11aa %}", "{% increment %}",
		"{% decrement foo bar %}", "{% decrement 11aa %}", "{% decrement %}",
	} {
		var syntax *liquid.SyntaxError
		if _, err := parse514(t, src, "strict2"); !errors.As(err, &syntax) {
			t.Errorf("%s in strict2: %v, want a syntax error", src, err)
		}
	}
	for _, c := range []struct{ src, mode, want string }{
		{"{% assign my_var = 'hello' %}{{ my_var }}", "strict2", "hello"},
		{"{% assign my-var = 'hello' %}{{ my-var }}", "strict2", "hello"},
		{"{% assign my_var = 'hello' | upcase %}{{ my_var }}", "strict2", "HELLO"},
		{"{% assign (a(b(c) = 1234 %}{{ self['(a(b(c)'] }}", "lax", "1234"},
		{"{% capture my_var %}hello{% endcapture %}{{ my_var }}", "strict2", "hello"},
		{"{% capture my-var %}hello{% endcapture %}{{ my-var }}", "strict2", "hello"},
		{"{% capture 'var' %}test string{% endcapture %}{{var}}", "lax", "test string"},
		// Ruby splits a lookup without regard to quotes, so self['(x[y']
		// reads nothing (differential case tags-27)
		{"{% capture (x[y %}hello{% endcapture %}{{ self['(x[y'] }}", "lax", ""},
		{"{% increment my-var %}", "strict2", "0"},
		{"{% decrement my-var %}", "strict2", "-1"},
	} {
		tmpl, err := parse514(t, c.src, liquid.ErrorMode(c.mode))
		if err != nil {
			t.Errorf("%s in %s: %v", c.src, c.mode, err)
			continue
		}
		if res, err := tmpl.Render(gocontext.Background(), nil); err != nil || res.Output != c.want {
			t.Errorf("%s in %s: %q (%v), want %q", c.src, c.mode, res.Output, err, c.want)
		}
	}
}

// toLiquidItem is an item whose Liquid value is another, as Shopify's
// CustomToLiquidDrop.
type toLiquidItem struct{ value interface{} }

func (d toLiquidItem) ToLiquid() interface{} { return d.value }

// condition_unit_test.rb and case_tag_unit_test.rb: blank and empty as
// Shopify Liquid 5.14 compares them, without ActiveSupport; and
// standard_filter_test.rb: join reads the Liquid value of each item.
func TestBlankAndEmpty(t *testing.T) {
	for _, c := range []struct {
		value        interface{}
		blank, empty bool
	}{
		{"   ", true, false},
		{"", true, true},
		{[]interface{}{}, true, true},
		{map[string]interface{}{}, true, true},
		{nil, true, false},
		// deliberate, for Notifuse's display conditions: false is a value,
		// not blank (Shopify Liquid 5.14 makes it blank; differential case
		// empty-blank-01)
		{false, false, false},
		// deliberate: a string of Unicode whitespace is blank, a no-break
		// space included (Ruby counts ASCII whitespace only)
		{"\u00a0\u2003", true, false},
		{true, false, false},
		{42, false, false},
		{0, false, false},
		{"hello", false, false},
		{[]interface{}{1, 2, 3}, false, false},
		{map[string]interface{}{"a": 1}, false, false},
		{[]string{}, true, true},
	} {
		tmpl, _ := parse514(t, "{% if x == blank %}b{% endif %}{% if x == empty %}e{% endif %}", "lax")
		out, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"x": c.value})
		want := ""
		if c.blank {
			want += "b"
		}
		if c.empty {
			want += "e"
		}
		if err != nil || out.Output != want {
			t.Errorf("%#v: %q (%v), want %q", c.value, out.Output, err, want)
		}
	}

	const caseWhenEmpty = "{%- case x -%}{%- when 2 or empty -%}2 or empty{%- else -%}not 2 or empty{%- endcase -%}"
	for _, mode := range []liquid.ErrorMode{"lax", "strict", "strict2"} {
		tmpl, err := parse514(t, caseWhenEmpty, mode)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range []struct {
			x    interface{}
			want string
		}{
			{2, "2 or empty"}, {map[string]interface{}{}, "2 or empty"}, {[]interface{}{}, "2 or empty"},
			{map[string]interface{}{"a": "b"}, "not 2 or empty"}, {[]interface{}{"a"}, "not 2 or empty"}, {4, "not 2 or empty"},
		} {
			res, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"x": c.x})
			if err != nil || res.Output != c.want {
				t.Errorf("case %#v in %s: %q (%v), want %q", c.x, mode, res.Output, err, c.want)
			}
		}
	}

	tmpl, _ := parse514(t, `{{ items | join: ", " }}`, "lax")
	res, err := tmpl.Render(gocontext.Background(), map[string]interface{}{"items": []interface{}{toLiquidItem{"i did it"}, toLiquidItem{"i did it"}}})
	if err != nil || res.Output != "i did it, i did it" {
		t.Errorf("join of ToLiquid items: %q (%v)", res.Output, err)
	}
}
