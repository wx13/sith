package file

import (
	"regexp"
	"strconv"
	"unicode"

	"github.com/nsf/termbox-go"
)

func (file *File) EnforceColBounds() {
	for idx, cursor := range file.MultiCursor.Cursors() {
		if cursor.Col() > file.buffer.RowLength(cursor.Row()) {
			file.MultiCursor.SetCol(idx, file.buffer.RowLength(cursor.Row()))
		}
		if cursor.Col() < 0 {
			file.MultiCursor.SetCol(idx, 0)
		}
	}
}

func (file *File) EnforceRowBounds() {
	for idx, cursor := range file.MultiCursor.Cursors() {
		if cursor.Row() >= file.buffer.Length() {
			file.MultiCursor.SetRow(idx, file.buffer.Length()-1)
		}
		if cursor.Row() < 0 {
			file.MultiCursor.SetRow(idx, 0)
		}
	}
}

func (file *File) MakeCursorNotAtTopBottom() {
	row := file.MultiCursor.GetRow(0)
	_, rows := termbox.Size()
	bottom := file.rowOffset + rows - 1
	if row >= bottom {
		file.rowOffset += (row - bottom) + rows/8
	}
}

func (file *File) CursorGoTo(row, col int) {
	file.MultiCursor.Set(row, col, col)
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.MakeCursorNotAtTopBottom()
}

func (file *File) PageDown() {
	_, rows := termbox.Size()
	file.CursorDown(rows/2 - 1)
}

func (file *File) PageUp() {
	_, rows := termbox.Size()
	file.CursorUp(rows/2 - 1)
}

func (file *File) CursorUp(n int) {
	row, _, colwant := file.MultiCursor.GetCursorRCC(0)
	row -= n
	if row < 0 {
		row = 0
	}
	file.MultiCursor.SetCursor(0, row, colwant, colwant)
	file.EnforceColBounds()
}

func (file *File) CursorDown(n int) {
	row, _, colwant := file.MultiCursor.GetCursorRCC(0)
	row += n
	if row >= file.buffer.Length() {
		row = file.buffer.Length() - 1
	}
	file.MultiCursor.SetCursor(0, row, colwant, colwant)
	file.EnforceColBounds()
}

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
	file.EnforceRowBounds()
	file.EnforceColBounds()
}

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
				row -= 1
				col = file.buffer.RowLength(row)
				file.MultiCursor.SetCursor(idx, row, col, col)
			}
		}
		file.MultiCursor.SetColwant(idx, -1)
	}
}

func (file *File) GetCursor(idx int) (int, int) {
	file.EnforceRowBounds()
	file.EnforceColBounds()
	row, col, _ := file.MultiCursor.GetCursorRCC(idx)
	line := file.buffer.GetRow(row).Slice(0, col).Tabs2spaces()
	n := file.screen.StringDispLen(line.ToString())
	return row - file.rowOffset, n - file.colOffset
}

func (file *File) ScrollLeft() {
	file.colOffset += 1
}

func (file *File) ScrollRight() {
	if file.colOffset > 0 {
		file.colOffset -= 1
	}
}

func (file *File) ScrollUp() {
	if file.rowOffset < file.buffer.Length()-1 {
		file.rowOffset += 1
	}
}

func (file *File) ScrollDown() {
	if file.rowOffset > 0 {
		file.rowOffset -= 1
	}
}

func (file *File) UpdateOffsets(nRows, nCols int) {

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
		for idx, _ := range file.MultiCursor.Cursors() {
			file.MultiCursor.SetCol(idx, 0)
			file.MultiCursor.SetColwant(idx, -1)
		}
	}
}

func (file *File) EndOfLine() {
	for idx, _ := range file.MultiCursor.Cursors() {
		row := file.MultiCursor.GetRow(idx)
		line := file.buffer.GetRow(row)
		file.MultiCursor.SetCol(idx, line.Length())
		file.MultiCursor.SetColwant(idx, -1)
	}
}

func (file *File) NextWord() {
	file.PrevNextWord(1)
}

func (file *File) PrevWord() {
	file.PrevNextWord(-1)
}

func isLetter(r rune) bool {
	return !(unicode.IsPunct(r) || unicode.IsSpace(r))
}

func (file *File) PrevNextWord(incr int) {
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
