package main

import (
	"fmt"
	"os"

	"github.com/wx13/sith/editor"
	"github.com/wx13/sith/version"
)

func main() {
	args := os.Args[1:]

	// Handle --version flag
	if len(args) > 0 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Println("sith", version.Get())
		return
	}

	// Check for --resume flag
	resume := false
	fileArgs := []string{}
	for _, arg := range args {
		if arg == "-r" || arg == "--resume" {
			resume = true
		} else {
			fileArgs = append(fileArgs, arg)
		}
	}

	e := editor.NewEditor()
	defer e.Quit()

	// Try to restore session if:
	// - --resume flag is set, OR
	// - No files specified and a saved session exists
	restored := false
	if resume || (len(fileArgs) == 0 && e.HasSavedSession()) {
		restored = e.TryRestoreSession(resume)
	}

	if !restored {
		// Normal startup with specified files (or empty)
		e.OpenFiles(fileArgs)
	}

	e.Flush()
	e.KeepFlushed()
	e.Listen()
}
