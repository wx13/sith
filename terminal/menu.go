package terminal

import (
	"strings"

	"github.com/nsf/termbox-go"
)

// Menu helps create a searchable, flexible, on-screen menu.
type Menu struct {
	cols, rows  int
	col0, row0  int
	screen      *Screen
	keyboard    *Keyboard
	selection   int
	rowShift    int
	borderColor termbox.Attribute
	choices     []string
}

// NewMenu creates a new Menu object.
func NewMenu(screen *Screen) *Menu {
	menu := Menu{}
	menu.setDims()
	menu.screen = screen
	menu.keyboard = NewKeyboard()
	menu.borderColor = termbox.ColorBlue
	return &menu
}

func (menu *Menu) setDims() {
	cols, rows := termbox.Size()
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
	menu.screen.WriteStringColor(menu.row0-1, menu.col0, searchStr, termbox.ColorWhite|termbox.AttrBold, borderColor)
}

// Show displays a menu of choices on the screen.
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

// Choose is the main interaction loop for the menu.
func (menu *Menu) Choose(choices []string, idx int) int {
	menu.choices = choices
	menu.setDims()
	menu.selection = idx
	searchStr := ""
loop:
	for {
		menu.Show(choices)
		menu.showSearchStr(searchStr)
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

// Search searches menu options for a partial string match.
func (menu *Menu) Search(choices []string, searchStr string) int {
	for index := 0; index < len(choices); index++ {
		if strings.Contains(strings.ToLower(choices[index]), strings.ToLower(searchStr)) {
			return index
		}
	}
	return menu.selection
}

// SearchNext searches menu options from the current option on.
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
