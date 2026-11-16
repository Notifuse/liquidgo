package tags

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TagConstructor builds a tag; it is liquid.TagConstructor.
type TagConstructor = liquid.TagConstructor

// node returns the tag a constructor built, or its error: a nil *T with an
// error must not become a non-nil Node.
func node[T liquid.Node](tag T, err error) (liquid.Node, error) {
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// RegisterStandardTags registers all standard tags with the environment.
func RegisterStandardTags(env *liquid.Environment) {
	// Simple tags
	env.RegisterTag("assign", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewAssignTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("echo", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewEchoTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("increment", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewIncrementTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("decrement", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewDecrementTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("break", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewBreakTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("continue", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewContinueTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("cycle", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCycleTag(tagName, markup, parseContext))
	}))

	// Block tags
	env.RegisterTag("comment", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCommentTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("raw", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewRawTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("#", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewInlineCommentTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("doc", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewDocTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("capture", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCaptureTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("if", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewIfTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("unless", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewUnlessTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("for", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewForTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("ifchanged", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewIfchangedTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("case", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCaseTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("tablerow", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewTableRowTag(tagName, markup, parseContext))
	}))
	// Note: snippet tag was removed in Shopify Liquid v5.11.0

	// Include/render tags
	env.RegisterTag("include", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewIncludeTag(tagName, markup, parseContext))
	}))
	env.RegisterTag("render", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewRenderTag(tagName, markup, parseContext))
	}))
}
