package liquid

import (
	"reflect"
	"runtime"
	"sync"
)

// OutputBuffer collects the output of a render. A buffer made for a render
// checks each write against its resource limits before it grows: the writes
// of a capture count in the assign score, the others against the render
// length limit.
type OutputBuffer struct {
	limits *ResourceLimits
	buf    []byte
	mode   bufferMode
}

type bufferMode int

const (
	// checkLength refuses a write that takes the buffer past the render
	// length limit.
	checkLength bufferMode = iota
	// countAssign counts each write in the assign score, as the body of a
	// capture.
	countAssign
	// checkAssign refuses a write that takes the buffer past the assign
	// score limit, and does not count it: the capture it is written to
	// counts it.
	checkAssign
)

// NewOutputBuffer returns an empty buffer that checks no limit.
func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{}
}

// NewOutputBufferFor returns an empty buffer whose writes are checked
// against the render length limit of context, as Ruby checks the string a
// tag renders to. Inside a capture, they are checked against the assign
// score limit instead, and counted once they are written to the capture.
func NewOutputBufferFor(context TagContext) *OutputBuffer {
	limits := limitsOf(context)
	if limits != nil && limits.inCapture() {
		return &OutputBuffer{limits: limits, mode: checkAssign}
	}
	return &OutputBuffer{limits: limits, mode: checkLength}
}

// RenderCaptured calls render with a buffer whose writes count in the assign
// score of context, as the body of a capture tag, and returns what it wrote.
// A buffer that NewOutputBufferFor makes during render is checked as part of
// the capture.
func RenderCaptured(context TagContext, render func(*OutputBuffer)) string {
	limits := limitsOf(context)
	if limits != nil {
		limits.captures++
		defer func() { limits.captures-- }()
	}
	buf := &OutputBuffer{limits: limits, mode: countAssign}
	render(buf)
	return buf.String()
}

// Scratch returns an empty buffer under the limits of o whose writes are not
// counted, as Ruby's block_output: a tag renders to it what it may not keep,
// and writes to o what it keeps, which o counts.
func (o *OutputBuffer) Scratch() *OutputBuffer {
	if o.mode == countAssign {
		return &OutputBuffer{limits: o.limits, mode: checkAssign}
	}
	return &OutputBuffer{limits: o.limits, mode: o.mode}
}

func limitsOf(context TagContext) *ResourceLimits {
	if context == nil {
		return nil
	}
	if ctx, ok := context.(*Context); ok && ctx == nil {
		return nil
	}
	return context.ResourceLimits()
}

// WriteString appends s, or raises a MemoryError if s would take the buffer
// over its limit.
func (o *OutputBuffer) WriteString(s string) {
	if len(s) == 0 {
		return
	}
	if rl := o.limits; rl != nil {
		size := len(o.buf) + len(s)
		switch o.mode {
		case countAssign:
			rl.IncrementAssignScore(len(s))
		case checkAssign:
			if rl.assignScoreLimit != nil && size > *rl.assignScoreLimit {
				rl.raiseLimitsReached()
			}
		default:
			if rl.renderLengthLimit != nil && size > *rl.renderLengthLimit {
				rl.raiseLimitsReached()
			}
		}
	}
	o.buf = append(o.buf, s...)
}

// String returns what was written.
func (o *OutputBuffer) String() string {
	return string(o.buf)
}

// Len returns the number of bytes written.
func (o *OutputBuffer) Len() int {
	return len(o.buf)
}

// RenderToString calls render with a new buffer and appends what it wrote to
// output. A tag that keeps a RenderToOutputBuffer method with a *string
// calls its RenderToOutput this way.
func RenderToString(context TagContext, output *string, render func(*OutputBuffer)) {
	buf := NewOutputBufferFor(context)
	render(buf)
	*output += buf.String()
}

// outputRenderer is a node that writes to an OutputBuffer. A tag declares
// RenderToOutput to render in linear time; one that declares only Render or
// RenderToOutputBuffer is rendered as before, into a string of its own.
type outputRenderer interface {
	RenderToOutput(context TagContext, output *OutputBuffer)
}

type stringBufferRenderer interface {
	RenderToOutputBuffer(context TagContext, output *string)
}

type stringRenderer interface {
	Render(context TagContext) string
}

type renderKind int

const (
	planNothing renderKind = iota
	planOutput
	planStringBuffer
	planRender
)

// renderPlan says which method renders a node type: the one the type
// declares, or the one it inherits. A type that declares Render renders
// with it, and with fallback when Render returns nothing or the node is
// blank, as in v5.11.0.
type renderPlan struct {
	fallback *renderPlan
	kind     renderKind
}

var (
	renderPlans        sync.Map // reflect.Type -> *renderPlan
	outputRendererType = reflect.TypeOf((*outputRenderer)(nil)).Elem()
	stringBufferType   = reflect.TypeOf((*stringBufferRenderer)(nil)).Elem()
	stringRendererType = reflect.TypeOf((*stringRenderer)(nil)).Elem()
	renderMethodNames  = []string{"RenderToOutput", "RenderToOutputBuffer", "Render"}
	renderMethodIfaces = []reflect.Type{outputRendererType, stringBufferType, stringRendererType}
	planKindsByDeclare = []renderKind{planOutput, planStringBuffer, planRender}
)

// planFor returns the render plan of the type of node.
func planFor(node interface{}) *renderPlan {
	t := reflect.TypeOf(node)
	if p, ok := renderPlans.Load(t); ok {
		return p.(*renderPlan)
	}
	p := buildPlan(t, 0)
	renderPlans.Store(t, p)
	return p
}

// declares reports whether the type t declares the method name with the
// signature of iface, on its pointer or on its value receiver, rather than
// inheriting it from an embedded field. Go builds the methods a type inherits,
// and the pointer methods of a value receiver, as wrappers that it marks
// as autogenerated.
func declares(t reflect.Type, name string, iface reflect.Type) bool {
	if !t.Implements(iface) {
		return false
	}
	sets := []reflect.Type{t}
	if t.Kind() == reflect.Ptr {
		sets = append(sets, t.Elem())
	}
	for _, set := range sets {
		if m, ok := set.MethodByName(name); ok && !isWrapper(m) {
			return true
		}
	}
	return false
}

// isWrapper reports whether m is a method Go generated: one promoted from
// an embedded field, or the pointer method of a value receiver.
func isWrapper(m reflect.Method) bool {
	fn := runtime.FuncForPC(m.Func.Pointer())
	if fn == nil {
		return true
	}
	file, _ := fn.FileLine(fn.Entry())
	return file == "<autogenerated>"
}

func buildPlan(t reflect.Type, depth int) *renderPlan {
	if depth > 16 {
		return genericPlan(t)
	}
	for i, name := range renderMethodNames {
		if !declares(t, name, renderMethodIfaces[i]) {
			continue
		}
		if planKindsByDeclare[i] != planRender {
			return &renderPlan{kind: planKindsByDeclare[i]}
		}
		return &renderPlan{kind: planRender, fallback: inheritedPlan(t, depth)}
	}
	return inheritedPlan(t, depth)
}

// inheritedPlan returns the plan of the one embedded field that gives t its
// render methods, as long as t still has the methods that plan calls.
func inheritedPlan(t reflect.Type, depth int) *renderPlan {
	elem := t
	if t.Kind() == reflect.Ptr {
		elem = t.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return genericPlan(t)
	}
	var provider reflect.Type
	providers := 0
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		if !f.Anonymous {
			continue
		}
		ft := f.Type
		if ft.Kind() != reflect.Ptr && t.Kind() == reflect.Ptr {
			ft = reflect.PointerTo(ft)
		}
		if ft.Implements(outputRendererType) || ft.Implements(stringBufferType) {
			provider = ft
			providers++
		}
	}
	if providers != 1 {
		return genericPlan(t)
	}
	p := buildPlan(provider, depth+1)
	if !callable(t, p) {
		return genericPlan(t)
	}
	return p
}

// callable reports whether t has every method p calls: two embedded fields
// with a method of the same name hide it from t.
func callable(t reflect.Type, p *renderPlan) bool {
	for ; p != nil; p = p.fallback {
		switch p.kind {
		case planOutput:
			if !t.Implements(outputRendererType) {
				return false
			}
		case planStringBuffer:
			if !t.Implements(stringBufferType) {
				return false
			}
		case planRender:
			if !t.Implements(stringRendererType) {
				return false
			}
		}
	}
	return true
}

func genericPlan(t reflect.Type) *renderPlan {
	switch {
	case t.Implements(outputRendererType):
		return &renderPlan{kind: planOutput}
	case t.Implements(stringBufferType):
		return &renderPlan{kind: planStringBuffer}
	}
	return &renderPlan{kind: planNothing}
}

// renderWithPlan renders node into output as plan says.
func renderWithPlan(p *renderPlan, node interface{}, context TagContext, output *OutputBuffer) {
	for p != nil {
		switch p.kind {
		case planOutput:
			node.(outputRenderer).RenderToOutput(context, output)
			return
		case planStringBuffer:
			var s string
			node.(stringBufferRenderer).RenderToOutputBuffer(context, &s)
			output.WriteString(s)
			return
		case planRender:
			if blanker, ok := node.(interface{ Blank() bool }); !ok || !blanker.Blank() {
				if s := node.(stringRenderer).Render(context); s != "" {
					output.WriteString(s)
					return
				}
			}
			p = p.fallback
		default:
			return
		}
	}
}
