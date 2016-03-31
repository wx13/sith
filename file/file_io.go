package file

import "io/ioutil"
import "os"
import "github.com/nsf/termbox-go"
import "strings"

func (file *File) Flush() {
	file.ComputeIndent()
	cols, rows := termbox.Size()
	slice := file.Slice(rows-1, cols)
	file.screen.Clear()
	for row, str := range slice {
		file.screen.WriteString(row, 0, str)
		strLine := file.buffer[row+file.rowOffset].Tabs2spaces().ToString()
		file.screen.Colorize(row, file.SyntaxRules.Colorize(strLine), file.colOffset)
	}
	for row := len(slice); row < rows-1; row++ {
		file.screen.WriteString(row, 0, "~")
	}
}

func (file *File) setNewline(buffer string) {
	file.newline = "\n"
	count := strings.Count(buffer, "\n")
	c := strings.Count(buffer, "\r")
	if c > count {
		count = c
		file.newline = "\r"
	}
	for _, newline := range []string{"\n\r", "\r\n"} {
		c := strings.Count(buffer, newline)
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
		file.buffer = MakeBuffer([]string{""})
	} else {
		file.fileMode = fileInfo.Mode()
		stringBuf := []string{""}

		byteBuf, err := ioutil.ReadFile(name)
		if err == nil {
			file.setNewline(string(byteBuf))
			stringBuf = strings.Split(string(byteBuf), file.newline)
		}

		file.buffer = MakeBuffer(stringBuf)
	}

	file.ForceSnapshot()
	file.SnapshotSaved()
	file.savedBuffer = file.buffer.DeepDup()

	file.RequestFlush()

}

func (file *File) RequestFlush() {
	select {
	case file.flushChan <- struct{}{}:
	default:
	}
}

func (file *File) RequestSave() {
	select {
	case file.saveChan <- struct{}{}:
	default:
	}
}

func (file *File) ProcessSaveRequests() {
	for {
		<-file.saveChan
		file.Save()
	}
}

func (file *File) Save() {
	file.SnapshotSaved()
	contents := file.toString()
	err := ioutil.WriteFile(file.Name, []byte(contents), file.fileMode)
	if err != nil {
		file.NotifyUser("Could not save to file: " + file.Name)
	} else {
		file.savedBuffer = file.buffer.DeepDup()
		file.NotifyUser("Saved to file: " + file.Name)
	}
}
