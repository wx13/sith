package file

import "go/format"
import "strings"
import "regexp"
import "errors"

func (file *File) replaceBuffer(newBuffer Buffer) {
	for k, line := range newBuffer {
		if k > len(file.Buffer) {
			file.Buffer = append(file.Buffer, line)
		} else {
			if file.Buffer[k].ToString() != line.ToString() {
				file.Buffer[k] = line
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
		if len(file.Buffer[cursor.row]) > maxLineLen {
			maxLineLen = len(file.Buffer[cursor.row])
		}
	}
	for idx, cursor := range file.MultiCursor {
		col, row := cursor.col, cursor.row
		if maxCol > 0 && col == 0 {
			continue
		}
		line := file.Buffer[row]
		if (ch == ' ' || ch == '\t') && col == 0 && len(line) == 0 && maxLineLen > 0 {
			continue
		}
		insertStr := string(ch)
		if ch == '\t' && file.autoTab && file.tabString != "\t" {
			insertStr = file.tabString
		}
		file.Buffer[row] = Line(string(line[0:col]) + insertStr + string(line[col:]))
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
			if row+1 >= len(file.Buffer) {
				return
			}
			col = len(file.Buffer[row])
			file.Buffer[row] = append(file.Buffer[row], file.Buffer[row+1]...)
			file.Buffer = append(file.Buffer[0:row+1], file.Buffer[row+2:]...)
			file.MultiCursor[idx].col = col
			file.MultiCursor[idx].row = row
		} else {
			line := file.Buffer[row]
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

			file.Buffer[row] = Line(string(line[0:col-nDel]) + string(line[col:]))
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
	for idx, cursor := range file.MultiCursor {
		col, row := cursor.col, cursor.row
		lineStart := file.Buffer[row][0:col]
		lineEnd := file.Buffer[row][col:]
		file.Buffer[row] = lineStart
		file.Buffer = append(file.Buffer, Line(""))
		copy(file.Buffer[row+2:], file.Buffer[row+1:])
		file.Buffer[row+1] = lineEnd
		file.MultiCursor[idx].row = row + 1
		file.MultiCursor[idx].col = 0
		if file.autoIndent {
			file.DoAutoIndent(idx)
		}
	}
	file.Snapshot()
}

func (file *File) DoAutoIndent(cursorIdx int) {

	row := file.MultiCursor[cursorIdx].row
	if row == 0 {
		return
	}

	origLine := file.Buffer[row].Dup()

	// Whitespace-only indent.
	re, _ := regexp.Compile("^[ \t]+")
	ws := Line(re.FindString(file.Buffer[row-1].ToString()))
	if len(ws) > 0 {
		file.Buffer[row] = append(ws, file.Buffer[row]...)
		file.MultiCursor[cursorIdx].col += len(ws)
		if len(file.Buffer[row-1]) == len(ws) {
			file.Buffer[row-1] = Line("")
		}
	}

	if row < 2 {
		return
	}

	// Non-whitespace indent.
	indent := file.Buffer[row-1].CommonStart(file.Buffer[row-2])
	if len(indent) > len(ws) {
		file.Snapshot()
		file.Buffer[row] = append(indent, origLine...)
		file.MultiCursor[cursorIdx].col += len(indent) - len(ws)
	}

}

func (file *File) Justify() {
	minRow, maxRow := file.MultiCursor.MinMaxRow()
	for row := minRow; row <= maxRow; row++ {
		if len(file.Buffer[row]) > 72 {
			col := 72
			for ; col >= 0; col-- {
				r := file.Buffer[row][col]
				if r == ' ' || r == '\t' {
					break
				}
			}
			if col <= 0 {
				continue
			}
			line := file.Buffer[row].Dup()
			file.Buffer[row] = line[:col]
			for file.Buffer[row][0] == ' ' {
				file.Buffer[row] = file.Buffer[row][1:]
			}
			if row == maxRow {
				rest := line[col:]
				file.Buffer = append(file.Buffer, Line(""))
				copy(file.Buffer[row+2:], file.Buffer[row+1:])
				file.Buffer[row+1] = rest
				for file.Buffer[row+1][0] == ' ' {
					file.Buffer[row+1] = file.Buffer[row+1][1:]
				}
				if len(file.Buffer[row+1]) > 72 {
					maxRow++
				}
			} else {
				rest := append(line[col:], ' ')
				file.Buffer[row+1] = append(rest, file.Buffer[row+1].Dup()...)
				for file.Buffer[row+1][0] == ' ' {
					file.Buffer[row+1] = file.Buffer[row+1][1:]
				}
			}
		}
	}
	file.MultiCursor = file.MultiCursor.Clear()
	file.Snapshot()
}

func (file *File) Cut() Buffer {
	row := file.MultiCursor[0].row
	cutBuffer := file.Buffer[row : row+1].Dup()
	if len(file.Buffer) == 1 {
		file.Buffer = MakeBuffer([]string{""})
	} else if row == 0 {
		file.Buffer = file.Buffer[1:]
	} else if row < len(file.Buffer)-1 {
		file.Buffer = append(file.Buffer[:row], file.Buffer[row+1:]...)
	} else {
		file.Buffer = file.Buffer[:row]
	}
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
	return cutBuffer
}

func (file *File) Paste(buffer Buffer) {
	row := file.MultiCursor[0].row
	newBuffer := file.Buffer[:row].Dup()
	for _, line := range buffer {
		newBuffer = append(newBuffer, line.Dup())
	}
	file.Buffer = append(newBuffer, file.Buffer[row:].Dup()...)
	file.CursorDown(len(buffer))
	file.EnforceRowBounds()
	file.EnforceColBounds()
	file.Snapshot()
}
