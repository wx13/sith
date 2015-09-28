package main

import "errors"
import "github.com/nsf/termbox-go"
import "strings"

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
	termbox.Close()
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

func (screen *Screen) WriteString(row, col int, s string) {
	screen.WriteStringColor(row, col, s, screen.fg, screen.bg)
}

func (screen *Screen) Colorize(row int, colors []LineColor) {
	cells := termbox.CellBuffer()
	cols, _ := termbox.Size()
	for _, lc := range colors {
		for col := lc.start; col < lc.end; col++ {
			j := row * cols + col
			if j < 0 || j >= len(cells) {
				continue
			}
			cells[j].Bg, cells[j].Fg = lc.bg, lc.fg
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



