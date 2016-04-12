package file

import (
	"os"
	"path"
	"sync"

	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/file/cursor"
	"github.com/wx13/sith/syntaxcolor"
	"github.com/wx13/sith/terminal"
)

type File struct {
	buffer      buffer.Buffer
	MultiCursor cursor.MultiCursor
	savedBuffer buffer.Buffer

	timer   Timer
	maxRate float64

	buffHist *BufferHist
	gotoHist []string

	Name        string
	SyntaxRules *syntaxcolor.SyntaxRules
	fileMode    os.FileMode
	autoIndent  bool

	autoTab   bool
	tabString string
	tabHealth bool

	newline string

	rowOffset int
	colOffset int
	screen    *terminal.Screen
	flushChan chan struct{}
	saveChan  chan struct{}

	notification      string
	clearNotification bool

	statusMutex *sync.Mutex
}

func NewFile(name string, flushChan chan struct{}, screen *terminal.Screen) *File {
	file := &File{
		Name:        name,
		screen:      screen,
		fileMode:    os.FileMode(int(0644)),
		buffer:      buffer.MakeBuffer([]string{""}),
		savedBuffer: buffer.MakeBuffer([]string{""}),
		MultiCursor: cursor.MakeMultiCursor(),
		flushChan:   flushChan,
		saveChan:    make(chan struct{}, 1),
		SyntaxRules: syntaxcolor.NewSyntaxRules(""),
		autoIndent:  true,
		autoTab:     true,
		tabString:   "\t",
		newline:     "\n",
		tabHealth:   true,
		timer:       MakeTimer(),
		maxRate:     100.0,
		statusMutex: &sync.Mutex{},
	}
	file.buffHist = NewBufferHist(file.buffer, file.MultiCursor)
	go file.ProcessSaveRequests()
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

func (file *File) ToggleAutoTab() {
	file.autoTab = file.autoTab != true
}

func (file *File) ComputeIndent() {
	if file.autoTab {
		file.tabString, file.tabHealth = file.buffer.GetIndent()
	}
}

func (file *File) Refresh() {
	file.screen.ReallyClear()
}

func (file *File) ClearCursors() {
	file.MultiCursor.Clear()
}

func (file *File) AddCursor() {
	file.MultiCursor.Snapshot()
}

func (file *File) AddCursorCol() {
	file.MultiCursor.SetColumn()
}

func (file *File) ToString() string {
	return file.buffer.ToString(file.newline)
}

// Slice returns a 2D slice of the buffer.
func (file *File) Slice(nRows, nCols int) []string {

	file.UpdateOffsets(nRows, nCols)

	startRow := file.rowOffset
	endRow := nRows + file.rowOffset
	startCol := file.colOffset
	endCol := nCols + file.colOffset
	if endRow > file.buffer.Length() {
		endRow = file.buffer.Length()
	}
	if endRow <= startRow {
		return []string{}
	}

	return file.buffer.StrSlab(startRow, endRow, startCol, endCol)

}

func (file *File) ForceSnapshot() {
	file.buffHist.ForceSnapshot(file.buffer, file.MultiCursor)
}

func (file *File) Snapshot() {
	file.buffHist.Snapshot(file.buffer, file.MultiCursor)
}

func (file *File) SnapshotSaved() {
	file.buffHist.SnapshotSaved()
}

func (file *File) Undo() {
	buffer, mc := file.buffHist.Prev()
	file.MultiCursor.ReplaceMC(mc)
	file.buffer.ReplaceBuffer(buffer)
}

func (file *File) Redo() {
	buffer, mc := file.buffHist.Next()
	file.buffer.ReplaceBuffer(buffer)
	file.MultiCursor.ReplaceMC(mc)
}

func (file *File) UndoSaved() {
	buffer, mc := file.buffHist.PrevSaved()
	file.MultiCursor.ReplaceMC(mc)
	file.buffer.ReplaceBuffer(buffer)
}

func (file *File) RedoSaved() {
	buffer, mc := file.buffHist.NextSaved()
	file.MultiCursor.ReplaceMC(mc)
	file.buffer.ReplaceBuffer(buffer)
}

func (file *File) AskReplace(searchTerm, replaceTerm string, row, col int, replaceAll bool) error {

	file.CursorGoTo(row, col)
	_, screenCol := file.GetCursor(0)
	startColOffset := screenCol - col

	var doReplace bool
	var err error
	if replaceAll {
		doReplace = true
	} else {
		file.Flush()
		var startCol, endCol int
		startCol, endCol = file.buffer.GetRow(row).Search(searchTerm, col, -1)
		for c := startCol + startColOffset; c < endCol+startColOffset; c++ {
			file.screen.Highlight(row-file.rowOffset, c)
		}
		doReplace, err = file.screen.AskYesNo("Replace this instance?")
		for c := startCol; c < endCol; c++ {
			file.screen.Highlight(row-file.rowOffset, c)
		}
		if err != nil {
			return err
		}
	}
	if doReplace {
		file.buffer.ReplaceWord(searchTerm, replaceTerm, row, col)
		file.screen.WriteString(row, 0, file.buffer.GetRow(row).ToString())
		file.CursorGoTo(row, col+len(replaceTerm))
	}
	return nil

}

func (file *File) Length() int {
	return file.buffer.Length()
}

func (file *File) MarkedSearch(searchTerm string, loop bool) (row, col int, err error) {
	file.MultiCursor.OuterMost()
	_, maxRow := file.MultiCursor.MinMaxRow()
	subBuffer := file.buffer.InclSlice(0, maxRow)
	firstCursor := file.MultiCursor.GetFirstCursor()
	row, col, err = subBuffer.Search(searchTerm, firstCursor, false)
	return
}

func (file *File) SearchFromCursor(searchTerm string) (row, col int, err error) {
	loop := false
	cursor := file.MultiCursor.GetCursor(0)
	row, col, err = file.buffer.Search(searchTerm, cursor, loop)
	return
}

func (file *File) SearchFromStart(searchTerm string) (row, col int, err error) {
	loop := false
	cursor := cursor.MakeCursor(0, -1)
	row, col, err = file.buffer.Search(searchTerm, cursor, loop)
	return
}
