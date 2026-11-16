package tags

import "github.com/Notifuse/liquidgo/v5/liquid"

// The children a parse tree visitor walks for each tag, as the
// ParseTreeVisitor classes of Ruby's tags (liquid.ParseTreeNode).

// ParseTreeChildren returns the conditions of the if, each with its body.
func (i *IfTag) ParseTreeChildren() []interface{} {
	children := make([]interface{}, 0, len(i.blocks))
	for _, b := range i.blocks {
		children = append(children, b)
	}
	return children
}

// ParseTreeChildren returns the value the case compares and its conditions.
func (c *CaseTag) ParseTreeChildren() []interface{} {
	children := []interface{}{c.left}
	for _, b := range c.blocks {
		children = append(children, b)
	}
	return children
}

// ParseTreeChildren returns the bodies of the loop, then its limit, offset
// and collection.
func (f *ForTag) ParseTreeChildren() []interface{} {
	return liquid.CompactChildren(f.forBlock, f.elseBlock, f.limit, f.from, f.collectionName)
}

// ParseTreeChildren returns the body of the tablerow, its attributes and its
// collection.
func (t *TableRowTag) ParseTreeChildren() []interface{} {
	children := append([]interface{}{}, t.Nodelist()...)
	children = append(children, liquid.AttributeChildren(t.attributes)...)
	return append(children, t.collectionName)
}

// ParseTreeChildren returns the values the cycle goes through.
func (c *CycleTag) ParseTreeChildren() []interface{} {
	return c.variables
}

// ParseTreeChildren returns the variable the assign evaluates.
func (a *AssignTag) ParseTreeChildren() []interface{} {
	return []interface{}{a.from}
}

// ParseTreeChildren returns the variable the echo writes.
func (e *EchoTag) ParseTreeChildren() []interface{} {
	return []interface{}{e.variable}
}

// ParseTreeChildren returns the name of the partial, the variable given to
// it and the values of its attributes.
func (i *IncludeTag) ParseTreeChildren() []interface{} {
	return append([]interface{}{i.templateNameExpr, i.variableNameExpr}, liquid.AttributeChildren(i.attributes)...)
}

// ParseTreeChildren returns the name of the partial, the variable given to
// it and the values of its attributes.
func (r *RenderTag) ParseTreeChildren() []interface{} {
	return append([]interface{}{r.templateNameExpr, r.variableNameExpr}, liquid.AttributeChildren(r.attributes)...)
}
