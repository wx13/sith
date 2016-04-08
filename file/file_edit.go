package file

import (
	"errors"
	"go/format"
	"regexp"
	"strings"
)

func (file *File) replaceBuffer(newBuffer Buffer) {
	n := len(newBuffer)
	file.buffer = file.buffer[:n]
	for k, line := range newBuffer {
		if k > len(file.buffer) {
			file.buffer = append(file.buffer, line)
		} else {
			if file.buffer[k].ToString() != line.ToString() {
				file.buffer[k] = line
			}
		}
	}
}

func (file *File) GoFmt() error {
	filetype := file.SyntaxRules.GetFileType(file.Name)
	if filetype != "go" {
		return errors.New("Will not gofmt a non-go file.")
	}
	contents := file.toString()
	bytes, err := format.Source([]byte(contents))
	if err == nil {
		stringBuf := strings.Split(string(bytes), file.newline)
		newBuffer := MakeBuffer(stringBuf)
		file.replaceBuffer(newBuffer)
	}
	file.Snapshot()
	return nil
}

func (file *File) InsertChar(ch rune) {
	maxCol := 0
	maxLineLen := 0
	for _, cursor := range file.MultiCursor {
		if cursor.col > maxCol {
			maxCol = cursor.col
		}
		if len(file.buffer[cursor.row]) > maxLineLen {
			maxLineLen = len(file.buffer[cursor.row])
		}
	}
	for idx, cursor := range file.MultiCursor {
		col, row := cursor.col, cursor.row
		if maxCol > 0 && col == 0 {
			continue
		}
		line := file.buffer[row]
		if (ch == ' ' || ch == '\t') && col == 0 && len(line) == 0 && maxLineLen > 0 {
			continue
		}
		insertStr := string(ch)
		if ch == '\t' && file.autoTab && file.tabString != "\t" {
			insertStr = file.tabString
		}
		file.buffer[row] = Line(string(line[0:col]) + insertStr + string(line[col:]))
		file.MultiCursor[idx].col += len(insertStr)
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
	}
	file.Snapshot()
}

func (file *File) Backspace() {
	for idx, cursor := range file.MultiCursor {
		col, row := cursor.col, cursor.row
		if col == 0 {
			if len(file.MultiCursor) > 1 {
				continue
			}
			if row == 0 {
				return
			}
			row -= 1
			if row+1 >= len(file.buffer) {
				return
			}
			col = len(file.buffer[row])
			file.buffer[row] = append(file.buffer[row], file.buffer[row+1]...)
			file.buffer = append(file.buffer[0:row+1], file.buffer[row+2:]...)
			file.MultiCursor[idx].col = col
			file.MultiCursor[idx].row = row
		} else {
			line := file.buffer[row]
			if col > len(line) {
				continue
			}

			// Handle multi-char indents.
			nDel := 1
			if file.autoTab && len(file.tabString) > 0 {
				if string(line[0:col]) == strings.Repeat(" ", col) {
					n := len(file.tabString)
					if n*(col/n) == col {
						nDel = n
					}
				}
			}

			file.buffer[row] = Line(string(line[0:col-nDel]) + string(line[col:]))
			file.MultiCursor[idx].col = col - nDel
			file.MultiCursor[idx].row = row
		}
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
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

	for idx, cursor := range file.MultiCursor {

		col, row := cursor.col, cursor.row
		lineStart := file.buffer[row][0:col]
		lineEnd := file.buffer[row][col:]

		file.buffer = append(file.buffer, Line(""))
		copy(file.buffer[row+2:], file.buffer[row+1:])
		file.buffer[row+1] = lineEnd

		file.MultiCursor[idx].row = row + 1
		file.MultiCursor[idx].col = 0

		if file.autoIndent && rate < file.maxRate && len(lineEnd) == 0 {
			file.DoAutoIndent(idx)
		}

		file.buffer[row] = lineStart.RemoveTrailingWhitespace()

	}

	file.Snapshot()
}

func (file *File) DoAutoIndent(cursorIdx int) {

	row := file.MultiCursor[cursorIdx].row
	if row == 0 {
		return
	}

	origLine := file.buffer[row].Dup()

	// Whitespace-only indent.
	re, _ := regexp.Compile("^[ \t]+")
	ws := Line(re.FindString(file.buffer[row-1].ToString()))
	if len(ws) > 0 {
		file.buffer[row] = append(ws, file.buffer[row]...)
		file.MultiCursor[cursorIdx].col += len(ws)
		if len(file.buffer[row-1]) == len(ws) {
			file.buffer[row-1] = Line("")
		}
	}

	if row < 2 {
		return
	}

	// Non-whitespace indent.
	indent := file.buffer[row-1].CommonStart(file.buffer[row-2])
	if len(indent) > len(ws) {
		file.ForceSnapshot()
		file.buffer[row] = append(indent, origLine...)
		file.MultiCursor[cursorIdx].col += len(indent) - len(ws)
	}

}

func (file *File) Justify(lineLen int) {
	minRow, maxRow := file.MultiCursor.MinMaxRow()
	lines := file.buffer[minRow : maxRow+1]
	bigString := lines.ToString(" ")
	lines = MakeSplitBuffer(bigString, lineLen)
	file.buffer = file.buffer.ReplaceLines(lines, minRow, maxRow)
	file.MultiCursor = file.MultiCursor.Clear()
	file.Snapshot()
}

func (file *File) Cut() Buffer {
	row := file.MultiCursor[0].row
	cutBuffer := file.buffer[row : row+1].Dup()
	if len(file.buffer) == 1 {
		file.buffer = MakeBuffer([]string{""})
	} else if row == 0 {
		file.buffer = file.buffer[1:]
	} else if row < len(file.buffer)-1 {
		file.buffer = append(file.buffer[:row], file.buffer[row+1:]...)
	} else {
		file.buffer = file.buffer[:row]
	}
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
	return cutBuffer
}

func (file *File) Paste(buffer Buffer) {
	row := file.MultiCursor[0].row
	newBuffer := file.buffer[:row].Dup()
	for _, line := range buffer {
		newBuffer = append(newBuffer, line.Dup())
	}
	file.buffer = append(newBuffer, file.buffer[row:].Dup()...)
	file.CursorDown(len(buffer))
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
}

func (file *File) CutToStartOfLine() {
	for idx, _ := range file.MultiCursor {
		row := file.MultiCursor[idx].row
		col := file.MultiCursor[idx].col
		file.buffer[row] = file.buffer[row][col:]
		file.MultiCursor[idx].col = 0
	}
	file.Snapshot()
}

func (file *File) CutToEndOfLine() {
	for idx, _ := range file.MultiCursor {
		row := file.MultiCursor[idx].row
		col := file.MultiCursor[idx].col
		file.buffer[row] = file.buffer[row][:col]
	}
	file.Snapshot()
}
