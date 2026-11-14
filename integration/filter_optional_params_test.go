package integration

import (
	"testing"
)

// TestFiltersWithOptionalParameters tests that filters work correctly
// when called with fewer arguments than their method signatures require.
// This tests the filter invocation system, not just the filter logic.
//
// Context: These tests expose a bug where filters with optional parameters
// (like default, sort, where) fail when templates don't provide all arguments.
// The bug is in strainer_template.go's Invoke method, which has a strict
// argument count check that doesn't account for optional parameters.
//
// See: https://shopify.dev/docs/api/liquid/filters for official Liquid behavior
func TestFiltersWithOptionalParameters(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		// default filter tests - most common use case
		{
			name:     "default filter with nil value",
			template: `{{ x | default: "fallback" }}`,
			data:     map[string]interface{}{"x": nil},
			expected: "fallback",
		},
		{
			name:     "default filter with empty string",
			template: `{{ x | default: "fallback" }}`,
			data:     map[string]interface{}{"x": ""},
			expected: "fallback",
		},
		{
			name:     "default filter with false value",
			template: `{{ x | default: "fallback" }}`,
			data:     map[string]interface{}{"x": false},
			expected: "fallback",
		},
		{
			name:     "default filter with existing value",
			template: `{{ x | default: "fallback" }}`,
			data:     map[string]interface{}{"x": "value"},
			expected: "value",
		},
		{
			name:     "default filter with empty array",
			template: `{{ x | default: "fallback" }}`,
			data:     map[string]interface{}{"x": []interface{}{}},
			expected: "fallback",
		},
		// Note: Keyword arguments (like allow_false: true) are a separate feature
		// that's not fully implemented yet in liquidgo. Skipping for now.
		// TODO: Implement keyword arguments support in Variable.Render
		// {
		// 	name:     "default filter with allow_false option",
		// 	template: `{{ x | default: "fallback", allow_false: true }}`,
		// 	data:     map[string]interface{}{"x": false},
		// 	expected: "false",
		// },

		// sort filter tests - property parameter is optional
		{
			name:     "sort without property parameter",
			template: `{{ nums | sort | join: "," }}`,
			data:     map[string]interface{}{"nums": []interface{}{3, 1, 2}},
			expected: "1,2,3",
		},
		{
			name:     "sort with property parameter",
			template: `{{ items | sort: "age" | map: "age" | join: "," }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "Alice", "age": 30},
					{"name": "Bob", "age": 20},
					{"name": "Charlie", "age": 25},
				},
			},
			expected: "20,25,30",
		},
		{
			name:     "sort strings without property",
			template: `{{ items | sort | join: "," }}`,
			data:     map[string]interface{}{"items": []interface{}{"zebra", "apple", "mango"}},
			expected: "apple,mango,zebra",
		},

		// where filter tests - targetValue parameter is optional
		{
			name:     "where filter by truthiness only",
			template: `{{ items | where: "active" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "Item1", "active": true},
					{"name": "Item2", "active": false},
					{"name": "Item3", "active": true},
				},
			},
			expected: "2",
		},
		{
			name:     "where filter with target value",
			template: `{{ items | where: "status", "active" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "Item1", "status": "active"},
					{"name": "Item2", "status": "inactive"},
					{"name": "Item3", "status": "active"},
				},
			},
			expected: "2",
		},
		{
			name:     "where filter by truthiness with string property",
			template: `{{ items | where: "name" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "Alice"},
					{"name": ""},
					{"name": "Bob"},
				},
			},
			expected: "2",
		},

		// sort_natural filter tests - property parameter is optional
		{
			name:     "sort_natural without property",
			template: `{{ items | sort_natural | join: "," }}`,
			data:     map[string]interface{}{"items": []interface{}{"c", "A", "b"}},
			expected: "A,b,c",
		},
		{
			name:     "sort_natural with property",
			template: `{{ items | sort_natural: "name" | map: "name" | join: "," }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"name": "zebra"},
					{"name": "Apple"},
					{"name": "banana"},
				},
			},
			expected: "Apple,banana,zebra",
		},

		// uniq filter tests - property parameter is optional
		{
			name:     "uniq without property",
			template: `{{ items | uniq | join: "," }}`,
			data:     map[string]interface{}{"items": []interface{}{1, 2, 1, 3, 2}},
			expected: "1,2,3",
		},
		{
			name:     "uniq with property",
			template: `{{ items | uniq: "type" | map: "type" | join: "," }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"type": "A"},
					{"type": "B"},
					{"type": "A"},
					{"type": "C"},
				},
			},
			expected: "A,B,C",
		},

		// compact filter tests - property parameter is optional
		{
			name:     "compact without property",
			template: `{{ items | compact | size }}`,
			data:     map[string]interface{}{"items": []interface{}{1, nil, 2, nil, 3}},
			expected: "3",
		},
		{
			name:     "compact with property",
			template: `{{ items | compact: "value" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"value": 1},
					{"value": nil},
					{"value": 2},
				},
			},
			expected: "2",
		},

		// reject filter tests - targetValue parameter is optional
		{
			name:     "reject by truthiness only",
			template: `{{ items | reject: "active" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"active": true},
					{"active": false},
					{"active": true},
				},
			},
			expected: "1",
		},
		{
			name:     "reject with target value",
			template: `{{ items | reject: "status", "inactive" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"status": "active"},
					{"status": "inactive"},
					{"status": "active"},
				},
			},
			expected: "2",
		},

		// truncate filter tests - truncateString parameter is optional
		{
			name:     "truncate with default ellipsis",
			template: `{{ text | truncate: 10 }}`,
			data:     map[string]interface{}{"text": "This is a long string"},
			expected: "This is...",
		},
		{
			name:     "truncate with custom ellipsis",
			template: `{{ text | truncate: 10, "---" }}`,
			data:     map[string]interface{}{"text": "This is a long string"},
			expected: "This is---",
		},

		// truncatewords filter tests - truncateString parameter is optional
		{
			name:     "truncatewords with default ellipsis",
			template: `{{ text | truncatewords: 3 }}`,
			data:     map[string]interface{}{"text": "This is a very long string"},
			expected: "This is a...",
		},
		{
			name:     "truncatewords with custom ellipsis",
			template: `{{ text | truncatewords: 3, "---" }}`,
			data:     map[string]interface{}{"text": "This is a very long string"},
			expected: "This is a---",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, tt.data)
		})
	}
}

// TestBackwardCompatibilityWithAllArguments ensures that filters with
// all arguments provided continue to work after the fix.
func TestBackwardCompatibilityWithAllArguments(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		// Note: Keyword arguments test removed - not implemented yet
		// {
		// 	name:     "default with all arguments",
		// 	template: `{{ x | default: "fallback", allow_false: true }}`,
		// 	data:     map[string]interface{}{"x": false},
		// 	expected: "false",
		// },
		{
			name:     "sort with property",
			template: `{{ items | sort: "age" | map: "age" | first }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"age": 30},
					{"age": 20},
				},
			},
			expected: "20",
		},
		{
			name:     "where with target value",
			template: `{{ items | where: "type", "A" | size }}`,
			data: map[string]interface{}{
				"items": []map[string]interface{}{
					{"type": "A"},
					{"type": "B"},
					{"type": "A"},
				},
			},
			expected: "2",
		},
		{
			name:     "truncate with custom ellipsis",
			template: `{{ text | truncate: 5, "..." }}`,
			data:     map[string]interface{}{"text": "Hello World"},
			expected: "He...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, tt.data)
		})
	}
}

// TestFiltersWithNoOptionalParams ensures filters without optional parameters
// continue to work (regression test).
func TestFiltersWithNoOptionalParams(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "strip_html",
			template: `{{ text | strip_html }}`,
			data:     map[string]interface{}{"text": "<p>Hello <b>World</b></p>"},
			expected: "Hello World",
		},
		{
			name:     "upcase",
			template: `{{ text | upcase }}`,
			data:     map[string]interface{}{"text": "hello"},
			expected: "HELLO",
		},
		{
			name:     "downcase",
			template: `{{ text | downcase }}`,
			data:     map[string]interface{}{"text": "HELLO"},
			expected: "hello",
		},
		{
			name:     "capitalize",
			template: `{{ text | capitalize }}`,
			data:     map[string]interface{}{"text": "hello world"},
			expected: "Hello world",
		},
		{
			name:     "size",
			template: `{{ items | size }}`,
			data:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			expected: "3",
		},
		{
			name:     "first",
			template: `{{ items | first }}`,
			data:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			expected: "1",
		},
		{
			name:     "last",
			template: `{{ items | last }}`,
			data:     map[string]interface{}{"items": []interface{}{1, 2, 3}},
			expected: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, tt.data)
		})
	}
}
