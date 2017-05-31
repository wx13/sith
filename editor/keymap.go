package editor

import (
	"fmt"

	"github.com/wx13/sith/terminal"
)

// Action defines a keyboard action.
type Action struct {
	Func func()
	Name string
}

// KeyMap is a list of named actions.
type KeyMap map[string]Action

// MakeKeyMap initializes the KeyMap.
func (editor *Editor) MakeKeyMap() KeyMap {
	km := make(KeyMap)
	km.Add("backspace", func() { editor.file.Backspace() }, "")
	km.Add("delete", func() { editor.file.Delete() }, "")
	km.Add("ctrlD", func() { editor.file.Delete() }, "")
	km.Add("space", func() { editor.file.InsertChar(' ') }, "")
	km.Add("tab", func() { editor.file.InsertChar('\t') }, "")
	km.Add("enter", func() { editor.file.Newline() }, "")
	km.Add("arrowLeft", func() { editor.file.CursorLeft() }, "")
	km.Add("arrowRight", func() { editor.file.CursorRight() }, "")
	km.Add("arrowUp", func() { editor.file.CursorUp(1) }, "")
	km.Add("arrowDown", func() { editor.file.CursorDown(1) }, "")
	km.Add("ctrlJ", func() { editor.file.ScrollUp() }, "Scroll Up")
	km.Add("ctrlK", func() { editor.file.ScrollDown() }, "Scroll Down")
	km.Add("ctrlP", func() { editor.file.ScrollRight() }, "Scroll Right")
	km.Add("ctrlO", func() { editor.file.ScrollLeft() }, "Scroll Left")
	km.Add("pageDown", func() { editor.file.PageDown() }, "")
	km.Add("ctrlN", func() { editor.file.PageDown() }, "")
	km.Add("pageUp", func() { editor.file.PageUp() }, "")
	km.Add("ctrlB", func() { editor.file.PageUp() }, "")
	km.Add("ctrlG", func() { editor.file.GoToLine() }, "Go to line number")
	km.Add("altL", func() { editor.file.Refresh() }, "Refresh screen")
	km.Add("altO", editor.OpenNewFile, "Open new file")
	km.Add("altQ", editor.Quit, "Quit editor")
	km.Add("altW", func() { editor.CloseFile() }, "Close file")
	km.Add("altZ", func() { editor.Suspend(); editor.keyboard = terminal.NewKeyboard() }, "Suspend")
	km.Add("altN", editor.NextFile, "Next file buffer")
	km.Add("altB", editor.PrevFile, "Previous file buffer")
	km.Add("altK", editor.LastFile, "Toggle between recent buffers")
	km.Add("altM", editor.SelectFile, "Select file buffer from menu")
	km.Add("ctrlX", func() { editor.file.AddCursor() }, "Add cursor")
	km.Add("altC", func() { editor.file.AddCursorCol() }, "Create column cursor")
	km.Add("altX", func() { editor.file.ClearCursors() }, "Clear multi-cursor")
	km.Add("ctrlU", func() { editor.file.Undo() }, "Undo")
	km.Add("ctrlY", func() { editor.file.Redo() }, "Redo")
	km.Add("altU", func() { editor.file.UndoSaved() }, "Macro undo")
	km.Add("altY", func() { editor.file.RedoSaved() }, "Macro redo")
	km.Add("ctrlS", editor.Save, "Save file")
	km.Add("altS", editor.SaveAs, "Save as...")
	km.Add("ctrlA", func() { editor.file.StartOfLine() }, "Move to start of line")
	km.Add("ctrlE", func() { editor.file.EndOfLine() }, "Move to end of line")
	km.Add("altA", func() { editor.file.CutToStartOfLine() }, "Cut to start of line")
	km.Add("altE", func() { editor.file.CutToEndOfLine() }, "Cut to end of line")
	km.Add("ctrlW", func() { editor.file.NextWord() }, "Move cursor to next word")
	km.Add("ctrlQ", func() { editor.file.PrevWord() }, "Move cursor to previous word")
	km.Add("ctrlF", func() { editor.Search(false) }, "Search")
	km.Add("ctrlR", func() { editor.Search(true) }, "Multi-file search")
	km.Add("altF", func() { editor.SearchAndReplace(false) }, "Search and replace")
	km.Add("altR", func() { editor.SearchAndReplace(true) }, "Multi-file search and replace")
	km.Add("ctrlC", editor.Cut, "Cut line")
	km.Add("ctrlV", editor.Paste, "Paste")
	km.Add("altV", editor.PasteFromMenu, "Paste from menu")
	km.Add("altG", editor.Fmt, "Run code formatter")
	km.Add("altJ", func() { editor.file.Justify(72) }, "Justify")
	km.Add("altH", func() { editor.file.Justify(0) }, "Unjustify")
	km.Add("altI", func() { editor.file.ToggleAutoIndent() }, "Toggle auto-indent")
	km.Add("altT", func() { editor.file.ToggleAutoTab() }, "Toggle Auto-tab")
	km.Add("alt6", editor.ExtraMode, "Extra mode")
	km.Add("ctrlSlash", editor.CmdMenu, "Display command menu")
	return km
}

// MakeExtraKeyMap initializes the "extra" keys map.
func (editor *Editor) MakeExtraKeyMap() KeyMap {
	km := make(KeyMap)
	km.Add("c", editor.SetCharMode, "Change character display mode")
	km.Add("a", func() { editor.file.CursorAlign() }, "Align cursor")
	km.Add("A", func() { editor.file.CursorUnalign() }, "Unalign cursor")
	km.Add("w", func() { editor.SearchLineFo() }, "Search from cursor to end of line")
	km.Add("q", func() { editor.SearchLineBa() }, "Search from cursor to start of line")
	km.Add("t", func() { editor.file.SetTabStr() }, "Manually set the indentation string")
	km.Add("T", func() { editor.file.UnsetTabStr() }, "(Re)Enable auto tab string detection")
	km.Add("s", editor.SaveAll, "Save all files")
	km.Add("r", func() { editor.file.Reload() }, "Reload file from disk")
	km.Add("R", editor.ReloadAll, "Reload file from disk")
	km.Add("g", func() { editor.file.ToggleAutoFmt() }, "Toggle auto fmt on save")
	return km
}

// Add inserts a new action into the keymap.
func (km KeyMap) Add(key string, f func(), name string) {
	km[key] = Action{f, name}
}

// Run runs the specified action.
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

// Keys shows a list of available actions.
func (km KeyMap) Keys() []string {
	keys := []string{}
	for key, action := range km {
		if len(action.Name) > 0 {
			keys = append(keys, key)
		}
	}
	return keys
}

// DisplayNames returns pretty-formatted keymap action names.
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
