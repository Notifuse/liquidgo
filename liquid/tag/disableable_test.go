package tag

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// CustomTag is a simple tag that renders its tag name.
// It embeds Disableable to check if it's disabled before rendering.
type CustomTag struct {
	*liquid.Tag
	Disableable
}

func NewCustomTag(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	return &CustomTag{
		Tag:         liquid.NewTag(tagName, markup, parseContext),
		Disableable: Disableable{},
	}, nil
}

func (c *CustomTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	c.Disableable.RenderToOutputBuffer(
		c.TagName(),
		context,
		c.LineNumber(),
		c.ParseContext(),
		output,
		func() {
			// Render tag name
			*output += c.TagName()
		},
	)
}

// Custom2Tag is another simple tag that renders its tag name.
// It embeds Disableable to check if it's disabled before rendering.
type Custom2Tag struct {
	*liquid.Tag
	Disableable
}

func NewCustom2Tag(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	return &Custom2Tag{
		Tag:         liquid.NewTag(tagName, markup, parseContext),
		Disableable: Disableable{},
	}, nil
}

func (c *Custom2Tag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	c.Disableable.RenderToOutputBuffer(
		c.TagName(),
		context,
		c.LineNumber(),
		c.ParseContext(),
		output,
		func() {
			// Render tag name
			*output += c.TagName()
		},
	)
}

// DisableCustomBlock is a block tag that disables "custom" tag.
// It embeds Disabler to wrap rendering with disabled tags.
type DisableCustomBlock struct {
	*liquid.Block
	Disabler
}

func NewDisableCustomBlock(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	return &DisableCustomBlock{
		Block:    liquid.NewBlock(tagName, markup, parseContext),
		Disabler: Disabler{},
	}, nil
}

func (d *DisableCustomBlock) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	disabledTags := []string{"custom"}
	d.Disabler.RenderToOutputBuffer(
		disabledTags,
		context,
		output,
		func() {
			// Render block body
			if d.Block != nil {
				d.Block.RenderToOutputBuffer(context, output)
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

func NewDisableBothBlock(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
	return &DisableBothBlock{
		Block:    liquid.NewBlock(tagName, markup, parseContext),
		Disabler: Disabler{},
	}, nil
}

func (d *DisableBothBlock) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	disabledTags := []string{"custom", "custom2"}
	d.Disabler.RenderToOutputBuffer(
		disabledTags,
		context,
		output,
		func() {
			// Render block body
			if d.Block != nil {
				d.Block.RenderToOutputBuffer(context, output)
			}
		},
	)
}

func TestBlockTagDisablingNestedTag(t *testing.T) {
	// Create environment
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Register custom tags
	env.RegisterTag("custom", tags.TagConstructor(NewCustomTag))
	env.RegisterTag("custom2", tags.TagConstructor(NewCustom2Tag))
	env.RegisterTag("disable", tags.TagConstructor(NewDisableCustomBlock))

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
	contextConfig := liquid.ContextConfig{
		Environment: env,
	}
	ctx := liquid.BuildContext(contextConfig)

	var output string
	tmpl.RenderToOutputBuffer(ctx, &output)

	// Expected: "Liquid error (line 1): custom usage is not allowed in this context;custom2"
	expected := "Liquid error (line 1): custom usage is not allowed in this context;custom2"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestBlockTagDisablingMultipleNestedTags(t *testing.T) {
	// Create environment
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)

	// Register custom tags
	env.RegisterTag("custom", tags.TagConstructor(NewCustomTag))
	env.RegisterTag("custom2", tags.TagConstructor(NewCustom2Tag))
	env.RegisterTag("disable", tags.TagConstructor(NewDisableBothBlock))

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
	contextConfig := liquid.ContextConfig{
		Environment: env,
	}
	ctx := liquid.BuildContext(contextConfig)

	var output string
	tmpl.RenderToOutputBuffer(ctx, &output)

	// Expected: "Liquid error (line 1): custom usage is not allowed in this context;Liquid error (line 1): custom2 usage is not allowed in this context"
	expected := "Liquid error (line 1): custom usage is not allowed in this context;Liquid error (line 1): custom2 usage is not allowed in this context"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
