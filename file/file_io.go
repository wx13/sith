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
		strLine := file.Buffer[row+file.rowOffset].tabs2spaces().ToString()
		file.screen.Colorize(row, file.SyntaxRules.Colorize(strLine), file.colOffset)
	}
	for row := len(slice); row < rows-1; row++ {
		file.screen.WriteString(row, 0, "~")
	}
}

// ReadFile reads in a file (if it exists).
func (file *File) ReadFile(name string) {

	fileInfo, err := os.Stat(name)
	if err != nil {
		file.Buffer = MakeBuffer([]string{""})
	} else {
		file.fileMode = fileInfo.Mode()

		byteBuf, err := ioutil.ReadFile(name)
		stringBuf := []string{""}
		if err == nil {
			stringBuf = strings.Split(string(byteBuf), "\n")
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

func (file *File) Save() string {
	contents := file.toString()
	err := ioutil.WriteFile(file.Name, []byte(contents), file.fileMode)
	if err != nil {
		return ("Could not save to file: " + file.Name)
	} else {
		file.savedBuffer = file.Buffer.DeepDup()
		return ("Saved to: " + file.Name)
	}
}
