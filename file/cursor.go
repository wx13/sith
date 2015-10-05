package file

type Cursor struct {
	row, col, colwant int
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

func (mc MultiCursor) Clear() MultiCursor {
	return mc[0:1]
}

func (mc MultiCursor) Add() MultiCursor {
	cursor := Cursor{
		row:     mc[0].row,
		col:     mc[0].col,
		colwant: mc[0].colwant,
	}
	mc = append(mc, cursor)
	return mc
}

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
