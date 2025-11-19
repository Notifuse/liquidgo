package liquid

// ForloopDrop provides information about a parent for loop.
type ForloopDrop struct {
	*Drop
	parentloop *ForloopDrop
	name       string
	length     int
	index      int
}

// NewForloopDrop creates a new ForloopDrop.
func NewForloopDrop(name string, length int, parentloop *ForloopDrop) *ForloopDrop {
	return &ForloopDrop{
		Drop:       NewDrop(),
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

// InvokeDrop invokes a method on the forloop drop.
func (f *ForloopDrop) InvokeDrop(methodOrKey string) interface{} {
	return InvokeDropOn(f, methodOrKey)
}
