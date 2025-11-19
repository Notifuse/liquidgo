package tags

import (
	"github.com/Notifuse/liquidgo/liquid"
)

// TagConstructor is a function type that creates a tag instance.
type TagConstructor func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error)

// RegisterStandardTags registers all standard tags with the environment.
func RegisterStandardTags(env *liquid.Environment) {
	// Simple tags
	env.RegisterTag("assign", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewAssignTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("echo", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewEchoTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("increment", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewIncrementTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("decrement", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewDecrementTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("break", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewBreakTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("continue", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewContinueTag(tagName, markup, parseContext), nil
	}))
	env.RegisterTag("cycle", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewCycleTag(tagName, markup, parseContext)
	}))

	// Block tags
	env.RegisterTag("comment", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewCommentTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("raw", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewRawTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("#", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewInlineCommentTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("doc", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewDocTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("capture", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewCaptureTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("if", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewIfTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("unless", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewUnlessTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("for", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewForTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("ifchanged", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewIfchangedTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("case", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewCaseTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("tablerow", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewTableRowTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("snippet", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewSnippetTag(tagName, markup, parseContext)
	}))

	// Include/render tags
	env.RegisterTag("include", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewIncludeTag(tagName, markup, parseContext)
	}))
	env.RegisterTag("render", TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (interface{}, error) {
		return NewRenderTag(tagName, markup, parseContext)
	}))
}
