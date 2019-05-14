package ui

import (
	"strings"

	"github.com/wx13/sith/terminal"
)

// Menu helps create a searchable, flexible, on-screen menu.
type Menu struct {
	cols, rows  int
	col0, row0  int
	screen      Screen
	keyboard    Keyboard
	selections  []int
	cursor      int
	rowShift    int
	borderColor terminal.Attribute
	choices     []string
}

// NewMenu creates a new Menu object.
func NewMenu(screen Screen, keyboard Keyboard) *Menu {
	menu := Menu{}
	menu.screen = screen
	menu.keyboard = keyboard
	menu.setDims()
	menu.borderColor = terminal.ColorBlue
	menu.selections = []int{}
	return &menu
}

func (menu *Menu) setDims() {
	cols, rows := menu.screen.Size()
	menu.rows = rows - 8
	menu.col0 = 4
	menu.row0 = 4
	if len(menu.choices) < menu.rows {
		menu.rows = len(menu.choices)
	}
	menu.cols = 4
	for _, choice := range menu.choices {
		if len(choice)+2 > menu.cols {
			menu.cols = len(choice) + 2
		}
	}
	if menu.cols > cols-8 {
		menu.cols = cols - 8
	}
}

// Clear clears the on-screen menu.
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

func (menu *Menu) showSearchStr(searchStr string) {
	borderColor := menu.borderColor
	menu.screen.WriteStringColor(menu.row0-1, menu.col0, searchStr, terminal.ColorWhite|terminal.AttrBold, borderColor)
}

// Show displays a menu of choices on the screen.
func (menu *Menu) Show(choices []string) {
	menu.Clear()
	if menu.cursor >= menu.rows-1+menu.rowShift {
		menu.rowShift = menu.cursor - menu.rows + 1
	}
	if menu.cursor < menu.rowShift {
		menu.rowShift = menu.cursor
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
	for _, row := range menu.selections {
		r := menu.row0 + row - menu.rowShift
		menu.screen.ColorRange(r, r, menu.col0, menu.col0+menu.cols-1, terminal.ColorGreen, terminal.ColorDefault)
	}
	r := menu.row0 + menu.cursor - menu.rowShift
	menu.screen.HighlightRange(r, r, menu.col0, menu.col0+menu.cols-1)
}

// Toggles the row in the selections list and returns the list.
func (menu *Menu) toggleSelection(row int) {
	for i, r := range menu.selections {
		if r == row {
			menu.selections = append(menu.selections[:i], menu.selections[i+1:]...)
			return
		}
	}
	menu.selections = append(menu.selections, row)
}

func (menu *Menu) appendSelection(row int) []int {
	for _, r := range menu.selections {
		if r == row {
			return menu.selections
		}
	}
	return append(menu.selections, row)
}

// Choose is the main interaction loop for the menu. It takes three required
// inputs: a list of choices (strings), a current-selection-index (int), and
// an initial search string (often ""). Optionally you can also pass a list
// of "keys" (strings) to listen for.
//
// The function returns two things: the integer index of the current selection,
// and the string description of the key that caused the program to exit.
func (menu *Menu) Choose(choices []string, idx int, searchStr string,
	keys ...string) (int, string) {
	all, str := menu.ChooseMulti(choices, idx, searchStr, keys...)
	return all[0], str
}
func (menu *Menu) ChooseMulti(choices []string, idx int, searchStr string,
	keys ...string) ([]int, string) {

	menu.choices = choices
	menu.setDims()
	menu.cursor = idx
	for {
		menu.Show(choices)
		menu.showSearchStr(searchStr)
		menu.screen.Flush()
		cmd, r := menu.keyboard.GetKey()
		switch cmd {
		case "enter":
			return menu.appendSelection(menu.cursor), ""
		case "ctrlC":
			return menu.appendSelection(menu.cursor), "cancel"
		case "arrowDown":
			if menu.cursor < len(choices)-1 {
				menu.cursor++
			}
		case "arrowUp":
			if menu.cursor > 0 {
				menu.cursor--
			}
		case "pageDown":
			menu.cursor += 10
			if menu.cursor >= len(choices) {
				menu.cursor = len(choices) - 1
			}
		case "pageUp":
			menu.cursor -= 10
			if menu.cursor < 0 {
				menu.cursor = 0
			}
		case "unknown":
		case "char":
			searchStr += string(r)
			menu.cursor = menu.Search(choices, searchStr)
		case "backspace":
			if len(searchStr) > 0 {
				searchStr = searchStr[:len(searchStr)-1]
				menu.cursor = menu.Search(choices, searchStr)
			}
		case "ctrlU":
			searchStr = ""
		case "ctrlN":
			menu.cursor = menu.SearchNext(choices, searchStr)
		case "altS":
			menu.toggleSelection(menu.cursor)
		case "altC":
			menu.selections = []int{}
		default:
		}
		// User keys
		for _, key := range keys {
			if cmd == key {
				return menu.appendSelection(menu.cursor), key
			}
		}
	}
}

// Search searches menu options for a partial string match.
func (menu *Menu) Search(choices []string, searchStr string) int {
	for index := 0; index < len(choices); index++ {
		if strings.Contains(strings.ToLower(choices[index]), strings.ToLower(searchStr)) {
			return index
		}
	}
	return menu.cursor
}

// SearchNext searches menu options from the current option on.
func (menu *Menu) SearchNext(choices []string, searchStr string) int {
	index := menu.cursor
	for {
		index++
		if index >= len(choices) {
			index = 0
		}
		if index == menu.cursor {
			break
		}
		if strings.Contains(strings.ToLower(choices[index]), strings.ToLower(searchStr)) {
			break
		}
	}
	return index
}
