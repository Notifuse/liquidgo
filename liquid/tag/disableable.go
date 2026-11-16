package tag

import (
	"github.com/Notifuse/liquidgo/v5/liquid"
)

// Disableable provides functionality for tags that can be disabled.
// Tags that embed this struct will check if they're disabled before rendering.
type Disableable struct{}

// disabledError creates a DisabledError and returns the error message string.
func (d *Disableable) disabledError(
	tagName string,
	context liquid.TagContext,
	lineNumber *int,
	parseContext liquid.ParseContextInterface,
) string {
	// Get locale from parse context
	var locale *liquid.I18n
	if pc, ok := parseContext.(*liquid.ParseContext); ok {
		locale = pc.Locale()
	} else {
		// Fallback: create default locale if parse context doesn't have one
		locale = liquid.NewI18n("")
	}

	// Translate error message with panic recovery
	var errorText string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If translation panics, use fallback
				errorText = "usage is not allowed in this context"
			}
		}()
		errorText = locale.T("errors.disabled.tag", nil)
		if errorText == "errors.disabled.tag" {
			// Fallback if translation not found (key returned as-is)
			errorText = "usage is not allowed in this context"
		}
	}()

	// Create error message: "{tag_name} {error_text}"
	errorMsg := tagName + " " + errorText

	// Create DisabledError
	err := liquid.NewDisabledError(errorMsg)
	if err.Err.LineNumber == nil {
		err.Err.LineNumber = lineNumber
	}

	// Handle error through context
	return context.HandleError(err, lineNumber)
}

// Render calls renderFn to render the tag, unless the tag is disabled in
// context (see Disabler), when it writes the disabled error instead. A tag
// that embeds Disableable calls it from its own RenderToOutput: the name is
// not RenderToOutput, so that it never hides the tag's Node method.
func (d *Disableable) Render(
	tagName string,
	context liquid.TagContext,
	lineNumber *int,
	parseContext liquid.ParseContextInterface,
	output *liquid.OutputBuffer,
	renderFn func(),
) {
	if context.TagDisabled(tagName) {
		output.WriteString(d.disabledError(tagName, context, lineNumber, parseContext))
		return
	}
	renderFn()
}
