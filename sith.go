package main

import (
	"os"

	"github.com/wx13/sith/editor"
)

func main() {

	session := editor.NewEditor()
	defer session.Quit()

	session.OpenFiles(os.Args[1:])

	session.Flush()
	session.KeepFlushed()
	session.Listen()

}
