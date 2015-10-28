package file

import "github.com/nsf/termbox-go"
import "strings"
import "regexp"
import "strconv"

func (file *File) EnforceColBounds() {
	for idx, cursor := range file.MultiCursor {
		if cursor.col > len(file.Buffer[cursor.row]) {
			file.MultiCursor[idx].col = len(file.Buffer[cursor.row])
		}
		if cursor.col < 0 {
			file.MultiCursor[idx].col = 0
		}
	}
}

func (file *File) EnforceRowBounds() {
	for idx, cursor := range file.MultiCursor {
		if cursor.row >= len(file.Buffer) {
			file.MultiCursor[idx].row = len(file.Buffer) - 1
		}
		if cursor.row < 0 {
			file.MultiCursor[idx].row = 0
		}
	}
}

func (file *File) MakeCursorNotAtTopBottom() {
	row := file.MultiCursor[0].row
	_, rows := termbox.Size()
	bottom := file.rowOffset + rows - 1
	if row >= bottom {
		file.rowOffset += (row - bottom) + rows/8
	}
}

func (file *File) CursorGoTo(row, col int) {
	file.MultiCursor[0].row = row
	file.MultiCursor[0].col = col
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
	file.MultiCursor[0].row -= n
	if file.MultiCursor[0].row < 0 {
		file.MultiCursor[0].row = 0
	}
	file.MultiCursor[0].col = file.MultiCursor[0].colwant
	file.EnforceColBounds()
}

func (file *File) CursorDown(n int) {
	file.MultiCursor[0].row += n
	if file.MultiCursor[0].row >= len(file.Buffer) {
		file.MultiCursor[0].row = len(file.Buffer) - 1
	}
	file.MultiCursor[0].col = file.MultiCursor[0].colwant
	file.EnforceColBounds()
}

func (file *File) CursorRight() {
	for idx, cursor := range file.MultiCursor {
		if cursor.col < len(file.Buffer[cursor.row]) {
			file.MultiCursor[idx].col += 1
		} else {
			if len(file.MultiCursor) > 1 {
				continue
			}
			if cursor.row < len(file.Buffer)-1 {
				file.MultiCursor[idx].row += 1
				file.MultiCursor[idx].col = 0
			}
		}
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
	}
	file.EnforceRowBounds()
	file.EnforceColBounds()
}

func (file *File) CursorLeft() {
	for idx, cursor := range file.MultiCursor {
		if cursor.col > 0 {
			file.MultiCursor[idx].col -= 1
		} else {
			if len(file.MultiCursor) > 1 {
				continue
			}
			if cursor.row > 0 {
				file.MultiCursor[idx].row -= 1
				file.MultiCursor[idx].col = len(file.Buffer[file.MultiCursor[idx].row])
			}
		}
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
	}
}

func (file *File) GetCursor(idx int) (int, int) {
	file.EnforceRowBounds()
	file.EnforceColBounds()
	line := file.Buffer[file.MultiCursor[idx].row][0:file.MultiCursor[idx].col]
	strLine := string(line)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return file.MultiCursor[idx].row - file.rowOffset, len(strLine) - file.colOffset
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
	if file.rowOffset < len(file.Buffer)-1 {
		file.rowOffset += 1
	}
}

func (file *File) ScrollDown() {
	if file.rowOffset > 0 {
		file.rowOffset -= 1
	}
}

func (file *File) UpdateOffsets(nRows, nCols int) {

	if file.MultiCursor[0].row < file.rowOffset {
		file.rowOffset = file.MultiCursor[0].row
	}
	if file.MultiCursor[0].row >= file.rowOffset+nRows-1 {
		file.rowOffset = file.MultiCursor[0].row - nRows + 1
	}

	if file.MultiCursor[0].col < file.colOffset {
		file.colOffset = file.MultiCursor[0].col
	}
	if file.MultiCursor[0].col >= file.colOffset+nCols-1 {
		file.colOffset = file.MultiCursor[0].col - nCols + 1
	}

}

func (file *File) StartOfLine() {
	allAtZero := true
	for _, cursor := range file.MultiCursor {
		if cursor.col != 0 {
			allAtZero = false
			break
		}
	}
	if allAtZero {
		re := regexp.MustCompile("^[ \t]*")
		for idx, cursor := range file.MultiCursor {
			row := cursor.row
			line := file.Buffer[row]
			match := re.FindStringIndex(line.toString())
			file.MultiCursor[idx].col = match[1]
			file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
		}
	} else {
		for idx, _ := range file.MultiCursor {
			file.MultiCursor[idx].col = 0
			file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
		}
	}
}

func (file *File) EndOfLine() {
	for idx, _ := range file.MultiCursor {
		row := file.MultiCursor[idx].row
		file.MultiCursor[idx].col = len(file.Buffer[row])
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
	}
}

func (file *File) NextWord() {
	for idx, cursor := range file.MultiCursor {
		row := cursor.row
		line := file.Buffer[row]
		col := cursor.col
		re := regexp.MustCompile("[\t ][^\t ]")
		offset := re.FindStringIndex(line[col:].toString())
		if offset == nil {
			col = len(line)
		} else {
			col += offset[0] + 1
		}
		file.MultiCursor[idx].col = col
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
	}
}

func (file *File) PrevWord() {
	for idx, cursor := range file.MultiCursor {
		row := cursor.row
		line := file.Buffer[row]
		col := cursor.col
		re := regexp.MustCompile("[\t ][^\t ]")
		offsets := re.FindAllStringIndex(line[:col].toString(), -1)
		if offsets == nil {
			col = 0
		} else {
			offset := offsets[len(offsets)-1]
			col = offset[0] + 1
		}
		file.MultiCursor[idx].col = col
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
	}
}

func (file *File) GoToLine() {
	lineNo := file.GetPromptAnswer("goto:", &file.gotoHist)
	if lineNo == "" {
		return
	}
	row, err := strconv.Atoi(lineNo)
	if err == nil {
		file.CursorGoTo(row, 0)
	}
}