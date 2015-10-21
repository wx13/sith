package file

import "io/ioutil"
import "strings"
import "sync"
import "os"
import "fmt"
import "go/format"
import "github.com/nsf/termbox-go"
import "path"
import "regexp"
import "github.com/wx13/sith/syntaxcolor"
import "github.com/wx13/sith/terminal"
import "strconv"

type File struct {
	Buffer      Buffer
	buffMutex   *sync.Mutex
	MultiCursor MultiCursor
	savedBuffer Buffer

	buffHist    *BufferHist
	searchHist  []string
	replaceHist []string
	gotoHist    []string

	Name        string
	SyntaxRules *syntaxcolor.SyntaxRules
	fileMode    os.FileMode
	autoIndent  bool

	rowOffset int
	colOffset int
	screen    *terminal.Screen
	flushChan chan struct{}
}

func NewFile(name string, flushChan chan struct{}, screen *terminal.Screen) *File {
	file := &File{
		Name:        name,
		screen:      screen,
		fileMode:    os.FileMode(int(0644)),
		Buffer:      MakeBuffer([]string{""}),
		buffMutex:   &sync.Mutex{},
		MultiCursor: MakeMultiCursor(),
		flushChan:   flushChan,
		SyntaxRules: syntaxcolor.NewSyntaxRules(""),
		autoIndent:  true,
	}
	file.buffHist = NewBufferHist(file.Buffer, file.MultiCursor)
	go file.ReadFile(name)
	switch path.Ext(name) {
	case ".md", ".txt", ".csv", ".C":
		file.autoIndent = false
	}
	return file
}

func (file *File) Close() bool {
	if file.IsModified() {
		doClose, _ := file.screen.AskYesNo("File has been modified. Close anyway?")
		if !doClose {
			return false
		}
	}
	return true
}

func (file *File) ToggleAutoIndent() {
	file.autoIndent = file.autoIndent != true
}

func (file *File) Flush() {
	cols, rows := termbox.Size()
	slice := file.Slice(rows-1, cols)
	file.screen.Clear()
	for row, str := range slice {
		file.screen.WriteString(row, 0, str)
		strLine := tabs2spaces(file.Buffer[row+file.rowOffset]).toString()
		file.screen.Colorize(row, file.SyntaxRules.Colorize(strLine), file.colOffset)
	}
	for row := len(slice); row < rows-1; row++ {
		file.screen.WriteString(row, 0, "~")
	}
}

func (file *File) Refresh() {
	file.screen.ReallyClear()
}

func (file *File) ClearCursors() {
	file.MultiCursor = file.MultiCursor.Clear()
}

func (file *File) AddCursor() {
	file.MultiCursor = file.MultiCursor.Add()
}

func (file *File) AddCursorCol() {
	file.MultiCursor = file.MultiCursor.SetColumn()
}

// ReadFile reads in a file (if it exists).
func (file *File) ReadFile(name string) {

	fileInfo, err := os.Stat(name)
	if err != nil {
		file.Buffer = MakeBuffer([]string{""})
	} else {
		file.fileMode = fileInfo.Mode()

		byteBuf, err := ioutil.ReadFile(name)
		stringBuf := []string{""}
		if err == nil {
			stringBuf = strings.Split(string(byteBuf), "\n")
		}

		file.Buffer = MakeBuffer(stringBuf)
	}

	file.Snapshot()
	file.savedBuffer = file.Buffer.DeepDup()

	select {
	case file.flushChan <- struct{}{}:
	default:
	}

}

func (file *File) toString() string {
	return file.Buffer.ToString()
}

func (file *File) Save() string {
	contents := file.toString()
	err := ioutil.WriteFile(file.Name, []byte(contents), file.fileMode)
	if err != nil {
		return ("Could not save to file: " + file.Name)
	} else {
		file.savedBuffer = file.Buffer.DeepDup()
		return ("Saved to: " + file.Name)
	}
}

func (file *File) replaceBuffer(newBuffer Buffer) {
	for k, line := range newBuffer {
		if k > len(file.Buffer) {
			file.Buffer = append(file.Buffer, line)
		} else {
			if file.Buffer[k].toString() != line.toString() {
				file.Buffer[k] = line
			}
		}
	}
}

func (file *File) GoFmt() {
	contents := file.toString()
	bytes, err := format.Source([]byte(contents))
	if err == nil {
		stringBuf := strings.Split(string(bytes), "\n")
		newBuffer := MakeBuffer(stringBuf)
		file.replaceBuffer(newBuffer)
	}
	file.Snapshot()
}

func (file *File) IsModified() bool {
	if len(file.Buffer) != len(file.savedBuffer) {
		return true
	}
	for row, _ := range file.Buffer {
		if file.Buffer[row].toString() != file.savedBuffer[row].toString() {
			return true
		}
	}
	return false
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
		file.Buffer[row] = Line(string(line[0:col]) + string(ch) + string(line[col:]))
		file.MultiCursor[idx].col += 1
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
			file.Buffer[row] = Line(string(line[0:col-1]) + string(line[col:]))
			file.MultiCursor[idx].col = col - 1
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

func (file *File) CursorGoTo(row, col int) {
	file.MultiCursor[0].row = row
	file.MultiCursor[0].col = col
	file.EnforceRowBounds()
	file.EnforceColBounds()
}

func (file *File) PageDown() {
	_, rows := termbox.Size()
	file.CursorDown(rows/2-1)
}

func (file *File) PageUp() {
	_, rows := termbox.Size()
	file.CursorUp(rows/2-1)
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
	ws := Line(re.FindString(file.Buffer[row-1].toString()))
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

func tabs2spaces(line Line) Line {
	strLine := string(line)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return Line(strLine)
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

// Slice returns a 2D slice of the buffer.
func (file *File) Slice(nRows, nCols int) []string {

	file.UpdateOffsets(nRows, nCols)

	startRow := file.rowOffset
	endRow := nRows + file.rowOffset
	startCol := file.colOffset
	endCol := nCols + file.colOffset
	if endRow > len(file.Buffer) {
		endRow = len(file.Buffer)
	}
	if endRow <= startRow {
		return []string{}
	}

	slice := make([]string, endRow-startRow)
	for row := startRow; row < endRow; row++ {
		line := tabs2spaces(file.Buffer[row])
		rowEndCol := endCol
		if rowEndCol > len(line) {
			rowEndCol = len(line)
		}
		if rowEndCol <= startCol {
			slice[row-startRow] = ""
		} else {
			slice[row-startRow] = string(line[startCol:rowEndCol])
		}
	}

	return slice

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
			if col > 0 {
				line := file.Buffer[row].Dup()
				file.Buffer[row] = line[:col]
				for file.Buffer[row][0] == ' ' {
					file.Buffer[row] = file.Buffer[row][1:]
				}
				if row+1 == len(file.Buffer) {
					file.Buffer = append(file.Buffer, line[col:])
				} else {
					rest := append(line[col:], ' ')
					file.Buffer[row+1] = append(rest, file.Buffer[row+1].Dup()...)
					for file.Buffer[row+1][0] == ' ' {
						file.Buffer[row+1] = file.Buffer[row+1][1:]
					}
				}
			}
		}
	}
}

func (file *File) StartOfLine() {
	for idx, _ := range file.MultiCursor {
		file.MultiCursor[idx].col = 0
		file.MultiCursor[idx].colwant = file.MultiCursor[idx].col
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
			col += offset[0]+1
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

func (file *File) Snapshot() {
	file.buffHist.Snapshot(file.Buffer, file.MultiCursor)
}

func (file *File) Undo() {
	file.Buffer, file.MultiCursor = file.buffHist.Prev()
}

func (file *File) Redo() {
	file.Buffer, file.MultiCursor = file.buffHist.Next()
}

func (file *File) GetPromptAnswer(question string, history *[]string) string {
	answer, err := file.screen.Ask(question, *history)
	if err != nil {
		return ""
	}
	if answer == "" {
		if len(*history) == 0 {
			return ""
		}
		answer = (*history)[0]
	} else {
		*history = append([]string{answer}, *history...)
	}
	return answer
}

func (file *File) GoToLine() {
	lineNo := file.GetPromptAnswer("goto:", &file.gotoHist)
	if lineNo == "" {
		return
	}
	row, err := strconv.Atoi(lineNo)
	if err == nil {
		file.CursorGoTo(row,0)
	}
}

func (file *File) Search() {
	searchTerm := file.GetPromptAnswer("search:", &file.searchHist)
	if searchTerm == "" {
		return
	}
	row, col, err := file.Buffer.Search(searchTerm, file.MultiCursor[0])
	if err == nil {
		file.CursorGoTo(row, col)
	}
}

func (file *File) SearchAndReplace() {

	searchTerm := file.GetPromptAnswer("search:", &file.searchHist)
	if searchTerm == "" {
		return
	}

	replaceTerm := file.GetPromptAnswer("replace:", &file.replaceHist)
	if replaceTerm == "" {
		return
	}

	for {
		row, col, err := file.Buffer.Search(searchTerm, file.MultiCursor[0])
		if err != nil {
			break
		}

		file.CursorGoTo(row, col)
		_, screenCol := file.GetCursor(0)
		startColOffset := screenCol - col
		file.Flush()
		var startCol, endCol int
		startCol, endCol = file.Buffer[row].Search(searchTerm, col, -1)
		for c := startCol + startColOffset; c < endCol+startColOffset; c++ {
			file.screen.Highlight(row-file.rowOffset, c)
		}
		doReplace, err := file.screen.AskYesNo("Replace this instance?")
		for c := startCol; c < endCol; c++ {
			file.screen.Highlight(row-file.rowOffset, c)
		}
		if err != nil {
			break
		}
		if doReplace {
			file.Buffer.Replace(searchTerm, replaceTerm, row, col)
			file.screen.WriteString(row, 0, file.Buffer[row].toString())
		}
	}
}

func (file *File) ModStatus() string {
	if file.IsModified() {
		return "Modified"
	} else {
		return ""
	}
}

func (file *File) WriteStatus(row, col int) {

	status := file.ModStatus()
	col -= len(status) + 2
	fg := termbox.ColorYellow
	bg := termbox.ColorBlack
	file.screen.WriteStringColor(row, col, status, fg, bg)

	if len(file.MultiCursor) > 1 {
		status = fmt.Sprintf("%dC", len(file.MultiCursor))
		col -= len(status) + 2
		fg := termbox.ColorBlack
		bg := termbox.ColorRed
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}

	if file.autoIndent {
		status = "->"
		col -= len(status) + 2
		fg := termbox.ColorRed | termbox.AttrBold
		bg := termbox.ColorBlack
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}

}
