package liquid

// outputEscapeExempt are the filters whose result an output writes as it is
// when the output escape is on: raw, and those that have just escaped it.
var outputEscapeExempt = map[string]bool{"raw": true, "escape": true, "escape_once": true, "h": true}

// SetOutputEscape turns the output escape of the environment on or off. On,
// every {{ }} output and echo writes its text HTML-escaped, as LiquidJS's
// outputEscape option does, so that a value from the render data cannot add
// markup to an HTML template. It is off by default, as in Shopify Liquid.
//
// What is not escaped:
//   - an output whose last filter is raw ({{ feed.html | raw }}), escape,
//     escape_once or h: raw writes HTML on purpose, the others have escaped;
//   - the text of a capture of the render, however it is reached
//     ({{ block }}, or {{ copy }} after an assign): its outputs were escaped
//     when it was captured, and the rest is the template's own markup. A
//     filter that changes that text ({{ block | strip }}) makes it ordinary
//     text again, escaped unless raw follows.
//
// Values are not escaped where they are read, so assign, capture, conditions
// and filters see the data as it is. Turning the option on registers the raw
// filter, which Shopify Liquid does not have.
func (e *Environment) SetOutputEscape(on bool) {
	e.outputEscape = on
	if on && !e.rawFilterRegistered {
		e.rawFilterRegistered = true
		if err := e.RegisterFilter(&rawFilter{}); err != nil {
			panic(err)
		}
	}
}

// OutputEscape reports whether the output escape of the environment is on.
func (e *Environment) OutputEscape() bool {
	return e != nil && e.outputEscape
}

// rawFilter holds the raw filter: it returns its input, and as the last
// filter of an output it keeps the output escape from escaping it.
type rawFilter struct{}

// Raw returns input as it is.
func (*rawFilter) Raw(input interface{}) interface{} {
	return input
}

// OutputEscape reports whether the outputs of the render are HTML-escaped.
func (c *Context) OutputEscape() bool {
	return c != nil && c.environment.OutputEscape()
}

// MarkSafeHTML records text as the text of a capture of the render: an output
// of exactly this text is not escaped. The record is shared by the
// subcontexts of the render.
func (c *Context) MarkSafeHTML(text string) {
	r := c.reporter()
	if r.safeHTML == nil {
		r.safeHTML = make(map[string]struct{})
	}
	r.safeHTML[text] = struct{}{}
}

func (c *Context) isSafeHTML(text string) bool {
	_, ok := c.reporter().safeHTML[text]
	return ok
}
