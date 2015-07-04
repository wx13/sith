package main

import "io/ioutil"
import "strings"
import "sync"
import "os"
import "fmt"

type File struct {
	buffer      Buffer
	buffMutex   *sync.Mutex
	multiCursor MultiCursor
	savedBuffer Buffer

	buffHist    *BufferHist
	searchHist  []string
	replaceHist []string

	name      string
	fileMode  os.FileMode
	rowOffset int
	colOffset int
	screen    *Screen
	flushChan chan struct{}
}

func NewFile(name string, flushChan chan struct{}, screen *Screen) *File {
	file := &File{
		name:        name,
		screen:      screen,
		fileMode:    os.FileMode(int(0644)),
		buffer:      MakeBuffer([]string{""}),
		buffMutex:   &sync.Mutex{},
		multiCursor: MakeMultiCursor(),
		flushChan:   flushChan,
	}
	file.buffHist = NewBufferHist(file.buffer, file.multiCursor)
	go file.ReadFile(name)
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

func (file *File) Refresh() {
	file.screen.Clear()
	file.screen.Flush()
}

func (file *File) ClearCursors() {
	file.multiCursor = file.multiCursor.Clear()
}

func (file *File) AddCursor() {
	file.multiCursor = file.multiCursor.Add()
}

func (file *File) AddCursorCol() {
	file.multiCursor = file.multiCursor.SetColumn()
}

// ReadFile reads in a file (if it exists).
func (file *File) ReadFile(name string) {

	fileInfo, err := os.Stat(name)
	if err != nil {
		file.buffer = MakeBuffer([]string{""})
		return
	}
	file.fileMode = fileInfo.Mode()

	byteBuf, err := ioutil.ReadFile(name)
	stringBuf := []string{""}
	if err == nil {
		stringBuf = strings.Split(string(byteBuf), "\n")
	}

	file.buffer = MakeBuffer(stringBuf)
	file.Snapshot()
	file.savedBuffer = file.buffer.DeepDup()

	select {
	case file.flushChan <- struct{}{}:
	default:
	}

}

func (file *File) toString() string {
	return file.buffer.ToString()
}

func (file *File) Save() string {
	contents := file.toString()
	err := ioutil.WriteFile(file.name, []byte(contents), file.fileMode)
	if err != nil {
		return ("Could not save to file: " + file.name)
	} else {
		file.savedBuffer = file.buffer.DeepDup()
		return ("Saved to: " + file.name)
	}
}

func (file *File) IsModified() bool {
	if len(file.buffer) != len(file.savedBuffer) {
		return true
	}
	for row, _ := range file.buffer {
		if file.buffer[row].toString() != file.savedBuffer[row].toString() {
			return true
		}
	}
	return false
}

// AddChar inserts a character at the current cursor position.
func (file *File) InsertChar(ch rune) {
	for idx, cursor := range file.multiCursor {
		col, row := cursor.col, cursor.row
		line := file.buffer[row]
		file.buffer[row] = Line(string(line[0:col]) + string(ch) + string(line[col:]))
		file.multiCursor[idx].col += 1
		file.multiCursor[idx].colwant = file.multiCursor[idx].col
	}
	file.Snapshot()
}

func (file *File) Backspace() {
	for idx, cursor := range file.multiCursor {
		col, row := cursor.col, cursor.row
		if col == 0 {
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
			file.multiCursor[idx].col = col
			file.multiCursor[idx].row = row
		} else {
			line := file.buffer[row]
			if col > len(line) {
				continue
			}
			file.buffer[row] = Line(string(line[0:col-1]) + string(line[col:]))
			file.multiCursor[idx].col = col - 1
			file.multiCursor[idx].row = row
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

func (file *File) EnforceColBounds() {
	for idx, cursor := range file.multiCursor {
		if cursor.col > len(file.buffer[cursor.row]) {
			file.multiCursor[idx].col = len(file.buffer[cursor.row])
		}
		if cursor.col < 0 {
			file.multiCursor[idx].col = 0
		}
	}
}

func (file *File) EnforceRowBounds() {
	for idx, cursor := range file.multiCursor {
		if cursor.row >= len(file.buffer) {
			file.multiCursor[idx].row = len(file.buffer) - 1
		}
		if cursor.row < 0 {
			file.multiCursor[idx].row = 0
		}
	}
}

func (file *File) CursorGoTo(row, col int) {
	file.multiCursor[0].row = row
	file.multiCursor[0].col = col
	file.EnforceRowBounds()
	file.EnforceColBounds()
}

func (file *File) CursorUp(n int) {
	file.multiCursor[0].row -= n
	if file.multiCursor[0].row < 0 {
		file.multiCursor[0].row = 0
	}
	file.multiCursor[0].col = file.multiCursor[0].colwant
	file.EnforceColBounds()
}

func (file *File) CursorDown(n int) {
	file.multiCursor[0].row += n
	if file.multiCursor[0].row >= len(file.buffer) {
		file.multiCursor[0].row = len(file.buffer) - 1
	}
	file.multiCursor[0].col = file.multiCursor[0].colwant
	file.EnforceColBounds()
}

func (file *File) CursorRight() {
	for idx, cursor := range file.multiCursor {
		if cursor.col < len(file.buffer[cursor.row]) {
			file.multiCursor[idx].col += 1
		} else {
			if cursor.row < len(file.buffer)-1 {
				file.multiCursor[idx].row += 1
				file.multiCursor[idx].col = 0
			}
		}
		file.multiCursor[idx].colwant = file.multiCursor[idx].col
	}
	file.EnforceRowBounds()
	file.EnforceColBounds()
}

func (file *File) CursorLeft() {
	for idx, cursor := range file.multiCursor {
		if cursor.col > 0 {
			file.multiCursor[idx].col -= 1
		} else {
			if cursor.row > 0 {
				file.multiCursor[idx].row -= 1
				file.multiCursor[idx].col = len(file.buffer[file.multiCursor[idx].row])
			}
		}
		file.multiCursor[idx].colwant = file.multiCursor[idx].col
	}
}

func (file *File) Newline() {
	for idx, cursor := range file.multiCursor {
		col, row := cursor.col, cursor.row
		lineStart := file.buffer[row][0:col]
		lineEnd := file.buffer[row][col:]
		file.buffer[row] = lineStart
		file.buffer = append(file.buffer, Line(""))
		copy(file.buffer[row+2:], file.buffer[row+1:])
		file.buffer[row+1] = lineEnd
		file.multiCursor[idx].row = row + 1
		file.multiCursor[idx].col = 0
	}
	file.Snapshot()
}

func (file *File) GetCursor(idx int) (int, int) {
	file.EnforceRowBounds()
	file.EnforceColBounds()
	line := file.buffer[file.multiCursor[idx].row][0:file.multiCursor[idx].col]
	strLine := string(line)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return file.multiCursor[idx].row - file.rowOffset, len(strLine) - file.colOffset
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
	if file.rowOffset < len(file.buffer)-1 {
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

// Slice returns a 2D slice of the buffer.
func (file *File) Slice(nRows, nCols int) []string {

	if file.multiCursor[0].row < file.rowOffset {
		file.rowOffset = file.multiCursor[0].row
	}
	if file.multiCursor[0].row >= file.rowOffset+nRows-1 {
		file.rowOffset = file.multiCursor[0].row - nRows + 1
	}

	if file.multiCursor[0].col < file.colOffset {
		file.colOffset = file.multiCursor[0].col
	}
	if file.multiCursor[0].col >= file.colOffset+nCols-1 {
		file.colOffset = file.multiCursor[0].col - nCols + 1
	}

	startRow := file.rowOffset
	endRow := nRows + file.rowOffset
	startCol := file.colOffset
	endCol := nCols + file.colOffset
	if endRow > len(file.buffer) {
		endRow = len(file.buffer)
	}
	if endRow <= startRow {
		return []string{}
	}

	slice := make([]string, endRow-startRow)
	for row := startRow; row < endRow; row++ {
		line := tabs2spaces(file.buffer[row])
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

func (file *File) StartOfLine() {
	for idx, _ := range file.multiCursor {
		file.multiCursor[idx].col = 0
	}
}

func (file *File) EndOfLine() {
	for idx, _ := range file.multiCursor {
		row := file.multiCursor[idx].row
		file.multiCursor[idx].col = len(file.buffer[row])
	}
}

func (file *File) NextWord() {
	for idx, cursor := range file.multiCursor {
		row := cursor.row
		line := file.buffer[row]
		col := cursor.col
		for col < len(line)-1 {
			col++
			s := string(line[col])
			if s == " " || s == "\t" {
				break
			}
		}
		file.multiCursor[idx].col = col
	}
}

func (file *File) PrevWord() {
	for idx, cursor := range file.multiCursor {
		row := cursor.row
		line := file.buffer[row]
		col := cursor.col
		for col > 0 {
			col--
			s := string(line[col])
			if s == " " || s == "\t" {
				break
			}
		}
		file.multiCursor[idx].col = col
	}
}

func (file *File) Cut() Buffer {
	row := file.multiCursor[0].row
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
	row := file.multiCursor[0].row
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

func (file *File) Snapshot() {
	file.buffHist.Snapshot(file.buffer, file.multiCursor)
}

func (file *File) Undo() {
	file.buffer, file.multiCursor = file.buffHist.Prev()
}

func (file *File) Redo() {
	file.buffer, file.multiCursor = file.buffHist.Next()
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

func (file *File) Search() {
	searchTerm := file.GetPromptAnswer("search:", &file.searchHist)
	if searchTerm == "" {
		return
	}
	row, col, err := file.buffer.Search(searchTerm, file.multiCursor[0])
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
		row, col, err := file.buffer.Search(searchTerm, file.multiCursor[0])
		if err != nil {
			break
		}

		file.CursorGoTo(row, col)
		var startCol, endCol int
		startCol, endCol = file.buffer[row].Search(searchTerm, col, -1)
		for c := startCol; c < endCol; c++ {
			file.screen.Highlight(row, c)
		}
		doReplace, err := file.screen.AskYesNo("Replace this instance?")
		for c := startCol; c < endCol; c++ {
			file.screen.Highlight(row, c)
		}
		if err != nil {
			break
		}
		if doReplace {
			file.buffer.Replace(searchTerm, replaceTerm, row, col)
			file.screen.WriteString(row, 0, file.buffer[row].toString())
		}
	}
}

func (file *File) Status() string {
	if file.IsModified() {
		return "Modified"
	} else {
		return ""
	}
}

func (file *File) WriteStatus(row, col int) {
	status := file.Status()
	col -= len(status) + 2
	fg := file.screen.colors["yellow"]
	bg := file.screen.colors["black"]
	file.screen.WriteStringColor(row, col, status, fg, bg)
	if len(file.multiCursor) > 1 {
		status = fmt.Sprintf("%dC", len(file.multiCursor))
		col -= len(status) + 2
		fg := file.screen.colors["white"]
		bg := file.screen.colors["red"]
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}
}
