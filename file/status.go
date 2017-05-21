package file

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

// IsModified checks to see if a file has been modified.
func (file *File) IsModified() bool {
	return !file.buffer.Equals(&file.savedBuffer)
}

// WriteStatus writes the status line.
func (file *File) WriteStatus(row, col int) {

	status := ""
	if file.MultiCursor.Length() > 1 {
		status = fmt.Sprintf("%dC", file.MultiCursor.Length())
		file.addToStatus(status, row, &col, termbox.ColorBlack, termbox.ColorRed)
	}

	if file.autoIndent {
		file.addToStatus("->", row, &col, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault)
	}

	if file.autoTab {
		if file.tabString == "\t" {
			status = "1t"
		} else {
			status = fmt.Sprintf("%ds", len(file.tabString))
		}
		if !file.tabDetect {
			status += "*"
		}
		file.addToStatus(status, row, &col, termbox.ColorGreen, termbox.ColorDefault)
	}

	if !file.tabHealth {
		file.addToStatus("MixedIndent", row, &col, termbox.ColorRed, termbox.ColorDefault)
	}

	if file.newline != "\n" {
		status = strings.Replace(file.newline, "\n", "\\n", -1)
		status = strings.Replace(status, "\r", "\\r", -1)
		file.addToStatus(status, row, &col, termbox.ColorYellow, termbox.ColorDefault)
	}

	file.statusMutex.Lock()

	if file.notification != "" {
		file.addToStatus(file.notification, row, &col, termbox.ColorCyan, termbox.ColorDefault)
	}

	if file.clearNotification {
		file.clearNotification = false
		file.notification = ""
	} else {
		file.clearNotification = true
	}

	file.statusMutex.Unlock()

}

func (file *File) addToStatus(msg string, row int, col *int, fg, bg termbox.Attribute) {
	*col -= len(msg) + 2
	file.screen.WriteStringColor(row, *col, msg, fg, bg)
}

// NotifyUser displays a message to the user.
func (file *File) NotifyUser(msg string) {
	file.statusMutex.Lock()
	if len(file.notification) > 0 {
		file.notification += " | "
	}
	file.notification += msg
	file.clearNotification = false
	file.statusMutex.Unlock()
	file.RequestFlush()
}
