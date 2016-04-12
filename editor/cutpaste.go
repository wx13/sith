package editor

import (
	"github.com/wx13/sith/terminal"
)

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

func (editor *Editor) Paste() {
	editor.file.Paste(editor.copyBuffer)
}

func (editor *Editor) PasteFromMenu() {
	menu := terminal.NewMenu(editor.screen)
	items := []string{}
	for _, buffer := range editor.copyHist {
		str := buffer[0]
		items = append(items, str)
	}
	idx := menu.Choose(items)
	if idx < 0 || idx >= len(editor.copyHist) {
		return
	}
	editor.file.Paste(editor.copyHist[idx])
}
