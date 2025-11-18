package liquid

import "reflect"

// ParseTreeVisitorCallback is a function type for visitor callbacks.
type ParseTreeVisitorCallback func(node interface{}, context interface{}) (interface{}, interface{})

// ParseTreeVisitor visits nodes in a parse tree.
type ParseTreeVisitor struct {
	node      interface{}
	callbacks map[reflect.Type]ParseTreeVisitorCallback
}

// NewParseTreeVisitor creates a new ParseTreeVisitor.
func NewParseTreeVisitor(node interface{}, callbacks map[reflect.Type]ParseTreeVisitorCallback) *ParseTreeVisitor {
	if callbacks == nil {
		callbacks = make(map[reflect.Type]ParseTreeVisitorCallback)
	}
	return &ParseTreeVisitor{
		node:      node,
		callbacks: callbacks,
	}
}

// For creates a ParseTreeVisitor for a node, using node-specific visitor if available.
func ForParseTreeVisitor(node interface{}, callbacks map[reflect.Type]ParseTreeVisitorCallback) *ParseTreeVisitor {
	nodeType := reflect.TypeOf(node)
	if nodeType == nil {
		return NewParseTreeVisitor(node, callbacks)
	}
	
	// Check for node-specific ParseTreeVisitor type
	// In Ruby: if defined?(node.class::ParseTreeVisitor)
	// In Go, we check if the node type has a method that returns a ParseTreeVisitor constructor
	nodeValue := reflect.ValueOf(node)
	if nodeValue.IsValid() {
		// Check if there's a method that can create a node-specific visitor
		// Some node types might have a ParseTreeVisitorType method or similar
		// For now, we'll use the default visitor, but check for node-specific implementations
		// by looking for a method that returns *ParseTreeVisitor
		parseTreeVisitorMethod := nodeValue.MethodByName("ParseTreeVisitor")
		if parseTreeVisitorMethod.IsValid() {
			results := parseTreeVisitorMethod.Call([]reflect.Value{reflect.ValueOf(callbacks)})
			if len(results) > 0 {
				if ptv, ok := results[0].Interface().(*ParseTreeVisitor); ok {
					return ptv
				}
			}
		}
	}
	
	// Use default visitor
	return NewParseTreeVisitor(node, callbacks)
}

// AddCallbackFor adds a callback for specific node types.
func (ptv *ParseTreeVisitor) AddCallbackFor(types []interface{}, callback ParseTreeVisitorCallback) *ParseTreeVisitor {
	for _, t := range types {
		typ := reflect.TypeOf(t)
		if typ != nil {
			ptv.callbacks[typ] = callback
		}
	}
	return ptv
}

// Visit visits the parse tree and returns results.
func (ptv *ParseTreeVisitor) Visit(context interface{}) []interface{} {
	children := ptv.children()
	result := make([]interface{}, 0, len(children))
	
	for _, child := range children {
		childType := reflect.TypeOf(child)
		callback, ok := ptv.callbacks[childType]
		if !ok {
			// Default callback: return node as-is
			callback = func(node interface{}, ctx interface{}) (interface{}, interface{}) {
				return node, ctx
			}
		}
		
		item, newContext := callback(child, context)
		childVisitor := ForParseTreeVisitor(child, ptv.callbacks)
		childResults := childVisitor.Visit(newContext)
		
		result = append(result, []interface{}{item, childResults})
	}
	
	return result
}

func (ptv *ParseTreeVisitor) children() []interface{} {
	// Check if node has Nodelist method
	nodeValue := reflect.ValueOf(ptv.node)
	if !nodeValue.IsValid() {
		return EMPTY_ARRAY
	}
	
	nodelistMethod := nodeValue.MethodByName("Nodelist")
	if nodelistMethod.IsValid() {
		result := nodelistMethod.Call(nil)
		if len(result) > 0 {
			if nodelist, ok := result[0].Interface().([]interface{}); ok {
				return nodelist
			}
		}
	}
	
	return EMPTY_ARRAY
}

