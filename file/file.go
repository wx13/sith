package file

import "sync"
import "os"
import "path"
import "github.com/wx13/sith/syntaxcolor"
import "github.com/wx13/sith/terminal"

type File struct {
	Buffer      Buffer
	buffMutex   *sync.Mutex
	MultiCursor MultiCursor
	savedBuffer Buffer

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
		saveChan:    make(chan struct{}, 1),
		SyntaxRules: syntaxcolor.NewSyntaxRules(""),
		autoIndent:  true,
		autoTab:     true,
		tabString:   "\t",
		newline:     "\n",
		tabHealth:   true,
	}
	file.buffHist = NewBufferHist(file.Buffer, file.MultiCursor)
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
		file.tabString, file.tabHealth = file.Buffer.GetIndent()
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

func (file *File) toString() string {
	return file.Buffer.ToString(file.newline)
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

	return file.Buffer.slice(startRow, endRow, startCol, endCol)

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
		startCol, endCol = file.Buffer[row].Search(searchTerm, col, -1)
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
		file.Buffer.Replace(searchTerm, replaceTerm, row, col)
		file.screen.WriteString(row, 0, file.Buffer[row].ToString())
	}
	return nil

}
