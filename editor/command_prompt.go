package editor

import (
	"sort"
	"strings"

	"github.com/wx13/sith/autocomplete"
	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
)

func (editor *Editor) buildCommandMap() map[string]func() {
	return map[string]func(){
		// Navigation
		"scroll-up":    func() { editor.file.ScrollUp() },
		"scroll-down":  func() { editor.file.ScrollDown() },
		"scroll-right": func() { editor.file.ScrollRight() },
		"scroll-left":  func() { editor.file.ScrollLeft() },
		"goto":         func() { editor.GoToLine() },
		"start-of-line": func() { editor.file.StartOfLine() },
		"end-of-line":   func() { editor.file.EndOfLine() },
		"next-word":     func() { editor.file.NextWord() },
		"prev-word":     func() { editor.file.PrevWord() },
		"next-change":   func() { editor.file.NextChange() },
		"prev-change":   func() { editor.file.PrevChange() },

		// Files
		"open":        editor.OpenNewFile,
		"close":       func() { editor.CloseFile() },
		"quit":        editor.Quit,
		"suspend":     func() { editor.Suspend(); editor.keyboard = terminal.NewKeyboard() },
		"next":        editor.NextFile,
		"prev":        editor.PrevFile,
		"last":        editor.LastFile,
		"select-file": editor.SelectFile,
		"refresh":     func() { editor.file.Refresh() },

		// Save
		"save":     editor.Save,
		"save-as":  editor.SaveAs,
		"save-all": editor.SaveAll,

		// Edit
		"undo":       func() { editor.file.Undo() },
		"redo":       func() { editor.file.Redo() },
		"macro-undo": func() { editor.file.UndoSaved() },
		"macro-redo": func() { editor.file.RedoSaved() },
		"cut":        editor.Cut,
		"paste":      editor.Paste,
		"paste-menu": editor.PasteFromMenu,
		"cut-to-start": func() { editor.file.CutToStartOfLine() },
		"cut-to-end":   func() { editor.file.CutToEndOfLine() },
		"cut-word":       func() { editor.file.CutWord(0) },
		"cut-word-left":  func() { editor.file.CutWord(-1) },
		"cut-word-right": func() { editor.file.CutWord(1) },
		"justify":   func() { editor.file.Justify() },
		"unjustify": func() { editor.file.UnJustify() },

		// Search
		"search":      func() { editor.Search(false) },
		"search-all":  func() { editor.Search(true) },
		"replace":     func() { editor.SearchAndReplace(false) },
		"replace-all": func() { editor.SearchAndReplace(true) },
		"search-line-fwd":  func() { editor.SearchLineFo() },
		"search-line-back": func() { editor.SearchLineBa() },
		"cursors-line-fwd":  func() { editor.AllLineFo() },
		"cursors-line-back": func() { editor.AllLineBa() },

		// Multi-cursor
		"add-cursor":    func() { editor.file.AddCursor() },
		"column-cursor": func() { editor.file.AddCursorCol() },
		"clear-cursors": func() { editor.file.ClearCursors() },
		"toggle-mc":     func() { editor.file.ToggleMCMode() },
		"align":         func() { editor.file.CursorAlign() },
		"unalign":       func() { editor.file.CursorUnalign() },

		// Settings
		"toggle-indent": func() { editor.file.ToggleAutoIndent() },
		"toggle-tab":    func() { editor.file.ToggleAutoTab() },
		"toggle-fmt":    func() { editor.file.ToggleAutoFmt() },
		"char-mode":     editor.SetCharMode,
		"set-tab-str":   func() { editor.file.SetTabStr() },
		"auto-tab-str":  func() { editor.file.UnsetTabStr() },
		"set-tab-width": func() { editor.file.SetTabWidth() },
		"set-line-len":  func() { editor.file.SetLineLen() },

		// Format
		"fmt":            func() { editor.file.Fmt() },
		"fmt-selection":  func() { editor.file.Fmt(true) },
		"fmt-code-block": func() { editor.FmtCodeBlock() },

		// Bookmarks & history
		"bookmark":      func() { editor.Bookmark() },
		"bookmark-menu": func() { editor.BookmarkMenu() },
		"history":       func() { editor.file.ShowHistory() },
		"show-diff":     func() { editor.file.ShowLineDiff() },

		// Reload
		"reload":     func() { editor.file.Reload() },
		"reload-all": editor.ReloadAll,

		// Meta
		"commands": editor.CmdMenu,
	}
}

// CommandPrompt shows a tmux-style ":" prompt for typing commands.
func (editor *Editor) CommandPrompt() {
	p := ui.MakePrompt(editor.screen, editor.keyboard)
	cmd := p.GetAnswer(":", &editor.cmdHist, editor.CommandComplete)
	if cmd == "" {
		return
	}
	action, ok := editor.commandMap[cmd]
	if !ok {
		editor.screen.Notify("Unknown command: " + cmd)
		return
	}
	action()
}

// CommandComplete returns tab-completion suffixes for command names.
func (editor *Editor) CommandComplete(input string) []string {
	input = strings.ToLower(input)
	var matches []string
	for name := range editor.commandMap {
		if strings.HasPrefix(name, input) && name != input {
			matches = append(matches, name)
		}
	}
	if len(matches) == 0 {
		return nil
	}
	sort.Strings(matches)
	common := autocomplete.GetCommonPrefix(matches)
	suffix := common[len(input):]
	if len(suffix) > 0 {
		return []string{suffix}
	}
	var suffixes []string
	for _, m := range matches {
		suffixes = append(suffixes, m[len(input):])
	}
	return suffixes
}
