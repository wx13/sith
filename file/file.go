package file

import (
	"crypto/md5"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wx13/sith/config"
	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/file/cursor"
	"github.com/wx13/sith/syntaxcolor"
	"github.com/wx13/sith/terminal"
)

type File struct {
	buffer      buffer.Buffer
	MultiCursor cursor.MultiCursor
	savedBuffer buffer.Buffer

	// Check for file system changes.
	md5sum  [16]byte
	modTime time.Time

	timer   Timer
	maxRate float64

	buffHist *BufferHist
	gotoHist []string

	Name        string
	SyntaxRules *syntaxcolor.SyntaxRules
	fileMode    os.FileMode
	autoIndent  bool

	autoTab   bool
	tabDetect bool
	tabString string
	tabHealth bool
	tabWidth  int

	newline string

	autoFmt bool
	fmtCmd  string

	rowOffset int
	colOffset int
	screen    *terminal.Screen
	flushChan chan struct{}
	saveChan  chan struct{}

	notification      string
	clearNotification bool

	statusMutex *sync.Mutex
}

func NewFile(name string, flushChan chan struct{}, screen *terminal.Screen, cfg config.Config) *File {
	file := &File{
		Name:        name,
		screen:      screen,
		fileMode:    os.FileMode(int(0644)),
		buffer:      buffer.MakeBuffer([]string{""}),
		savedBuffer: buffer.MakeBuffer([]string{""}),
		MultiCursor: cursor.MakeMultiCursor(),
		flushChan:   flushChan,
		saveChan:    make(chan struct{}, 1),
		autoIndent:  true,
		autoTab:     true,
		tabDetect:   true,
		tabString:   "\t",
		tabWidth:    4,
		newline:     "\n",
		tabHealth:   true,
		timer:       MakeTimer(),
		maxRate:     100.0,
		statusMutex: &sync.Mutex{},
		modTime:     time.Now(),
		md5sum:      md5.Sum([]byte("")),
		autoFmt:     true,
	}
	file.buffHist = NewBufferHist(file.buffer, file.MultiCursor)
	file.ingestConfig(cfg)
	go file.processSaveRequests()
	go file.ReadFile(name)
	return file
}

func GetFileExt(filename string) string {
	ext := path.Ext(filename)
	if len(ext) == 0 {
		basename := path.Base(filename)
		if basename == "COMMIT_EDITMSG" {
			basename = "git"
		}
		return strings.ToLower(basename)
	}
	ext = ext[1:]
	return ext
}

func (file *File) ingestConfig(cfg config.Config) {
	ext := GetFileExt(file.Name)
	cfg = cfg.ForExt(ext)
	file.autoTab = cfg.AutoTab
	file.tabDetect = cfg.TabDetect
	file.tabWidth = cfg.TabWidth
	file.tabString = cfg.TabString
	file.SyntaxRules = syntaxcolor.NewSyntaxRules(cfg)
	file.fmtCmd = cfg.FmtCmd
}

func (file *File) Reload() {
	if file.IsModified() {
		ok, _ := file.screen.AskYesNo("Changes will be lost. Reload anyway?")
		if !ok {
			return
		}
	}
	go file.ReadFile(file.Name)
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

func (file *File) ToggleAutoFmt() {
	file.autoFmt = file.autoFmt != true
}

// SetTabWidth sets the tab display width.
func (file *File) SetTabWidth() {
	p := terminal.MakePrompt(file.screen)
	str, err := p.Ask("tab width:", nil)
	if err != nil {
		return
	}
	width, err := strconv.Atoi(str)
	if err == nil {
		file.tabWidth = width
	}
}

// SetTabStr manually sets the tab string, and disables auto-tab-detection.
func (file *File) SetTabStr() {
	p := terminal.MakePrompt(file.screen)
	str, err := p.Ask("tab string:", nil)
	if err == nil {
		file.tabString = str
		file.tabDetect = false
	}
}

// UnsetTabStr (re)enables auto-tab detaction.
func (file *File) UnsetTabStr() {
	file.tabDetect = true
	file.ComputeIndent()
}

func (file *File) ComputeIndent() {
	if file.tabDetect {
		file.tabString, file.tabHealth = file.buffer.GetIndent()
	} else if file.autoTab {
		_, file.tabHealth = file.buffer.GetIndent()
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

	file.updateOffsets(nRows, nCols)

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

	return file.buffer.StrSlab(startRow, endRow, startCol, endCol, file.tabWidth)

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

func (file *File) SearchLineFo(term string) {
	file.MultiCursor.OnePerLine()
	for idx, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		line := file.buffer.GetRow(row)
		c0, _ := line.Search(term, col, -1)
		if c0 >= 0 {
			file.MultiCursor.SetCol(idx, c0)
		}
	}
}

func (file *File) SearchLineBa(term string) {
	file.MultiCursor.OnePerLine()
	for idx, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		line := file.buffer.GetRow(row)
		c0, _ := line.Search(term, col, 0)
		if c0 >= 0 {
			file.MultiCursor.SetCol(idx, c0)
		}
	}
}

func (file *File) AllLineFo(term string) {
	file.MultiCursor.OnePerLine()
	positions := make(map[int][]int)
	found := 0
	for _, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		cols := file.buffer.GetRow(row).SearchAll(term, col, -1)
		positions[row] = cols
		found += len(cols)
	}
	if found > 0 {
		file.MultiCursor.ResetRowsCols(positions)
	}
}

func (file *File) AllLineBa(term string) {
	file.MultiCursor.OnePerLine()
	positions := make(map[int][]int)
	found := 0
	for _, cursor := range file.MultiCursor.Cursors() {
		row, col := cursor.RowCol()
		cols := file.buffer.GetRow(row).SearchAll(term, col, 0)
		positions[row] = cols
		found += len(cols)
	}
	if found > 0 {
		file.MultiCursor.ResetRowsCols(positions)
	}
}
