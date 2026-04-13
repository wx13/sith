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

	session := editor.NewEditor()
	defer session.Quit()

	// Try to restore session if:
	// - --resume flag is set, OR
	// - No files specified and a saved session exists
	if resume || (len(fileArgs) == 0 && session.HasSavedSession()) {
		files, activeIdx := session.RestoreSession()
		if len(files) > 0 {
			if !resume {
				// Ask user if they want to restore
				fmt.Printf("Restore previous session (%s, %d files)? [Y/n] ", session.SessionAge(), len(files))
				var response string
				fmt.Scanln(&response)
				if response != "" && response[0] != 'y' && response[0] != 'Y' {
					files = nil
				}
			}
			if len(files) > 0 {
				session.OpenFiles(files)
				session.SwitchFile(activeIdx)
				session.RestoreCursorPositions()
				session.Flush()
				session.KeepFlushed()
				session.Listen()
				return
			}
		}
	}

	// Normal startup with specified files (or empty)
	session.OpenFiles(fileArgs)

	session.Flush()
	session.KeepFlushed()
	session.Listen()
}
