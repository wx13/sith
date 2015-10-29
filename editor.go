package main

import "github.com/nsf/termbox-go"
import "fmt"
import "os"
import "io/ioutil"
import "path/filepath"
import "github.com/wx13/sith/syntaxcolor"
import "github.com/wx13/sith/terminal"
import "github.com/wx13/sith/file"

type Editor struct {
	screen     *terminal.Screen
	file       *file.File
	files      []*file.File
	fileIdx    int
	fileIdxPrv int
	keyboard   *terminal.Keyboard
	flushChan  chan struct{}
	msg        string

	searchHist  []string
	replaceHist []string

	copyBuffer file.Buffer
	copyContig int
	copyHist   []file.Buffer
}

func NewEditor() *Editor {
	return &Editor{
		flushChan:  make(chan struct{}, 1),
		screen:     terminal.NewScreen(),
		copyBuffer: file.MakeBuffer([]string{}),
		copyContig: 0,
		copyHist:   []file.Buffer{},
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
		os.Exit(0)
	}
	editor.NextFile()
	return true
}

func (editor *Editor) Listen() {

	keyboard := terminal.NewKeyboard()
	for {
		cmd, r := keyboard.GetKey()
		editor.msg = ""
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
		case "arrowLeft", "ctrlO":
			editor.file.CursorLeft()
		case "arrowRight", "ctrlL":
			editor.file.CursorRight()
		case "arrowUp", "ctrlK":
			editor.file.CursorUp(1)
		case "arrowDown", "ctrlJ":
			editor.file.CursorDown(1)
		case "ctrlU":
			editor.file.ScrollUp()
		case "ctrlP":
			editor.file.ScrollDown()
		case "altP":
			editor.file.ScrollRight()
		case "altU":
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
		case "altZ":
			editor.Suspend()
			keyboard = terminal.NewKeyboard()
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
		case "ctrlZ":
			editor.Undo()
		case "ctrlY":
			editor.Redo()
		case "ctrlS":
			editor.Save()
		case "ctrlA":
			editor.file.StartOfLine()
		case "ctrlE":
			editor.file.EndOfLine()
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
			editor.file.Justify()
		case "altI":
			editor.file.ToggleAutoIndent()
		case "altT":
			editor.file.ToggleAutoTab()
		case "unknown":
			editor.msg = "Unknown keypress"
		case "char":
			editor.file.InsertChar(r)
		default:
			editor.msg = "Unknown keypress"
		}
		editor.copyContig--
		editor.RequestFlush()
	}

}

func (editor *Editor) Cut() {
	if editor.copyContig > 0 {
		editor.copyBuffer = append(editor.copyBuffer, editor.file.Cut()...)
		editor.copyHist[0] = editor.copyBuffer.Dup()
	} else {
		editor.copyBuffer = editor.file.Cut()
		editor.copyHist = append([]file.Buffer{editor.copyBuffer.Dup()}, editor.copyHist...)
	}
	editor.copyContig = 2
}

func (editor *Editor) Search(multiFile bool) {
	searchTerm := editor.screen.GetPromptAnswer("search:", &editor.searchHist)
	if searchTerm == "" {
		editor.msg = "Cancelled"
		return
	}
	editor.MultiFileSearch(searchTerm, multiFile)
}

func (editor *Editor) MultiFileSearch(searchTerm string, multiFile bool) (int, int, error) {

	// Search remainder of current file.
	row, col, err := editor.file.Buffer.Search(searchTerm, editor.file.MultiCursor[0], false)
	if err == nil {
		editor.file.CursorGoTo(row, col)
		return row, col, err
	}

	// Search other files.
	if multiFile {
		for idx := editor.fileIdx + 1; idx != editor.fileIdx; idx++ {
			if idx >= len(editor.files) {
				idx = 0
			}
			theFile := editor.files[idx]
			row, col, err := theFile.Buffer.Search(searchTerm, file.MakeCursor(0,-1), false)
			if err == nil {
				editor.SwitchFile(idx)
				editor.file.CursorGoTo(row, col)
				return row, col, err
			}
		}
	}

	// Search start of current file.
	row, col, err = editor.file.Buffer.Search(searchTerm, file.MakeCursor(0,-1), false)
	if err == nil {
		editor.file.CursorGoTo(row, col)
		return row, col, err
	}

	editor.msg = "Not Found"
	return row, col, err
}

func (editor *Editor) SearchAndReplace(multiFile bool) {
	searchTerm := editor.screen.GetPromptAnswer("search:", &editor.searchHist)
	if searchTerm == "" {
		editor.msg = "Cancelled"
		return
	}

	replaceTerm := editor.screen.GetPromptAnswer("replace:", &editor.replaceHist)

	replaceAll, err := editor.screen.AskYesNo("Replace All?")
	if err != nil {
		editor.msg = "Cancelled"
		return
	}

	editor.MultiFileSearchAndReplace(searchTerm, replaceTerm, multiFile, replaceAll)
}

func (editor *Editor) MultiFileSearchAndReplace(searchTerm, replaceTerm string, multiFile, replaceAll bool) {

	var idx0, row0, col0 int
	idx0 = -1
	numMatches := 0
	for {
		row, col, err := editor.MultiFileSearch(searchTerm, multiFile)
		if err == nil {
			if idx0 < 0 {
				idx0 = editor.fileIdx
				row0, col0 = row, col
			} else if idx0 == editor.fileIdx && row0 == row && col0 == col {
				break
			}
			numMatches++
		} else {
			break
		}
	}

	for {

		row, col, err := editor.MultiFileSearch(searchTerm, multiFile)
		numMatches--
		if numMatches < 0 {
			break
		}

		if err == nil {
			err := editor.file.AskReplace(searchTerm, replaceTerm, row, col, replaceAll)
			if err != nil {
				editor.msg = "Cancelled"
				return
			}
		} else {
			editor.msg = "Not Found"
			break
		}

	}
}

func (editor *Editor) Paste() {
	editor.file.Paste(editor.copyBuffer)
}

func (editor *Editor) PasteFromMenu() {
	menu := terminal.NewMenu(editor.screen)
	items := []string{}
	for _, buffer := range editor.copyHist {
		str := buffer[0].ToString()
		items = append(items, str)
	}
	idx := menu.Choose(items)
	editor.file.Paste(editor.copyHist[idx])
}

func (editor *Editor) Undo() {
	editor.file.Undo()
}

func (editor *Editor) Redo() {
	editor.file.Redo()
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

func (editor *Editor) Save() {
	editor.msg = editor.file.Save()
}

func (editor *Editor) GoFmt() {
	editor.msg = editor.file.GoFmt()
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
	for k, _ := range editor.file.MultiCursor[1:] {
		r, c := editor.file.GetCursor(k + 1)
		j := r*cols + c
		if j < 0 || j >= len(cells) {
			continue
		}
		cell := cells[j]
		cells[j].Bg, cells[j].Fg = cell.Fg, cell.Bg
	}
}

func (editor *Editor) Flush() {
	editor.file.Flush()
	editor.HighlightCursors()
	editor.UpdateStatus()
	editor.screen.WriteMessage(editor.msg)
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
	message := fmt.Sprintf("%s (%d/%d)   %d/%d,%d",
		editor.file.Name,
		editor.fileIdx,
		len(editor.files),
		editor.file.MultiCursor[0].Row(),
		len(editor.file.Buffer)-1,
		editor.file.MultiCursor[0].Col(),
	)
	col := cols - len(message)
	editor.screen.WriteString(rows-1, col, message)
	editor.screen.WriteString(rows-1, 0, "[ Sith 0.2 ]")
	editor.screen.DecorateStatusLine()
	editor.file.WriteStatus(rows-1, col)
	editor.screen.SetCursor(editor.file.GetCursor(0))
}
