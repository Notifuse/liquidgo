package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestBlockTag is a simple block tag for testing
type TestBlockTag struct {
	*liquid.Block
}

func NewTestBlockTag(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	block := liquid.NewBlock(tagName, markup, parseContext)
	return &TestBlockTag{Block: block}, nil
}

func (t *TestBlockTag) Render(ctx liquid.TagContext) string {
	return "hello"
}

func TestUnexpectedEndTag(t *testing.T) {
	source := `{% if true %}{% endunless %}`
	// Ruby error: "'endunless' is not a valid delimiter for if tags. use endif"
	// Go error format may differ, so match key parts
	assertMatchSyntaxError(t, "endunless.*(not a valid delimiter|invalid delimiter).*if.*endif", source)
}

func TestWithCustomTag(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterTag("testtag", tags.TagConstructor(NewTestBlockTag))

	tmpl, err := liquid.ParseTemplate(`{% testtag %} {% endtesttag %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}
	if tmpl == nil {
		t.Error("Expected template, got nil")
	}
}

func TestCustomBlockTagsHaveADefaultRenderToOutputBufferMethodForBackwardsCompatibility(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterTag("blabla", tags.TagConstructor(NewTestBlockTag))

	tmpl, err := liquid.ParseTemplate(`{% blabla %} bla {% endblabla %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	result := tmpl.Render(nil, &liquid.RenderOptions{})
	if result != "hello" {
		t.Errorf("Expected 'hello', got %q", result)
	}

	// Test with output buffer
	var buf string
	output := tmpl.Render(nil, &liquid.RenderOptions{Output: &buf})
	if output != "hello" {
		t.Errorf("Expected 'hello', got %q", output)
	}
	if buf != "hello" {
		t.Errorf("Expected buf 'hello', got %q", buf)
	}

	// Test with inheritance - TestBlockTag2 extends TestBlockTag1
	env2 := liquid.NewEnvironment()
	tags.RegisterStandardTags(env2)
	env2.RegisterTag("blabla", tags.TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		block1, _ := NewTestBlockTag(tagName, markup, parseContext)
		return &TestBlockTag2{TestBlockTag: block1.(*TestBlockTag)}, nil
	}))

	tmpl2, err := liquid.ParseTemplate(`{% blabla %} foo {% endblabla %}`, &liquid.TemplateOptions{Environment: env2})
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	result = tmpl2.Render(nil, &liquid.RenderOptions{})
	if result != "foohellobar" {
		t.Errorf("Expected 'foohellobar', got %q", result)
	}

	var buf2 string
	output = tmpl2.Render(nil, &liquid.RenderOptions{Output: &buf2})
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

func (t *TestBlockTag2) Render(ctx liquid.TagContext) string {
	// Ruby: 'foo' + super + 'bar' where super calls parent's render
	// Parent (TestBlockTag) returns "hello"
	// So result is "foo" + "hello" + "bar" = "foohellobar"
	return "foo" + t.TestBlockTag.Render(ctx) + "bar"
}
