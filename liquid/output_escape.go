package liquid

// outputEscapeSafeFilters are the filters whose result is HTML the output
// escape writes as it is: raw on purpose, the others because they have just
// escaped their input.
var outputEscapeSafeFilters = map[string]bool{"raw": true, "escape": true, "escape_once": true, "h": true}

// outputEscapeKeepingFilters are the filters that keep HTML safe: they only
// remove whitespace, cut the text short or add <br />, so they cannot turn
// escaped text or a capture into markup from the data. Any other filter makes
// its result ordinary text, escaped unless a safe filter follows.
var outputEscapeKeepingFilters = map[string]bool{
	"strip": true, "lstrip": true, "rstrip": true, "strip_newlines": true,
	"newline_to_br": true, "truncate": true, "truncatewords": true,
}

// SetOutputEscape turns the output escape of the environment on or off. On,
// every {{ }} output, echo and cycle writes its text HTML-escaped, as
// LiquidJS's outputEscape option does, so that a value of the render data
// cannot add markup to an HTML template. It is off by default, as in Shopify
// Liquid.
//
// An output is written as it is when its value is HTML on purpose:
//   - the result of raw ({{ feed.html | raw }}), or of escape, escape_once
//     or h, which have escaped it already;
//   - the text of a capture of the render, however it is reached ({{ block }},
//     or {{ copy }} after an assign): its outputs were escaped when it was
//     captured, and the rest is the template's own markup;
//   - either of these through strip, lstrip, rstrip, strip_newlines,
//     newline_to_br, truncate or truncatewords ({{ bio | escape |
//     newline_to_br }}).
//
// Any other filter after them makes the result ordinary text again. Values are
// not escaped where they are read, so assign, capture, conditions and filters
// see the data as it is. Turning the option on registers the raw filter, which
// Shopify Liquid does not have.
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

// rawFilter holds the raw filter: it returns its input, and the output escape
// writes its result as it is.
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
// subcontexts of the render, and Reset clears it.
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

// escapeTracker follows a variable through its filters to tell whether its
// value is HTML on purpose (see SetOutputEscape). Its methods do nothing on a
// nil tracker, which Variable.Render passes.
type escapeTracker struct {
	ctx  *Context
	safe bool
}

func (t *escapeTracker) start(value interface{}) {
	if t == nil {
		return
	}
	text, ok := value.(string)
	t.safe = ok && t.ctx.isSafeHTML(text)
}

func (t *escapeTracker) filter(name string) {
	if t == nil {
		return
	}
	switch {
	case outputEscapeSafeFilters[name]:
		t.safe = true
	case outputEscapeKeepingFilters[name]:
	default:
		t.safe = false
	}
}

func (t *escapeTracker) unsafe() {
	if t != nil {
		t.safe = false
	}
}

// WriteOutput writes value to output as an output of the template: escaped
// when the output escape is on, unless it is the text of a capture. Tags that
// write a value, such as cycle, use it.
func WriteOutput(context TagContext, output *OutputBuffer, value interface{}) {
	ctx, ok := context.Context().(*Context)
	if !ok || !ctx.OutputEscape() {
		WriteValue(output, value)
		return
	}
	if text, ok := value.(string); ok && ctx.isSafeHTML(text) {
		output.WriteString(text)
		return
	}
	writeValue(escapingWriter{output}, value, nil)
}

// escapeChunk is how much text escapingWriter escapes at a time, so that the
// render length limit stops a long output before all of it is escaped.
const escapeChunk = 64 << 10

// escapingWriter HTML-escapes what it writes to an output buffer, a chunk at a
// time: each write checks the render length limit.
type escapingWriter struct{ output *OutputBuffer }

func (w escapingWriter) WriteString(s string) {
	for len(s) > escapeChunk {
		w.output.WriteString(htmlEscaper.Replace(s[:escapeChunk]))
		s = s[escapeChunk:]
	}
	w.output.WriteString(htmlEscaper.Replace(s))
}
