package file

import (
	"errors"
	"go/format"
	"regexp"
	"strings"

	"github.com/wx13/sith/file/buffer"
)

// Fmt runs a code formatter on the text buffer and updates the buffer.
func (file *File) Fmt() error {
	filetype := file.SyntaxRules.GetFileType(file.Name)
	if filetype != "go" {
		return errors.New("Will not gofmt a non-go file.")
	}
	contents := file.ToString()
	bytes, err := format.Source([]byte(contents))
	if err == nil {
		stringBuf := strings.Split(string(bytes), file.newline)
		newBuffer := buffer.MakeBuffer(stringBuf)
		file.buffer.ReplaceBuffer(newBuffer)
	}
	file.Snapshot()
	return err
}

func getMaxCol(rows map[int][]int) int {
	maxCol := 0
	for _, cols := range rows {
		for _, col := range cols {
			if col > maxCol {
				maxCol = col
			}
		}
	}
	return maxCol
}

func (file File) allBlankLines(rows map[int][]int) bool {
	for row, _ := range rows {
		if file.buffer.GetRow(row).Length() > 0 {
			return false
		}
	}
	return true
}

func (file File) removeBlankLineCursors(rows map[int][]int) (map[int][]int, []int) {

	if file.allBlankLines(rows) {
		return rows, []int{}
	}

	blankRows := []int{}
	for row, _ := range rows {
		if file.buffer.GetRow(row).Length() == 0 {
			blankRows = append(blankRows, row)
			delete(rows, row)
			continue
		}
	}

	return rows, blankRows
}

// InsertChar insters a character (rune) into the current cursor position.
func (file *File) InsertChar(ch rune) {

	str := string(ch)
	if ch == '\t' && file.autoTab && file.tabString != "\t" {
		str = file.tabString
	}

	rows := file.MultiCursor.GetRowsCols()
	var blankRows []int
	rows, blankRows = file.removeBlankLineCursors(rows)
	rows = file.buffer.InsertStr(str, rows)
	for _, row := range blankRows {
		rows[row] = []int{0}
	}
	file.MultiCursor.ResetCursors(rows)

	file.Snapshot()

}

func allColsZero(rows map[int][]int) bool {
	for _, cols := range rows {
		for _, col := range cols {
			if col > 0 {
				return false
			}
		}
	}
	return true
}

// Backspace removes the character before the cursor.
func (file *File) Backspace() {
	rows := file.MultiCursor.GetRowsCols()
	if allColsZero(rows) {
		rows = file.buffer.DeleteNewlines(rows)
	} else {
		rows = file.buffer.DeleteChars(-1, rows)
	}
	file.MultiCursor.ResetCursors(rows)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
}

// Delete deletes the character under the cursor.
func (file *File) Delete() {
	file.CursorRight()
	file.Backspace()
}

// Newline breaks the current line into two.
func (file *File) Newline() {

	if len(file.MultiCursor.Cursors()) == 1 {
		rate := file.timer.Tick()
		cursor := file.MultiCursor.Cursors()[0]
		row, col := cursor.RowCol()
		lineStart := file.buffer.RowSlice(row, 0, col)
		lineEnd := file.buffer.RowSlice(row, col, -1)
		newLines := []buffer.Line{lineStart, lineEnd}

		file.buffer.ReplaceLines(newLines, row, row)

		file.MultiCursor.SetCursor(0, row+1, 0, 0)

		if file.autoIndent && rate < file.maxRate && lineEnd.Length() == 0 {
			file.doAutoIndent(0)
		}

		file.buffer.SetRow(row, lineStart.RemoveTrailingWhitespace())

	} else {
		rows := file.MultiCursor.GetRowsCols()
		rows = file.buffer.InsertNewlines(rows)
		file.MultiCursor.ResetCursors(rows)
	}

	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
}

func (file *File) doAutoIndent(idx int) {

	row := file.MultiCursor.GetRow(idx)
	if row == 0 {
		return
	}

	origLine := file.buffer.GetRow(row).Dup()

	// Whitespace-only indent.
	re, _ := regexp.Compile("^[ \t]+")
	prevLineStr := file.buffer.GetRow(row - 1).ToString()
	ws := re.FindString(prevLineStr)
	if len(ws) > 0 {
		newLineStr := ws + file.buffer.GetRow(row).ToString()
		file.buffer.SetRow(row, buffer.MakeLine(newLineStr))
		col := file.MultiCursor.GetCol(idx) + len(ws)
		file.MultiCursor.SetCursor(idx, row, col, col)
		if file.buffer.GetRow(row-1).Length() == len(ws) {
			file.buffer.SetRow(row-1, buffer.MakeLine(""))
		}
	}

	if row < 2 {
		return
	}

	// Non-whitespace indent.
	indent := file.buffer.GetRow(row - 1).CommonStart(file.buffer.GetRow(row - 2))
	if indent.Length() > len(ws) {
		file.ForceSnapshot()
		newLineStr := indent.ToString() + origLine.ToString()
		file.buffer.SetRow(row, buffer.MakeLine(newLineStr))
		col := file.MultiCursor.GetCol(idx) + indent.Length() - len(ws)
		file.MultiCursor.SetCursor(idx, row, col, col)
	}

}

// Justify justifies the marked text.
func (file *File) Justify(lineLen int) {
	minRow, maxRow := file.MultiCursor.MinMaxRow()
	bigString := file.buffer.InclSlice(minRow, maxRow).ToString(" ")
	splitBuf := buffer.MakeSplitBuffer(bigString, lineLen)
	file.buffer.ReplaceLines(splitBuf.Lines(), minRow, maxRow)
	file.MultiCursor.Clear()
	file.Snapshot()
}

// Cut cuts the current line and adds to the copy buffer.
func (file *File) Cut() []string {
	row := file.MultiCursor.GetRow(0)
	cutLines := file.buffer.InclSlice(row, row).Dup()
	strs := make([]string, cutLines.Length())
	for idx, line := range cutLines.Lines() {
		strs[idx] = line.ToString()
	}
	file.buffer.DeleteRow(row)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
	return strs
}

// Paste inserts the copy buffer into buffer at the current line.
func (file *File) Paste(strs []string) {
	row := file.MultiCursor.GetRow(0)
	pasteLines := make([]buffer.Line, len(strs))
	for idx, str := range strs {
		pasteLines[idx] = buffer.MakeLine(str)
	}
	file.buffer.InsertAfter(row-1, pasteLines...)
	file.CursorDown(len(pasteLines))
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
}

// CutToStartOfLine cuts the text from the cursor to the start of the line.
func (file *File) CutToStartOfLine() {
	for idx := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		line := file.buffer.GetRow(row).Slice(col, -1)
		file.buffer.SetRow(row, line)
		file.MultiCursor.SetCursor(idx, row, 0, 0)
	}
	file.Snapshot()
}

// CutToEndOfLine cuts the text from the cursor to the end of the line.
func (file *File) CutToEndOfLine() {
	for idx := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		line := file.buffer.GetRow(row).Slice(0, col)
		file.buffer.SetRow(row, line)
	}
	file.Snapshot()
}

// CursorAlign inserts spaces into each cursor position, in order to
// align the cursors vertically.
func (file *File) CursorAlign() {
	rows := file.MultiCursor.GetRowsCols()
	rows = file.buffer.Align(rows)
	file.MultiCursor.ResetCursors(rows)
	file.Snapshot()
}

// CursorUnalign removes whitespace (except for 1 space) immediately preceding
// each cursor position.  Effectively, it undoes a CursorAlign.
func (file *File) CursorUnalign() {
	for idx, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		col = file.buffer.CompressPriorSpaces(row, col)
		file.MultiCursor.SetCursor(idx, row, col, col)
	}
}
