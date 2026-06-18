package terminal

import (
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/wx13/sith/syntaxcolor"
)

type charMode int

// Attribute combines a tcell.Color with style attributes.
type Attribute uint64

const (
	ColorBlue     = Attribute(tcell.ColorBlue)
	ColorRed      = Attribute(tcell.ColorRed)
	ColorGreen    = Attribute(tcell.ColorGreen)
	ColorYellow   = Attribute(tcell.ColorYellow)
	ColorCyan     = Attribute(tcell.ColorTeal)
	ColorMagenta  = Attribute(tcell.ColorPurple)
	ColorWhite    = Attribute(tcell.ColorWhite)
	ColorDefault  = Attribute(tcell.ColorDefault)
	AttrBold      = Attribute(1 << 48)
	AttrReverse   = Attribute(1 << 49)
	AttrUnderline = Attribute(1 << 50)
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

	tcell   tcell.Screen
	tbMutex *sync.Mutex

	charMode charMode

	// gutterWidth reserves columns on the left for indicators (code blocks, git status, etc.)
	gutterWidth int
}

// toStyle converts fg/bg Attributes to a tcell.Style.
func toStyle(fg, bg Attribute) tcell.Style {
	style := tcell.StyleDefault

	// Extract color (lower 48 bits contain tcell.Color)
	fgColor := tcell.Color(fg & 0xFFFFFFFFFFFF)
	bgColor := tcell.Color(bg & 0xFFFFFFFFFFFF)

	style = style.Foreground(fgColor).Background(bgColor)

	// Apply attributes from either fg or bg
	attrs := fg | bg
	if attrs&AttrBold != 0 {
		style = style.Bold(true)
	}
	if attrs&AttrReverse != 0 {
		style = style.Reverse(true)
	}
	if attrs&AttrUnderline != 0 {
		style = style.Underline(true)
	}

	return style
}

// NewScreen creates a new screen object.
func NewScreen() *Screen {
	screen := Screen{
		row:         0,
		col:         0,
		bg:          ColorDefault,
		fg:          ColorDefault,
		flushChan:   make(chan struct{}, 1),
		dieChan:     make(chan struct{}, 1),
		tbMutex:     &sync.Mutex{},
		charMode:    charModeFullUnicode,
		gutterWidth: 1, // Reserve 1 column for indicators (code blocks, git status, etc.)
	}
	screen.tbMutex.Lock()
	var err error
	screen.tcell, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.tcell.Init(); err != nil {
		panic(err)
	}
	screen.tcell.EnableMouse()
	screen.tbMutex.Unlock()
	screen.handleRequests()
	return &screen
}

// Size returns the usable screen size (col, row), accounting for the gutter.
func (screen *Screen) Size() (int, int) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	cols, rows := screen.tcell.Size()
	return cols - screen.gutterWidth, rows
}

// SetGutterWidth sets the width of the left gutter (reserved for indicators).
func (screen *Screen) SetGutterWidth(width int) {
	screen.gutterWidth = width
}

// GutterWidth returns the current gutter width.
func (screen *Screen) GutterWidth() int {
	return screen.gutterWidth
}

// Return the screen size.
func Size() (cols, rows int) {
	// This is a package-level function used before screen is created.
	// Return a reasonable default; actual size comes from Screen.Size().
	return 80, 24
}

// Row returns the current cursor row.
func (screen *Screen) Row() int {
	return screen.row
}

// Col returns the current cursor column.
func (screen *Screen) Col() int {
	return screen.col
}

// Suspend suspends the screen interaction to let the user
// access the terminal.
func (screen *Screen) Suspend() {
	screen.Clear()
	screen.tbMutex.Lock()
	screen.tcell.Show()
	screen.tcell.Fini()
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
	var err error
	screen.tcell, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.tcell.Init(); err != nil {
		panic(err)
	}
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
				screen.tcell.Show()
				screen.tbMutex.Unlock()
			case <-screen.dieChan:
				screen.Clear()
				screen.tbMutex.Lock()
				screen.tcell.Show()
				screen.tcell.Fini()
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
	screen.tcell.ShowCursor(c+screen.gutterWidth, r)
	screen.tbMutex.Unlock()
}

// Clear clears the screen.
func (screen *Screen) Clear() {
	screen.tbMutex.Lock()
	screen.tcell.Clear()
	screen.tbMutex.Unlock()
	cols, rows := screen.Size()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(" ", cols))
	}
}

// ReallyClear writes a repeaded character to the screen and then
// clears it, to make sure all terminal garbage is gone.
func (screen *Screen) ReallyClear() {
	cols, rows := screen.Size()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(".", cols))
	}
	screen.tbMutex.Lock()
	screen.tcell.Show()
	screen.tbMutex.Unlock()
	for row := 0; row < rows; row++ {
		screen.WriteString(row, 0, strings.Repeat(" ", cols))
	}
	screen.Flush()
}

// DecorateStatusLine colors the status line text.
func (screen *Screen) DecorateStatusLine() {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	cols, rows := screen.tcell.Size()
	style := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	for col := 0; col < cols; col++ {
		mainc, combc, _, _ := screen.tcell.GetContent(col, rows-1)
		screen.tcell.SetContent(col, rows-1, mainc, combc, style)
	}
}

// WriteString write a string to the screen in the default color scheme.
func (screen *Screen) WriteString(row, col int, s string) {
	screen.WriteStringColor(row, col, s, screen.fg, screen.bg)
}

// Underlines a piece of text.
func (screen *Screen) Underline(row, start_col, end_col, offset int) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	cols, _ := screen.tcell.Size()
	for col := start_col; col < end_col; col++ {
		adjCol := col - offset + screen.gutterWidth
		if adjCol > cols || adjCol < 0 {
			continue
		}
		mainc, combc, style, _ := screen.tcell.GetContent(adjCol, row)
		screen.tcell.SetContent(adjCol, row, mainc, combc, style.Underline(true))
	}
}

// Colorize changes the color of text on the screen.
func (screen *Screen) Colorize(row int, colors []syntaxcolor.LineColor, offset int) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	cols, _ := screen.tcell.Size()
	for _, lc := range colors {
		style := tcell.StyleDefault.Foreground(lc.Fg).Background(lc.Bg)
		for col := lc.Start; col < lc.End; col++ {
			adjCol := col - offset + screen.gutterWidth
			if adjCol > cols || adjCol < 0 {
				continue
			}
			mainc, combc, _, _ := screen.tcell.GetContent(adjCol, row)
			screen.tcell.SetContent(adjCol, row, mainc, combc, style)
		}
	}
}

// DrawLeftBar draws a vertical bar character at the left edge of a row.
// Used to visually indicate code blocks in markdown files.
func (screen *Screen) DrawLeftBar(row int, color tcell.Color) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	style := tcell.StyleDefault.Foreground(color)
	screen.tcell.SetContent(0, row, '▌', nil, style)
}

// PrintableRune uses the charMode to convert the rune into
// a printable rune.
func (screen *Screen) PrintableRune(c rune) (rune, int) {
	if c < 32 {
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
	style := toStyle(fg, bg)
	k := 0
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	for _, c := range s {
		r, n := screen.PrintableRune(c)
		if n <= 0 {
			continue
		}
		screen.tcell.SetContent(col+k+screen.gutterWidth, row, r, nil, style)
		k += n
	}
}

// WriteMessage writes a status-line message.
func (screen *Screen) WriteMessage(msg string) {
	if len(msg) == 0 {
		return
	}
	_, rows := screen.Size()
	screen.WriteString(rows-1, 0, msg+"  ")
}

// Notify writes a status-line notification.
func (screen *Screen) Notify(msg string) {
	cols, rows := screen.Size()
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
	adjCol := col + screen.gutterWidth
	mainc, combc, style, _ := screen.tcell.GetContent(adjCol, row)
	screen.tcell.SetContent(adjCol, row, mainc, combc, style.Reverse(true))
}

// HighlightRange reverses the screen color over a range of rows/columns.
func (screen *Screen) HighlightRange(startRow, endRow, startCol, endCol int) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			adjCol := col + screen.gutterWidth
			mainc, combc, style, _ := screen.tcell.GetContent(adjCol, row)
			screen.tcell.SetContent(adjCol, row, mainc, combc, style.Reverse(true))
		}
	}
}

// ColorRange colors a range of cells.
func (screen *Screen) ColorRange(startRow, endRow, startCol, endCol int, fg, bg Attribute) {
	screen.tbMutex.Lock()
	defer screen.tbMutex.Unlock()
	style := toStyle(fg, bg)
	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			adjCol := col + screen.gutterWidth
			mainc, combc, _, _ := screen.tcell.GetContent(adjCol, row)
			screen.tcell.SetContent(adjCol, row, mainc, combc, style)
		}
	}
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

// GetTcell returns the underlying tcell.Screen for keyboard polling.
func (screen *Screen) GetTcell() tcell.Screen {
	return screen.tcell
}
