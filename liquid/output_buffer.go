package liquid

import "unsafe"

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

// reserve makes room for n bytes, the output a template wrote at its last
// render, within the render length limit.
func (o *OutputBuffer) reserve(n int) {
	if rl := o.limits; rl != nil && rl.renderLengthLimit != nil && n > *rl.renderLengthLimit {
		n = *rl.renderLengthLimit
	}
	if n > cap(o.buf) {
		o.buf = make([]byte, 0, n)
	}
}

// result returns what was written, without a copy: the buffer must not be
// written again.
func (o *OutputBuffer) result() string {
	if len(o.buf) == 0 {
		return ""
	}
	return unsafe.String(&o.buf[0], len(o.buf))
}

// String returns what was written.
func (o *OutputBuffer) String() string {
	return string(o.buf)
}

// Len returns the number of bytes written.
func (o *OutputBuffer) Len() int {
	return len(o.buf)
}
