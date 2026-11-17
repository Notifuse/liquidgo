package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// TestBlockTag is a simple block tag for testing
type TestBlockTag struct {
	*liquid.Block
}

func NewTestBlockTag(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)
	return &TestBlockTag{Block: block}, nil
}

// RenderToOutput is Ruby's render of the test, which returns 'hello': a tag
// has one render method, liquid.Node's.
func (t *TestBlockTag) RenderToOutput(_ liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("hello")
}

func TestUnexpectedEndTag(t *testing.T) {
	source := `{% if true %}{% endunless %}`
	// Ruby error: "'endunless' is not a valid delimiter for if tags. use endif"
	// Go error format may differ, so match key parts
	assertMatchSyntaxError(t, "endunless.*(not a valid delimiter|invalid delimiter).*if.*endif", source)
}

func TestWithCustomTag(t *testing.T) {
	envTags := tags.Standard()
	envTags["testtag"] = tags.TagConstructor(NewTestBlockTag)
	env := mustEnv(liquid.EnvironmentConfig{Tags: envTags})

	tmpl, err := liquid.ParseTemplate(`{% testtag %} {% endtesttag %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	if tmpl == nil {
		t.Error("Expected template, got nil")
	}
}

func TestCustomBlockTagsHaveADefaultRenderToOutputBufferMethodForBackwardsCompatibility(t *testing.T) {
	envTags := tags.Standard()
	envTags["blabla"] = tags.TagConstructor(NewTestBlockTag)
	env := mustEnv(liquid.EnvironmentConfig{Tags: envTags})

	tmpl, err := liquid.ParseTemplate(`{% blabla %} bla {% endblabla %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	result := renderRuby(tmpl, nil, &rubyOptions{})
	if result != "hello" {
		t.Errorf("Expected 'hello', got %q", result)
	}

	// Test with output buffer
	var buf string
	output := renderRuby(tmpl, nil, &rubyOptions{Output: &buf})
	if output != "hello" {
		t.Errorf("Expected 'hello', got %q", output)
	}
	if buf != "hello" {
		t.Errorf("Expected buf 'hello', got %q", buf)
	}

	// Test with inheritance - TestBlockTag2 extends TestBlockTag1
	env2Tags := tags.Standard()
	env2Tags["blabla"] = tags.TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		block1, _ := NewTestBlockTag(tagName, markup, parseContext)
		return &TestBlockTag2{TestBlockTag: block1.(*TestBlockTag)}, nil
	})
	env2 := mustEnv(liquid.EnvironmentConfig{Tags: env2Tags})

	tmpl2, err := liquid.ParseTemplate(`{% blabla %} foo {% endblabla %}`, &liquid.TemplateOptions{Environment: env2})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	result = renderRuby(tmpl2, nil, &rubyOptions{})
	if result != "foohellobar" {
		t.Errorf("Expected 'foohellobar', got %q", result)
	}

	var buf2 string
	output = renderRuby(tmpl2, nil, &rubyOptions{Output: &buf2})
	if output != "foohellobar" {
		t.Errorf("Expected 'foohellobar', got %q", output)
	}
	if buf2 != "foohellobar" {
		t.Errorf("Expected buf 'foohellobar', got %q", buf2)
	}
}

// TestBlockTag2 extends TestBlockTag
type TestBlockTag2 struct {
	*TestBlockTag
}

func (t *TestBlockTag2) RenderToOutput(ctx liquid.TagContext, output *liquid.OutputBuffer) {
	// Ruby: 'foo' + super + 'bar' where super calls parent's render
	// Parent (TestBlockTag) returns "hello"
	// So result is "foo" + "hello" + "bar" = "foohellobar"
	output.WriteString("foo")
	t.TestBlockTag.RenderToOutput(ctx, output)
	output.WriteString("bar")
}
