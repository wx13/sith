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
	"github.com/wx13/sith/ui"
)

// File contains all the details about a given file. This includes:
// name, buffer, buffer history, file-specific settings, etc.
type File struct {
	buffer      buffer.Buffer
	MultiCursor cursor.MultiCursor
	savedBuffer buffer.Buffer

	// Check for file system changes.
	md5sum  [16]byte
	modTime time.Time

	// For autocompletion. Passed in by editor.
	AutoComplete func(prefix string) []string
	lastTab      time.Time
	doubleTab    time.Duration

	timer   Timer
	maxRate float64

	buffHist *BufferHist

	Name        string
	SyntaxRules *syntaxcolor.SyntaxRules
	fileMode    os.FileMode
	autoIndent  bool

	autoTab   bool
	tabDetect bool
	tabString string
	tabHealth bool
	tabWidth  int
	lineLen   int

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

// NewFile creates a new File object. It reads in the specified file and ingests
// the specified configuration.
func NewFile(name string, flushChan chan struct{}, screen *terminal.Screen,
	cfg config.Config, wg *sync.WaitGroup) *File {

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
		lineLen:     80,
		newline:     "\n",
		tabHealth:   true,
		timer:       MakeTimer(),
		maxRate:     100.0,
		statusMutex: &sync.Mutex{},
		modTime:     time.Now(),
		md5sum:      md5.Sum([]byte("")),
		autoFmt:     true,
	}
	file.ingestConfig(cfg)
	go file.processSaveRequests()
	go func() {
		// Read file async, so we don't have to wait for it.
		file.ReadFile(name, wg)
		// Once the file is read, initialize the buffer history.
		file.buffHist = NewBufferHist(file.buffer, file.MultiCursor)
	}()
	return file
}

func (file *File) SetCompleter(f func(prefix string) []string) {
	file.AutoComplete = f
	file.lastTab = time.Now()
	file.doubleTab = time.Second
}

// GetFileExt returns the filename extension.
func GetFileExt(filename string) string {
	ext := path.Ext(filename)
	if len(ext) == 0 {
		basename := path.Base(filename)
		return strings.ToLower(basename)
	}
	ext = ext[1:]
	return ext
}

// ingestConfig uses the specified config file to set internal parameters.
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

// Reload re-reads a file from disk.
func (file *File) Reload(wgs ...*sync.WaitGroup) {
	if file.IsModified() {
		prompt := ui.MakePrompt(file.screen, terminal.NewKeyboard())
		ok, _ := prompt.AskYesNo("Changes will be lost. Reload anyway?")
		if !ok {
			return
		}
	}
	go file.ReadFile(file.Name, wgs...)
}

// Close doesn't actually close anything (b/c garbage collection will take care
// of it. Close just checks with the user and returns true if the file should
// close.
func (file *File) Close() bool {
	if file.IsModified() {
		prompt := ui.MakePrompt(file.screen, terminal.NewKeyboard())
		doClose, _ := prompt.AskYesNo("File has been modified. Close anyway?")
		if !doClose {
			return false
		}
	}
	return true
}

// ToggleMCMode cycles among the available multicursor navigation modes.
func (file *File) ToggleMCMode() {
	file.MultiCursor.CycleNavMode()
}

// ToggleAutoIndent toggles the autoindent setting.
func (file *File) ToggleAutoIndent() {
	file.autoIndent = file.autoIndent != true
}

// ToggleAutoTab toggles the autotab setting.
func (file *File) ToggleAutoTab() {
	file.autoTab = file.autoTab != true
}

// ToggleAutoFmt toggles the auto-format setting.
func (file *File) ToggleAutoFmt() {
	file.autoFmt = file.autoFmt != true
}

// SetTabWidth sets the tab display width.
func (file *File) SetTabWidth() {
	p := ui.MakePrompt(file.screen, terminal.NewKeyboard())
	str, err := p.Ask("tab width:", nil)
	if err != nil {
		return
	}
	width, err := strconv.Atoi(str)
	if err == nil {
		file.tabWidth = width
	}
}

func (file *File) SetLineLen() {
	p := ui.MakePrompt(file.screen, terminal.NewKeyboard())
	str, err := p.Ask("Line length:", nil)
	if err != nil {
		return
	}
	lineLen, err := strconv.Atoi(str)
	if err == nil {
		file.lineLen = lineLen
	}
}

// SetTabStr manually sets the tab string, and disables auto-tab-detection.
func (file *File) SetTabStr() {
	p := ui.MakePrompt(file.screen, terminal.NewKeyboard())
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

// ComputeIndent sets the tabString and tabHealth based on the current indentation.
func (file *File) ComputeIndent() {
	if file.tabDetect {
		file.tabString, file.tabHealth = file.buffer.GetIndent()
	} else if file.autoTab {
		_, file.tabHealth = file.buffer.GetIndent()
	}
}

// Refresh redraws the screen.
func (file *File) Refresh() {
	file.screen.ReallyClear()
}

// ClearCursors clears out the multicursors.
func (file *File) ClearCursors() {
	file.MultiCursor.Clear()
}

// AddCursor sets the current main cursor as a new multicursor member.
func (file *File) AddCursor() {
	file.MultiCursor.Snapshot()
}

// AddCursorCol creates a multicursor set along a column.
func (file *File) AddCursorCol() {
	file.MultiCursor.SetColumn()
}

// ToString returns a string representation of the text buffer. It uses the
// original newline (from the input file) as the line separator.
func (file *File) ToString() string {
	return file.buffer.ToString(file.newline)
}

// ToCorpus returns a string representation of the text buffer, with the current
// token removed. It is used for autocomplete.
func (file *File) ToCorpus(cursors map[int][]int) string {
	return file.buffer.ToCorpus(cursors)
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

// Snapshot saves a snapshot of the buffer state, but only if it has changed.
func (file *File) Snapshot() {
	if file.buffHist != nil {
		file.buffHist.Snapshot(file.buffer, file.MultiCursor)
	}
}

// ForceSnapshot saves a snapshot of the buffer state, even if it hasn't changed
// since the last snapshot.
func (file *File) ForceSnapshot() {
	if file.buffHist != nil {
		file.buffHist.ForceSnapshot(file.buffer, file.MultiCursor)
	}
}

// SnapshotSaved saves a special "saved" snapshot when the user saves (or opens)
// a file.
func (file *File) SnapshotSaved() {
	if file.buffHist != nil {
		file.buffHist.SnapshotSaved()
	}
}

// Undo reverts the buffer state to the last snapshot.
func (file *File) Undo() {
	if file.buffHist == nil {
		return
	}
	buffer, mc := file.buffHist.Prev()
	file.MultiCursor.ReplaceMC(mc)
	file.buffer.ReplaceBuffer(buffer)
}

// Redo sets the buffer state ahead one in the buffer history.
func (file *File) Redo() {
	if file.buffHist == nil {
		return
	}
	buffer, mc := file.buffHist.Next()
	file.buffer.ReplaceBuffer(buffer)
	file.MultiCursor.ReplaceMC(mc)
}

// UndoSaved reverts the buffer state to the last *saved* snapshot.
func (file *File) UndoSaved() {
	if file.buffHist == nil {
		return
	}
	buffer, mc := file.buffHist.PrevSaved()
	file.MultiCursor.ReplaceMC(mc)
	file.buffer.ReplaceBuffer(buffer)
}

// RedoSaved is like UndoSaved, but the other direction in time.
func (file *File) RedoSaved() {
	if file.buffHist == nil {
		return
	}
	buffer, mc := file.buffHist.NextSaved()
	file.MultiCursor.ReplaceMC(mc)
	file.buffer.ReplaceBuffer(buffer)
}

// AskReplace replaces each instance of searchTerm with replaceTerm, asking
// the user for confirmation each time.
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
		prompt := ui.MakePrompt(file.screen, terminal.NewKeyboard())
		doReplace, err = prompt.AskYesNo("Replace this instance?")
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

// Length returns the number of lines in the buffer.
func (file *File) Length() int {
	return file.buffer.Length()
}

// MarkedSearch searches for searchTerm in the text limited by the muticursor
// extent.
func (file *File) MarkedSearch(searchTerm string, loop bool) (row, col int, err error) {
	file.MultiCursor.OuterMost()
	_, maxRow := file.MultiCursor.MinMaxRow()
	subBuffer := file.buffer.InclSlice(0, maxRow)
	firstCursor := file.MultiCursor.GetFirstCursor()
	row, col, err = subBuffer.Search(searchTerm, firstCursor, false)
	return
}

// SearchFromCursor searches the current buffer, starting from the cursor position.
func (file *File) SearchFromCursor(searchTerm string) (row, col int, err error) {
	loop := false
	cursor := file.MultiCursor.GetCursor(0)
	row, col, err = file.buffer.Search(searchTerm, cursor, loop)
	return
}

// SearchFromStart searches the buffer from the start of the file.
func (file *File) SearchFromStart(searchTerm string) (row, col int, err error) {
	loop := false
	cursor := cursor.MakeCursor(0, -1)
	row, col, err = file.buffer.Search(searchTerm, cursor, loop)
	return
}

// SearchLineFo searches forward (to the right) on each multicursor marked line.
// It sets a cursor on each line where a match was found.
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

// SearchLineBa searches backward (to the left) on each marked line.
// It sets a cursor on each line where a match was found.
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

// AllLineFo is like SearchLineFo, but can return multiple matches on the same line.
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

// AllLineBa is like SearchLineBa, but can return multiple matches on the same line.
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
