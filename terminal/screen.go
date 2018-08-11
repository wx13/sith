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

type charMode int

type Attribute termbox.Attribute

const (
	ColorBlue     = Attribute(termbox.ColorBlue)
	ColorRed      = Attribute(termbox.ColorRed)
	ColorGreen    = Attribute(termbox.ColorGreen)
	ColorYellow   = Attribute(termbox.ColorYellow)
	ColorCyan     = Attribute(termbox.ColorCyan)
	ColorMagenta  = Attribute(termbox.ColorMagenta)
	ColorWhite    = Attribute(termbox.ColorWhite)
	ColorDefault  = Attribute(termbox.ColorDefault)
	AttrBold      = Attribute(termbox.AttrBold)
	AttrReverse   = Attribute(termbox.AttrReverse)
	AttrUnderline = Attribute(termbox.AttrUnderline)
)

const (
	charModeASCII charMode = iota
	charModeSomeUnicode
	charModeNarrowUnicode
	charModeFullUnicode
)

// Screen is an interface the the terminal screen.
type Screen struct {
	row, col int
	fg, bg   Attribute
	colors   map[string]Attribute

	flushChan chan struct{}
	dieChan   chan struct{}

	tbMutex *sync.Mutex

	charMode charMode
}

// NewScreen creates a new screen object.
func NewScreen() *Screen {
	screen := Screen{
		row:       0,
		col:       0,
		bg:        ColorDefault,
		fg:        ColorDefault,
		flushChan: make(chan struct{}, 1),
		dieChan:   make(chan struct{}, 1),
		tbMutex:   &sync.Mutex{},
		charMode:  charModeFullUnicode,
	}
	screen.tbMutex.Lock()
	termbox.Init()
	screen.tbMutex.Unlock()
	screen.handleRequests()
	return &screen
}

// Size returns the screen size (col, row).
func (screen *Screen) Size() (int, int) {
	return termbox.Size()
}

// Return the screen size.
func Size() (cols, rows int) {
	return termbox.Size()
}

func (screen *Screen) Row() int {
	return screen.row
}

func (screen *Screen) Col() int {
	return screen.col
}

// Suspend suspends the screen interaction to let the user
// access the terminal.
func (screen *Screen) Suspend() {
	screen.Clear()
	screen.tbMutex.Lock()
	termbox.Flush()
	termbox.Close()
	screen.tbMutex.Unlock()
}

// Close ends the terminal session.
func (screen *Screen) Close() {
	select {
	case screen.dieChan <- struct{}{}:
	default:
	}
}

// Open starts the terminal session.
func (screen *Screen) Open() {
	screen.tbMutex.Lock()
	termbox.Init()
	screen.tbMutex.Unlock()
}

// Flush *requests* a terminal flush event (async).
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

// SetCursor moves the cursor to a position.
func (screen *Screen) SetCursor(r, c int) {
	screen.row = r
	screen.col = c
	screen.tbMutex.Lock()
	termbox.SetCursor(c, r)
	screen.tbMutex.Unlock()
}

// Clear clears the screen.
func (screen *Screen) Clear() {
	screen.tbMutex.Lock()
	termbox.Clear(termbox.Attribute(screen.fg), termbox.Attribute(screen.bg))
	screen.tbMutex.Unlock()
	cols, rows := termbox.Size()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(" ", cols))
	}
}

// ReallyClear writes a repeaded character to the screen and then
// clears it, to make sure all terminal garbage is gone.
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

// DecorateStatusLine colors the status line text.
func (screen *Screen) DecorateStatusLine() {
	cells := termbox.CellBuffer()
	cols, rows := termbox.Size()
	for col := 0; col < cols; col++ {
		j := (rows-1)*cols + col
		cells[j].Fg = termbox.ColorBlue
	}
}

// WriteString write a string to the screen in the default color scheme.
func (screen *Screen) WriteString(row, col int, s string) {
	screen.WriteStringColor(row, col, s, screen.fg, screen.bg)
}

// Underlines a piece of text.
func (screen *Screen) Underline(row, start_col, end_col, offset int) {
	cells := termbox.CellBuffer()
	screen.tbMutex.Lock()
	cols, _ := termbox.Size()
	screen.tbMutex.Unlock()
	for col := start_col; col < end_col; col++ {
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
		cells[j].Fg |= termbox.AttrUnderline
	}
}

// Colorize changes the color of text on the screen.
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

// PrintableRune uses the charMode to convert the rune into
// a printable rune.
func (screen *Screen) PrintableRune(c rune) (rune, int) {
	if !unicode.IsPrint(c) {
		return c, 0
	}
	if screen.charMode == charModeASCII {
		if c >= 127 {
			c = '*'
		}
	}
	if screen.charMode == charModeSomeUnicode {
		if c >= 734 {
			c = 183
		}
	}
	if screen.charMode == charModeNarrowUnicode {
		w := runewidth.RuneWidth(c)
		if w != 1 {
			c = 183
		}
	}
	return c, runewidth.RuneWidth(c)
}

// StringDispLen estimates the display length of a string.
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

// WriteStringColor writes a colored string to the screen.
func (screen *Screen) WriteStringColor(row, col int, s string, fg, bg Attribute) {
	k := 0
	for _, c := range s {
		r, n := screen.PrintableRune(c)
		if n <= 0 {
			continue
		}
		screen.tbMutex.Lock()
		termbox.SetCell(col+k, row, r, termbox.Attribute(fg), termbox.Attribute(bg))
		screen.tbMutex.Unlock()
		k += n
	}
}

// WriteMessage writes a status-line message.
func (screen *Screen) WriteMessage(msg string) {
	if len(msg) == 0 {
		return
	}
	_, rows := termbox.Size()
	screen.WriteString(rows-1, 0, msg+"  ")
}

// Notify writes a status-line notification.
func (screen *Screen) Notify(msg string) {
	cols, rows := termbox.Size()
	screen.WriteString(rows-1, (cols-len(msg))/2, msg+"  ")
}

// Alert is the same as Notify.
func (screen *Screen) Alert(msg string) {
	screen.Notify(msg)
}

// Highlight reverses the screen color.
func (screen *Screen) Highlight(row, col int) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	j := row*cols + col
	cells[j].Bg |= termbox.AttrReverse
	cells[j].Fg |= termbox.AttrReverse
}

// SetCharMode sets the character display mode.
func (screen *Screen) SetCharMode(c int) {
	screen.charMode = charMode(c)
}

// ListCharModes lists the available character display modes.
func (screen *Screen) ListCharModes() []string {
	return []string{
		"ASCII only",
		"Some unicode characters",
		"Narrow unicode characters",
		"Full unicode",
	}
}
