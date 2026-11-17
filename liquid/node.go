package liquid

// Node is a node of a parsed template that renders: a variable, a tag or a
// block. A tag of a host application implements it to write its output.
type Node interface {
	RenderToOutput(context TagContext, output *OutputBuffer)
}

// TagConstructor builds the tag a template names, from its name and markup:
// EnvironmentConfig.Tags maps a tag name to it. An error is a syntax error
// of the template, and so is no tag (nil, or a nil pointer).
type TagConstructor func(tagName, markup string, parseContext ParseContextInterface) (Node, error)
