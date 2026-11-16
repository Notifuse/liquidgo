package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tag"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// The tags of other packages render through liquid.Node: a tag that embeds
// Tag or Block and declares RenderToOutput, one that embeds nothing, one that
// embeds a tag of its own package, one that embeds a standard tag, and ones
// that declare RenderToOutput on a value receiver.

type renderOnlyTag struct{ *liquid.Tag }

func (t *renderOnlyTag) RenderToOutput(_ liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("R")
}

type emptyRenderTag struct{ *liquid.Tag }

func (t *emptyRenderTag) RenderToOutput(liquid.TagContext, *liquid.OutputBuffer) {}

type bufferOnlyTag struct{ *liquid.Tag }

func (t *bufferOnlyTag) RenderToOutput(_ liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("B")
}

type blockRenderTag struct{ *liquid.Block }

func (b *blockRenderTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("[")
	b.Block.RenderToOutput(context, output)
	output.WriteString("]")
}

type blockBufferTag struct{ *liquid.Block }

func (b *blockBufferTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("<")
	b.Block.RenderToOutput(context, output)
	output.WriteString(">")
}

type plainBlockTag struct{ *liquid.Block }

// unembeddedTag embeds nothing.
type unembeddedTag struct{}

func (u *unembeddedTag) RenderToOutput(_ liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("U")
}

func (u *unembeddedTag) Blank() bool { return false }

// nestedBufferTag embeds a tag that declares RenderToOutput.
type nestedBufferTag struct{ *blockBufferTag }

// wrappedIfTag embeds the standard if tag.
type wrappedIfTag struct{ *tags.IfTag }

// wrappedForTag embeds the standard for tag and declares RenderToOutput.
type wrappedForTag struct{ *tags.ForTag }

func (w *wrappedForTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("(")
	w.ForTag.RenderToOutput(context, output)
	output.WriteString(")")
}

// valueBufferTag declares RenderToOutput on its value receiver.
type valueBufferTag struct{ *liquid.Tag }

func (t valueBufferTag) RenderToOutput(_ liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("V")
}

// valueBlockTag declares RenderToOutput on its value receiver.
type valueBlockTag struct{ *liquid.Block }

func (b valueBlockTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("<")
	b.Block.RenderToOutput(context, output)
	output.WriteString(">")
}

// disableableTag embeds Tag and tag.Disableable, whose Render does not hide
// the RenderToOutput of Tag.
type disableableTag struct {
	*liquid.Tag
	tag.Disableable
}

func tagOf(make func(*liquid.Tag) liquid.Node) liquid.TagConstructor {
	return func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return make(liquid.NewTag(name, markup, pc)), nil
	}
}

func blockOf(make func(*liquid.Block) liquid.Node) liquid.TagConstructor {
	return func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		return make(liquid.NewBlock(name, markup, pc)), nil
	}
}

func customTagEnvironment() *liquid.Environment {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterTag("renderonly", tagOf(func(t *liquid.Tag) liquid.Node { return &renderOnlyTag{t} }))
	env.RegisterTag("emptyrender", tagOf(func(t *liquid.Tag) liquid.Node { return &emptyRenderTag{t} }))
	env.RegisterTag("bufferonly", tagOf(func(t *liquid.Tag) liquid.Node { return &bufferOnlyTag{t} }))
	env.RegisterTag("unembedded", tagOf(func(*liquid.Tag) liquid.Node { return &unembeddedTag{} }))
	env.RegisterTag("blockrender", blockOf(func(b *liquid.Block) liquid.Node { return &blockRenderTag{b} }))
	env.RegisterTag("blockbuffer", blockOf(func(b *liquid.Block) liquid.Node { return &blockBufferTag{b} }))
	env.RegisterTag("plainblock", blockOf(func(b *liquid.Block) liquid.Node { return &plainBlockTag{b} }))
	env.RegisterTag("nestedbuffer", blockOf(func(b *liquid.Block) liquid.Node { return &nestedBufferTag{&blockBufferTag{b}} }))
	env.RegisterTag("valuebuffer", tagOf(func(t *liquid.Tag) liquid.Node { return &valueBufferTag{t} }))
	env.RegisterTag("valuebufferbyvalue", tagOf(func(t *liquid.Tag) liquid.Node { return valueBufferTag{t} }))
	env.RegisterTag("valueblock", blockOf(func(b *liquid.Block) liquid.Node { return &valueBlockTag{b} }))
	env.RegisterTag("disableable", tagOf(func(t *liquid.Tag) liquid.Node { return &disableableTag{Tag: t} }))
	env.RegisterTag("wrappedif", func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		t, err := tags.NewIfTag(name, markup, pc)
		if err != nil {
			return nil, err
		}
		return &wrappedIfTag{t}, nil
	})
	env.RegisterTag("wrappedfor", func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		t, err := tags.NewForTag(name, markup, pc)
		if err != nil {
			return nil, err
		}
		return &wrappedForTag{t}, nil
	})
	return env
}

func TestCustomTagsRenderThroughNode(t *testing.T) {
	env := customTagEnvironment()
	loop := "{% for i in (1..3) %}{{ i }}{% endfor %}"
	for _, c := range []struct{ source, want string }{
		{"a{% renderonly %}b", "aRb"},
		{"a{% emptyrender %}b", "ab"},
		{"a{% bufferonly %}b", "aBb"},
		{"a{% unembedded %}b", "aUb"},
		{"a{% blockrender %}x{% endblockrender %}b", "a[x]b"},
		{"a{% blockbuffer %}" + loop + "{% endblockbuffer %}b", "a<123>b"},
		{"a{% plainblock %}" + loop + "{% endplainblock %}b", "a123b"},
		{"a{% nestedbuffer %}" + loop + "{% endnestedbuffer %}b", "a<123>b"},
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
