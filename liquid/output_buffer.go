package liquid

import (
	"reflect"
	"sync"
)

// OutputBuffer collects the output of a render. A buffer made for a render
// checks each write against its resource limits before it grows: the writes
// of a capture count in the assign score, the others against the render
// length limit.
type OutputBuffer struct {
	limits  *ResourceLimits
	buf     []byte
	capture bool
}

// NewOutputBuffer returns an empty buffer that checks no limit.
func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{}
}

// NewOutputBufferFor returns an empty buffer whose writes are checked
// against the render length limit of context.
// Inside ResourceLimits.WithCapture, its writes count in the assign score.
func NewOutputBufferFor(context TagContext) *OutputBuffer {
	limits := limitsOf(context)
	return &OutputBuffer{limits: limits, capture: limits != nil && limits.lastCaptureLength != nil}
}

// NewCaptureBuffer returns an empty buffer whose writes are counted in the
// assign score of context, as the body of a capture tag is.
func NewCaptureBuffer(context TagContext) *OutputBuffer {
	return &OutputBuffer{limits: limitsOf(context), capture: true}
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
	if o.limits != nil {
		if o.capture {
			o.limits.IncrementAssignScore(len(s))
		} else if limit := o.limits.renderLengthLimit; limit != nil && len(o.buf)+len(s) > *limit {
			o.limits.raiseLimitsReached()
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

// Since returns what was written after the first n bytes.
func (o *OutputBuffer) Since(n int) string {
	return string(o.buf[n:])
}

// Truncate keeps the first n bytes.
func (o *OutputBuffer) Truncate(n int) {
	o.buf = o.buf[:n]
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

// declares reports whether the pointer type t declares the method name with
// the signature of iface, rather than inheriting it from an embedded pointer:
// a method declared on *S is not in the method set of S, a promoted one is.
func declares(t reflect.Type, name string, iface reflect.Type) bool {
	if _, ok := t.MethodByName(name); !ok || !t.Implements(iface) {
		return false
	}
	_, inElem := t.Elem().MethodByName(name)
	return !inElem
}

func buildPlan(t reflect.Type, depth int) *renderPlan {
	if t.Kind() != reflect.Ptr || depth > 16 {
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
// render methods.
func inheritedPlan(t reflect.Type, depth int) *renderPlan {
	elem := t.Elem()
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
		if ft.Kind() != reflect.Ptr {
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
	return buildPlan(provider, depth+1)
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
