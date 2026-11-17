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

// Standard returns the standard tags of Liquid, by name, in a map of the
// caller's own, for EnvironmentConfig.Tags: an entry added or replaced
// changes the tags of that environment only.
func Standard() map[string]TagConstructor {
	tags := make(map[string]TagConstructor, 24)
	// Simple tags
	tags["assign"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewAssignTag(tagName, markup, parseContext))
	})
	tags["echo"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewEchoTag(tagName, markup, parseContext), nil
	})
	tags["increment"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewIncrementTag(tagName, markup, parseContext)
	})
	tags["decrement"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewDecrementTag(tagName, markup, parseContext)
	})
	tags["break"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewBreakTag(tagName, markup, parseContext), nil
	})
	tags["continue"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return NewContinueTag(tagName, markup, parseContext), nil
	})
	tags["cycle"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCycleTag(tagName, markup, parseContext))
	})

	// Block tags
	tags["comment"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCommentTag(tagName, markup, parseContext))
	})
	tags["raw"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewRawTag(tagName, markup, parseContext))
	})
	tags["#"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewInlineCommentTag(tagName, markup, parseContext))
	})
	tags["doc"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewDocTag(tagName, markup, parseContext))
	})
	tags["capture"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCaptureTag(tagName, markup, parseContext))
	})
	tags["if"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewIfTag(tagName, markup, parseContext))
	})
	tags["unless"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewUnlessTag(tagName, markup, parseContext))
	})
	tags["for"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewForTag(tagName, markup, parseContext))
	})
	tags["ifchanged"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewIfchangedTag(tagName, markup, parseContext))
	})
	tags["case"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewCaseTag(tagName, markup, parseContext))
	})
	tags["tablerow"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewTableRowTag(tagName, markup, parseContext))
	})
	// Note: snippet tag was removed in Shopify Liquid v5.11.0

	// Include/render tags
	tags["include"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewIncludeTag(tagName, markup, parseContext))
	})
	tags["render"] = TagConstructor(func(tagName, markup string, parseContext liquid.ParseContextInterface) (liquid.Node, error) {
		return node(NewRenderTag(tagName, markup, parseContext))
	})
	return tags
}
