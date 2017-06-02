package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/file"
	"github.com/wx13/sith/syntaxcolor"
	"github.com/wx13/sith/terminal"
)

// Editor is the main editor object.  It orchestrates the terminal,
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

	searchHist  []string
	replaceHist []string

	copyBuffer []string
	copyContig int
	copyHist   [][]string
}

// NewEditor creates a new Editor object.
func NewEditor() *Editor {
	return &Editor{
		flushChan:  make(chan struct{}, 1),
		screen:     terminal.NewScreen(),
		copyBuffer: []string{},
		copyContig: 0,
		copyHist:   [][]string{},
	}
}

// OpenNewFile offers a file selection menu to choose a new file to open.
func (editor *Editor) OpenNewFile() {
	dir, _ := os.Getwd()
	dir += "/"
	names := []string{}
	idx := 0
	files := []os.FileInfo{}
	filename := ""
	for {
		files, _ = ioutil.ReadDir(dir)
		dotdot, err := os.Stat("../")
		if err == nil {
			files = append([]os.FileInfo{dotdot}, files...)
		}
		names = []string{}
		for _, file := range files {
			if file.IsDir() {
				names = append(names, file.Name()+"/")
			} else {
				names = append(names, file.Name())
			}
		}
		menu := terminal.NewMenu(editor.screen)
		key := ""
		idx, key = menu.Choose(names, 0, "ctrlO")
		editor.Flush()
		if idx < 0 || key == "cancel" {
			return
		}
		if key == "ctrlO" {
			var err error
			p := terminal.MakePrompt(editor.screen)
			filename, err = p.Ask(dir, nil)
			if err != nil {
				editor.screen.Notify("Unknown answer")
				return
			}
			break
		}
		chosenFile := files[idx]
		if chosenFile.IsDir() {
			dir = filepath.Clean(dir+chosenFile.Name()) + "/"
		} else {
			filename = names[idx]
			break
		}
	}
	cwd, _ := os.Getwd()
	chosenFile, _ := filepath.Rel(cwd, dir+filename)
	editor.OpenFile(chosenFile)
	editor.fileIdxPrv = editor.fileIdx
	editor.fileIdx = len(editor.files) - 1
	editor.file = editor.files[editor.fileIdx]
}

// OpenFile opens a specified file.
func (editor *Editor) OpenFile(name string) {
	file := file.NewFile(name, editor.flushChan, editor.screen)
	file.SyntaxRules = syntaxcolor.NewSyntaxRules(name)
	editor.files = append(editor.files, file)
}

// OpenFiles opens a set of specified files.
func (editor *Editor) OpenFiles(fileNames []string) {
	for _, name := range fileNames {
		editor.OpenFile(name)
	}
	if len(editor.files) == 0 {
		editor.files = append(editor.files, file.NewFile("", editor.flushChan, editor.screen))
	}
	editor.fileIdx = 0
	editor.fileIdxPrv = 0
	editor.file = editor.files[0]
}

func (editor *Editor) ReloadAll() {
	for _, file := range editor.files {
		file.Reload()
	}
}

// Quit closes all the files and exits the editor.
func (editor *Editor) Quit() {
	for range editor.files {
		if !editor.CloseFile() {
			editor.NextFile()
		}
	}
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
	editor.keymap = editor.MakeKeyMap()
	editor.xKeymap = editor.MakeExtraKeyMap()
	for {
		cmd, r := editor.keyboard.GetKey()
		editor.handleCmd(cmd, r)
		editor.copyContig--
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
	p := terminal.MakePrompt(editor.screen)
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
	names := []string{}
	for _, file := range editor.files {
		status := ""
		if file.IsModified() {
			status = "*"
		}
		if file.FileChanged() {
			status += "+"
		}
		names = append(names, status+file.Name)
	}
	menu := terminal.NewMenu(editor.screen)
	idx, cmd := menu.Choose(names, editor.fileIdx)
	if idx >= 0 && cmd == "" {
		editor.SwitchFile(idx)
	}
}

// SetCharMode offers a menu for selecting the character
// display mode.
func (editor *Editor) SetCharMode() {
	modes := editor.screen.ListCharModes()
	menu := terminal.NewMenu(editor.screen)
	idx, cmd := menu.Choose(modes, 0)
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

	menu := terminal.NewMenu(editor.screen)
	idx, cancel := menu.Choose(names, 0)
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
	p := terminal.MakePrompt(editor.screen)
	filename, err := p.Ask("Save to:", nil)
	if err != nil {
		editor.screen.Notify("Cancelled")
		return
	}
	editor.file.Name = filename
	editor.Save()
}

// Fmt runs the code formatter on the buffer text.
func (editor *Editor) Fmt() {
	editor.file.Fmt()
}

func intMod(a, n int) int {
	if a < 0 {
		return a - n*((a-n+1)/n)
	}
	return a - n*(a/n)
}

// SwitchFile changes to a new file buffer.
func (editor *Editor) SwitchFile(n int) {
	n = intMod(n, len(editor.files))
	editor.fileIdxPrv = editor.fileIdx
	editor.fileIdx = n
	editor.file = editor.files[n]
}

// HighlightCursors highlights all the multi-cursors.
func (editor *Editor) HighlightCursors() {
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	r0, c0 := editor.file.GetCursor(0)
	for k := range editor.file.MultiCursor.Cursors()[1:] {
		r, c := editor.file.GetCursor(k + 1)
		j := r*cols + c
		if j < 0 || j >= len(cells) {
			continue
		}
		if r == r0 && c == c0 {
			cells[j].Bg |= termbox.AttrBold
			cells[j].Fg |= termbox.AttrBold | termbox.ColorYellow
		} else {
			cells[j].Bg |= termbox.AttrReverse
			cells[j].Fg |= termbox.AttrReverse
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
		editor.screen.WriteStringColor(row, col-3, "M  ", termbox.ColorRed, termbox.ColorDefault)
		return 3
	}
	for _, file := range editor.files {
		if file.IsModified() {
			editor.screen.WriteStringColor(row, col-3, "M  ", termbox.ColorYellow, termbox.ColorDefault)
			return 3
		}
	}
	return 0
}

func (editor *Editor) writeSyncStatus(row, col int) int {
	if editor.file.FileChanged() {
		editor.screen.WriteStringColor(row, col-3, "S  ", termbox.ColorRed, termbox.ColorDefault)
		return 3
	}
	for _, file := range editor.files {
		if file.FileChanged() {
			editor.screen.WriteStringColor(row, col-3, "S  ", termbox.ColorYellow, termbox.ColorDefault)
			return 3
		}
	}
	return 0
}

// UpdateStatus updates the status line.
func (editor *Editor) UpdateStatus() {
	cols, rows := termbox.Size()

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
	editor.screen.WriteString(rows-1, 0, "[ Sith 0.5.0 ]")
	editor.screen.DecorateStatusLine()
	col -= editor.writeModStatus(rows-1, col)
	col -= editor.writeSyncStatus(rows-1, col)
	editor.file.WriteStatus(rows-1, col)
	editor.screen.SetCursor(editor.file.GetCursor(0))
}
