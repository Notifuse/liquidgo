package integration

import (
	gocontext "context"
	"runtime"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func renderWithOutputEscape(t *testing.T, on bool, template string, data map[string]interface{}) string {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: on})
	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", template, err)
	}
	return renderRuby(tmpl, data, nil)
}

// TestOutputEscape is the contract of EnvironmentConfig.OutputEscape: with it on,
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
		{"a capture through strip", `{% capture block %} <i>{{ name }}</i> {% endcapture %}{{ block | strip }}`, `<i>&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;</i>`},
		{"a capture through a filter that changes it", `{% capture block %}<i>x</i>{% endcapture %}{{ block | upcase }}`, `&lt;I&gt;X&lt;/I&gt;`},
		{"a capture through a filter that changes it, then raw", `{% capture block %}<i>x</i>{% endcapture %}{{ block | upcase | raw }}`, `<I>X</I>`},
		{"escape then newline_to_br", "{{ '<a>\nb' | escape | newline_to_br }}", "&lt;a&gt;<br />\nb"},
		{"escape then truncate", `{{ name | escape | truncate: 13 }}`, `&lt;b&gt;A...`},
		{"escape then lstrip", `{{ '  <b>' | escape | lstrip }}`, `&lt;b&gt;`},
		{"escape then truncate with a literal ellipsis", `{{ 'Hello world here' | escape | truncate: 13, '<i>…</i>' }}`, `Hello<i>…</i>`},
		{"escape then truncate with an ellipsis from the data", `{{ 'Hello world here' | escape | truncate: 8, name }}`, `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"escape then truncatewords with an ellipsis from the data", `{{ 'Hello world' | escape | truncatewords: 1, name }}`, `Hello&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"escape then append", `{{ name | escape | append: '<br>' }}`, `&amp;lt;b&amp;gt;Ann&amp;lt;/b&amp;gt; &amp;amp; &amp;quot;co&amp;quot; &amp;#39;x&amp;#39;&lt;br&gt;`},
		{"cycle", `{% cycle name, 'y' %}`, `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"cycle of a literal", `{% cycle '<i>', 'y' %}`, `&lt;i&gt;`},
		{"cycle in a liquid tag", "{% liquid\ncycle name, 'y'\n%}", `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
		{"cycle of a capture", `{% capture c %}<i>x</i>{% endcapture %}{% cycle c, 'y' %}`, `<i>x</i>`},
		{"a capture of a cycle", `{% capture c %}{% cycle name %}{% endcapture %}{{ c }}`, `&lt;b&gt;Ann&lt;/b&gt; &amp; &quot;co&quot; &#39;x&#39;`},
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

// TestOutputEscapeIsOffByDefault: without OutputEscape, outputs are written
// as Shopify Liquid writes them, and raw is not a filter.
func TestOutputEscapeIsOffByDefault(t *testing.T) {
	data := map[string]interface{}{"html": `<p>Feed</p>`}
	if got := renderWithOutputEscape(t, false, `{{ html }}`, data); got != `<p>Feed</p>` {
		t.Errorf("got %q", got)
	}
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(`{{ html | raw }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	renderRuby(tmpl, data, &rubyOptions{StrictFilters: true})
	if len(rubyErrors(tmpl)) == 0 {
		t.Error("raw is a filter without the output escape; Shopify Liquid has none")
	}
}

// TestOutputEscapeCountsInTheLimits: the escaped text is what the render length
// limit counts.
func TestOutputEscapeCountsInTheLimits(t *testing.T) {
	limit := 50
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: true, Limits: liquid.ResourceLimitsConfig{RenderLengthLimit: &limit}})
	tmpl, err := liquid.ParseTemplate(`{{ q }}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	// 10 quotes: 10 bytes of data, 60 bytes once escaped.
	renderRuby(tmpl, map[string]interface{}{"q": `""""""""""`}, nil)
	if !rubyLimits(tmpl).Reached() {
		t.Error("the escaped output passes the render length limit without reaching it")
	}
}

// The output escape holds in partials, which share the capture record of the
// render.
func TestOutputEscapeInPartials(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: true})
	partials := recursivePartials{
		"show":    `[{{ name }}]`,
		"capture": `{% capture c %}<u>{{ name }}</u>{% endcapture %}{{ c }}`,
	}
	for template, want := range map[string]string{
		`{% render 'show', name: name %}`:                                     `[&lt;b&gt;]`,
		`{% include 'show' %}`:                                                `[&lt;b&gt;]`,
		`{% render 'capture', name: name %}`:                                  `<u>&lt;b&gt;</u>`,
		`{% capture c %}<u>x</u>{% endcapture %}{% render 'show', name: c %}`: `[<u>x</u>]`,
		`{% include 'capture' %}`:                                             `<u>&lt;b&gt;</u>`,
	} {
		tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		tmpl.Registers()["file_system"] = partials
		if got := renderRuby(tmpl, map[string]interface{}{"name": "<b>"}, nil); got != want {
			t.Errorf("%s: got %q, want %q", template, got, want)
		}
	}
}

// A list is escaped and written item by item, so the render length limit stops
// it before all of it is built: the render allocates about what it does with the
// option off, where escaping the whole list first allocated thousands of times
// more on long lists.
func TestOutputEscapeOfAListStaysWithinTheLimit(t *testing.T) {
	allocated := func(on bool) uint64 {
		limit := 1 << 20
		env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: on, Limits: liquid.ResourceLimitsConfig{RenderLengthLimit: &limit}})
		items := make([]interface{}, 400000)
		for i := range items {
			items[i] = "<b>0123456789</b>"
		}
		tmpl, err := liquid.ParseTemplate(`{{ items }}`, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		renderRuby(tmpl, map[string]interface{}{"items": items}, nil)
		runtime.ReadMemStats(&after)
		if !rubyLimits(tmpl).Reached() {
			t.Fatalf("output escape %v: the list passes the render length limit without reaching it", on)
		}
		return after.TotalAlloc - before.TotalAlloc
	}
	off, on := allocated(false), allocated(true)
	if on > 2*off {
		t.Errorf("the escaped list allocated %d bytes, %d without the output escape", on, off)
	}
}

// A long string is escaped a chunk at a time, so the render length limit stops it
// before all of it is escaped. (Without the option, the limit refuses it before
// any copy.)
func TestOutputEscapeOfALongStringStaysWithinTheLimit(t *testing.T) {
	allocated := func(on bool) uint64 {
		limit := 1 << 20
		env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: on, Limits: liquid.ResourceLimitsConfig{RenderLengthLimit: &limit}})
		long := strings.Repeat("<", 8<<20)
		tmpl, err := liquid.ParseTemplate(`{{ long }}`, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		renderRuby(tmpl, map[string]interface{}{"long": long}, nil)
		runtime.ReadMemStats(&after)
		if !rubyLimits(tmpl).Reached() {
			t.Fatalf("output escape %v: the string passes the render length limit without reaching it", on)
		}
		return after.TotalAlloc - before.TotalAlloc
	}
	// Escaped whole, the 8 MB string would take 32 MB before the limit refuses it;
	// a chunk at a time, about the 1 MB limit's worth of chunks.
	if on := allocated(true); on > 8<<20 {
		t.Errorf("the escaped string allocated %d bytes before the limit stopped it", on)
	}
}

// With the output escape on, a value of the data cannot leave the HTML
// destination it is written in (category 5 of the security check): the text
// of an element, an attribute in double or single quotes, a comment and an
// Outlook conditional comment. Each case closes its destination in the
// value, and the value stays inside.
func TestOutputEscapeKeepsAValueInItsDestination(t *testing.T) {
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), OutputEscape: true})
	data := map[string]interface{}{"v": `x"'></a><img src=x onerror=alert(1)>--><![endif]-->&`}
	escaped := `x&quot;&#39;&gt;&lt;/a&gt;&lt;img src=x onerror=alert(1)&gt;--&gt;&lt;![endif]--&gt;&amp;`
	for _, c := range []struct{ destination, template string }{
		{"element text", `<p>{{ v }}</p>`},
		{"attribute in double quotes", `<a title="{{ v }}">t</a>`},
		{"attribute in single quotes", `<a title='{{ v }}'>t</a>`},
		{"comment", `<!-- {{ v }} -->`},
		{"Outlook conditional comment", `<!--[if mso]>{{ v }}<![endif]-->`},
	} {
		tmpl, err := liquid.ParseTemplate(c.template, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatal(err)
		}
		res, err := tmpl.Render(gocontext.Background(), data)
		want := strings.Replace(c.template, "{{ v }}", escaped, 1)
		if err != nil || res.Output != want {
			t.Errorf("%s: %q (%v), want %q", c.destination, res.Output, err, want)
		}
		for _, closer := range []string{`"`, `'`, `>`, `<`, `-->`} {
			if strings.Count(res.Output, closer) != strings.Count(c.template, closer)-strings.Count("{{ v }}", closer) {
				t.Errorf("%s: the value adds a %s to %q", c.destination, closer, res.Output)
			}
		}
	}
}
