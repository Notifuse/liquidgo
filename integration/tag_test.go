package integration

import (
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// TestTag1 is a simple tag that returns "hello"
type TestTag1 struct {
	*liquid.Tag
}

func NewTestTag1(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	return &TestTag1{Tag: liquid.NewTag(tagName, markup, parseContext)}, nil
}

// RenderToOutput is Ruby's render of the test, which returns 'hello': a tag
// has one render method, liquid.Node's.
func (t *TestTag1) RenderToOutput(_ liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("hello")
}

// TestTag2 extends TestTag1 and adds prefix/suffix
type TestTag2 struct {
	*TestTag1
}

func NewTestTag2(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	tag1, _ := NewTestTag1(tagName, markup, parseContext)
	return &TestTag2{TestTag1: tag1.(*TestTag1)}, nil
}

func (t *TestTag2) RenderToOutput(ctx liquid.TagContext, output *liquid.OutputBuffer) {
	output.WriteString("foo")
	t.TestTag1.RenderToOutput(ctx, output)
	output.WriteString("bar")
}

func TestCustomTagsHaveADefaultRenderToOutputBufferMethodForBackwardsCompatibility(t *testing.T) {
	envTags := tags.Standard()
	envTags["blabla"] = tags.TagConstructor(NewTestTag1)
	env := mustEnv(liquid.EnvironmentConfig{Tags: envTags})

	tmpl, err := liquid.ParseTemplate(`{% blabla %}`, &liquid.TemplateOptions{Environment: env})
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
	// In Go, we can't check object identity like Ruby does in the original test
	// The output and buf are separate string values in Go

	// Test with inheritance
	env2Tags := tags.Standard()
	env2Tags["blabla"] = tags.TagConstructor(NewTestTag2)
	env2 := mustEnv(liquid.EnvironmentConfig{Tags: env2Tags})

	tmpl2, err := liquid.ParseTemplate(`{% blabla %}`, &liquid.TemplateOptions{Environment: env2})
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
