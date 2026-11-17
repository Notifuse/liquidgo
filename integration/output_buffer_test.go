package integration

import (
	"errors"
	"strings"
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

// disableableTag embeds Tag and tag.Disableable, and renders nothing through
// Disableable.Render.
type disableableTag struct {
	*liquid.Tag
	tag.Disableable
}

func (d *disableableTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	d.Render(d.TagName(), context, d.LineNumber(), d.ParseContext(), output, func() {})
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
	return mustEnv(customTagConfig())
}

// customTagConfig is the config of the environment of the custom tags.
func customTagConfig() liquid.EnvironmentConfig {
	envTags := tags.Standard()
	envTags["renderonly"] = tagOf(func(t *liquid.Tag) liquid.Node { return &renderOnlyTag{t} })
	envTags["emptyrender"] = tagOf(func(t *liquid.Tag) liquid.Node { return &emptyRenderTag{t} })
	envTags["bufferonly"] = tagOf(func(t *liquid.Tag) liquid.Node { return &bufferOnlyTag{t} })
	envTags["unembedded"] = tagOf(func(*liquid.Tag) liquid.Node { return &unembeddedTag{} })
	envTags["blockrender"] = blockOf(func(b *liquid.Block) liquid.Node { return &blockRenderTag{b} })
	envTags["blockbuffer"] = blockOf(func(b *liquid.Block) liquid.Node { return &blockBufferTag{b} })
	envTags["plainblock"] = blockOf(func(b *liquid.Block) liquid.Node { return &plainBlockTag{b} })
	envTags["nestedbuffer"] = blockOf(func(b *liquid.Block) liquid.Node { return &nestedBufferTag{&blockBufferTag{b}} })
	envTags["valuebuffer"] = tagOf(func(t *liquid.Tag) liquid.Node { return &valueBufferTag{t} })
	envTags["valuebufferbyvalue"] = tagOf(func(t *liquid.Tag) liquid.Node { return valueBufferTag{t} })
	envTags["valueblock"] = blockOf(func(b *liquid.Block) liquid.Node { return &valueBlockTag{b} })
	envTags["disableable"] = tagOf(func(t *liquid.Tag) liquid.Node { return &disableableTag{Tag: t} })
	envTags["wrappedif"] = func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		t, err := tags.NewIfTag(name, markup, pc)
		if err != nil {
			return nil, err
		}
		return &wrappedIfTag{t}, nil
	}
	envTags["wrappedfor"] = func(name, markup string, pc liquid.ParseContextInterface) (liquid.Node, error) {
		t, err := tags.NewForTag(name, markup, pc)
		if err != nil {
			return nil, err
		}
		return &wrappedForTag{t}, nil
	}
	return liquid.EnvironmentConfig{Tags: envTags}
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

// A constructor that returns neither a tag nor an error gives a syntax error
// that names the tag, and in warn mode a warning, the tag staying as text, as
// a tag whose constructor fails in warn mode.
func TestAConstructorThatGivesNoTagIsASyntaxError(t *testing.T) {
	noTag := func(string, string, liquid.ParseContextInterface) (liquid.Node, error) { return nil, nil }

	envTags := tags.Standard()
	envTags["notag"] = noTag
	env := mustEnv(liquid.EnvironmentConfig{Tags: envTags})
	if _, err := liquid.ParseTemplate("a{% notag %}b", &liquid.TemplateOptions{Environment: env}); err == nil ||
		!strings.Contains(err.Error(), "notag") {
		t.Errorf("parse error %v, want a syntax error that names notag", err)
	}
	// a nil *T in a Node is no tag either
	envTags["typednil"] = func(string, string, liquid.ParseContextInterface) (liquid.Node, error) {
		var tag *renderOnlyTag
		return tag, nil
	}
	env = mustEnv(liquid.EnvironmentConfig{Tags: envTags})
	_, err := liquid.ParseTemplate("a{% typednil %}b", &liquid.TemplateOptions{Environment: env})
	var syntaxErr *liquid.SyntaxError
	if !errors.As(err, &syntaxErr) || !strings.Contains(err.Error(), "typednil") {
		t.Errorf("parse error %v (%T), want a syntax error that names typednil", err, err)
	}

	warnTags := tags.Standard()
	warnTags["notag"] = noTag
	warn := mustEnv(liquid.EnvironmentConfig{Tags: warnTags, ErrorMode: "warn"})
	tmpl, err := liquid.ParseTemplate("a{% notag %}b", &liquid.TemplateOptions{Environment: warn})
	if err != nil {
		t.Fatalf("warn mode: %v", err)
	}
	if ws := tmpl.Warnings(); len(ws) != 1 || !strings.Contains(ws[0].Error(), "notag") {
		t.Errorf("warnings %v, want one that names notag", ws)
	}
	if got, want := renderRuby(tmpl, nil, nil), "a{% notag %}b"; got != want {
		t.Errorf("renders %q, want %q", got, want)
	}
}
