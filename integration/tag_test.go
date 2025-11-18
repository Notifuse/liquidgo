package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// TestTag1 is a simple tag that returns "hello"
type TestTag1 struct {
	*liquid.Tag
}

func NewTestTag1(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	return &TestTag1{Tag: liquid.NewTag(tagName, markup, parseContext)}, nil
}

func (t *TestTag1) Render(ctx liquid.TagContext) string {
	return "hello"
}

// TestTag2 extends TestTag1 and adds prefix/suffix
type TestTag2 struct {
	*TestTag1
}

func NewTestTag2(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	tag1, _ := NewTestTag1(tagName, markup, parseContext)
	return &TestTag2{TestTag1: tag1.(*TestTag1)}, nil
}

func (t *TestTag2) Render(ctx liquid.TagContext) string {
	return "foo" + t.TestTag1.Render(ctx) + "bar"
}

func TestCustomTagsHaveADefaultRenderToOutputBufferMethodForBackwardsCompatibility(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterTag("blabla", tags.TagConstructor(NewTestTag1))

	tmpl, err := liquid.ParseTemplate(`{% blabla %}`, &liquid.TemplateOptions{Environment: env})
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
	// In Go, we can't check object identity like Ruby does in the original test
	// The output and buf are separate string values in Go

	// Test with inheritance
	env2 := liquid.NewEnvironment()
	tags.RegisterStandardTags(env2)
	env2.RegisterTag("blabla", tags.TagConstructor(NewTestTag2))

	tmpl2, err := liquid.ParseTemplate(`{% blabla %}`, &liquid.TemplateOptions{Environment: env2})
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

