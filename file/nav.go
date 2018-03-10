package file

import (
	"regexp"

	"github.com/nsf/termbox-go"
)

func (file *File) enforceColBounds(indexes ...int) {
	if len(indexes) == 0 {
		for idx, _ := range file.MultiCursor.Cursors() {
			indexes = append(indexes, idx)
		}
	}
	for _, idx := range indexes {
		cursor := file.MultiCursor.GetCursor(idx)
		if cursor.Col() > file.buffer.RowLength(cursor.Row()) {
			file.MultiCursor.SetCol(idx, file.buffer.RowLength(cursor.Row()))
		}
		if cursor.Col() < 0 {
			file.MultiCursor.SetCol(idx, 0)
		}
	}
}

func (file *File) enforceRowBounds(indexes ...int) {
	if len(indexes) == 0 {
		for idx, _ := range file.MultiCursor.Cursors() {
			indexes = append(indexes, idx)
		}
	}
	for _, idx := range indexes {
		cursor := file.MultiCursor.GetCursor(idx)
		if cursor.Row() >= file.buffer.Length() {
			file.MultiCursor.SetRow(idx, file.buffer.Length()-1)
		}
		if cursor.Row() < 0 {
			file.MultiCursor.SetRow(idx, 0)
		}
	}
}

func (file *File) makeCursorNotAtTopBottom() {
	row := file.MultiCursor.GetRow(0)
	_, rows := termbox.Size()
	bottom := file.rowOffset + rows - 1
	if row >= bottom {
		file.rowOffset += (row - bottom) + rows/8
	}
}

// CursorGoTo moves the cursor to a row, col position. If row is negative, then
// it specifies from the end of the file.
func (file *File) CursorGoTo(row, col int) {
	if row < 0 {
		row = file.Length() + row
	}
	file.MultiCursor.Set(row, col, col)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.makeCursorNotAtTopBottom()
}

// PageDown moves the cursor half a screen down.
func (file *File) PageDown() {
	_, rows := termbox.Size()
	file.CursorDown(rows/2 - 1)
}

// PageUp moves the cursor have a screen up.
func (file *File) PageUp() {
	_, rows := termbox.Size()
	file.CursorUp(rows/2 - 1)
}

// CursorUp moves the cursor up n rows.
func (file *File) CursorUp(n int) {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() || file.MultiCursor.NavModeIsColumn() {
		cursors = cursors[:1]
	}
	for idx := range cursors {
		row, _, colwant := file.MultiCursor.GetCursorRCC(idx)
		row -= n
		if row < 0 {
			row = 0
		}
		file.MultiCursor.SetCursor(idx, row, colwant, colwant)
	}
	file.enforceRowBounds()
	file.enforceColBounds()
}

// CursorDown moves the cursor down n rows.
func (file *File) CursorDown(n int) {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() || file.MultiCursor.NavModeIsColumn() {
		cursors = cursors[:1]
	}
	for idx := range cursors {
		row, _, colwant := file.MultiCursor.GetCursorRCC(idx)
		row += n
		if row >= file.buffer.Length() {
			row = file.buffer.Length() - 1
		}
		file.MultiCursor.SetCursor(idx, row, colwant, colwant)
	}
	file.enforceRowBounds()
	file.enforceColBounds()
}

// CursorRight moves the cursor one column to the right.
func (file *File) CursorRight() {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx, cursor := range cursors {
		row, col := cursor.RowCol()
		if col < file.buffer.RowLength(row) {
			file.MultiCursor.SetCol(idx, col+1)
		} else {
			if file.MultiCursor.Length() > 1 {
				continue
			}
			if row < file.buffer.Length()-1 {
				file.MultiCursor.SetRow(idx, row+1)
				file.MultiCursor.SetCol(idx, 0)
			}
		}
		file.MultiCursor.SetColwant(idx, -1)
	}
	file.enforceRowBounds()
	file.enforceColBounds()
}

// CursorLeft moves the cursor one column to the left.
func (file *File) CursorLeft() {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx, cursor := range cursors {
		row, col := cursor.RowCol()
		if col > 0 {
			file.MultiCursor.SetCol(idx, col-1)
		} else {
			if file.MultiCursor.Length() > 1 {
				continue
			}
			if row > 0 {
				row--
				col = file.buffer.RowLength(row)
				file.MultiCursor.SetCursor(idx, row, col, col)
			}
		}
		file.MultiCursor.SetColwant(idx, -1)
	}
}

// GetCursor returns the row, col position for the specified multi-cursor index.
func (file *File) GetCursor(idx int) (int, int) {
	file.enforceRowBounds(idx)
	file.enforceColBounds(idx)
	row, col, _ := file.MultiCursor.GetCursorRCC(idx)
	line := file.buffer.GetRow(row).Slice(0, col).Tabs2spaces(file.tabWidth)
	n := file.screen.StringDispLen(line.ToString())
	return row - file.rowOffset, n - file.colOffset
}

func (file *File) GetRowCol(idx int) (int, int) {
	file.enforceRowBounds(idx)
	file.enforceColBounds(idx)
	row, col, _ := file.MultiCursor.GetCursorRCC(idx)
	return row, col
}

// ScrollLeft shifts the view screen to the left.
func (file *File) ScrollLeft() {
	file.colOffset++
}

// ScrollRight shifts the view screen to the right.
func (file *File) ScrollRight() {
	if file.colOffset > 0 {
		file.colOffset--
	}
}

// ScrollUp shifts the screen up one row.
func (file *File) ScrollUp() {
	if file.rowOffset < file.buffer.Length()-1 {
		file.rowOffset++
	}
}

// ScrollDown shifts the screen down one row.
func (file *File) ScrollDown() {
	if file.rowOffset > 0 {
		file.rowOffset--
	}
}

func (file *File) updateOffsets(nRows, nCols int) {

	row := file.MultiCursor.GetRow(0)
	if row < file.rowOffset {
		file.rowOffset = row
	}
	if row >= file.rowOffset+nRows-1 {
		file.rowOffset = row - nRows + 1
	}

	_, col := file.GetCursor(0)
	col += file.colOffset
	if col < file.colOffset {
		file.colOffset = col
	}
	if col >= file.colOffset+nCols-1 {
		file.colOffset = col - nCols + 1
	}

}

// StartOfLine moves the cursors to the start of the line.
// If they are already at the start, the moves them to the first
// non-whitespace character.
func (file *File) StartOfLine() {
	allAtZero := true
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for _, cursor := range cursors {
		if cursor.Col() != 0 {
			allAtZero = false
			break
		}
	}
	if allAtZero {
		re := regexp.MustCompile("^[ \t]*")
		for idx, cursor := range cursors {
			row := cursor.Row()
			line := file.buffer.GetRow(row)
			match := re.FindStringIndex(line.ToString())
			file.MultiCursor.SetCol(idx, match[1])
			file.MultiCursor.SetColwant(idx, -1)
		}
	} else {
		for idx := range cursors {
			file.MultiCursor.SetCol(idx, 0)
			file.MultiCursor.SetColwant(idx, -1)
		}
	}
}

// EndOfLine moves the cursors to the end of the line.
func (file *File) EndOfLine() {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx := range cursors {
		row := file.MultiCursor.GetRow(idx)
		line := file.buffer.GetRow(row)
		file.MultiCursor.SetCol(idx, line.Length())
		file.MultiCursor.SetColwant(idx, -1)
	}
}

// NextWord moves the cursor to the next word.
func (file *File) NextWord() {
	file.prevNextWord(1)
}

// PrevWord moves the cursor to the previous word.
func (file *File) PrevWord() {
	file.prevNextWord(-1)
}

func (file *File) prevNextWord(incr int) {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx, cursor := range cursors {
		row := cursor.Row()
		col := cursor.Col()
		line := file.buffer.GetRow(row)
		col = line.PrevNextWord(col, incr)
		file.MultiCursor.SetCol(idx, col)
		file.MultiCursor.SetColwant(idx, -1)
	}
}
