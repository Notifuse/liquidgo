package liquid

import "regexp"

// Const contains constants used throughout the Liquid package.
var (
	// FilterSeparator is the regex pattern for filter separator (|)
	FilterSeparator = regexp.MustCompile(`\|`)

	// ArgumentSeparator is the character used to separate arguments
	ArgumentSeparator = ','

	// FilterArgumentSeparator is the character used to separate filter arguments
	FilterArgumentSeparator = ':'

	// VariableAttributeSeparator is the character used to separate variable attributes
	VariableAttributeSeparator = '.'

	// WhitespaceControl is the character used for whitespace control
	WhitespaceControl = '-'

	// TagStart is the regex pattern for tag start ({%)
	TagStart = regexp.MustCompile(`\{\%`)

	// TagEnd is the regex pattern for tag end (%})
	TagEnd = regexp.MustCompile(`\%\}`)

	// TagName is the regex pattern for tag names
	TagName = regexp.MustCompile(`#|\w+`)

	// VariableSignature is the regex pattern for variable signatures
	VariableSignature = regexp.MustCompile(`\(?[\w\-\.\[\]]\)?`)

	// VariableSegment is the regex pattern for variable segments
	VariableSegment = regexp.MustCompile(`[\w\-]`)

	// VariableStart is the regex pattern for variable start ({{)
	VariableStart = regexp.MustCompile(`\{\{`)

	// VariableEnd is the regex pattern for variable end (}})
	VariableEnd = regexp.MustCompile(`\}\}`)

	// VariableIncompleteEnd is the regex pattern for incomplete variable end
	VariableIncompleteEnd = regexp.MustCompile(`\}\}?`)

	// QuotedString is the regex pattern for quoted strings
	QuotedString = regexp.MustCompile(`"[^"]*"|'[^']*'`)

	// QuotedFragment is the regex pattern for quoted fragments
	QuotedFragment = regexp.MustCompile(`"[^"]*"|'[^']*'|(?:[^\s,\|'"]|"[^"]*"|'[^']*')+`)

	// TagAttributes is the regex pattern for tag attributes
	TagAttributes = regexp.MustCompile(`(\w[\w-]*)\s*\:\s*("[^"]*"|'[^']*'|(?:[^\s,\|'"]|"[^"]*"|'[^']*')+)`)

	// AnyStartingTag is the regex pattern for any starting tag
	AnyStartingTag = regexp.MustCompile(`\{\%|\{\{`)

	// PartialTemplateParser is the regex pattern for partial template parsing
	PartialTemplateParser = regexp.MustCompile(`(?s)\{\%.*?\%\}|\{\{.*?\}\}?`)

	// TemplateParser is the regex pattern for template parsing
	TemplateParser = regexp.MustCompile(`(?s)(\{\%.*?\%\}|\{\{.*?\}\}?|\{\%|\{\{)`)

	// VariableParser is the regex pattern for variable parsing
	// Note: Go regexp doesn't support atomic groups or recursion, so we use a simpler pattern
	// that matches brackets with content or word characters with optional question mark
	VariableParser = regexp.MustCompile(`\[[^\[\]]*\]|[\w\-]+\??`)
)

// Const contains empty constants similar to Ruby's frozen empty collections.
var (
	// EMPTY_HASH is an empty map that can be reused
	EMPTY_HASH = map[string]interface{}{}

	// EMPTY_ARRAY is an empty slice that can be reused
	EMPTY_ARRAY = []interface{}{}
)
