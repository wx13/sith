package editor

import "github.com/wx13/sith/terminal"
import "github.com/wx13/sith/file"

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

