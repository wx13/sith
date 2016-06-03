package editor

import (
	"fmt"

	"github.com/wx13/sith/terminal"
)

type Action struct {
	Func func()
	Name string
}

type KeyMap map[string]Action

func (editor *Editor) MakeKeyMap() KeyMap {
	km := make(KeyMap)
	km.Add("backspace", editor.file.Backspace, "")
	km.Add("delete", editor.file.Delete, "")
	km.Add("ctrlD", editor.file.Delete, "")
	km.Add("space", func() { editor.file.InsertChar(' ') }, "")
	km.Add("tab", func() { editor.file.InsertChar('\t') }, "")
	km.Add("enter", editor.file.Newline, "")
	km.Add("arrowLeft", editor.file.CursorLeft, "")
	km.Add("arrowRight", editor.file.CursorRight, "")
	km.Add("arrowUp", func() { editor.file.CursorUp(1) }, "")
	km.Add("arrowDown", func() { editor.file.CursorDown(1) }, "")
	km.Add("ctrlJ", editor.file.ScrollUp, "Scroll Up")
	km.Add("ctrlK", editor.file.ScrollDown, "Scroll Down")
	km.Add("ctrlP", editor.file.ScrollRight, "Scroll Right")
	km.Add("ctrlO", editor.file.ScrollLeft, "Scroll Left")
	km.Add("pageDown", editor.file.PageDown, "")
	km.Add("ctrlN", editor.file.PageDown, "")
	km.Add("pageUp", editor.file.PageUp, "")
	km.Add("ctrlB", editor.file.PageUp, "")
	km.Add("ctrlG", editor.file.GoToLine, "Go to line number")
	km.Add("altL", editor.file.Refresh, "Refresh screen")
	km.Add("altO", editor.OpenNewFile, "Open new file")
	km.Add("altQ", editor.Quit, "Quit editor")
	km.Add("altW", func() { editor.CloseFile() }, "Close file")
	km.Add("altS", func() { editor.Suspend(); editor.keyboard = terminal.NewKeyboard() }, "Suspend")
	km.Add("altN", editor.NextFile, "Next file buffer")
	km.Add("altB", editor.PrevFile, "Previous file buffer")
	km.Add("altK", editor.LastFile, "Toggle between recent buffers")
	km.Add("altM", editor.SelectFile, "Select file buffer from menu")
	km.Add("ctrlX", editor.file.AddCursor, "Add cursor")
	km.Add("altC", editor.file.AddCursorCol, "Create column cursor")
	km.Add("altX", editor.file.ClearCursors, "Clear multi-cursor")
	km.Add("ctrlU", editor.Undo, "Undo")
	km.Add("ctrlY", editor.Redo, "Redo")
	km.Add("altU", editor.UndoSaved, "Macro undo")
	km.Add("altY", editor.RedoSaved, "Macro redo")
	km.Add("ctrlS", editor.Save, "Save file")
	km.Add("ctrlA", editor.file.StartOfLine, "Move to start of line")
	km.Add("ctrlE", editor.file.EndOfLine, "Move to end of line")
	km.Add("altA", editor.file.CutToStartOfLine, "Cut to start of line")
	km.Add("altE", editor.file.CutToEndOfLine, "Cut to end of line")
	km.Add("ctrlW", editor.file.NextWord, "Move cursor to next word")
	km.Add("ctrlQ", editor.file.PrevWord, "Move cursor to previous word")
	km.Add("ctrlF", func() { editor.Search(false) }, "Search")
	km.Add("ctrlR", func() { editor.Search(true) }, "Multi-file search")
	km.Add("altF", func() { editor.SearchAndReplace(false) }, "Search and replace")
	km.Add("altR", func() { editor.SearchAndReplace(true) }, "Multi-file search and replace")
	km.Add("ctrlC", editor.Cut, "Cut line")
	km.Add("ctrlV", editor.Paste, "Paste")
	km.Add("altV", editor.PasteFromMenu, "Paste from menu")
	km.Add("altG", editor.GoFmt, "Go fmt")
	km.Add("altJ", func() { editor.file.Justify(72) }, "Justify")
	km.Add("altH", func() { editor.file.Justify(0) }, "Unjustify")
	km.Add("altI", editor.file.ToggleAutoIndent, "Toggle auto-indent")
	km.Add("altT", editor.file.ToggleAutoTab, "Toggle Auto-tab")
	km.Add("alt6", editor.ExtraMode, "Extra mode")
	km.Add("ctrlSlash", editor.CmdMenu, "Display command menu")
	return km
}

func (editor *Editor) MakeExtraKeyMap() KeyMap {
	km := make(KeyMap)
	km.Add("c", editor.SetCharMode, "Change character display mode")
	km.Add("a", editor.file.CursorAlign, "Align cursor")
	km.Add("A", editor.file.CursorUnalign, "Unalign cursor")
	return km
}

func (km KeyMap) Add(key string, f func(), name string) {
	km[key] = Action{f, name}
}

func (km KeyMap) Run(key string) string {
	action, ok := km[key]
	if ok {
		action.Func()
		return ""
	}
	if key == "char" {
		return "char"
	}
	return "unknown"
}

func (km KeyMap) Keys() []string {
	keys := []string{}
	for key, action := range km {
		if len(action.Name) > 0 {
			keys = append(keys, key)
		}
	}
	return keys
}

func (km KeyMap) DisplayNames(keys []string, prefix string) []string {
	names := make([]string, len(keys))
	for idx, key := range keys {
		action, ok := km[key]
		if ok {
			names[idx] = fmt.Sprintf("%s%-10s  %s", prefix, key, action.Name)
			continue
		}
	}
	return names
}
