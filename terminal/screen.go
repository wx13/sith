package terminal

import "errors"
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

// ------------------------------------
// Menu
// ------------------------------------

type Menu struct {
	cols, rows  int
	col0, row0  int
	screen      *Screen
	keyboard    *Keyboard
	selection   int
	rowShift    int
	borderColor termbox.Attribute
}

func NewMenu(screen *Screen) *Menu {
	menu := Menu{}
	menu.cols, menu.rows = termbox.Size()
	menu.cols -= 8
	menu.rows -= 8
	if menu.cols > 70 {
		menu.cols = 70
	}
	menu.col0 = 4
	menu.row0 = 4
	menu.screen = screen
	menu.keyboard = NewKeyboard()
	menu.borderColor = termbox.ColorBlue
	return &menu
}

func (menu *Menu) Clear() {
	borderColor := menu.borderColor
	menu.screen.WriteStringColor(menu.row0-1, menu.col0-2, strings.Repeat(" ", menu.cols+4), borderColor, borderColor)
	menu.screen.WriteStringColor(menu.row0+menu.rows, menu.col0-2, strings.Repeat(" ", menu.cols+4), borderColor, borderColor)
	for row := 0; row < menu.rows; row++ {
		menu.screen.WriteStringColor(menu.row0+row, menu.col0-2, "  ", borderColor, borderColor)
		menu.screen.WriteStringColor(menu.row0+row, menu.col0+menu.cols, "  ", borderColor, borderColor)
		menu.screen.WriteString(menu.row0+row, menu.col0, strings.Repeat(" ", menu.cols))
	}
}

func (menu *Menu) ShowSearchStr(searchStr string) {
	borderColor := menu.borderColor
	menu.screen.WriteStringColor(menu.row0-1, menu.col0, searchStr, termbox.ColorBlack, borderColor)
}

func (menu *Menu) Show(choices []string) {
	menu.Clear()
	if menu.selection >= menu.rows-1+menu.rowShift {
		menu.rowShift = menu.selection - menu.rows + 1
	}
	if menu.selection < menu.rowShift {
		menu.rowShift = menu.selection
	}
	for row := 0; row < menu.rows; row++ {
		idx := menu.rowShift + row
		if idx >= len(choices) {
			break
		}
		line := choices[idx]
		if len(line) >= menu.cols {
			line = line[:menu.cols]
		}
		menu.screen.WriteString(menu.row0+row, menu.col0, line)
	}
	for col := 0; col < menu.cols; col++ {
		menu.screen.Highlight(menu.row0+menu.selection-menu.rowShift, menu.col0+col)
	}
}

func (menu *Menu) Choose(choices []string) int {
	if len(choices) < menu.rows {
		menu.rows = len(choices)
	}
	menu.selection = 0
	searchStr := ""
loop:
	for {
		menu.Show(choices)
		menu.ShowSearchStr(searchStr)
		menu.screen.Flush()
		cmd, r := menu.keyboard.GetKey()
		switch cmd {
		case "enter":
			break loop
		case "ctrlC":
			return -1
		case "arrowDown":
			if menu.selection < len(choices)-1 {
				menu.selection++
			}
		case "arrowUp":
			if menu.selection > 0 {
				menu.selection--
			}
		case "pageDown":
			menu.selection += 10
			if menu.selection >= len(choices) {
				menu.selection = len(choices) - 1
			}
		case "pageUp":
			menu.selection -= 10
			if menu.selection < 0 {
				menu.selection = 0
			}
		case "unknown":
		case "char":
			searchStr += string(r)
			menu.selection = menu.Search(choices, searchStr)
		case "backspace":
			if len(searchStr) > 0 {
				searchStr = searchStr[:len(searchStr)-1]
				menu.selection = menu.Search(choices, searchStr)
			}
		case "ctrlU":
			searchStr = ""
		case "ctrlN":
			menu.selection = menu.SearchNext(choices, searchStr)
		default:
		}
	}
	return menu.selection
}

func (menu *Menu) Search(choices []string, searchStr string) int {
	for index := 0; index < len(choices); index++ {
		if strings.Contains(strings.ToLower(choices[index]), strings.ToLower(searchStr)) {
			return index
		}
	}
	return menu.selection
}

func (menu *Menu) SearchNext(choices []string, searchStr string) int {
	index := menu.selection
	for {
		index++
		if index >= len(choices) {
			index = 0
		}
		if index == menu.selection {
			break
		}
		if strings.Contains(strings.ToLower(choices[index]), strings.ToLower(searchStr)) {
			break
		}
	}
	return index
}

// ------------------------------------
// Prompt
// ------------------------------------

type Prompt struct {
	oldRow, oldCol   int
	row, col         int
	question, answer string
	screen           *Screen
	keyboard         *Keyboard
}

func MakePrompt(screen *Screen) Prompt {
	_, rows := termbox.Size()
	row := rows - 1
	return Prompt{screen: screen, row: row, keyboard: NewKeyboard()}
}

func (prompt *Prompt) AskYesNo(question string) (bool, error) {
	prompt.screen.WriteMessage(question)
	prompt.screen.Flush()
	ev := termbox.PollEvent()
	if strings.ToLower(string(ev.Ch)) == "y" {
		return true, nil
	} else if strings.ToLower(string(ev.Ch)) == "n" {
		return false, nil
	} else {
		return false, errors.New("Cancel")
	}
}

func (prompt *Prompt) SaveCursor() {
	prompt.oldRow = prompt.screen.row
	prompt.oldCol = prompt.screen.col
}

func (prompt *Prompt) RestoreCursor() {
	prompt.screen.SetCursor(prompt.oldRow, prompt.oldCol)
}

func (prompt *Prompt) Show() {
	prompt.screen.WriteMessage(prompt.question + " " + prompt.answer)
	prompt.screen.SetCursor(prompt.row, prompt.col+len(prompt.question)+1)
	prompt.screen.Flush()
}

func (prompt *Prompt) Clear() {
	spaces := strings.Repeat(" ", len(prompt.answer))
	prompt.screen.WriteString(prompt.row, len(prompt.question)+1, spaces)
}

func (prompt *Prompt) Delete() {
	prompt.answer = prompt.answer[:prompt.col] + prompt.answer[prompt.col+1:]
	prompt.screen.WriteString(prompt.row, len(prompt.question)+1+len(prompt.answer), " ")
}

func (prompt *Prompt) Ask(question string, history []string) (string, error) {

	prompt.SaveCursor()
	prompt.question = question

	prompt.screen.WriteMessage(question)
	prompt.screen.Flush()

	histIdx := -1

loop:
	for {

		prompt.Show()

		cmd, r := prompt.keyboard.GetKey()
		switch cmd {
		case "backspace":
			if prompt.col > 0 {
				prompt.col -= 1
				prompt.Delete()
			}
		case "delete":
			if prompt.col < len(prompt.answer) {
				prompt.Delete()
			}
		case "space":
			prompt.answer += " "
			prompt.col += 1
		case "tab":
		case "enter":
			break loop
		case "arrowLeft":
			if prompt.col > 0 {
				prompt.col -= 1
			}
		case "arrowRight":
			if prompt.col < len(prompt.answer) {
				prompt.col += 1
			}
		case "arrowUp":
			prompt.Clear()
			if histIdx < len(history)-1 {
				histIdx++
				prompt.answer = history[histIdx]
			}
		case "arrowDown":
			prompt.Clear()
			if histIdx > 0 {
				histIdx--
				prompt.answer = history[histIdx]
			}
		case "ctrlC":
			prompt.answer = ""
			prompt.RestoreCursor()
			return "", errors.New("Cancel")
		case "ctrlK":
			prompt.Clear()
			prompt.answer = prompt.answer[:prompt.col]
		case "ctrlU":
			prompt.Clear()
			prompt.answer = prompt.answer[prompt.col:]
			prompt.col = 0
		case "ctrlL":
			prompt.Clear()
			prompt.answer = ""
			prompt.col = 0
		case "unknown":
		case "char":
			prompt.answer = prompt.answer[:prompt.col] + string(r) + prompt.answer[prompt.col:]
			prompt.col += 1
		default:
		}
	}
	prompt.Clear()
	prompt.RestoreCursor()
	return prompt.answer, nil
}
