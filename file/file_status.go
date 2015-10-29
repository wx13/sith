package file

import "github.com/nsf/termbox-go"
import "fmt"

func (file *File) IsModified() bool {
	if len(file.Buffer) != len(file.savedBuffer) {
		return true
	}
	for row, _ := range file.Buffer {
		if file.Buffer[row].ToString() != file.savedBuffer[row].ToString() {
			return true
		}
	}
	return false
}

func (file *File) ModStatus() string {
	if file.IsModified() {
		return "Modified"
	} else {
		return ""
	}
}

func (file *File) WriteStatus(row, col int) {

	status := file.ModStatus()
	col -= len(status) + 2
	fg := termbox.ColorYellow
	bg := termbox.ColorDefault
	file.screen.WriteStringColor(row, col, status, fg, bg)

	if len(file.MultiCursor) > 1 {
		status = fmt.Sprintf("%dC", len(file.MultiCursor))
		col -= len(status) + 2
		fg := termbox.ColorBlack
		bg := termbox.ColorRed
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}

	if file.autoIndent {
		status = "->"
		col -= len(status) + 2
		fg := termbox.ColorRed | termbox.AttrBold
		bg := termbox.ColorDefault
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}

	if file.autoTab {
		if file.tabString == "\t" {
			status = "1t"
		} else {
			status = fmt.Sprintf("%ds", len(file.tabString))
		}
		col -= len(status) + 2
		fg := termbox.ColorGreen
		bg := termbox.ColorDefault
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}

	if !file.tabHealth {
		status = "MixedIndent"
		col -= len(status) + 2
		fg := termbox.ColorRed
		bg := termbox.ColorDefault
		file.screen.WriteStringColor(row, col, status, fg, bg)
	}

}
