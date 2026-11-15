package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func renderWithOutputEscape(t *testing.T, on bool, template string, data map[string]interface{}) string {
	t.Helper()
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.SetOutputEscape(on)
	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", template, err)
	}
	return tmpl.Render(data, nil)
}

// TestOutputEscape is the contract of Environment.SetOutputEscape: with it on,
// a value of the render data cannot add markup to the output; raw writes HTML
// on purpose; an output escaped by its own last filter is not escaped twice;
// the text of a capture is the template's own markup; and assign, conditions
// and filters see the data as it is.
func TestOutputEscape(t *testing.T) {
	data := map[string]interface{}{
		"name": `<b>Ann</b> & "co" 'x'`,
		"html": `<p>Feed</p>`,
		"list": []interface{}{"<i>", "&"},
	}
	cases := []struct{ name, template, want string }{
		{"an output", `{{ name }}`, `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"echo", `{% echo name %}`, `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"after a filter", `{{ name | upcase }}`, `&lt;B&gt;ANN&lt;/B&gt; &amp; &quot;CO&quot; &#39;X&#39;`},
		{"a list", `{{ list }}`, `&lt;i&gt;&amp;`},
		{"raw", `{{ html | raw }}`, `<p>Feed</p>`},
		{"raw before another filter", `{{ html | raw | upcase }}`, `&lt;P&gt;FEED&lt;/P&gt;`},
		{"escape is not escaped twice", `{{ name | escape }}`, `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"escape_once", `{{ '&amp; <' | escape_once }}`, `&amp; &lt;`},
		{"h", `{{ html | h }}`, `&lt;p&gt;Feed&lt;/p&gt;`},
		{"text of the template", `<p>{{ 'x' }}</p>`, `<p>x</p>`},
		{"a string literal", `{{ '<br>' }}`, `&lt;br&gt;`},
		{"a capture", `{% capture block %}<i>{{ name }}</i>{% endcapture %}{{ block }}`, `<i>&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;</i>`},
		{"a capture through an assign", `{% capture block %}<i>x</i>{% endcapture %}{% assign copy = block %}{{ copy }}`, `<i>x</i>`},
		{"a capture changed by a filter", `{% capture block %} <i>x</i> {% endcapture %}{{ block | strip }}`, `&lt;i&gt;x&lt;/i&gt;`},
		{"a capture changed by a filter, then raw", `{% capture block %} <i>x</i> {% endcapture %}{{ block | strip | raw }}`, `<i>x</i>`},
		{"raw in an assign", `{% assign v = html | raw %}{{ v }}`, `&lt;p&gt;Feed&lt;/p&gt;`},
		{"conditions see the data", `{% if html == '<p>Feed</p>' %}same{% endif %}`, `same`},
		{"url_encode sees the data", `{{ html | url_encode }}`, `%3Cp%3EFeed%3C%2Fp%3E`},
		{"size sees the data", `{{ html | size }}`, `11`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := renderWithOutputEscape(t, true, tc.template, data); got != tc.want {
				t.Errorf("%s\n got: %s\nwant: %s", tc.template, got, tc.want)
			}
		})
	}
}

// TestOutputEscapeIsOffByDefault: without SetOutputEscape, outputs are written
// as Shopify Liquid writes them, and raw is not a filter.
func TestOutputEscapeIsOffByDefault(t *testing.T) {
	data := map[string]interface{}{"html": `<p>Feed</p>`}
	if got := renderWithOutputEscape(t, false, `{{ html }}`, data); got != `<p>Feed</p>` {
		t.Errorf("got %q", got)
	}
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tmpl, err := liquid.ParseTemplate(`{{ html | raw }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	tmpl.Render(data, &liquid.RenderOptions{StrictFilters: true})
	if len(tmpl.Errors()) == 0 {
		t.Error("raw is a filter without the output escape; Shopify Liquid has none")
	}
}

// TestOutputEscapeCountsInTheLimits: the escaped text is what the render length
// limit counts.
func TestOutputEscapeCountsInTheLimits(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.SetOutputEscape(true)
	env.SetDefaultResourceLimits(map[string]interface{}{"render_length_limit": 50})
	tmpl, err := liquid.ParseTemplate(`{{ q }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	// 10 quotes: 10 bytes of data, 60 bytes once escaped.
	tmpl.Render(map[string]interface{}{"q": `""""""""""`}, nil)
	if !tmpl.ResourceLimits().Reached() {
		t.Error("the escaped output passes the render length limit without reaching it")
	}
}
