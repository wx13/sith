package terminal

import "github.com/nsf/termbox-go"
import "strings"
import "github.com/wx13/sith/syntaxcolor"

type Screen struct {
	row, col int
	fg, bg   termbox.Attribute
	colors   map[string]termbox.Attribute
}

func NewScreen() *Screen {
	screen := Screen{
		row: 0,
		col: 0,
		bg:  termbox.ColorBlack,
		fg:  termbox.ColorWhite,
		colors: map[string]termbox.Attribute{
			"yellow":  termbox.ColorYellow,
			"black":   termbox.ColorBlack,
			"blue":    termbox.ColorBlue,
			"green":   termbox.ColorGreen,
			"magenta": termbox.ColorMagenta,
			"white":   termbox.ColorWhite,
			"red":     termbox.ColorRed,
			"cyan":    termbox.ColorCyan,
		},
	}
	termbox.Init()
	return &screen
}

func (screen *Screen) Close() {
	screen.Clear()
	termbox.Flush()
	termbox.Close()
}

func (screen *Screen) Open() {
	termbox.Init()
}

func (screen *Screen) Flush() {
	termbox.Flush()
}

func (screen *Screen) SetCursor(r, c int) {
	screen.row = r
	screen.col = c
	termbox.SetCursor(c, r)
}

func (screen *Screen) Clear() {
	termbox.Clear(screen.fg, screen.bg)
	cols, rows := termbox.Size()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(" ", cols))
	}
}

func (screen *Screen) DecorateStatusLine() {
	cells := termbox.CellBuffer()
	cols, rows := termbox.Size()
	for col := 0; col < cols; col++ {
		j := (rows-1)*cols + col
		cells[j].Fg = termbox.ColorBlue
	}
}

func (screen *Screen) WriteString(row, col int, s string) {
	screen.WriteStringColor(row, col, s, screen.fg, screen.bg)
}

func (screen *Screen) Colorize(row int, colors []syntaxcolor.LineColor) {
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	for _, lc := range colors {
		for col := lc.Start; col < lc.End; col++ {
			j := row*cols + col
			if j < 0 || j >= len(cells) {
				continue
			}
			cells[j].Bg, cells[j].Fg = lc.Bg, lc.Fg
		}
	}
}

func (screen *Screen) WriteStringColor(row, col int, s string, fg, bg termbox.Attribute) {
	for k, c := range s {
		termbox.SetCell(col+k, row, c, fg, bg)
	}
}

func (screen *Screen) WriteMessage(msg string) {
	_, rows := termbox.Size()
	screen.WriteString(rows-1, 0, msg)
}

func (screen *Screen) AskYesNo(question string) (bool, error) {
	prompt := MakePrompt(screen)
	return prompt.AskYesNo(question)
}

func (screen *Screen) Ask(question string, history []string) (string, error) {
	prompt := MakePrompt(screen)
	return prompt.Ask(question, history)
}

func (screen *Screen) Highlight(row, col int) {
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	j := row*cols + col
	cell := cells[j]
	cells[j].Bg, cells[j].Fg = cell.Fg, cell.Bg
}

