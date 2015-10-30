package file

import "github.com/nsf/termbox-go"
import "fmt"
import "strings"

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
	file.AddToStatus(status, row, &col, termbox.ColorYellow, termbox.ColorDefault)

	if len(file.MultiCursor) > 1 {
		status = fmt.Sprintf("%dC", len(file.MultiCursor))
		file.AddToStatus(status, row, &col, termbox.ColorBlack, termbox.ColorRed)
	}

	if file.autoIndent {
		file.AddToStatus("->", row, &col, termbox.ColorRed | termbox.AttrBold, termbox.ColorDefault)
	}

	if file.autoTab {
		if file.tabString == "\t" {
			status = "1t"
		} else {
			status = fmt.Sprintf("%ds", len(file.tabString))
		}
		file.AddToStatus(status, row, &col, termbox.ColorGreen, termbox.ColorDefault)
	}

	if !file.tabHealth {
		file.AddToStatus("MixedIndent", row, &col, termbox.ColorRed, termbox.ColorDefault)
	}

	if file.newline != "\n" {
		status = strings.Replace(file.newline, "\n", "\\n", -1)
		status = strings.Replace(status, "\r", "\\r", -1)
		file.AddToStatus(status, row, &col, termbox.ColorYellow, termbox.ColorDefault)
	}

}

func (file *File) AddToStatus(msg string, row int, col *int, fg, bg termbox.Attribute) {
	*col -= len(msg) + 2
	file.screen.WriteStringColor(row, *col, msg, fg, bg)
}
