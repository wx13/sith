package file

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

func (file *File) IsModified() bool {
	return !file.buffer.Equals(file.savedBuffer)
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
	if len(status) > 0 {
		file.AddToStatus(status, row, &col, termbox.ColorYellow, termbox.ColorDefault)
	}

	if file.MultiCursor.Length() > 1 {
		status = fmt.Sprintf("%dC", file.MultiCursor.Length())
		file.AddToStatus(status, row, &col, termbox.ColorBlack, termbox.ColorRed)
	}

	if file.autoIndent {
		file.AddToStatus("->", row, &col, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault)
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

	if file.notification != "" {
		file.AddToStatus(file.notification, row, &col, termbox.ColorCyan, termbox.ColorDefault)
	}

	if file.clearNotification {
		file.clearNotification = false
		file.notification = ""
	} else {
		file.clearNotification = true
	}

}

func (file *File) AddToStatus(msg string, row int, col *int, fg, bg termbox.Attribute) {
	*col -= len(msg) + 2
	file.screen.WriteStringColor(row, *col, msg, fg, bg)
}

func (file *File) NotifyUser(msg string) {
	if len(file.notification) > 0 {
		file.notification += " | "
	}
	file.notification += msg
	file.clearNotification = false
	file.RequestFlush()
}
