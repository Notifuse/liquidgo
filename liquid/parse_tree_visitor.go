package liquid

import (
	"reflect"
	"sort"
)

// ParseTreeVisitorCallback is called for each node of a type it was added for,
// with the context of its parent; it returns the item that stands for the node
// in the result, and the context of its children (nil keeps the parent's).
type ParseTreeVisitorCallback func(node interface{}, context interface{}) (interface{}, interface{})

// ParseTreeNode is a node of a parse tree whose children the visitor walks:
// its nodes and its expressions, as Ruby's per-class ParseTreeVisitor. A block
// without it gives its nodelist; any other node has no children.
type ParseTreeNode interface {
	ParseTreeChildren() []interface{}
}

// ParseTreeVisitor walks a parse tree, as Ruby's Liquid::ParseTreeVisitor: a
// callback added for VariableLookup, for example, lists the variables a
// template reads.
type ParseTreeVisitor struct {
	node      interface{}
	callbacks map[reflect.Type]ParseTreeVisitorCallback
}

// ForParseTreeVisitor returns a visitor of node with callbacks (nil for
// none yet).
func ForParseTreeVisitor(node interface{}, callbacks map[reflect.Type]ParseTreeVisitorCallback) *ParseTreeVisitor {
	if callbacks == nil {
		callbacks = make(map[reflect.Type]ParseTreeVisitorCallback)
	}
	return &ParseTreeVisitor{node: node, callbacks: callbacks}
}

// AddCallbackFor calls callback for the nodes of the types of the given
// values, such as (*VariableLookup)(nil).
func (ptv *ParseTreeVisitor) AddCallbackFor(types []interface{}, callback ParseTreeVisitorCallback) *ParseTreeVisitor {
	for _, t := range types {
		if typ := reflect.TypeOf(t); typ != nil {
			ptv.callbacks[typ] = callback
		}
	}
	return ptv
}

// Visit walks the children of the node: for each, the pair of its item (nil
// without a callback for its type) and the result of visiting it.
func (ptv *ParseTreeVisitor) Visit(context interface{}) []interface{} {
	children := parseTreeChildren(ptv.node)
	result := make([]interface{}, 0, len(children))
	for _, child := range children {
		var item, childContext interface{}
		if callback, ok := ptv.callbacks[reflect.TypeOf(child)]; ok {
			item, childContext = callback(child, context)
		}
		if childContext == nil {
			childContext = context
		}
		result = append(result, []interface{}{item, ForParseTreeVisitor(child, ptv.callbacks).Visit(childContext)})
	}
	return result
}

// parseTreeChildren returns the children of node, as Ruby's children.
func parseTreeChildren(node interface{}) []interface{} {
	switch n := node.(type) {
	case nil:
		return nil
	case ParseTreeNode:
		return n.ParseTreeChildren()
	case interface{ Nodelist() []interface{} }:
		return n.Nodelist()
	}
	return nil
}

// CompactChildren returns children without the nil ones, a nil pointer
// included, as Ruby's compact.
func CompactChildren(children ...interface{}) []interface{} {
	out := make([]interface{}, 0, len(children))
	for _, c := range children {
		if c == nil {
			continue
		}
		if v := reflect.ValueOf(c); v.Kind() == reflect.Ptr && v.IsNil() {
			continue
		}
		out = append(out, c)
	}
	return out
}

// AttributeChildren returns the values of attributes, by key, so that a
// visit is the same at each run.
func AttributeChildren(attributes map[string]interface{}) []interface{} {
	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		out = append(out, attributes[k])
	}
	return out
}

// ParseTreeChildren returns the name and the filters, flattened, of the
// variable, as Ruby's Variable::ParseTreeVisitor.
func (v *Variable) ParseTreeChildren() []interface{} {
	children := []interface{}{v.name}
	for _, f := range v.filters {
		children = append(children, flatten(f)...)
	}
	return children
}

// ParseTreeChildren returns the lookups of the variable lookup.
func (vl *VariableLookup) ParseTreeChildren() []interface{} {
	return vl.lookups
}

// ParseTreeChildren returns the two ends of the range.
func (rl *RangeLookup) ParseTreeChildren() []interface{} {
	return []interface{}{rl.startObj, rl.endObj}
}

// ParseTreeChildren returns the operands, the chained condition and the
// body of the condition.
func (c *Condition) ParseTreeChildren() []interface{} {
	return CompactChildren(c.left, c.right, c.childCondition, c.attachment)
}

// flatten returns the items of nested lists, as Ruby's flatten.
func flatten(v interface{}) []interface{} {
	list, ok := v.([]interface{})
	if !ok {
		return []interface{}{v}
	}
	var out []interface{}
	for _, e := range list {
		out = append(out, flatten(e)...)
	}
	return out
}
