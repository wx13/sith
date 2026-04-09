package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/wx13/sith/editor"
)

func printVersion() {
	version := "(unknown)"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		version = info.Main.Version
	}
	fmt.Println("sith", version)
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		printVersion()
		return
	}

	session := editor.NewEditor()
	defer session.Quit()

	session.OpenFiles(os.Args[1:])

	session.Flush()
	session.KeepFlushed()
	session.Listen()
}
