Sith
====

Sith is a text editor written in go.  MIT license.

![screenshot](http://www.wx13.com/sithscreenshot.png)

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
| alt-K    | Toggle file     | Toggle between most recent files
| alt-M    | File selct menu | Bring up a menu to select file/window.
| alt-O    | File open menu  | Bring up a file browser menu to select a new file
| ctrl-S   | Save
| Alt-S    | Suspend
| **Multi-Cursor**
| ctrl-X   | Add Cursor      | Adds another cursor to the multi-cursor.
| alt-X    | Clear cursors   | Removes all but the main cursor from the multicursor.
| alt-C    | Column cursor   | Populates multicursor with a set of virtical cursors.
| **Editing**
| ctrl-U   | Undo
| ctrl-Y   | Redo
| Alt-U    | Macro undo      | Like undo, but go the the next "saved" state.
| Alt-Y    | Macro redo      | Like redo, but go the the next "saved" state.
| ctrl-C   | Cut line        | Cut the current line and place in copy buffer
| ctrl-C   | Cut to end      | Cut the current line(s) from cursor to end on line.
| ctrl-C   | Cut to start    | Cut the current line(s) from cursor to start on line.
| ctrl-V   | Paste           | Insert the copy buffer above the current line
| alt-V    | Menu paste      | Menu select from copy/paste history
| alt-J    | Justify         | Justify (72 cols) all lines from lowest to highest cursor
| alt-H    | Unjustify       | Unjustify all lines from lowest to highest cursor into one long line
| alt-G    | go fmt          | Run go formatter on code.
| ctrl-D   | delete
| **Navigation**
| ctrl-A   | Start of line   | Go to the first column in the current line (all cursors)
| ctrl-E   | End of line     | go to the last column of the current line (all cursors)
| ctrl-W   | Next word       | move to the next whitespace character in the current line
| ctrl-Q   | Previous workd  | move to the previous whitespace character in the current line
| ctrl-J   | Scroll Up       | Scroll the entire screen up by one line
| ctrl-K   | Scroll Down     | Scroll the entire screen down by one line
| ctrl-O   | Scroll Left     | Scroll the entire screen left by one column
| ctrl-P   | Scroll Right    | Scroll the entire screen right by one column
| ctrl-N/B | Page down/up    | PageDn / PageUp also supported
| ctrl-G   | Goto line number| Editor will prompt for line number (with per-file history)
| **Toggles**
| alt-I    | Toggle auto-indent
| alt-T    | Toogle auto-tab | Toggle auto-detection of indentation string.
| **Search**
| ctrl-F   | Search
| alt-F    | Search and replace
| ctrl-R   | Multi-file search
| alt-R    | Multi-file search and replace
| **Misc**
| alt-L    | Refresh screen  | Redraw all pixels on screen.
