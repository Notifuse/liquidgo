package liquid

// errorFromPanic returns the error a recovered panic carries: a Liquid error
// as it is, and anything else as an InternalError, whose message reveals
// nothing of the cause.
func errorFromPanic(r interface{}) error {
	switch e := r.(type) {
	case LiquidError:
		return e
	case *Error:
		return e
	}
	return NewInternalError("internal")
}

// stopsRender reports whether err must end the render rather than be
// written in place of the node: a resource limit, or a render that was
// cancelled.
func stopsRender(err error) bool {
	switch err.(type) {
	case *MemoryError, *CanceledError:
		return true
	}
	return false
}

// rescueRenderNode handles a panic in the render of node, as Ruby's
// BlockBody#rescue_render_node: a limit or a cancellation goes on up; an
// undefined variable, drop method or filter is recorded and writes nothing;
// any other error is recorded and its message written in place of the node,
// unless the node is a blank tag and blankErrorText is false (a body parsed
// outside strict2, as Shopify Liquid 5.14).
func rescueRenderNode(context TagContext, output *OutputBuffer, node interface{}, r interface{}, blankErrorText bool) {
	if p := fatalPanic(context, r); p != nil {
		panic(p)
	}
	err := errorFromPanic(r)
	if stopsRender(err) {
		panic(r)
	}
	var lineNumber *int
	if n, ok := node.(interface{ LineNumber() *int }); ok {
		lineNumber = n.LineNumber()
	}
	switch err.(type) {
	case *UndefinedVariable, *UndefinedDropMethod, *UndefinedFilter:
		context.HandleError(err, lineNumber)
		return
	}
	message := context.HandleError(err, lineNumber)
	if _, isVariable := node.(*Variable); !isVariable && !blankErrorText {
		if b, ok := node.(interface{ Blank() bool }); ok && b.Blank() {
			return
		}
	}
	output.WriteString(message)
}
