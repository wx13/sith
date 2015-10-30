package file

import "io/ioutil"
import "os"
import "github.com/nsf/termbox-go"
import "strings"
import "errors"

func (file *File) Flush() {
	file.ComputeIndent()
	cols, rows := termbox.Size()
	slice := file.Slice(rows-1, cols)
	file.screen.Clear()
	for row, str := range slice {
		file.screen.WriteString(row, 0, str)
		strLine := file.Buffer[row+file.rowOffset].tabs2spaces().ToString()
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
		file.Buffer = MakeBuffer([]string{""})
	} else {
		file.fileMode = fileInfo.Mode()
		stringBuf := []string{""}

		byteBuf, err := ioutil.ReadFile(name)
		if err == nil {
			file.setNewline(string(byteBuf))
			stringBuf = strings.Split(string(byteBuf), file.newline)
		}

		file.Buffer = MakeBuffer(stringBuf)
	}

	file.Snapshot()
	file.savedBuffer = file.Buffer.DeepDup()

	select {
	case file.flushChan <- struct{}{}:
	default:
	}

}

func (file *File) Save() error {
	contents := file.toString()
	err := ioutil.WriteFile(file.Name, []byte(contents), file.fileMode)
	if err != nil {
		return errors.New("Could not save to file: " + file.Name)
	} else {
		file.savedBuffer = file.Buffer.DeepDup()
		return nil
	}
}
