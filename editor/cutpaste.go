package editor

import (
	"fmt"
	"strings"

	"github.com/wx13/sith/ui"
)

// Keeps track of the current copy buffer and a history of copy buffers.
type CopyBuffer struct {
	// Current copy buffer is a set of lines.
	current []string
	// History is a set of buffers.
	history [][]string
	// contig keeps track of whether or not cuts are contiguous
	contig int
	// Maximum history elements to store.
	maxHist int
}

// Returns a new copy buffer object.
func NewCopyBuffer() *CopyBuffer {
	return &CopyBuffer{
		current: []string{},
		history: [][]string{},
		contig:  0,
		maxHist: 100,
	}
}

// This is called everytime the editor does something. If that something is not
// a copy operation, then the copies are no longer contiguous.
func (cb *CopyBuffer) NoOp() {
	cb.contig--
}

// Adds a set of lines to the copy buffer.
func (cb *CopyBuffer) Cut(lines ...string) {
	if cb.contig > 0 {
		cb.current = append(cb.current, lines...)
	} else {
		cb.Save()
		cb.current = lines
	}
	cb.contig = 2
}

// Saves the current copy buffer to history.
func (cb *CopyBuffer) Save() {
	if len(cb.current) == 0 {
		return
	}

	// Remove any duplicate buffers.
	tmp := cb.history[:0]
	cur := strings.Join(cb.current, "\n")
	for _, buf := range cb.history {
		if strings.Join(buf, "\n") != cur {
			tmp = append(tmp, buf)
		}
	}
	cb.history = tmp

	// Prepend the current buffer to the history list.
	cb.history = append([][]string{cb.current}, cb.history...)

	// Ensure the list is not too long.
	if len(cb.history) > cb.maxHist {
		cb.history = cb.history[:cb.maxHist]
	}
}

// Returns the current copy buffer.
func (cb *CopyBuffer) Paste() []string {
	return cb.current
}

// Returns the copy buffer history, which each buffer joined by the specified string.
func (cb *CopyBuffer) Join(delim string) []string {
	cb.Save()
	items := []string{}
	for _, buffer := range cb.history {
		str := strings.Join(buffer, delim)
		items = append(items, str)
	}
	return items
}

// Returns the Nth buffer from history.
func (cb *CopyBuffer) History(index int) ([]string, error) {
	if (index < 0) || (index > len(cb.history)) {
		return cb.history[0], fmt.Errorf("index out of range")
	}
	return cb.history[index], nil
}

// Cut cuts the current line and sticks it in the copy buffer.
func (editor *Editor) Cut() {
	editor.copyBuffer.Cut(editor.file.Cut()...)
}

// Paste pastes the current copy buffer into the main buffer.
func (editor *Editor) Paste() {
	editor.file.Paste(editor.copyBuffer.Paste())
}

// PasteFromMenu allows the user to select from the paste history.
func (editor *Editor) PasteFromMenu() {
	menu := ui.NewMenu(editor.screen, editor.keyboard)
	items := editor.copyBuffer.Join(" || ")
	idx, name := menu.Choose(items, 0, "")
	if name != "" {
		return
	}
	buf, err := editor.copyBuffer.History(idx)
	if err == nil {
		editor.file.Paste(buf)
		editor.copyBuffer.Cut(buf...)
	}
}
