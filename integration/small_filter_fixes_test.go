package integration

import "testing"

func TestSmallFilterFixes(t *testing.T) {
	assigns := map[string]interface{}{
		"list":   []interface{}{1, 2, 3},
		"codes":  []interface{}{"10", "2", "1"},
		"h":      map[string]interface{}{"a": 1, "b": 2},
		"hs":     map[string]interface{}{"size": "big"},
		"typed":  map[string]int{"a": 1},
		"quoted": `"q" 'a'`,
	}
	tests := []struct {
		template string
		expected string
	}{
		{`{{ list | join }}`, "1 2 3"},
		{`{{ list | join: ", " }}`, "1, 2, 3"},
		{`{{ "&amp; &" | escape_once }}`, "&amp; &amp;"},
		{`{{ "&lt;b&gt; <i> &#39; &#x27; &a1;" | escape_once }}`, "&lt;b&gt; &lt;i&gt; &#39; &amp;#x27; &amp;a1;"},
		{`{{ quoted | escape_once }}`, "&quot;q&quot; &#39;a&#39;"},
		{"{{ \"<div\nclass='x'>test</div>\" | strip_html }}", "test"},
		{`{{ codes | sort | join: "," }}`, "1,10,2"},
		{`{{ list | sort | join: "," }}`, "1,2,3"},
		{`{{ h.size }}`, "2"},
		{`{{ hs.size }}`, "big"},
		{`{{ typed.size }}`, "1"},
		{`{% if h.size == 2 %}two{% endif %}`, "two"},
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, assigns)
		})
	}
}
