package main

import (
	"github.com/wx13/sith/editor"
	"os"
)

func main() {

	session := editor.NewEditor()
	defer session.Quit()

	session.OpenFiles(os.Args[1:])

	session.Flush()
	session.KeepFlushed()
	session.Listen()

}
