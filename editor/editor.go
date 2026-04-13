package editor

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/wx13/sith/autocomplete"
	"github.com/wx13/sith/config"
	"github.com/wx13/sith/file"
	"github.com/wx13/sith/state"
	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
	"github.com/wx13/sith/version"
)

// Editor is the main editor object. It orchestrates the terminal,
// the buffer, etc.
type Editor struct {
	screen     *terminal.Screen
	file       *file.File
	files      []*file.File
	fileIdx    int
	fileIdxPrv int
	keyboard   *terminal.Keyboard
	flushChan  chan struct{}
	keymap     KeyMap
	xKeymap    KeyMap
	bookmarks  *Bookmarks

	completer *autocomplete.AutoComplete

	searchHist  []string
	replaceHist []string
	gotoHist    []string
	history     *state.History

	copyBuffer *CopyBuffer

	cfg config.Config
}

// NewEditor creates a new Editor object.
func NewEditor() *Editor {
	history := state.NewHistory()
	return &Editor{
		flushChan:   make(chan struct{}, 1),
		screen:      terminal.NewScreen(),
		copyBuffer:  NewCopyBuffer(),
		cfg:         config.CreateConfig(),
		completer:   autocomplete.New(),
		history:     history,
		searchHist:  history.GetSearch(),
		replaceHist: history.GetReplace(),
		gotoHist:    history.GetGoto(),
	}
}

// OpenNewFile offers a file selection menu to choose a new file to open.
func (editor *Editor) OpenNewFile() {
	dir, _ := os.Getwd()
	dir += "/"
	filenames := []string{}
	for {
		entries, _ := os.ReadDir(dir)
		hasDotDot := false
		if _, err := os.Stat("../"); err == nil {
			hasDotDot = true
		}
		names := []string{}
		if hasDotDot {
			names = append(names, "../")
		}
		for _, entry := range entries {
			if entry.IsDir() {
				names = append(names, entry.Name()+"/")
			} else {
				names = append(names, entry.Name())
			}
		}
		menu := ui.NewMenu(editor.screen, editor.keyboard)
		idxs, key := menu.ChooseMulti(names, 0, "", "ctrlO")
		editor.Flush()
		if len(idxs) == 0 || key == "cancel" {
			return
		}
		if key == "ctrlO" {
			var err error
			p := ui.MakePrompt(editor.screen, editor.keyboard)
			filename, err := p.Ask(dir, nil)
			if err != nil {
				editor.screen.Notify("Unknown answer")
				return
			}
			filenames = []string{filename}
			break
		}
		if len(idxs) == 1 {
			chosenName := names[idxs[0]]
			if fs.ValidPath(chosenName) && chosenName[len(chosenName)-1] == '/' {
				dir = filepath.Clean(dir+chosenName) + "/"
			} else {
				filenames = []string{chosenName}
				break
			}
		} else {
			for _, idx := range idxs {
				chosenName := names[idx]
				if fs.ValidPath(chosenName) && chosenName[len(chosenName)-1] == '/' {
					editor.screen.Notify("Can't open multiple directories")
					return
				}
				filenames = append(filenames, chosenName)
			}
			break
		}
	}
	cwd, _ := os.Getwd()
	for _, filename := range filenames {
		chosenFile, _ := filepath.Rel(cwd, dir+filename)
		editor.OpenFile(chosenFile)
		editor.fileIdxPrv = editor.fileIdx
		editor.fileIdx = len(editor.files) - 1
		editor.file = editor.files[editor.fileIdx]
	}
}

// OpenFile opens a specified file.
func (editor *Editor) OpenFile(name string) {
	var wg sync.WaitGroup
	wg.Add(1)
	file := file.NewFile(name, editor.flushChan, editor.screen, editor.cfg, &wg)
	file.SetCompleter(editor.AutoComplete)
	editor.files = append(editor.files, file)
}

func (editor *Editor) AutoComplete(prefix string) []string {
	text := editor.file.ToCorpus(editor.file.GetRowsCols())
	for i, file := range editor.files {
		if i != editor.fileIdx {
			text += "\n" + file.ToString()
		}
	}
	return editor.completer.Complete(prefix, text)
}

// OpenFiles opens a set of specified files.
func (editor *Editor) OpenFiles(fileNames []string) {
	var wg sync.WaitGroup
	wg.Add(len(fileNames))
	for _, name := range fileNames {
		file := file.NewFile(name, editor.flushChan, editor.screen, editor.cfg, &wg)
		file.SetCompleter(editor.AutoComplete)
		editor.files = append(editor.files, file)
	}
	if len(editor.files) == 0 {
		wg.Add(1)
		file := file.NewFile("", editor.flushChan, editor.screen, editor.cfg, &wg)
		file.SetCompleter(editor.AutoComplete)
		editor.files = append(editor.files, file)
	}
	editor.fileIdx = 0
	editor.fileIdxPrv = 0
	editor.file = editor.files[0]
}

// ReloadAll re-reads all open buffers.
func (editor *Editor) ReloadAll() {
	var wg sync.WaitGroup
	wg.Add(len(editor.files))
	for _, file := range editor.files {
		file.Reload(&wg)
	}
}

// Quit closes all the files and exits the editor.
func (editor *Editor) Quit() {
	// Check if any files are modified. If so, confirm with the user.
	mod := false
	for _, file := range editor.files {
		if file.IsModified() {
			mod = true
			break
		}
	}
	if mod {
		prompt := ui.MakePrompt(editor.screen, editor.keyboard)
		close, _ := prompt.AskYesNo("One or more files has been modified. Quit without saving?")
		if !close {
			return
		}
	}

	// Save history before exiting.
	editor.saveHistory()

	// Exit.
	editor.screen.Close()
}

// saveHistory saves the search/replace/goto history to disk.
func (editor *Editor) saveHistory() {
	if editor.history == nil {
		return
	}
	// Update history from current in-memory slices
	for _, term := range reverseStrings(editor.searchHist) {
		editor.history.AddSearch(term)
	}
	for _, term := range reverseStrings(editor.replaceHist) {
		editor.history.AddReplace(term)
	}
	for _, term := range reverseStrings(editor.gotoHist) {
		editor.history.AddGoto(term)
	}
	editor.history.Save()
}

// reverseStrings returns a reversed copy of the slice.
func reverseStrings(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// CloseFile closes the current file.
func (editor *Editor) CloseFile() bool {
	editor.Flush()
	idx := editor.fileIdx
	if !editor.files[idx].Close() {
		return false
	}
	editor.files = append(editor.files[:idx], editor.files[idx+1:]...)
	if len(editor.files) == 0 {
		editor.screen.Close()
		return true
	}
	editor.NextFile()
	return true
}

// Listen is the main editor loop.
func (editor *Editor) Listen() {

	editor.keyboard = terminal.NewKeyboard()
	editor.keyboard.SetScreen(editor.screen.GetTcell())
	editor.keymap = editor.MakeKeyMap()
	editor.xKeymap = editor.MakeExtraKeyMap()
	for {
		cmd, r := editor.keyboard.GetKey()
		editor.handleCmd(cmd, r)
		editor.copyBuffer.NoOp()
		editor.RequestFlush()
	}

}

func (editor *Editor) handleCmd(cmd string, r rune) {
	ans := editor.keymap.Run(cmd)
	if ans == "" {
		return
	}
	if ans == "char" {
		editor.file.InsertChar(r)
	} else {
		editor.screen.Notify("Unknown keypress")
	}
}

// ExtraMode allows for additional keypresses.
func (editor *Editor) ExtraMode() {
	p := ui.MakePrompt(editor.screen, editor.keyboard)
	r := p.GetRune("key:")
	ans := editor.xKeymap.Run(string(r))
	if len(ans) > 0 {
		editor.screen.Notify("Unknown command")
	}
}

// NextFile cycles to the next open file.
func (editor *Editor) NextFile() {
	editor.SwitchFile(editor.fileIdx + 1)
}

// PrevFile cycles to the previous open file.
func (editor *Editor) PrevFile() {
	editor.SwitchFile(editor.fileIdx - 1)
}

// LastFile toggles between the two most recent files.
func (editor *Editor) LastFile() {
	editor.SwitchFile(editor.fileIdxPrv)
}

// SelectFile offers a menu to select from open files.
func (editor *Editor) SelectFile() {
	menu := ui.NewMenu(editor.screen, editor.keyboard)
	idx := editor.fileIdx
	cmd := ""
	// Allow user to swap file order.
	for {
		names := []string{}
		for _, file := range editor.files {
			status := ""
			if file.IsModified() {
				status = "*"
			}
			changed, err := file.FileChanged()
			if err != nil {
				status += "!"
			}
			if changed {
				status += "+"
			}
			names = append(names, status+file.Name)
		}
		idx, cmd = menu.Choose(names, idx, "", "ctrlJ", "ctrlK")
		if cmd == "ctrlJ" {
			if idx >= len(editor.files)-1 {
				continue
			}
			file := editor.files[idx]
			editor.files[idx] = editor.files[idx+1]
			idx++
			editor.files[idx] = file
			continue
		}
		if cmd == "ctrlK" {
			if idx == 0 {
				continue
			}
			file := editor.files[idx]
			editor.files[idx] = editor.files[idx-1]
			idx--
			editor.files[idx] = file
			continue
		}
		break
	}
	// Switch file.
	if idx >= 0 && cmd == "" {
		editor.SwitchFile(idx)
	}
}

// SetCharMode offers a menu for selecting the character
// display mode.
func (editor *Editor) SetCharMode() {
	modes := editor.screen.ListCharModes()
	menu := ui.NewMenu(editor.screen, editor.keyboard)
	idx, cmd := menu.Choose(modes, 0, "")
	if idx >= 0 && cmd == "" {
		editor.screen.SetCharMode(idx)
	}
}

// CmdMenu offers a menu of available commands.
func (editor *Editor) CmdMenu() {

	keys := editor.keymap.Keys()
	sort.Strings(keys)
	names := editor.keymap.DisplayNames(keys, "")

	xkeys := editor.xKeymap.Keys()
	sort.Strings(xkeys)
	xnames := editor.xKeymap.DisplayNames(xkeys, "Alt-6 ")

	names = append(names, xnames...)

	menu := ui.NewMenu(editor.screen, editor.keyboard)
	idx, cancel := menu.Choose(names, 0, "")
	if idx < 0 || cancel != "" {
		return
	}

	if idx < len(keys) {
		key := keys[idx]
		editor.keymap.Run(key)
	} else {
		key := xkeys[idx-len(keys)]
		editor.xKeymap.Run(key)
	}

}

// Save saves the buffer to the file.
func (editor *Editor) Save() {
	editor.file.RequestSave()
}

// SaveAll saves all the open buffers.
func (editor *Editor) SaveAll() {
	for _, file := range editor.files {
		file.RequestSave()
	}
}

// SaveAs prompts for a file to save to.
func (editor *Editor) SaveAs() {
	p := ui.MakePrompt(editor.screen, editor.keyboard)
	filename, err := p.Ask("Save to:", nil)
	if err != nil {
		editor.screen.Notify("Cancelled")
		return
	}
	editor.file.Name = filename
	editor.Save()
}

func intMod(a, n int) int {
	if a < 0 {
		return a - n*((a-n+1)/n)
	}
	return a - n*(a/n)
}

// SwitchFileByName changes to the first file with a matching name.
func (editor *Editor) SwitchFileByName(name string) error {
	for n, file := range editor.files {
		if file.Name == name {
			editor.SwitchFile(n)
			return nil
		}
	}
	return fmt.Errorf("no such file")
}

// SwitchFile changes to a new file buffer.
func (editor *Editor) SwitchFile(n int) {
	n = intMod(n, len(editor.files))
	if n == editor.fileIdx {
		return
	}
	editor.fileIdxPrv = editor.fileIdx
	editor.fileIdx = n
	editor.file = editor.files[n]
}

// HighlightCursors highlights all the multi-cursors.
func (editor *Editor) HighlightCursors() {
	cols, rows := editor.screen.Size()
	r0, c0 := editor.file.GetCursor(0)
	if editor.file.MultiCursor.Length() <= 1 {
		return
	}
	for k := range editor.file.MultiCursor.Cursors()[1:] {
		r, c := editor.file.GetCursor(k + 1)
		if r < 0 || r > rows || c < 0 || c > cols {
			continue
		}
		if r == r0 && c == c0 {
			editor.screen.ColorRange(r, r, c, c,
				terminal.ColorYellow|terminal.AttrBold,
				terminal.AttrBold)
		} else {
			editor.screen.Highlight(r, c)
		}
	}
}

// Flush writes the current buffer to the screen.
func (editor *Editor) Flush() {
	editor.file.Flush()
	editor.HighlightCursors()
	editor.UpdateStatus()
	editor.screen.Flush()
}

// KeepFlushed waits for flush requests, and then flushes
// to the screen.
func (editor *Editor) KeepFlushed() {
	go func() {
		for {
			<-editor.flushChan
			editor.Flush()
		}
	}()
}

// RequestFlush requests a flush event (async).
func (editor *Editor) RequestFlush() {
	select {
	case editor.flushChan <- struct{}{}:
	default:
	}
}

func (editor *Editor) getFilename(maxNameLen int) string {
	name := editor.file.Name
	nameLen := len(name)
	if nameLen > maxNameLen {
		name = name[0:maxNameLen/2] + "..." + name[nameLen-maxNameLen/2:nameLen]
	}
	return name
}

func (editor *Editor) writeModStatus(row, col int) int {
	if editor.file.IsModified() {
		editor.screen.WriteStringColor(row, col-3, "M  ", terminal.ColorRed, terminal.ColorDefault)
		return 3
	}
	if len(editor.files) <= 1 {
		return 0
	}
	for _, file := range editor.files {
		if file.IsModified() {
			editor.screen.WriteStringColor(row, col-3, "M  ", terminal.ColorYellow, terminal.ColorDefault)
			return 3
		}
	}
	return 0
}

func (editor *Editor) writeSyncStatus(row, col int) int {
	changed, err := editor.file.FileChanged()
	if err != nil {
		editor.screen.WriteStringColor(row, col-3, "X", terminal.ColorDefault, terminal.ColorRed)
		return 3
	}
	if changed {
		editor.screen.WriteStringColor(row, col-3, "S  ", terminal.ColorRed, terminal.ColorDefault)
		return 3
	}
	for _, file := range editor.files {
		changed, err := file.FileChanged()
		if err != nil {
			editor.screen.WriteStringColor(row, col-3, "X", terminal.ColorDefault, terminal.ColorYellow)
		}
		if changed {
			editor.screen.WriteStringColor(row, col-3, "S  ", terminal.ColorYellow, terminal.ColorDefault)
			return 3
		}
	}
	return 0
}

// UpdateStatus updates the status line.
func (editor *Editor) UpdateStatus() {
	cols, rows := editor.screen.Size()

	name := editor.getFilename(cols / 3)
	message := fmt.Sprintf("%s (%d/%d)   %d/%d,%d",
		name,
		editor.fileIdx,
		len(editor.files),
		editor.file.MultiCursor.GetRow(0),
		editor.file.Length()-1,
		editor.file.MultiCursor.GetCol(0),
	)
	col := cols - len(message)
	editor.screen.WriteString(rows-1, col, message)
	banner := "[ Sith " + version.Get() + " ]"
	editor.screen.WriteString(rows-1, 0, banner)
	editor.screen.DecorateStatusLine()
	col -= editor.writeModStatus(rows-1, col)
	col -= editor.writeSyncStatus(rows-1, col)
	editor.file.WriteStatus(rows-1, col)
	editor.screen.SetCursor(editor.file.GetCursor(0))
}
