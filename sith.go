package main

import "os"

func main() {

	editor := NewEditor()
	defer editor.Quit()

	editor.OpenFiles(os.Args[1:])

	editor.Flush()
	editor.KeepFlushed()
	editor.Listen()

}
