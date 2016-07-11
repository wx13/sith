package file

import (
	"regexp"
	"strconv"
	"unicode"

	"github.com/nsf/termbox-go"
)

func (file *File) enforceColBounds() {
	for idx, cursor := range file.MultiCursor.Cursors() {
		if cursor.Col() > file.buffer.RowLength(cursor.Row()) {
			file.MultiCursor.SetCol(idx, file.buffer.RowLength(cursor.Row()))
		}
		if cursor.Col() < 0 {
			file.MultiCursor.SetCol(idx, 0)
		}
	}
}

func (file *File) enforceRowBounds() {
	for idx, cursor := range file.MultiCursor.Cursors() {
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

// CursorGoTo moves the cursor to a row, col position.
func (file *File) CursorGoTo(row, col int) {
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
	row, _, colwant := file.MultiCursor.GetCursorRCC(0)
	row -= n
	if row < 0 {
		row = 0
	}
	file.MultiCursor.SetCursor(0, row, colwant, colwant)
	file.enforceColBounds()
}

// CursorDown moves the cursor down n rows.
func (file *File) CursorDown(n int) {
	row, _, colwant := file.MultiCursor.GetCursorRCC(0)
	row += n
	if row >= file.buffer.Length() {
		row = file.buffer.Length() - 1
	}
	file.MultiCursor.SetCursor(0, row, colwant, colwant)
	file.enforceColBounds()
}

// CursorRight moves the cursor one column to the right.
func (file *File) CursorRight() {
	for idx, cursor := range file.MultiCursor.Cursors() {
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
	for idx, cursor := range file.MultiCursor.Cursors() {
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
	file.enforceRowBounds()
	file.enforceColBounds()
	row, col, _ := file.MultiCursor.GetCursorRCC(idx)
	line := file.buffer.GetRow(row).Slice(0, col).Tabs2spaces()
	n := file.screen.StringDispLen(line.ToString())
	return row - file.rowOffset, n - file.colOffset
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
	for _, cursor := range file.MultiCursor.Cursors() {
		if cursor.Col() != 0 {
			allAtZero = false
			break
		}
	}
	if allAtZero {
		re := regexp.MustCompile("^[ \t]*")
		for idx, cursor := range file.MultiCursor.Cursors() {
			row := cursor.Row()
			line := file.buffer.GetRow(row)
			match := re.FindStringIndex(line.ToString())
			file.MultiCursor.SetCol(idx, match[1])
			file.MultiCursor.SetColwant(idx, -1)
		}
	} else {
		for idx := range file.MultiCursor.Cursors() {
			file.MultiCursor.SetCol(idx, 0)
			file.MultiCursor.SetColwant(idx, -1)
		}
	}
}

// EndOfLine moves the cursors to the end of the line.
func (file *File) EndOfLine() {
	for idx := range file.MultiCursor.Cursors() {
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

func isLetter(r rune) bool {
	return !(unicode.IsPunct(r) || unicode.IsSpace(r))
}

func (file *File) prevNextWord(incr int) {
	for idx, cursor := range file.MultiCursor.Cursors() {
		row := cursor.Row()
		col := cursor.Col()
		line := file.buffer.GetRow(row)
		r := line.GetChar(col)
		if isLetter(r) {
			for ; col <= line.Length() && col >= 0; col += incr {
				r = line.GetChar(col)
				if !isLetter(r) {
					break
				}
			}
		}
		for ; col <= line.Length() && col >= 0; col += incr {
			r = line.GetChar(col)
			if isLetter(r) {
				break
			}
		}
		file.MultiCursor.SetCol(idx, col)
		file.MultiCursor.SetColwant(idx, -1)
	}
}

// GoToLine prompts the user for a row number, and the puts the cursor
// on that row.
func (file *File) GoToLine() {
	lineNo := file.screen.GetPromptAnswer("goto:", &file.gotoHist)
	if lineNo == "" {
		return
	}
	row, err := strconv.Atoi(lineNo)
	if err == nil {
		file.CursorGoTo(row, 0)
	}
}
