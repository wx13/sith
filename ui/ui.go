package ui

import (
	"github.com/wx13/sith/terminal"
)

type Screen interface {
	Flush()
	Highlight(row, col int)
	WriteString(row, col int, text string)
	WriteStringColor(row, col int, text string, fg, bg terminal.Attribute)
	Size() (cols, rows int)
	WriteMessage(msg string)
	Row() int
	Col() int
	SetCursor(r, c int)
}

type Keyboard interface {
	GetKey() (string, rune)
}
