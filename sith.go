package main

import (
	"fmt"
	"os"

	"github.com/wx13/sith/editor"
	"github.com/wx13/sith/version"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("sith", version.Get())
		return
	}

	session := editor.NewEditor()
	defer session.Quit()

	session.OpenFiles(os.Args[1:])

	session.Flush()
	session.KeepFlushed()
	session.Listen()
}
