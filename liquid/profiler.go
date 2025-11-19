package liquid

import (
	"time"
)

// Profiler enables support for profiling template rendering to help track down performance issues.
//
// To enable profiling, pass Profile: true option to Template.Parse.
// After Template.Render is called, the template object makes available an instance of this
// class via the Template.Profiler() method.
//
// This object contains all profiling information, containing information on what tags were rendered,
// where in the templates these tags live, and how long each tag took to render.
//
// This is a tree structure that keeps track of tags and rendering times
// inside of {% include %} tags.
//
// Profiler also exposes the total time of the template's render in Profiler.TotalRenderTime().
//
// All render times are in seconds. There is a small performance hit when profiling is enabled.
type Profiler struct {
	rootChildren    []*Timing
	currentChildren *[]*Timing // Pointer to current children slice being built
	totalTime       float64
}

// Timing represents a single timing node in the profiler tree.
type Timing struct {
	code         string
	templateName string
	lineNumber   *int
	children     []*Timing
	totalTime    float64
	selfTime     *float64 // Cached self time
}

// NewProfiler creates a new Profiler instance.
func NewProfiler() *Profiler {
	return &Profiler{
		rootChildren:    []*Timing{},
		currentChildren: nil,
		totalTime:       0.0,
	}
}

// Profile profiles a template render.
// Nested renders are done from a tag that already has a timing node.
func (p *Profiler) Profile(templateName string, fn func()) {
	if p.currentChildren != nil {
		// Already profiling, just execute
		fn()
		return
	}

	renderIdx := len(p.rootChildren)
	p.currentChildren = &p.rootChildren
	defer func() {
		p.currentChildren = nil
		if renderIdx < len(p.rootChildren) {
			if timing := p.rootChildren[renderIdx]; timing != nil {
				p.totalTime += timing.totalTime
			}
		}
	}()

	// Profile the entire template render as a node
	p.ProfileNode(templateName, "", nil, fn)
}

// ProfileNode profiles a single node (tag or variable).
func (p *Profiler) ProfileNode(templateName, code string, lineNumber *int, fn func()) {
	timing := &Timing{
		code:         code,
		templateName: templateName,
		lineNumber:   lineNumber,
		children:     []*Timing{},
	}

	// Store pointer to parent children slice for appending
	parentChildrenRef := p.currentChildren
	if parentChildrenRef == nil {
		parentChildrenRef = &p.rootChildren
	}

	startTime := time.Now()
	p.currentChildren = &timing.children
	defer func() {
		timing.totalTime = time.Since(startTime).Seconds()
		*parentChildrenRef = append(*parentChildrenRef, timing)
		p.currentChildren = parentChildrenRef
	}()

	fn()
}

// Children returns the profiler children.
// If there's only one child, return its children instead (to skip the root wrapper).
func (p *Profiler) Children() []*Timing {
	children := p.rootChildren
	if len(children) == 1 && children[0] != nil {
		return children[0].children
	}
	return children
}

// Length returns the number of children.
func (p *Profiler) Length() int {
	return len(p.Children())
}

// At returns the child at the given index.
func (p *Profiler) At(idx int) *Timing {
	children := p.Children()
	if idx < 0 || idx >= len(children) {
		return nil
	}
	return children[idx]
}

// TotalTime returns the total render time in seconds.
func (p *Profiler) TotalTime() float64 {
	return p.totalTime
}

// TotalRenderTime returns the total render time (alias for TotalTime).
func (p *Profiler) TotalRenderTime() float64 {
	return p.totalTime
}

// Code returns the code for this timing node.
func (t *Timing) Code() string {
	return t.code
}

// TemplateName returns the template name for this timing node.
func (t *Timing) TemplateName() string {
	return t.templateName
}

// Partial returns the template name (alias for TemplateName).
func (t *Timing) Partial() string {
	return t.templateName
}

// LineNumber returns the line number for this timing node.
func (t *Timing) LineNumber() *int {
	return t.lineNumber
}

// Children returns the children of this timing node.
func (t *Timing) Children() []*Timing {
	return t.children
}

// TotalTime returns the total time for this timing node in seconds.
func (t *Timing) TotalTime() float64 {
	return t.totalTime
}

// RenderTime returns the render time (alias for TotalTime).
func (t *Timing) RenderTime() float64 {
	return t.totalTime
}

// SelfTime returns the self time (total time minus children time).
func (t *Timing) SelfTime() float64 {
	if t.selfTime != nil {
		return *t.selfTime
	}

	totalChildrenTime := 0.0
	for _, child := range t.children {
		totalChildrenTime += child.totalTime
	}

	selfTime := t.totalTime - totalChildrenTime
	t.selfTime = &selfTime
	return selfTime
}
