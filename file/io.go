package file

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/file/buffer"
)

// Flush writes the buffer contents to the screen.
func (file *File) Flush() {
	file.ComputeIndent()
	cols, rows := termbox.Size()
	slice := file.Slice(rows-1, cols)
	file.screen.Clear()
	for row, str := range slice {
		file.screen.WriteString(row, 0, str)
		fullStr := file.buffer.GetRow(row + file.rowOffset).Tabs2spaces().ToString()
		file.screen.Colorize(row, file.SyntaxRules.Colorize(fullStr), file.colOffset)
	}
	for row := len(slice); row < rows-1; row++ {
		file.screen.WriteString(row, 0, "~")
	}
}

func (file *File) setNewline(bufferStr string) {
	file.newline = "\n"
	count := strings.Count(bufferStr, "\n")
	c := strings.Count(bufferStr, "\r")
	if c > count {
		count = c
		file.newline = "\r"
	}
	for _, newline := range []string{"\n\r", "\r\n"} {
		c := strings.Count(bufferStr, newline)
		if c > count/2 {
			count = c
			file.newline = newline
		}
	}
}

// ReadFile reads in a file (if it exists).
func (file *File) ReadFile(name string) {

	fileInfo, err := os.Stat(name)
	if err != nil {
		file.buffer.ReplaceBuffer(buffer.MakeBuffer([]string{""}))
	} else {
		file.fileMode = fileInfo.Mode()
		stringBuf := []string{""}

		byteBuf, err := ioutil.ReadFile(name)
		if err == nil {
			file.setNewline(string(byteBuf))
			stringBuf = strings.Split(string(byteBuf), file.newline)
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

func (file *File) Save() {
	file.SnapshotSaved()
	contents := file.ToString()
	err := ioutil.WriteFile(file.Name, []byte(contents), file.fileMode)
	if err != nil {
		file.NotifyUser("Could not save to file: " + file.Name)
	} else {
		file.savedBuffer.ReplaceBuffer(file.buffer.DeepDup())
		file.NotifyUser("Saved to file: " + file.Name)
	}
}
