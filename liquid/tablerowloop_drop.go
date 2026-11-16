package liquid

// TablerowloopDrop provides information about a parent tablerow loop.
type TablerowloopDrop struct {
	length int
	row    int
	col    int
	cols   int
	index  int
}

// NewTablerowloopDrop creates a new TablerowloopDrop.
func NewTablerowloopDrop(length int, cols int) *TablerowloopDrop {
	return &TablerowloopDrop{
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

// String writes the drop as Ruby writes it, by its class.
func (t *TablerowloopDrop) String() string { return "Liquid::TablerowloopDrop" }

// LiquidGet returns the key of tablerowloop a template reads, as Ruby's
// TablerowloopDrop.
func (t *TablerowloopDrop) LiquidGet(_ *Context, key string) (interface{}, bool, error) {
	switch key {
	case "length":
		return t.Length(), true, nil
	case "col":
		return t.Col(), true, nil
	case "row":
		return t.Row(), true, nil
	case "index":
		return t.Index(), true, nil
	case "index0":
		return t.Index0(), true, nil
	case "col0":
		return t.Col0(), true, nil
	case "rindex":
		return t.Rindex(), true, nil
	case "rindex0":
		return t.Rindex0(), true, nil
	case "first":
		return t.First(), true, nil
	case "last":
		return t.Last(), true, nil
	case "col_first":
		return t.ColFirst(), true, nil
	case "col_last":
		return t.ColLast(), true, nil
	}
	return nil, false, nil
}
