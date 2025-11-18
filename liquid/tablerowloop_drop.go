package liquid

// TablerowloopDrop provides information about a parent tablerow loop.
type TablerowloopDrop struct {
	*Drop
	length int
	row    int
	col    int
	cols   int
	index  int
}

// NewTablerowloopDrop creates a new TablerowloopDrop.
func NewTablerowloopDrop(length int, cols int) *TablerowloopDrop {
	return &TablerowloopDrop{
		Drop:   NewDrop(),
		length: length,
		row:    1,
		col:    1,
		cols:   cols,
		index:  0,
	}
}

// Length returns the total number of iterations in the loop.
func (t *TablerowloopDrop) Length() int {
	return t.length
}

// Col returns the 1-based index of the current column.
func (t *TablerowloopDrop) Col() int {
	return t.col
}

// Row returns the 1-based index of current row.
func (t *TablerowloopDrop) Row() int {
	return t.row
}

// Index returns the 1-based index of the current iteration.
func (t *TablerowloopDrop) Index() int {
	return t.index + 1
}

// Index0 returns the 0-based index of the current iteration.
func (t *TablerowloopDrop) Index0() int {
	return t.index
}

// Col0 returns the 0-based index of the current column.
func (t *TablerowloopDrop) Col0() int {
	return t.col - 1
}

// Rindex returns the 1-based index of the current iteration, in reverse order.
func (t *TablerowloopDrop) Rindex() int {
	return t.length - t.index
}

// Rindex0 returns the 0-based index of the current iteration, in reverse order.
func (t *TablerowloopDrop) Rindex0() int {
	return t.length - t.index - 1
}

// First returns true if the current iteration is the first.
func (t *TablerowloopDrop) First() bool {
	return t.index == 0
}

// Last returns true if the current iteration is the last.
func (t *TablerowloopDrop) Last() bool {
	return t.index == t.length-1
}

// ColFirst returns true if the current column is the first in the row.
func (t *TablerowloopDrop) ColFirst() bool {
	return t.col == 1
}

// ColLast returns true if the current column is the last in the row.
func (t *TablerowloopDrop) ColLast() bool {
	return t.col == t.cols
}

// Cols returns the number of columns.
func (t *TablerowloopDrop) Cols() int {
	return t.cols
}

// Increment increments the index and updates row/col.
func (t *TablerowloopDrop) Increment() {
	t.index++

	if t.col == t.cols {
		t.col = 1
		t.row++
	} else {
		t.col++
	}
}

// InvokeDrop invokes a method on the tablerowloop drop.
func (t *TablerowloopDrop) InvokeDrop(methodOrKey string) interface{} {
	return InvokeDropOn(t, methodOrKey)
}

