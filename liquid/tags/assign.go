package tags

import (
	"regexp"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// assignSyntax matches: variable_name = value
// Using Ruby Liquid regex: /(#{VariableSignature}+)\s*=\s*(.*)\s*/om, with
// VariableSignature = /\(?[\w\-\.\[\]]\)?/
var assignSyntax = regexp.MustCompile(`(?s)((?:\(?[\w\-\.\[\]]\)?)+)\s*=\s*(.*)\s*`)

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
		return nil, liquid.NewSyntaxErrorFor(parseContext, "errors.syntax.assign", nil)
	}
	if err := checkVariableName(matches[1], markup, parseContext); err != nil {
		return nil, err
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

// RenderToOutput renders the assign tag.
func (a *AssignTag) RenderToOutput(context liquid.TagContext, output *liquid.OutputBuffer) {
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
		return len(v)
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
