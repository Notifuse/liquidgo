package liquid

// selfKeyword names the SelfDrop in a template (Shopify Liquid 5.13).
const selfKeyword = "self"

// SelfDrop is the value of self in a template, when no variable of that name
// exists (Shopify Liquid 5.13): self.x and self['x'] read the variable x
// through the scopes and the data of the context self was read in, as a
// lookup of x would; a partial given self reads the variables of its caller.
type SelfDrop struct {
	context *Context
}

// LiquidGet reads the variable key of the context self was read in. A key no
// scope nor data defines is nil, or an UndefinedVariable with strict
// variables, as Ruby's lookup of a key self does not have.
func (s *SelfDrop) LiquidGet(ctx *Context, key string) (interface{}, bool, error) {
	if !s.context.variableDefined(key) {
		if ctx != nil && ctx.StrictVariables() {
			return nil, false, NewUndefinedVariable("undefined variable " + key)
		}
		return nil, true, nil
	}
	return s.context.FindVariable(key, false), true, nil
}

// String writes the drop as Ruby writes it, by its class.
func (s *SelfDrop) String() string { return "Liquid::SelfDrop" }
