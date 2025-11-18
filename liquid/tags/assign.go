package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/liquid"
)

// assignSyntax matches: variable_name = value
// VariableSignature matches a single character, so we need to match multiple
var assignSyntax = regexp.MustCompile(`([\w\-\.\[\]]+)\s*=\s*(.*)\s*`)

// AssignTag represents an assign tag that creates a new variable.
type AssignTag struct {
	*liquid.Tag
	to   string
	from *liquid.Variable
}

// NewAssignTag creates a new AssignTag.
func NewAssignTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*AssignTag, error) {
	matches := assignSyntax.FindStringSubmatch(markup)
	if len(matches) < 3 {
		// Get locale from parse context if it's a ParseContext struct
		var locale *liquid.I18n
		if pc, ok := parseContext.(*liquid.ParseContext); ok {
			locale = pc.Locale()
			msg := locale.Translate("errors.syntax.assign", map[string]interface{}{})
			return nil, liquid.NewSyntaxError(msg)
		}
		return nil, liquid.NewSyntaxError("Liquid syntax error: assign")
	}

	return &AssignTag{
		Tag:  liquid.NewTag(tagName, markup, parseContext),
		to:   matches[1],
		from: liquid.NewVariable(matches[2], parseContext),
	}, nil
}

// To returns the variable name being assigned to.
func (a *AssignTag) To() string {
	return a.to
}

// From returns the variable being assigned from.
func (a *AssignTag) From() *liquid.Variable {
	return a.from
}

// RenderToOutputBuffer renders the assign tag.
func (a *AssignTag) RenderToOutputBuffer(context liquid.TagContext, output *string) {
	val := a.from.Render(context)

	// Set in the last scope (outermost scope, matching Ruby's context.scopes.last[@to] = val)
	ctx := context.Context().(*liquid.Context)
	ctx.SetLast(a.to, val)

	// Increment assign score
	rl := context.ResourceLimits()
	if rl != nil {
		score := assignScoreOf(val)
		rl.IncrementAssignScore(score)
	}
}

// assignScoreOf calculates the assign score for resource limits.
func assignScoreOf(val interface{}) int {
	switch v := val.(type) {
	case string:
		return len([]byte(v))
	case []interface{}:
		sum := 1
		for _, child := range v {
			sum += assignScoreOf(child)
		}
		return sum
	case map[string]interface{}:
		sum := 1
		for key, entryValue := range v {
			sum += assignScoreOf(key)
			sum += assignScoreOf(entryValue)
		}
		return sum
	default:
		return 1
	}
}

// Blank returns true since assign tags are blank.
func (a *AssignTag) Blank() bool {
	return true
}
