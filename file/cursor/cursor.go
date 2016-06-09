package cursor

import (
	"sync"
)

type Cursor struct {
	row, col, colwant int
}

func MakeCursor(row, col int) Cursor {
	return Cursor{
		row:     row,
		col:     col,
		colwant: col,
	}
}

func (cursor Cursor) Dup() Cursor {
	return Cursor{
		row:     cursor.row,
		col:     cursor.col,
		colwant: cursor.colwant,
	}
}

func (cursor Cursor) Row() int {
	return cursor.row
}

func (cursor Cursor) Col() int {
	return cursor.col
}

func (cursor Cursor) Colwant() int {
	return cursor.colwant
}

func (cursor Cursor) RowCol() (int, int) {
	return cursor.row, cursor.col
}

func (cursor *Cursor) Set(row, col, colwant int) {
	cursor.row = row
	cursor.col = col
	cursor.colwant = col
}

type MultiCursor struct {
	cursors []Cursor
	mutex   *sync.Mutex
}

func MakeMultiCursor() MultiCursor {
	return MultiCursor{
		cursors: []Cursor{MakeCursor(0, 0)},
		mutex:   &sync.Mutex{},
	}
}

func (mc MultiCursor) GetCursor(idx int) Cursor {
	return mc.cursors[idx]
}

func (mc MultiCursor) GetCursorRCC(idx int) (row, col, colwant int) {
	c := mc.cursors[idx]
	return c.row, c.col, c.colwant
}

func (mc MultiCursor) GetRow(idx int) int {
	return mc.cursors[idx].row
}

func (mc MultiCursor) GetCol(idx int) int {
	return mc.cursors[idx].col
}

func (mc MultiCursor) GetRowCol(idx int) (int, int) {
	return mc.cursors[idx].RowCol()
}

func (mc *MultiCursor) SetCursor(idx, row, col, colwant int) {
	mc.cursors[idx].row = row
	mc.cursors[idx].col = col
	mc.cursors[idx].colwant = colwant
}

func (mc *MultiCursor) Set(row, col, colwant int) {
	mc.SetCursor(0, row, col, colwant)
}

func (mc *MultiCursor) SetRow(idx, row int) {
	mc.cursors[idx].row = row
}

func (mc *MultiCursor) SetCol(idx, col int) {
	mc.cursors[idx].col = col
}

func (mc *MultiCursor) SetColwant(idx, colwant int) {
	if colwant < 0 {
		mc.cursors[idx].colwant = mc.cursors[idx].col
	} else {
		mc.cursors[idx].colwant = colwant
	}
}

func (mc *MultiCursor) ReplaceMC(mc2 MultiCursor) {
	mc.cursors = mc2.cursors
}

func (mc MultiCursor) Dup() MultiCursor {
	newCursors := make([]Cursor, mc.Length())
	for k, cursor := range mc.cursors {
		newCursors[k] = cursor.Dup()
	}
	return MultiCursor{
		cursors: newCursors,
		mutex:   &sync.Mutex{},
	}
}

func (mc MultiCursor) Length() int {
	return len(mc.cursors)
}

// Clear keeps only the first cursor.
func (mc *MultiCursor) Clear() {
	mc.cursors = mc.cursors[0:1]
}

// Add appends another cursor.
func (mc *MultiCursor) Append(cursor Cursor) {
	mc.cursors = append(mc.cursors, cursor)
}

func (mc *MultiCursor) Snapshot() {
	cursor := mc.cursors[0].Dup()
	mc.cursors = append(mc.cursors, cursor)
}

func (mc *MultiCursor) OuterMost() {
	if mc.Length() < 2 {
		return
	}
	minCursor := mc.cursors[0].Dup()
	maxCursor := mc.cursors[0].Dup()
	for _, cursor := range mc.cursors {
		if cursor.row < minCursor.row {
			minCursor = cursor.Dup()
		}
		if cursor.row >= maxCursor.row {
			maxCursor = cursor.Dup()
		}
	}
	mc.cursors = []Cursor{minCursor, maxCursor}
}

func (mc MultiCursor) MaxCol() int {
	maxCol := 0
	for _, cursor := range mc.cursors {
		if cursor.col > maxCol {
			maxCol = cursor.col
		}
	}
	return maxCol
}

// Return the smallest and largest row from all the cursors.
func (mc MultiCursor) MinMaxRow() (minRow, maxRow int) {
	minRow = mc.cursors[0].row
	maxRow = mc.cursors[0].row
	for _, cursor := range mc.cursors {
		if cursor.row < minRow {
			minRow = cursor.row
		}
		if cursor.row > maxRow {
			maxRow = cursor.row
		}
	}
	return minRow, maxRow
}

// Get FirstCursor returns the first cursor (by row).
func (mc MultiCursor) GetFirstCursor() Cursor {
	firstCursor := mc.cursors[0]
	for _, cursor := range mc.cursors {
		if cursor.row < firstCursor.row {
			firstCursor = cursor
		}
	}
	return firstCursor
}

// SetColumn sets the col of each cursor.
func (mc *MultiCursor) SetColumn() {
	col := mc.cursors[0].col
	minRow, maxRow := mc.MinMaxRow()
	mc.cursors = []Cursor{}
	for row := minRow; row <= maxRow; row++ {
		cursor := Cursor{row: row, col: col, colwant: col}
		mc.Append(cursor)
	}
}

func (mc MultiCursor) Cursors() []Cursor {
	return mc.cursors
}
