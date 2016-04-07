package terminal

import (
	"errors"
	"github.com/nsf/termbox-go"
	"strings"
)

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
		case "enter":
			break loop
		case "ctrlE":
			prompt.col = len(prompt.answer)
		case "ctrlA":
			prompt.col = 0
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
		case "tab":
			prompt.answer = prompt.answer[:prompt.col] + "\t" + prompt.answer[prompt.col:]
			prompt.col += 1
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

func (screen *Screen) GetPromptAnswer(question string, history *[]string) string {
	answer, err := screen.Ask(question, *history)
	if err != nil {
		return ""
	}
	if answer == "" {
		if len(*history) == 0 {
			return ""
		}
		answer = (*history)[0]
	} else {
		*history = append([]string{answer}, *history...)
	}
	n := len(*history)
	if n > 10000 {
		*history = (*history)[(n - 10000):]
	}
	return answer
}
