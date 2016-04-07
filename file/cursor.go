package file

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

type MultiCursor []Cursor

func MakeMultiCursor() MultiCursor {
	return MultiCursor{Cursor{}}
}

func (mc MultiCursor) Dup() MultiCursor {
	mcCopy := make(MultiCursor, len(mc))
	for k, cursor := range mc {
		mcCopy[k] = cursor.Dup()
	}
	return mcCopy
}

// Clear keeps only the first cursor.
func (mc MultiCursor) Clear() MultiCursor {
	return mc[0:1]
}

// Add appends another cursor.
func (mc MultiCursor) Add() MultiCursor {
	cursor := Cursor{
		row:     mc[0].row,
		col:     mc[0].col,
		colwant: mc[0].colwant,
	}
	mc = append(mc, cursor)
	return mc
}

// Outermost returns a multicursor object with just the
// first and last cursor (by row, and in order).
func (mc MultiCursor) OuterMost() MultiCursor {
	if len(mc) == 1 {
		return mc
	}
	minCursor := mc[0].Dup()
	maxCursor := mc[0].Dup()
	for _, cursor := range mc {
		if cursor.row < minCursor.row {
			minCursor = cursor.Dup()
		}
		if cursor.row >= maxCursor.row {
			maxCursor = cursor.Dup()
		}
	}
	return []Cursor{minCursor, maxCursor}
}

// Return the smallest and largest row from all the cursors.
func (mc MultiCursor) MinMaxRow() (minRow, maxRow int) {
	minRow = mc[0].row
	maxRow = mc[0].row
	for _, cursor := range mc {
		if cursor.row < minRow {
			minRow = cursor.row
		}
		if cursor.row > maxRow {
			maxRow = cursor.row
		}
	}
	return
}

// Get FirstCursor returns the first cursor (by row).
func (mc MultiCursor) GetFirstCursor() Cursor {
	firstCursor := mc[0]
	for _, cursor := range mc {
		if cursor.row < firstCursor.row {
			firstCursor = cursor
		}
	}
	return firstCursor
}

// SetColumn sets the col of each cursor.
func (mc MultiCursor) SetColumn() MultiCursor {
	col := mc[0].col
	minRow, maxRow := mc.MinMaxRow()
	mc = MultiCursor{}
	for row := minRow; row <= maxRow; row++ {
		cursor := Cursor{row: row, col: col, colwant: col}
		mc = append(mc, cursor)
	}
	return mc
}
