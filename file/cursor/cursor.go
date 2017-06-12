// Package cursor handles all the cursor and multicursor stuff.
package cursor

import (
	"sync"
)

// Cursor keeps track of the row, col of a cursor,
// and also the wanted column (colwant) of the cursor.
// The wanted column is the column the cursor would like
// to be in, if the line were long enough.
type Cursor struct {
	row, col, colwant int
}

// MakeCursor creates a new Cursor object.
func MakeCursor(row, col int) Cursor {
	return Cursor{
		row:     row,
		col:     col,
		colwant: col,
	}
}

// CursorFromSlice creates a new Cursor object, from a slice of ints.
// The input slice is if the form [row, col, colwant].
func CursorFromSlice(s []int) Cursor {
	cursor := Cursor{}
	switch len(s) {
	case 3:
		cursor.colwant = s[2]
		fallthrough
	case 2:
		cursor.col = s[1]
		fallthrough
	case 1:
		cursor.row = s[0]
	}
	return cursor
}

// Dup duplicates a cursor object.
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

// RowCol returns the row and column of the cursor.
func (cursor Cursor) RowCol() (int, int) {
	return cursor.row, cursor.col
}

func (cursor *Cursor) Set(row, col, colwant int) {
	cursor.row = row
	cursor.col = col
	cursor.colwant = col
}

// MultiCursor is a set of Cursors.
type MultiCursor struct {
	cursors []Cursor
	mutex   *sync.Mutex
}

// MakeMultiCursor Creates a new MultiCursor.
func MakeMultiCursor() MultiCursor {
	return MultiCursor{
		cursors: []Cursor{MakeCursor(0, 0)},
		mutex:   &sync.Mutex{},
	}
}

// GetCursor returns a cursor by index.
func (mc MultiCursor) GetCursor(idx int) Cursor {
	return mc.cursors[idx]
}

// GetCursorRCC gets the (row, col, colwant) of a cursor by index.
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

// SetCursor sets the position of the cursor identified by index.
func (mc *MultiCursor) SetCursor(idx, row, col, colwant int) {
	mc.cursors[idx].row = row
	mc.cursors[idx].col = col
	mc.cursors[idx].colwant = colwant
}

// ResetCursors manually sets all the cursor positions. This is useful
// for a full cursor reset.
func (mc *MultiCursor) ResetCursors(rows map[int][]int) {
	mc.cursors = []Cursor{}
	for row, cols := range rows {
		for _, col := range cols {
			mc.cursors = append(mc.cursors, MakeCursor(row, col))
		}
	}
	if len(mc.cursors) == 0 {
		mc.cursors = []Cursor{MakeCursor(0, 0)}
	}
}

// ResetRowsCols creates a new MultiCursor object
// from a map of row, column positions.
func (mc *MultiCursor) ResetRowsCols(rowcol map[int][]int) {
	mc.cursors = []Cursor{}
	for row, cols := range rowcol {
		for _, col := range cols {
			mc.cursors = append(mc.cursors, MakeCursor(row, col))
		}
	}
	if len(mc.cursors) == 0 {
		mc.cursors = []Cursor{MakeCursor(0, 0)}
	}
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

// ReplaceMC sets the list of cursors to be the list of
// cursors from another MC object.
func (mc *MultiCursor) ReplaceMC(mc2 MultiCursor) {
	mc.cursors = mc2.cursors
}

// Dup duplicates a MC object.
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

// GetRows returns a list of (integer) rows that contain cursors.
func (mc MultiCursor) GetRows() []int {
	rowmap := make(map[int]bool)
	for _, cursor := range mc.cursors {
		rowmap[cursor.Row()] = true
	}
	rows := []int{}
	for row := range rowmap {
		rows = append(rows, row)
	}
	return rows
}

// GetRowsCols returns a map of cursor positions.
func (mc MultiCursor) GetRowsCols() map[int][]int {
	rows := map[int][]int{}
	for _, cursor := range mc.cursors {
		r, c := cursor.RowCol()
		_, exist := rows[r]
		if !exist {
			rows[r] = []int{}
		}
		rows[r] = append(rows[r], c)
	}
	return rows
}

// Length returns the number of cursors.
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

// Snapshot adds the current primary cursor row,col as a new
// cursor to the list.  It is the primary way the user will
// add cursors to the MC cursor list.
func (mc *MultiCursor) Snapshot() {
	cursor := mc.cursors[0].Dup()
	mc.cursors = append(mc.cursors, cursor)
}

// OuterMost removes all cursors except the the most extreme.
// Length will be 2 after this operation.
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

// MaxCol returns the largest column value among cursors.
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

// OnePerLine keeps only the first cursor on each line.
func (mc *MultiCursor) OnePerLine() {

	// Create a map of rows for deduplication.
	cols := make(map[int]int)
	for _, cursor := range mc.cursors {
		r, c := cursor.RowCol()
		_, exist := cols[r]
		if !exist {
			cols[r] = c
		}
		if c < cols[r] {
			cols[r] = c
		}
	}

	// Recreate the cursors from the map.
	mc.cursors = []Cursor{}
	for r, c := range cols {
		mc.cursors = append(mc.cursors, MakeCursor(r, c))
	}

}
