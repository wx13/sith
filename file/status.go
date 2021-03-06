package file

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/wx13/sith/terminal"
)

// IsModified checks to see if a file has been modified.
func (file *File) IsModified() bool {
	return !file.buffer.Equals(&file.savedBuffer)
}

// FileChanged checks to see if underlying file changed.
func (file *File) FileChanged() (bool, error) {
	fileInfo, err := os.Stat(file.Name)
	if err != nil {
		return false, err
	}
	if !fileInfo.ModTime().After(file.modTime) {
		return false, nil
	}
	byteBuf, err := ioutil.ReadFile(file.Name)
	if err != nil {
		return false, err
	}
	md5sum := md5.Sum(byteBuf)
	return md5sum != file.md5sum, nil
}

// WriteStatus writes the status line.
func (file *File) WriteStatus(row, col int) {

	if file.MultiCursor.Length() > 1 {
		status := fmt.Sprintf("%d%s", file.MultiCursor.Length(), file.MultiCursor.GetNavModeShort())
		file.addToStatus(status, row, &col,
			terminal.ColorGreen|terminal.AttrReverse|terminal.AttrBold,
			terminal.ColorDefault|terminal.AttrReverse)
	}

	if file.autoIndent {
		file.addToStatus("->", row, &col, terminal.ColorGreen, terminal.ColorDefault)
	}

	if file.autoFmt {
		ext := GetFileExt(file.Name)
		if file.fmtCmd != "" || ext == "go" {
			file.addToStatus("f", row, &col, terminal.ColorGreen, terminal.ColorDefault)
		}
	}

	if file.autoTab {
		var status string
		if file.tabString == "\t" {
			status = "1t"
		} else {
			status = fmt.Sprintf("%ds", len(file.tabString))
		}
		if !file.tabDetect {
			status += "*"
		}
		file.addToStatus(status, row, &col, terminal.ColorGreen, terminal.ColorDefault)
	}

	if !file.tabHealth {
		file.addToStatus("MixedIndent", row, &col, terminal.ColorRed, terminal.ColorDefault)
	}

	if file.newline != "\n" {
		status := strings.Replace(file.newline, "\n", "\\n", -1)
		status = strings.Replace(status, "\r", "\\r", -1)
		file.addToStatus(status, row, &col, terminal.ColorYellow, terminal.ColorDefault)
	}

	file.statusMutex.Lock()

	if file.notification != "" {
		file.addToStatus(file.notification, row, &col, terminal.ColorCyan, terminal.ColorDefault)
	}

	if file.clearNotification {
		file.clearNotification = false
		file.notification = ""
	} else {
		file.clearNotification = true
	}

	file.statusMutex.Unlock()

}

func (file *File) addToStatus(msg string, row int, col *int, fg, bg terminal.Attribute) {
	*col -= len(msg) + 1
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
