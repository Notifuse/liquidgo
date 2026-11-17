package tag

import (
	gocontext "context"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// CustomTag is a simple tag that renders its tag name.
// It embeds Disableable to check if it's disabled before rendering.
type CustomTag struct {
	*liquid.Tag
	Disableable
}

func NewCustomTag(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	return &CustomTag{
		Tag:         liquid.NewTag(tagName, markup, parseContext),
		Disableable: Disableable{},
	}, nil
}

func (c *CustomTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	c.Render(
		c.TagName(),
		context,
		c.LineNumber(),
		c.ParseContext(),
		output,
		func() {
			// Render tag name
			output.WriteString(c.TagName())
		},
	)
}

// Custom2Tag is another simple tag that renders its tag name.
// It embeds Disableable to check if it's disabled before rendering.
type Custom2Tag struct {
	*liquid.Tag
	Disableable
}

func NewCustom2Tag(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	return &Custom2Tag{
		Tag:         liquid.NewTag(tagName, markup, parseContext),
		Disableable: Disableable{},
	}, nil
}

func (c *Custom2Tag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	c.Render(
		c.TagName(),
		context,
		c.LineNumber(),
		c.ParseContext(),
		output,
		func() {
			// Render tag name
			output.WriteString(c.TagName())
		},
	)
}

// DisableCustomBlock is a block tag that disables "custom" tag.
// It embeds Disabler to wrap rendering with disabled tags.
type DisableCustomBlock struct {
	*liquid.Block
	Disabler
}

func NewDisableCustomBlock(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	return &DisableCustomBlock{
		Block:    liquid.NewBlock(tagName, markup, parseContext),
		Disabler: Disabler{},
	}, nil
}

func (d *DisableCustomBlock) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	disabledTags := []string{"custom"}
	d.Render(
		disabledTags,
		context,
		func() {
			// Render block body
			if d.Block != nil {
				d.Block.RenderToOutput(context, output)
			}
		},
	)
}

// DisableBothBlock is a block tag that disables both "custom" and "custom2" tags.
// It embeds Disabler to wrap rendering with disabled tags.
type DisableBothBlock struct {
	*liquid.Block
	Disabler
}

func NewDisableBothBlock(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
	return &DisableBothBlock{
		Block:    liquid.NewBlock(tagName, markup, parseContext),
		Disabler: Disabler{},
	}, nil
}

func (d *DisableBothBlock) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
	disabledTags := []string{"custom", "custom2"}
	d.Render(
		disabledTags,
		context,
		func() {
			// Render block body
			if d.Block != nil {
				d.Block.RenderToOutput(context, output)
			}
		},
	)
}

func TestBlockTagDisablingNestedTag(t *testing.T) {
	// Create environment

	// Register custom tags
	envTags := tags.Standard()
	envTags["custom"] = tags.TagConstructor(NewCustomTag)
	envTags["custom2"] = tags.TagConstructor(NewCustom2Tag)
	envTags["disable"] = tags.TagConstructor(NewDisableCustomBlock)
	env := mustEnv(liquid.EnvironmentConfig{Tags: envTags})

	// Parse template
	templateOptions := &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	}

	tmpl, err := liquid.ParseTemplate(`{% disable %}{% custom %};{% custom2 %}{% enddisable %}`, templateOptions)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// Render template
	res, _ := tmpl.Render(gocontext.Background(), nil)
	output := res.Output

	// Expected: "Liquid error (line 1): custom usage is not allowed in this context;custom2"
	expected := "Liquid error (line 1): custom usage is not allowed in this context;custom2"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestBlockTagDisablingMultipleNestedTags(t *testing.T) {
	// Create environment

	// Register custom tags
	envTags := tags.Standard()
	envTags["custom"] = tags.TagConstructor(NewCustomTag)
	envTags["custom2"] = tags.TagConstructor(NewCustom2Tag)
	envTags["disable"] = tags.TagConstructor(NewDisableBothBlock)
	env := mustEnv(liquid.EnvironmentConfig{Tags: envTags})

	// Parse template
	templateOptions := &liquid.TemplateOptions{
		Environment: env,
		LineNumbers: true,
	}

	tmpl, err := liquid.ParseTemplate(`{% disable %}{% custom %};{% custom2 %}{% enddisable %}`, templateOptions)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	// Render template
	res, _ := tmpl.Render(gocontext.Background(), nil)
	output := res.Output

	// Expected: "Liquid error (line 1): custom usage is not allowed in this context;Liquid error (line 1): custom2 usage is not allowed in this context"
	expected := "Liquid error (line 1): custom usage is not allowed in this context;Liquid error (line 1): custom2 usage is not allowed in this context"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
