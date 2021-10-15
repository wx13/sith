package ui

import (
	"errors"
	"strings"

	"github.com/wx13/sith/autocomplete"
)

// Prompt is a user-input prompt.
type Prompt struct {
	oldRow, oldCol   int
	row, col         int
	question, answer string
	screen           Screen
	keyboard         Keyboard
}

// MakePrompt creates a new prompt object.
func MakePrompt(screen Screen, keyboard Keyboard) Prompt {
	_, rows := screen.Size()
	row := rows - 1
	return Prompt{screen: screen, row: row, keyboard: keyboard}
}

// GetRune expects a single keypress answer and returns the rune.
func (prompt *Prompt) GetRune(question string) rune {

	prompt.screen.WriteMessage(question)
	prompt.screen.Flush()

	cmd, r := prompt.keyboard.GetKey()
	switch cmd {
	case "ctrlC":
		return 0
	case "char":
		return r
	default:
		return 0
	}

}

// AskYesNo expects either y or n as an answer.
func (prompt *Prompt) AskYesNo(question string) (bool, error) {
	prompt.screen.WriteMessage(question)
	prompt.screen.Flush()
	_, ch := prompt.keyboard.GetKey()
	if strings.ToLower(string(ch)) == "y" {
		return true, nil
	} else if strings.ToLower(string(ch)) == "n" {
		return false, nil
	} else {
		return false, errors.New("Cancel")
	}
}

func (prompt *Prompt) saveCursor() {
	prompt.oldRow = prompt.screen.Row()
	prompt.oldCol = prompt.screen.Col()
}

func (prompt *Prompt) restoreCursor() {
	prompt.screen.SetCursor(prompt.oldRow, prompt.oldCol)
}

// Show displays the prompt to the user.
func (prompt *Prompt) show() {
	prompt.screen.WriteMessage(prompt.question + " " + prompt.answer)
	prompt.screen.SetCursor(prompt.row, prompt.col+len(prompt.question)+1)
	prompt.screen.Flush()
}

func (prompt *Prompt) clear() {
	spaces := strings.Repeat(" ", len(prompt.answer))
	prompt.screen.WriteString(prompt.row, len(prompt.question)+1, spaces)
}

func (prompt *Prompt) delete() {
	prompt.answer = prompt.answer[:prompt.col] + prompt.answer[prompt.col+1:]
	prompt.screen.WriteString(prompt.row, len(prompt.question)+1+len(prompt.answer), " ")
}

// Ask asks the user a question, expecting a string response.
func (prompt *Prompt) Ask(question string, history []string, completers ...func(string) []string) (string, error) {

	prompt.saveCursor()
	prompt.question = question

	prompt.screen.WriteMessage(question)
	prompt.screen.Flush()

	histIdx := -1

loop:
	for {

		prompt.show()

		cmd, r := prompt.keyboard.GetKey()
		switch cmd {
		case "backspace":
			if prompt.col > 0 {
				prompt.col--
				prompt.delete()
			}
		case "delete":
			if prompt.col < len(prompt.answer) {
				prompt.delete()
			}
		case "enter":
			break loop
		case "ctrlE":
			prompt.col = len(prompt.answer)
		case "ctrlA":
			prompt.col = 0
		case "arrowLeft":
			if prompt.col > 0 {
				prompt.col--
			}
		case "arrowRight":
			if prompt.col < len(prompt.answer) {
				prompt.col++
			}
		case "arrowUp":
			prompt.clear()
			if histIdx < len(history)-1 {
				histIdx++
				prompt.answer = history[histIdx]
			}
		case "arrowDown":
			prompt.clear()
			if histIdx > 0 {
				histIdx--
				prompt.answer = history[histIdx]
			}
		case "ctrlC":
			prompt.answer = ""
			prompt.restoreCursor()
			return "", errors.New("Cancel")
		case "ctrlK":
			prompt.clear()
			prompt.answer = prompt.answer[:prompt.col]
		case "ctrlU":
			prompt.clear()
			prompt.answer = prompt.answer[prompt.col:]
			prompt.col = 0
		case "ctrlL":
			prompt.clear()
			prompt.answer = ""
			prompt.col = 0
		case "tab":
			if len(completers) > 0 && completers[0] != nil {
				words := strings.Fields(prompt.answer)
				token := words[len(words)-1]
				results := completers[0](token)
				prefix := autocomplete.GetCommonPrefix(results)
				if len(results) == 0 {
					break
				}
				ans := results[0]
				if len(prefix) > len(token) {
					ans = prefix
				} else if len(results) > 1 {
					continue
				}
				prompt.col += len(ans)
				prompt.answer += ans
			}
		case "unknown":
		case "ctrlT":
			prompt.answer = prompt.answer[:prompt.col] + "\t" + prompt.answer[prompt.col:]
			prompt.col++
		case "char":
			prompt.answer = prompt.answer[:prompt.col] + string(r) + prompt.answer[prompt.col:]
			prompt.col++
		case "space":
			prompt.answer = prompt.answer[:prompt.col] + " " + prompt.answer[prompt.col:]
			prompt.col++
		default:
		}
	}
	prompt.clear()
	prompt.restoreCursor()
	return prompt.answer, nil
}

// GetPromptAnswer is a wrapper arount Ask, which handles some history stuff.
func (prompt *Prompt) GetAnswer(question string, history *[]string, completers ...func(string) []string) string {
	answer, err := prompt.Ask(question, *history, completers...)
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
