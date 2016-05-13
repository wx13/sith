package terminal

import (
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/mattn/go-runewidth"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/syntaxcolor"
)

type CharMode int

const (
	CharModeAscii CharMode = iota
	CharModeSomeUnicode
	CharModeNarrowUnicode
	CharModeFullUnicode
)

type Screen struct {
	row, col int
	fg, bg   termbox.Attribute
	colors   map[string]termbox.Attribute

	flushChan chan struct{}
	dieChan   chan struct{}

	tbMutex *sync.Mutex

	charMode CharMode
}

func NewScreen() *Screen {
	screen := Screen{
		row:       0,
		col:       0,
		bg:        termbox.ColorDefault,
		fg:        termbox.ColorDefault,
		flushChan: make(chan struct{}, 1),
		dieChan:   make(chan struct{}, 1),
		tbMutex:   &sync.Mutex{},
		charMode:  CharModeFullUnicode,
	}
	screen.tbMutex.Lock()
	termbox.Init()
	screen.tbMutex.Unlock()
	screen.handleRequests()
	return &screen
}

func (screen *Screen) Suspend() {
	screen.Clear()
	screen.tbMutex.Lock()
	termbox.Flush()
	termbox.Close()
	screen.tbMutex.Unlock()
}

func (screen *Screen) Close() {
	select {
	case screen.dieChan <- struct{}{}:
	default:
	}
}

func (screen *Screen) Open() {
	screen.tbMutex.Lock()
	termbox.Init()
	screen.tbMutex.Unlock()
}

func (screen *Screen) Flush() {
	select {
	case screen.flushChan <- struct{}{}:
	default:
	}
}

func (screen *Screen) handleRequests() {
	go func() {
		for {
			select {
			case <-screen.flushChan:
				screen.tbMutex.Lock()
				termbox.Flush()
				screen.tbMutex.Unlock()
			case <-screen.dieChan:
				screen.Clear()
				screen.tbMutex.Lock()
				termbox.Flush()
				termbox.Close()
				screen.tbMutex.Unlock()
				os.Exit(0)
			}
		}
	}()
}

func (screen *Screen) SetCursor(r, c int) {
	screen.row = r
	screen.col = c
	screen.tbMutex.Lock()
	termbox.SetCursor(c, r)
	screen.tbMutex.Unlock()
}

func (screen *Screen) Clear() {
	screen.tbMutex.Lock()
	termbox.Clear(screen.fg, screen.bg)
	screen.tbMutex.Unlock()
	cols, rows := termbox.Size()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(" ", cols))
	}
}

func (screen *Screen) ReallyClear() {
	cols, rows := termbox.Size()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(".", cols))
	}
	screen.tbMutex.Lock()
	termbox.Flush()
	screen.tbMutex.Unlock()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(" ", cols))
	}
	screen.Flush()
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

func (screen *Screen) Colorize(row int, colors []syntaxcolor.LineColor, offset int) {
	cells := termbox.CellBuffer()
	screen.tbMutex.Lock()
	cols, _ := termbox.Size()
	screen.tbMutex.Unlock()
	for _, lc := range colors {
		for col := lc.Start; col < lc.End; col++ {
			if (col - offset) > cols {
				break
			}
			if (col - offset) < 0 {
				continue
			}
			j := row*cols + (col - offset)
			if j < 0 || j >= len(cells) {
				continue
			}
			cells[j].Bg, cells[j].Fg = lc.Bg, lc.Fg
		}
	}
}

func (screen *Screen) PrintableRune(c rune) (rune, int) {
	if !unicode.IsPrint(c) {
		return c, 0
	}
	if screen.charMode == CharModeAscii {
		if c >= 127 {
			c = '*'
		}
	}
	if screen.charMode == CharModeSomeUnicode {
		if c >= 734 {
			c = 183
		}
	}
	if screen.charMode == CharModeNarrowUnicode {
		w := runewidth.RuneWidth(c)
		if w != 1 {
			c = 183
		}
	}
	return c, runewidth.RuneWidth(c)
}

func (screen *Screen) StringDispLen(s string) int {
	N := 0
	for _, c := range s {
		_, n := screen.PrintableRune(c)
		if n > 0 {
			N += n
		}
	}
	return N
}

func (screen *Screen) WriteStringColor(row, col int, s string, fg, bg termbox.Attribute) {
	k := 0
	for _, c := range s {
		r, n := screen.PrintableRune(c)
		if n <= 0 {
			continue
		}
		screen.tbMutex.Lock()
		termbox.SetCell(col+k, row, r, fg, bg)
		screen.tbMutex.Unlock()
		k += n
	}
}

func (screen *Screen) WriteMessage(msg string) {
	if len(msg) == 0 {
		return
	}
	_, rows := termbox.Size()
	screen.WriteString(rows-1, 0, msg+"  ")
}

func (screen *Screen) Notify(msg string) {
	cols, rows := termbox.Size()
	screen.WriteString(rows-1, (cols-len(msg))/2, msg+"  ")
}

func (screen *Screen) Alert(msg string) {
	screen.Notify(msg)
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
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	j := row*cols + col
	cells[j].Bg |= termbox.AttrReverse
	cells[j].Fg |= termbox.AttrReverse
}

func (screen *Screen) SetCharMode(c int) {
	screen.charMode = CharMode(c)
}

func (screen *Screen) ListCharModes() []string {
	return []string{
		"ASCII only",
		"Some unicode characters",
		"Narrow unicode characters",
		"Full unicode",
	}
}
