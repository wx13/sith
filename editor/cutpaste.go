package editor

import (
	"strings"

	"github.com/wx13/sith/ui"
)

// Cut cuts the current line and sticks it in the copy buffer.
func (editor *Editor) Cut() {
	if editor.copyContig > 0 {
		editor.copyBuffer = append(editor.copyBuffer, editor.file.Cut()...)
		editor.copyHist[0] = editor.copyBuffer
	} else {
		editor.copyBuffer = editor.file.Cut()
		editor.copyHist = append([][]string{editor.copyBuffer}, editor.copyHist...)
	}
	editor.copyContig = 2
}

// Paste pastes the current copy buffer into the main buffer.
func (editor *Editor) Paste() {
	editor.file.Paste(editor.copyBuffer)
}

// PasteFromMenu allows the user to select from the paste history.
func (editor *Editor) PasteFromMenu() {
	menu := ui.NewMenu(editor.screen, editor.keyboard)
	items := []string{}
	for _, buffer := range editor.copyHist {
		str := strings.Join(buffer, " || ")
		items = append(items, str)
	}
	idx, name := menu.Choose(items, 0)
	if name != "" || idx < 0 || idx >= len(editor.copyHist) {
		return
	}
	editor.file.Paste(editor.copyHist[idx])
}
