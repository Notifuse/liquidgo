package integration

import (
	"reflect"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// Ported from test/unit/parse_tree_visitor_test.rb: the visitor walks the
// nodes and expressions of a parse tree, and a callback on VariableLookup
// lists the variables a template reads.

// variableTraversal is Ruby's traversal helper: a visitor of the root of
// source whose callback gives the name of each variable lookup.
func variableTraversal(t *testing.T, source string) *liquid.ParseTreeVisitor {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", source, err)
	}
	return liquid.ForParseTreeVisitor(tmpl.Root(), nil).AddCallbackFor([]interface{}{(*liquid.VariableLookup)(nil)},
		func(node, context interface{}) (interface{}, interface{}) {
			return node.(*liquid.VariableLookup).Name(), context
		})
}

// visitVariables is Ruby's visit helper: the result, flattened, without nil.
func visitVariables(t *testing.T, source string) []interface{} {
	t.Helper()
	return flattenCompact(variableTraversal(t, source).Visit(nil))
}

func flattenCompact(v interface{}) []interface{} {
	var out []interface{}
	switch x := v.(type) {
	case nil:
	case []interface{}:
		for _, e := range x {
			out = append(out, flattenCompact(e)...)
		}
	default:
		out = append(out, x)
	}
	return out
}

func TestParseTreeVisitorListsTheVariables(t *testing.T) {
	for _, c := range []struct {
		name, source string
		want         []interface{}
	}{
		{"test_variable", "{{ test }}", []interface{}{"test"}},
		{"test_varible_with_filter", "{{ test | split: infilter }}", []interface{}{"test", "infilter"}},
		{"test_dynamic_variable", "{{ test[inlookup] }}", []interface{}{"test", "inlookup"}},
		{"test_echo", "{% echo test %}", []interface{}{"test"}},
		{"test_if_condition", "{% if test %}{% endif %}", []interface{}{"test"}},
		{"test_complex_if_condition", "{% if 1 == 1 and 2 == test %}{% endif %}", []interface{}{"test"}},
		{"test_if_body", "{% if 1 == 1 %}{{ test }}{% endif %}", []interface{}{"test"}},
		{"test_unless_condition", "{% unless test %}{% endunless %}", []interface{}{"test"}},
		{"test_complex_unless_condition", "{% unless 1 == 1 and 2 == test %}{% endunless %}", []interface{}{"test"}},
		{"test_unless_body", "{% unless 1 == 1 %}{{ test }}{% endunless %}", []interface{}{"test"}},
		{"test_elsif_condition", "{% if 1 == 1 %}{% elsif test %}{% endif %}", []interface{}{"test"}},
		{"test_complex_elsif_condition", "{% if 1 == 1 %}{% elsif 1 == 1 and 2 == test %}{% endif %}", []interface{}{"test"}},
		{"test_elsif_body", "{% if 1 == 1 %}{% elsif 2 == 2 %}{{ test }}{% endif %}", []interface{}{"test"}},
		{"test_else_body", "{% if 1 == 1 %}{% else %}{{ test }}{% endif %}", []interface{}{"test"}},
		{"test_case_left", "{% case test %}{% endcase %}", []interface{}{"test"}},
		{"test_case_condition", "{% case 1 %}{% when test %}{% endcase %}", []interface{}{"test"}},
		{"test_case_when_body", "{% case 1 %}{% when 2 %}{{ test }}{% endcase %}", []interface{}{"test"}},
		{"test_case_else_body", "{% case 1 %}{% else %}{{ test }}{% endcase %}", []interface{}{"test"}},
		{"test_for_in", "{% for x in test %}{% endfor %}", []interface{}{"test"}},
		{"test_for_limit", "{% for x in (1..5) limit: test %}{% endfor %}", []interface{}{"test"}},
		{"test_for_offset", "{% for x in (1..5) offset: test %}{% endfor %}", []interface{}{"test"}},
		{"test_for_body", "{% for x in (1..5) %}{{ test }}{% endfor %}", []interface{}{"test"}},
		{"test_for_range", "{% for x in (1..test) %}{% endfor %}", []interface{}{"test"}},
		{"test_tablerow_in", "{% tablerow x in test %}{% endtablerow %}", []interface{}{"test"}},
		{"test_tablerow_limit", "{% tablerow x in (1..5) limit: test %}{% endtablerow %}", []interface{}{"test"}},
		{"test_tablerow_offset", "{% tablerow x in (1..5) offset: test %}{% endtablerow %}", []interface{}{"test"}},
		{"test_tablerow_body", "{% tablerow x in (1..5) %}{{ test }}{% endtablerow %}", []interface{}{"test"}},
		{"test_cycle", "{% cycle test %}", []interface{}{"test"}},
		{"test_assign", "{% assign x = test %}", []interface{}{"test"}},
		{"test_capture", "{% capture x %}{{ test }}{% endcapture %}", []interface{}{"test"}},
		{"test_include", "{% include test %}", []interface{}{"test"}},
		{"test_include_with", "{% include \"hai\" with test %}", []interface{}{"test"}},
		{"test_include_for", "{% include \"hai\" for test %}", []interface{}{"test"}},
		{"test_render_with", "{% render \"hai\" with test %}", []interface{}{"test"}},
		{"test_render_for", "{% render \"hai\" for test %}", []interface{}{"test"}},
	} {
		if got := visitVariables(t, c.source); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %q visits %v, want %v", c.name, c.source, got, c.want)
		}
	}
}

// Ported from test_preserve_tree_structure.
func TestParseTreeVisitorPreservesTheTreeStructure(t *testing.T) {
	got := variableTraversal(t, "{% for x in xs offset: test %}{{ other }}{% endfor %}").Visit(nil)
	leaf := func(name interface{}) []interface{} { return []interface{}{name, []interface{}{}} }
	want := []interface{}{[]interface{}{nil, []interface{}{
		[]interface{}{nil, []interface{}{[]interface{}{nil, []interface{}{leaf("other")}}}},
		leaf("test"),
		leaf("xs"),
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("visit gives %#v, want %#v", got, want)
	}
}
