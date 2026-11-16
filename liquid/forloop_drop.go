package liquid

// ForloopDrop provides information about a parent for loop.
type ForloopDrop struct {
	parentloop *ForloopDrop
	name       string
	length     int
	index      int
}

// NewForloopDrop creates a new ForloopDrop.
func NewForloopDrop(name string, length int, parentloop *ForloopDrop) *ForloopDrop {
	return &ForloopDrop{
		name:       name,
		length:     length,
		parentloop: parentloop,
		index:      0,
	}
}

// Name returns the name of the loop.
func (f *ForloopDrop) Name() string {
	return f.name
}

// Length returns the total number of iterations in the loop.
func (f *ForloopDrop) Length() int {
	return f.length
}

// Parentloop returns the parent forloop object.
// Returns nil if the current for loop isn't nested inside another for loop.
func (f *ForloopDrop) Parentloop() *ForloopDrop {
	return f.parentloop
}

// Index returns the 1-based index of the current iteration.
func (f *ForloopDrop) Index() int {
	return f.index + 1
}

// Index0 returns the 0-based index of the current iteration.
func (f *ForloopDrop) Index0() int {
	return f.index
}

// Rindex returns the 1-based index of the current iteration, in reverse order.
func (f *ForloopDrop) Rindex() int {
	return f.length - f.index
}

// Rindex0 returns the 0-based index of the current iteration, in reverse order.
func (f *ForloopDrop) Rindex0() int {
	return f.length - f.index - 1
}

// First returns true if the current iteration is the first.
func (f *ForloopDrop) First() bool {
	return f.index == 0
}

// Last returns true if the current iteration is the last.
func (f *ForloopDrop) Last() bool {
	return f.index == f.length-1
}

// Increment increments the index (protected method).
func (f *ForloopDrop) Increment() {
	f.index++
}

// String writes the drop as Ruby writes it, by its class.
func (f *ForloopDrop) String() string { return "Liquid::ForloopDrop" }

// LiquidGet returns the key of forloop a template reads, as Ruby's
// ForloopDrop.
func (f *ForloopDrop) LiquidGet(_ *Context, key string) (interface{}, bool, error) {
	switch key {
	case "name":
		return f.Name(), true, nil
	case "length":
		return f.Length(), true, nil
	case "parentloop":
		if f.parentloop == nil {
			return nil, true, nil
		}
		return f.parentloop, true, nil
	case "index":
		return f.Index(), true, nil
	case "index0":
		return f.Index0(), true, nil
	case "rindex":
		return f.Rindex(), true, nil
	case "rindex0":
		return f.Rindex0(), true, nil
	case "first":
		return f.First(), true, nil
	case "last":
		return f.Last(), true, nil
	}
	return nil, false, nil
}
