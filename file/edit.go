package file

import (
	"errors"
	"go/format"
	"regexp"
	"strings"

	"github.com/wx13/sith/file/buffer"
)

func (file *File) GoFmt() error {
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
	return nil
}

func (file *File) InsertChar(ch rune) {
	maxCol := 0
	maxLineLen := 0
	for _, cursor := range file.MultiCursor.Cursors() {
		if cursor.Col() > maxCol {
			maxCol = cursor.Col()
		}
		if file.buffer.RowLength(cursor.Row()) > maxLineLen {
			maxLineLen = file.buffer.RowLength(cursor.Row())
		}
	}
	for idx, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		if maxCol > 0 && col == 0 {
			continue
		}
		line := file.buffer.GetRow(row)
		if (ch == ' ' || ch == '\t') && col == 0 && line.Length() == 0 && maxLineLen > 0 {
			continue
		}
		insertStr := string(ch)
		if ch == '\t' && file.autoTab && file.tabString != "\t" {
			insertStr = file.tabString
		}
		newLine := buffer.MakeLine(line.Slice(0, col).ToString() + insertStr + line.Slice(col, -1).ToString())
		file.buffer.SetRow(row, newLine)
		col += len(insertStr)
		file.MultiCursor.SetCursor(idx, row, col, col)
	}
	file.Snapshot()
}

func (file *File) Backspace() {
	for idx, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		if col == 0 {
			if file.MultiCursor.Length() > 1 {
				continue
			}
			if row == 0 {
				return
			}
			row -= 1
			if row+1 >= file.buffer.Length() {
				return
			}
			col = file.buffer.RowLength(row)
			newLine := buffer.MakeLine(file.buffer.GetRow(row).ToString() + file.buffer.GetRow(row+1).ToString())
			file.buffer.ReplaceLine(newLine, row)
			file.buffer.DeleteRow(row + 1)
			file.MultiCursor.SetCursor(idx, row, col, col)
		} else {
			line := file.buffer.GetRow(row)
			if col > line.Length() {
				continue
			}

			// Handle multi-char indents.
			nDel := 1
			if file.autoTab && len(file.tabString) > 0 {
				if line.Slice(0, col).ToString() == strings.Repeat(" ", col) {
					n := len(file.tabString)
					if n*(col/n) == col {
						nDel = n
					}
				}
			}

			newLine := buffer.MakeLine(line.Slice(0, col-nDel).ToString() + line.Slice(col, -1).ToString())
			file.buffer.SetRow(row, newLine)
			col -= nDel
			file.MultiCursor.SetCursor(idx, row, col, col)
		}
	}
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
}

func (file *File) Delete() {
	file.CursorRight()
	file.Backspace()
}

func (file *File) Newline() {

	rate := file.timer.Tick()

	for idx, cursor := range file.MultiCursor.Cursors() {

		row, col := cursor.RowCol()
		lineStart := file.buffer.RowSlice(row, 0, col)
		lineEnd := file.buffer.RowSlice(row, col, -1)
		newLines := []buffer.Line{lineStart, lineEnd}

		file.buffer.ReplaceLines(newLines, row, row)

		file.MultiCursor.SetCursor(idx, row+1, 0, 0)

		if file.autoIndent && rate < file.maxRate && lineEnd.Length() == 0 {
			file.DoAutoIndent(idx)
		}

		file.buffer.SetRow(row, lineStart.RemoveTrailingWhitespace())

	}

	file.Snapshot()
}

func (file *File) DoAutoIndent(idx int) {

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

func (file *File) Justify(lineLen int) {
	minRow, maxRow := file.MultiCursor.MinMaxRow()
	bigString := file.buffer.InclSlice(minRow, maxRow).ToString(" ")
	splitBuf := buffer.MakeSplitBuffer(bigString, lineLen)
	file.buffer.ReplaceLines(splitBuf.Lines(), minRow, maxRow)
	file.MultiCursor.Clear()
	file.Snapshot()
}

func (file *File) Cut() []string {
	row := file.MultiCursor.GetRow(0)
	cutLines := file.buffer.InclSlice(row, row).Dup()
	strs := make([]string, cutLines.Length())
	for idx, line := range cutLines.Lines() {
		strs[idx] = line.ToString()
	}
	file.buffer.DeleteRow(row)
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
	return strs
}

func (file *File) Paste(strs []string) {
	row := file.MultiCursor.GetRow(0)
	pasteLines := make([]buffer.Line, len(strs))
	for idx, str := range strs {
		pasteLines[idx] = buffer.MakeLine(str)
	}
	file.buffer.InsertAfter(row-1, pasteLines...)
	file.CursorDown(len(pasteLines))
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
}

func (file *File) CutToStartOfLine() {
	for idx, _ := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		line := file.buffer.GetRow(row).Slice(col, -1)
		file.buffer.SetRow(row, line)
		file.MultiCursor.SetCursor(idx, row, 0, 0)
	}
	file.Snapshot()
}

func (file *File) CutToEndOfLine() {
	for idx, _ := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		line := file.buffer.GetRow(row).Slice(0, col)
		file.buffer.SetRow(row, line)
	}
	file.Snapshot()
}

func (file *File) CursorAlign() {
	maxCol := file.MultiCursor.MaxCol()
	for idx, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		nSpaces := maxCol - col
		spaces := strings.Repeat(" ", nSpaces)
		line := file.buffer.GetRow(row)
		newLine := buffer.MakeLine(line.Slice(0, col).ToString() + spaces + line.Slice(col, -1).ToString())
		file.buffer.SetRow(row, newLine)
		col += len(spaces)
		file.MultiCursor.SetCursor(idx, row, col, col)
	}
	file.Snapshot()
}
