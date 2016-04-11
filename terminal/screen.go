package terminal

import (
	"os"
	"strings"
	"sync"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/syntaxcolor"
)

type Screen struct {
	row, col int
	fg, bg   termbox.Attribute
	colors   map[string]termbox.Attribute

	flushChan chan struct{}
	dieChan   chan struct{}

	tbMutex *sync.Mutex
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
	screen.tbMutex.Unlock()
	termbox.Close()
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
				screen.tbMutex.Unlock()
				termbox.Close()
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
	cols, _ := termbox.Size()
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

func (screen *Screen) WriteStringColor(row, col int, s string, fg, bg termbox.Attribute) {
	for k, c := range s {
		screen.tbMutex.Lock()
		termbox.SetCell(col+k, row, c, fg, bg)
		screen.tbMutex.Unlock()
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
