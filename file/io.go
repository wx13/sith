package file

import (
	"crypto/md5"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/syntaxcolor"
)

// Flush writes the buffer contents to the screen.
func (file *File) Flush() {
	file.ComputeIndent()
	cols, rows := termbox.Size()
	slice := file.Slice(rows-1, cols)
	file.screen.Clear()
	for row, str := range slice {
		file.screen.WriteString(row, 0, str)
		fullStr := file.buffer.GetRow(row + file.rowOffset).Tabs2spaces(file.tabWidth).ToString()
		file.screen.Colorize(row, file.SyntaxRules.Colorize(fullStr), file.colOffset)
	}
	for row := len(slice); row < rows-1; row++ {
		file.screen.WriteString(row, 0, "~")
	}
	file.ColorBracketMatch(rows)
}

// ColorBracketMatch colorizes a matching bracket character.
func (file *File) ColorBracketMatch(rows int) {
	cursor := file.MultiCursor.GetCursor(0)
	row, col := cursor.RowCol()
	row, col, err := file.buffer.BracketMatch(row, col, row+rows)
	if err != nil {
		return
	}
	col = file.buffer.GetRow(row).TabCursorPos(col, file.tabWidth)
	lc := []syntaxcolor.LineColor{
		{
			Fg:    termbox.ColorRed | termbox.AttrBold,
			Start: col,
			End:   col + 1,
		},
	}

	file.screen.Colorize(row-file.rowOffset, lc, file.colOffset)
}

// setNewline determines (estimates) the newline string that the file uses.
// It simply looks for most used newline string.
func (file *File) setNewline(bufferStr string) {

	// Default to line feed.
	file.newline = "\n"
	count := strings.Count(bufferStr, "\n")

	// Check if carriage return is more popular.
	c := strings.Count(bufferStr, "\r")
	if c > count {
		count = c
		file.newline = "\r"
	}

	// Check for CRLF or LFCR.
	for _, newline := range []string{"\n\r", "\r\n"} {
		c := strings.Count(bufferStr, newline)
		// Factor of two to prevent overcounting.
		if c > count/2 {
			count = c
			file.newline = newline
		}
	}

}

// ReadFile reads in a file (if it exists).
func (file *File) ReadFile(name string, wgs ...*sync.WaitGroup) {

	for _, wg := range wgs {
		defer wg.Done()
	}

	file.md5sum = md5.Sum([]byte(""))

	fileInfo, err := os.Stat(name)
	if err != nil {
		file.buffer.ReplaceBuffer(buffer.MakeBuffer([]string{""}))
		file.modTime = time.Now()
	} else {
		file.fileMode = fileInfo.Mode()
		file.modTime = fileInfo.ModTime()
		stringBuf := []string{""}

		byteBuf, err := ioutil.ReadFile(name)
		if err == nil {
			file.setNewline(string(byteBuf))
			stringBuf = strings.Split(string(byteBuf), file.newline)
			file.md5sum = md5.Sum(byteBuf)
		}

		file.buffer.ReplaceBuffer(buffer.MakeBuffer(stringBuf))
	}

	file.ForceSnapshot()
	file.SnapshotSaved()
	file.savedBuffer.ReplaceBuffer(file.buffer.DeepDup())

	file.RequestFlush()

}

// RequestFlush places a flush request on the flush channel.
func (file *File) RequestFlush() {
	select {
	case file.flushChan <- struct{}{}:
	default:
	}
}

// RequestSave places a save request on the save channel.
func (file *File) RequestSave() {
	select {
	case file.saveChan <- struct{}{}:
	default:
	}
}

func (file *File) processSaveRequests() {
	for {
		<-file.saveChan
		file.Save()
	}
}

// Save saves a file.
func (file *File) Save() {
	if file.autoFmt {
		err := file.Fmt()
		if err != nil {
			file.NotifyUser(err.Error())
		}
	}
	file.SnapshotSaved()
	contents := []byte(file.ToString())
	err := ioutil.WriteFile(file.Name, contents, file.fileMode)
	if err != nil {
		file.NotifyUser("Save Failed: " + err.Error())
	} else {
		file.savedBuffer.ReplaceBuffer(file.buffer.DeepDup())
		file.NotifyUser("Saved.")
		file.modTime = time.Now()
		file.md5sum = md5.Sum(contents)
	}
}
