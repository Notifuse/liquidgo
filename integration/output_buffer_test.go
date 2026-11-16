package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tag"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// The tags of other packages render as they did before OutputBuffer: a tag
// that embeds Tag or Block and declares Render or RenderToOutputBuffer, one
// that embeds nothing, one that embeds a tag of its own package, and one that
// embeds a standard tag.

type renderOnlyTag struct{ *liquid.Tag }

func (t *renderOnlyTag) Render(liquid.TagContext) string { return "R" }

type emptyRenderTag struct{ *liquid.Tag }

func (t *emptyRenderTag) Render(liquid.TagContext) string { return "" }

type bufferOnlyTag struct{ *liquid.Tag }

func (t *bufferOnlyTag) RenderToOutputBuffer(_ liquid.TagContext, output *string) { *output += "B" }

type blockRenderTag struct{ *liquid.Block }

func (b *blockRenderTag) Render(context liquid.TagContext) string {
	return "[" + b.Block.Render(context) + "]"
}

type blockBufferTag struct{ *liquid.Block }

func (b *blockBufferTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	*output += "<"
	b.Block.RenderToOutputBuffer(context, output)
	*output += ">"
}

type plainBlockTag struct{ *liquid.Block }

type blankRenderBlockTag struct{ *liquid.Block }

func (b *blankRenderBlockTag) Render(liquid.TagContext) string { return "never" }

func (b *blankRenderBlockTag) Blank() bool { return true }

// unembeddedTag embeds nothing.
type unembeddedTag struct{}

func (u *unembeddedTag) RenderToOutputBuffer(_ liquid.TagContext, output *string) { *output += "U" }

func (u *unembeddedTag) Blank() bool { return false }

// nestedBufferTag embeds a tag that declares RenderToOutputBuffer.
type nestedBufferTag struct{ *blockBufferTag }

// nestedRenderTag declares Render on a tag that declares RenderToOutputBuffer.
type nestedRenderTag struct{ *blockBufferTag }

func (n *nestedRenderTag) Render(context liquid.TagContext) string {
	return "{" + n.Block.Render(context) + "}"
}

// wrappedIfTag embeds the standard if tag.
type wrappedIfTag struct{ *tags.IfTag }

// wrappedForTag embeds the standard for tag and declares RenderToOutputBuffer.
type wrappedForTag struct{ *tags.ForTag }

func (w *wrappedForTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	*output += "("
	w.ForTag.RenderToOutputBuffer(context, output)
	*output += ")"
}

// valueBufferTag declares RenderToOutputBuffer on its value receiver.
type valueBufferTag struct{ *liquid.Tag }

func (t valueBufferTag) RenderToOutputBuffer(_ liquid.TagContext, output *string) { *output += "V" }

// valueBlockTag declares RenderToOutputBuffer on its value receiver.
type valueBlockTag struct{ *liquid.Block }

func (b valueBlockTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	*output += "<"
	b.Block.RenderToOutputBuffer(context, output)
	*output += ">"
}

// disableableTag embeds Tag and tag.Disableable, whose methods of the same
// names hide each other.
type disableableTag struct {
	*liquid.Tag
	tag.Disableable
}

func (d *disableableTag) Render(liquid.TagContext) string { return "" }

func tagOf(make func(*liquid.Tag) interface{}) interface{} {
	return tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		return make(liquid.NewTag(name, markup, pc)), nil
	})
}

func blockOf(make func(*liquid.Block) interface{}) interface{} {
	return tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		return make(liquid.NewBlock(name, markup, pc)), nil
	})
}

func customTagEnvironment() *liquid.Environment {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterTag("renderonly", tagOf(func(t *liquid.Tag) interface{} { return &renderOnlyTag{t} }))
	env.RegisterTag("emptyrender", tagOf(func(t *liquid.Tag) interface{} { return &emptyRenderTag{t} }))
	env.RegisterTag("bufferonly", tagOf(func(t *liquid.Tag) interface{} { return &bufferOnlyTag{t} }))
	env.RegisterTag("unembedded", tagOf(func(*liquid.Tag) interface{} { return &unembeddedTag{} }))
	env.RegisterTag("blockrender", blockOf(func(b *liquid.Block) interface{} { return &blockRenderTag{b} }))
	env.RegisterTag("blockbuffer", blockOf(func(b *liquid.Block) interface{} { return &blockBufferTag{b} }))
	env.RegisterTag("plainblock", blockOf(func(b *liquid.Block) interface{} { return &plainBlockTag{b} }))
	env.RegisterTag("blankrender", blockOf(func(b *liquid.Block) interface{} { return &blankRenderBlockTag{b} }))
	env.RegisterTag("nestedbuffer", blockOf(func(b *liquid.Block) interface{} { return &nestedBufferTag{&blockBufferTag{b}} }))
	env.RegisterTag("nestedrender", blockOf(func(b *liquid.Block) interface{} { return &nestedRenderTag{&blockBufferTag{b}} }))
	env.RegisterTag("valuebuffer", tagOf(func(t *liquid.Tag) interface{} { return &valueBufferTag{t} }))
	env.RegisterTag("valuebufferbyvalue", tagOf(func(t *liquid.Tag) interface{} { return valueBufferTag{t} }))
	env.RegisterTag("valueblock", blockOf(func(b *liquid.Block) interface{} { return &valueBlockTag{b} }))
	env.RegisterTag("disableable", tagOf(func(t *liquid.Tag) interface{} { return &disableableTag{Tag: t} }))
	env.RegisterTag("wrappedif", tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		t, err := tags.NewIfTag(name, markup, pc)
		if err != nil {
			return nil, err
		}
		return &wrappedIfTag{t}, nil
	}))
	env.RegisterTag("wrappedfor", tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		t, err := tags.NewForTag(name, markup, pc)
		if err != nil {
			return nil, err
		}
		return &wrappedForTag{t}, nil
	}))
	return env
}

func TestCustomTagsRenderAsBeforeOutputBuffer(t *testing.T) {
	env := customTagEnvironment()
	loop := "{% for i in (1..3) %}{{ i }}{% endfor %}"
	for _, c := range []struct{ source, want string }{
		{"a{% renderonly %}b", "aRb"},
		{"a{% emptyrender %}b", "ab"},
		{"a{% bufferonly %}b", "aBb"},
		{"a{% unembedded %}b", "aUb"},
		// A block whose body holds only tags is blank: its Render is not
		// called.
		{"a{% blockrender %}" + loop + "{% endblockrender %}b", "a123b"},
		{"a{% blockrender %}x{% endblockrender %}b", "a[x]b"},
		{"a{% blockbuffer %}" + loop + "{% endblockbuffer %}b", "a<123>b"},
		{"a{% plainblock %}" + loop + "{% endplainblock %}b", "a123b"},
		{"a{% blankrender %}" + loop + "{% endblankrender %}b", "a123b"},
		{"a{% nestedbuffer %}" + loop + "{% endnestedbuffer %}b", "a<123>b"},
		{"a{% nestedrender %}" + loop + "{% endnestedrender %}b", "a<123>b"},
		{"a{% wrappedif true %}" + loop + "{% else %}no{% endwrappedif %}b", "a123b"},
		{"a{% wrappedfor j in (1..2) %}{{ j }}{% endwrappedfor %}b", "a(12)b"},
		{"{% capture c %}{% blockbuffer %}x{% endblockbuffer %}{% endcapture %}{{ c }}{{ c }}", "<x><x>"},
		{"a{% valuebuffer %}b", "aVb"},
		{"a{% valuebufferbyvalue %}b", "aVb"},
		{"a{% valueblock %}" + loop + "{% endvalueblock %}b", "a<123>b"},
		{"a{% disableable %}b", "ab"},
		{"{% for i in (1..2) %}{% blockbuffer %}{% plainblock %}{{ i }}{% endplainblock %}{% endblockbuffer %}{% endfor %}", "<1><2>"},
	} {
		tmpl, err := liquid.ParseTemplate(c.source, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Errorf("%q: %v", c.source, err)
			continue
		}
		if got := renderRuby(tmpl, nil, nil); got != c.want {
			t.Errorf("%q renders %q, want %q", c.source, got, c.want)
		}
	}
}
