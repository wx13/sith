package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/file"
	"github.com/wx13/sith/syntaxcolor"
	"github.com/wx13/sith/terminal"
)

type Editor struct {
	screen     *terminal.Screen
	file       *file.File
	files      []*file.File
	fileIdx    int
	fileIdxPrv int
	keyboard   *terminal.Keyboard
	flushChan  chan struct{}

	searchHist  []string
	replaceHist []string

	copyBuffer []string
	copyContig int
	copyHist   [][]string
}

func NewEditor() *Editor {
	return &Editor{
		flushChan:  make(chan struct{}, 1),
		screen:     terminal.NewScreen(),
		copyBuffer: []string{},
		copyContig: 0,
		copyHist:   [][]string{},
	}
}

func (editor *Editor) OpenNewFile() {
	dir, _ := os.Getwd()
	dir += "/"
	names := []string{}
	idx := 0
	files := []os.FileInfo{}
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
		idx = menu.Choose(names)
		editor.Flush()
		if idx < 0 {
			return
		}
		chosenFile := files[idx]
		if chosenFile.IsDir() {
			dir = filepath.Clean(dir+chosenFile.Name()) + "/"
		} else {
			break
		}
	}
	cwd, _ := os.Getwd()
	chosenFile, _ := filepath.Rel(cwd, dir+names[idx])
	editor.OpenFile(chosenFile)
	editor.fileIdxPrv = editor.fileIdx
	editor.fileIdx = len(editor.files) - 1
	editor.file = editor.files[editor.fileIdx]
}

func (editor *Editor) OpenFile(name string) {
	file := file.NewFile(name, editor.flushChan, editor.screen)
	file.SyntaxRules = syntaxcolor.NewSyntaxRules(name)
	editor.files = append(editor.files, file)
}

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

func (editor *Editor) Quit() {
	for _, _ = range editor.files {
		if !editor.CloseFile() {
			editor.NextFile()
		}
	}
}

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

func (editor *Editor) Listen() {

	editor.keyboard = terminal.NewKeyboard()
	for {
		cmd, r := editor.keyboard.GetKey()
		editor.HandleCmd(cmd, r)
		editor.copyContig--
		editor.RequestFlush()
	}

}

func (editor *Editor) HandleCmd(cmd string, r rune) {
	switch cmd {
	case "backspace":
		editor.file.Backspace()
	case "delete", "ctrlD":
		editor.file.Delete()
	case "space":
		editor.file.InsertChar(' ')
	case "tab":
		editor.file.InsertChar('\t')
	case "enter":
		editor.file.Newline()
	case "arrowLeft":
		editor.file.CursorLeft()
	case "arrowRight":
		editor.file.CursorRight()
	case "arrowUp":
		editor.file.CursorUp(1)
	case "arrowDown":
		editor.file.CursorDown(1)
	case "ctrlJ":
		editor.file.ScrollUp()
	case "ctrlK":
		editor.file.ScrollDown()
	case "ctrlP":
		editor.file.ScrollRight()
	case "ctrlO":
		editor.file.ScrollLeft()
	case "pageDown", "ctrlN":
		editor.file.PageDown()
	case "pageUp", "ctrlB":
		editor.file.PageUp()
	case "ctrlG":
		editor.file.GoToLine()
	case "altL":
		editor.file.Refresh()
	case "altO":
		editor.OpenNewFile()
	case "altQ":
		editor.Quit()
	case "altW":
		editor.CloseFile()
	case "altS":
		editor.Suspend()
		editor.keyboard = terminal.NewKeyboard()
	case "altN":
		editor.NextFile()
	case "altB":
		editor.PrevFile()
	case "altK":
		editor.LastFile()
	case "altM":
		editor.SelectFile()
	case "ctrlX":
		editor.file.AddCursor()
	case "altC":
		editor.file.AddCursorCol()
	case "altX":
		editor.file.ClearCursors()
	case "ctrlU":
		editor.Undo()
	case "ctrlY":
		editor.Redo()
	case "altU":
		editor.UndoSaved()
	case "altY":
		editor.RedoSaved()
	case "ctrlS":
		editor.Save()
	case "ctrlA":
		editor.file.StartOfLine()
	case "ctrlE":
		editor.file.EndOfLine()
	case "altA":
		editor.file.CutToStartOfLine()
	case "altE":
		editor.file.CutToEndOfLine()
	case "ctrlW":
		editor.file.NextWord()
	case "ctrlQ":
		editor.file.PrevWord()
	case "ctrlF":
		editor.Search(false)
	case "altF":
		editor.SearchAndReplace(false)
	case "ctrlR":
		editor.Search(true)
	case "altR":
		editor.SearchAndReplace(true)
	case "ctrlC":
		editor.Cut()
	case "ctrlV":
		editor.Paste()
	case "altV":
		editor.PasteFromMenu()
	case "altG":
		editor.GoFmt()
	case "altJ":
		editor.file.Justify(72)
	case "altH":
		editor.file.Justify(0)
	case "altI":
		editor.file.ToggleAutoIndent()
	case "altT":
		editor.file.ToggleAutoTab()
	case "alt6":
		editor.SetCharMode()
	case "unknown":
		editor.screen.Notify("Unknown keypress")
	case "char":
		editor.file.InsertChar(r)
	default:
		editor.screen.Notify("Unknown keypress")
	}
}

func (editor *Editor) Undo() {
	editor.file.Undo()
}

func (editor *Editor) Redo() {
	editor.file.Redo()
}

func (editor *Editor) UndoSaved() {
	editor.file.UndoSaved()
}

func (editor *Editor) RedoSaved() {
	editor.file.RedoSaved()
}

func (editor *Editor) NextFile() {
	editor.SwitchFile(editor.fileIdx + 1)
}

func (editor *Editor) PrevFile() {
	editor.SwitchFile(editor.fileIdx - 1)
}

func (editor *Editor) LastFile() {
	editor.SwitchFile(editor.fileIdxPrv)
}

func (editor *Editor) SelectFile() {
	names := []string{}
	for _, file := range editor.files {
		names = append(names, file.Name)
	}
	menu := terminal.NewMenu(editor.screen)
	idx := menu.Choose(names)
	if idx >= 0 {
		editor.SwitchFile(idx)
	}
}

func (editor *Editor) SetCharMode() {
	modes := editor.screen.ListCharModes()
	menu := terminal.NewMenu(editor.screen)
	idx := menu.Choose(modes)
	if idx >= 0 {
		editor.screen.SetCharMode(idx)
	}
}

func (editor *Editor) Save() {
	filetype := editor.file.SyntaxRules.GetFileType(editor.file.Name)
	if filetype == "go" {
		editor.GoFmt()
	}
	editor.file.RequestSave()
}

func (editor *Editor) GoFmt() {
	err := editor.file.GoFmt()
	if err == nil {
		editor.RequestFlush()
		editor.file.NotifyUser("GoFmt done")
	} else {
		editor.file.NotifyUser(err.Error())
	}
}

func intMod(a, n int) int {
	if a >= 0 {
		return a - n*(a/n)
	} else {
		return a - n*((a-n+1)/n)
	}
}

func (editor *Editor) SwitchFile(n int) {
	n = intMod(n, len(editor.files))
	editor.fileIdxPrv = editor.fileIdx
	editor.fileIdx = n
	editor.file = editor.files[n]
}

func (editor *Editor) HighlightCursors() {
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	for k, _ := range editor.file.MultiCursor.Cursors()[1:] {
		r, c := editor.file.GetCursor(k + 1)
		j := r*cols + c
		if j < 0 || j >= len(cells) {
			continue
		}
		cells[j].Bg |= termbox.AttrReverse
		cells[j].Fg |= termbox.AttrReverse
	}
}

func (editor *Editor) Flush() {
	editor.file.Flush()
	editor.HighlightCursors()
	editor.UpdateStatus()
	editor.screen.Flush()
}

func (editor *Editor) KeepFlushed() {
	go func() {
		for {
			<-editor.flushChan
			editor.Flush()
		}
	}()
}

func (editor *Editor) RequestFlush() {
	select {
	case editor.flushChan <- struct{}{}:
	default:
	}
}

func (editor *Editor) UpdateStatus() {
	cols, rows := termbox.Size()
	maxNameLen := cols / 3
	name := editor.file.Name
	nameLen := len(name)
	if nameLen > maxNameLen {
		name = name[0:maxNameLen/2] + "..." + name[nameLen-maxNameLen/2:nameLen]
	}
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
	editor.screen.WriteString(rows-1, 0, "[ Sith 0.4 ]")
	editor.screen.DecorateStatusLine()
	editor.file.WriteStatus(rows-1, col)
	editor.screen.SetCursor(editor.file.GetCursor(0))
}
