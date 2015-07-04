Sith
====

Sith is a text editor written in go.

Build
-----

	cd sith/
	go get .
	go build

Run
---

	./sith <list of files>

Commands
--------

| Keypress | command         | details
| -------- | -------         | -------
| **Files and windows**
| alt-Q    | Quit            | Close all windows and exit.  If any files are modified, check with user before closing.
| alt-W    | Close           | Close current window.  Check with user, if file is modified.
| alt-N    | Next file       | Switch to the next open file
| alt-B    | Previous file   | Switch to the previous open file
| ctrl-S   | Save
| **Multi-Cursor**
| ctrl-X   | Add Cursor      | Adds another cursor to the multi-cursor.
| alt-X    | Clear cursors   | Removes all but the main cursor from the multicursor.
| alt-C    | Column cursor   | Populates multicursor with a set of virtical cursors.
| **Editing**
| ctrl-Z   | Undo
| ctrl-Y   | Redo
| ctrl-C   | Cut line        | Cut the current line and place in copy buffer
| ctrl-V   | Paste           | Insert the copy buffer above the current line
| **Navigation**
| ctrl-A   | Start of line   | Go to the first column in the current line (all cursors)
| ctrl-E   | End of line     | go to the last column of the current line (all cursors)
| ctrl-W   | Next word       | move to the next whitespace character in the current line
| ctrl-Q   | Previous workd  | move to the previous whitespace character in the current line
| ctrl-F   | Search
| alt-F    | Search and replace



