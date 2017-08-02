package cursor

import (
	"sync"
)

type NavMode int

const (
	Column      NavMode = 0
	AllTogether NavMode = 1
	Detached    NavMode = 2
)

// MultiCursor is a set of Cursors.
type MultiCursor struct {
	cursors []Cursor
	mutex   *sync.Mutex
	navMode NavMode
}

// MakeMultiCursor Creates a new MultiCursor.
func MakeMultiCursor() MultiCursor {
	return MultiCursor{
		cursors: []Cursor{MakeCursor(0, 0)},
		mutex:   &sync.Mutex{},
		navMode: Column,
	}
}

// CycleNavMode cycles through the three nav modes.
func (mc *MultiCursor) CycleNavMode() {
	mc.navMode = (mc.navMode + 1) % 3
}

// GetNavMode returns the current navigation mode.
func (mc MultiCursor) GetNavMode() NavMode {
	return mc.navMode
}

// NavModeIsColumn returns true if navMode is Column.
func (mc MultiCursor) NavModeIsColumn() bool {
	return mc.navMode == Column
}

// NavModeIsAllTogether returns true if navMode is AllTogether.
func (mc MultiCursor) NavModeIsAllTogether() bool {
	return mc.navMode == AllTogether
}

// NavModeIsDetached returns true if navMode is Detached.
func (mc MultiCursor) NavModeIsDetached() bool {
	return mc.navMode == Detached
}

// GetNavModeShort returns a 1 character representation of the nav mode.
func (mc MultiCursor) GetNavModeShort() string {
	switch mc.navMode {
	case Column:
		return "C"
	case AllTogether:
		return "A"
	case Detached:
		return "D"
	default:
		return ""
	}
}

// GetCursor returns a cursor by index.
func (mc MultiCursor) GetCursor(idx int) Cursor {
	if idx > len(mc.cursors) {
		idx = len(mc.cursors) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return mc.cursors[idx]
}

// GetCursorRCC gets the (row, col, colwant) of a cursor by index.
func (mc MultiCursor) GetCursorRCC(idx int) (row, col, colwant int) {
	c := mc.GetCursor(idx)
	return c.row, c.col, c.colwant
}

func (mc MultiCursor) GetRow(idx int) int {
	return mc.GetCursor(idx).row
}

func (mc MultiCursor) GetCol(idx int) int {
	return mc.GetCursor(idx).col
}

func (mc MultiCursor) GetRowCol(idx int) (int, int) {
	return mc.GetCursor(idx).RowCol()
}

// SetCursor sets the position of the cursor identified by index.
func (mc *MultiCursor) SetCursor(idx, row, col, colwant int) {
	if idx < 0 || idx > len(mc.cursors) {
		return
	}
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

// Cursors returns the list of cursors.
func (mc MultiCursor) Cursors() []Cursor {
	return mc.cursors
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

// GetRowsCols returns a map of cursor positions. The map keys are rows (integers),
// and the values are lists of cursor positions.
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
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	cursor := mc.cursors[0].Dup()
	// If cursor matches an existing cursor, remove it instead of
	// adding a new one.
	if len(mc.cursors) > 1 {
		toggle := false
		for k, c := range mc.cursors {
			if k == 0 {
				continue
			}
			if c.row == cursor.row && c.col == cursor.col {
				toggle = true
				if k < len(mc.cursors)-1 {
					mc.cursors = append(mc.cursors[:k], mc.cursors[k+1:]...)
				} else {
					mc.cursors = mc.cursors[:k]
				}
			}
		}
		if toggle {
			return
		}
	}
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
