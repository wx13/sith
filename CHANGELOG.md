Change Log
==========

## [0.7.0] 2018-01-02

### Features
- Autocomplete -- text-based and triggered by tab
  - double-tab brings up a menu
  - Used in search prompt as well
- Uses version package to automatically stamp version numbers.
- Much improved text justification
  - handles comments and indentation
- Bookmarks -- mark a line of a file and return to it later.


## [0.6.2] 2017-09-20

### Features
- More flexibilty in fmtcmd
- Can now use code formatter on a select setof lines
- Syntax coloring for git-rebase-todo

### Bugfixes
- Fix two multicursor index-out-of-range errors
- Corrected space char insertion issue in prompt string
- Speed up multicursor highlighting for large multicursors


## [0.6.1] 2017-07-14

### Features
- Allow configuration specification of syntax rule which always applies.
  - Add "clober=true" and rule will always apply (even inside string or comment).
- MultiCursor navigation modes
  - AllTogether: all cursors move in unison
  - Column (default): cursor move together horizontally, but not vertically
  - Detached: all but main cursor are fixed
  - Alt-Z to cycle among the modes

## [0.6.0] 2017-07-14

### Features
- TOML configuration file
  - any of: ~/.sith.toml, ~/.sith/config.toml, or ~/.config/sith.toml
  - specify syntax coloring, inhertance, code formatter, etc.
- More multicursor stuff:
  - multiple cursors on same line
  - insert and delete newlines
  - align & unalign multiple cursors in a line (or multiple lines)



## [0.5.1] 2017-06-19

### Bugfixes

- check array bounds on GetRow()


## [0.5.0] 2017-06-01

### Features
- Better file status awareness
  - Status-line display if other open buffers are modified
  - Modification status in buffer selection menu
  - Checks if any files have changed (on disk) while open
  - Can now reload a file from disk
- Save-all and Save-as
- Can create a new file from the open-file menu
- Status bar formatting tweaks
- Make overlapping cursors more visible
- Highlight matching bracket character
- Allow toggling of auto-format-on-save


## [0.4.4] 2017-05-21

### Bugfixes
- Was getting caught in loop when multifile search yielded no results
- Ignore zero-length search matches (regex)
- Prevent index out-of-bounds in search

### Features
- Auto-tab improvements:
  - Allow manual override of tab detection
  - Never allow single-space indent string
- Report gofmt error to user
- Menu improvements:
  - Make menu search string easier to see
  - Make current buffer the default selection
- Improved syntax-coloring:
  - Better interaction between rules
  - Backtic quotes for Go
  - More C++ extensions


## [0.4.3] 2016-09-27

### Features
- In paste-from-menu, show concatenated lines, rather than just the first line.
- Improved next/prev word navigation

### Bugfixes
- prevent condition where array length changes after measuring
- fixed some golint issues


## [0.4.2] 2016-06-12

### Features
- Choose command from menu
- Align cursors: In multicursor mode, insert spaces before each cursor in order to
  bring the cursors into alignment.
- Multicursor forward/backward search within a line
  - useful, e.g., for aligning cursors a say an '='

### Bugfixes
- Marked search (and search/replace) work again.  Was a problem with the
  'outermost' method not working on two cursors.


## [0.4.1] 2016-05-15

### Bugfixes
- Fixed unicode handling
  - Try to guess character width in terminal
  - Allow user different character display modes: ascii, full unicode,
    narrow unicode chars only, or low-order unicode characters
    - This is a work-around for terminals which fall back to variable-width
      fonts for glyphs
  - Fix clipped unicode when shifted off edge of screen

### Features
- more stability in undo/redo cursor movement
  - keep cursor on changed spot


## [0.4] 2016-05-02

Big refactor.  Pushed buffer and cursor stuff into their own
sub-packages.  File no longer accesses buffer/cursor internals.

Removed all known race conditions.


## [0.3.4] 2016-04-07

### Bugfixes
- In replaceBuffer, truncate end of old buffer to length of new
  - This was causing problems when "gofmt" was shortening the
    file.
- Dup buffer on forced snapshot.
  - This was causing issues with undoing auto indent stuff.

### Features
- "macro" undo/redo
  - Each save marks a state, which can be reverted to.


## [0.3.3] 2016-03-30

### Bugfixes
- Search/replace must advance to end of word when replacing;
  otherwise, we may match the same word twice.
- must force flush for screen refresh, or else no changes get made
- Truncate prompt histories (search/replace), so they
  don't get too long

### Features
- prev/next word uses any non-alphanum char as a break point
- disable snapshotting on paste
- disable autoindent on paste
- remove trailing whitespace on newline
- run gofmt when saving a go file



## [0.3.2] 2015-11-30

### Bugfixes
- Justify was not handling long last lines.  Now it is.
- Justify was not handling lines without spaces.  Now it does.
- Justify was not working with short lines.  Now it merges short
  lines together.
- Make screen.Flush() async.  Just like editor, place empty struct
  on a flush chan.  This is to prevent race condition during user
  prompint.

### Features
- Search and S&R can now be restricted to a set of lines.
- Unjustify turns a set of lines into a single long line.




## [0.3.1] 2015-11-11

### Bugfixes
- Correctly account for tabs when positioning cursor
- Make nav column stickiness work better
- Prevent syntax coloring from wrapping on long lines
- Truncate long filenames in status bar
- Async notifications work now
- Menu width set by contents
- Catch bad values returned from menu in paste-from-menu
- Justify
  + remove trailing whitespace
  + maintain blank line after justified text
  + turn off multicursor when done


## [0.3] "The Good and the Bad" 2015-11-03

Some good new features (autodetection of indentation, multi-file
search/replace, toggle between files, cut/paste history), but one
uncovered problem (how to do async notifications).

### Bugfixes
- Search doesn't match at end of line.
- Use AttrReverse for multicursor, so that it works with new
  terminal default colors

### Features
- Use terminal default colors
- Autodetect indentation, supported by 'tab' and backspace.
- Toggle between two files
- Multi-file search and search/replace
  - including "replace all"
- Cut/Paste history
- Detect line endings (\n, \r, etc.)
  - display non-\n ending in status bar
- Async save and gofmt, with async notifications



## [0.2] "Editorish" 2015-10-21

### Bugfixes
- Page up/down now does half a screen
- Next/prev word now functions more like you would expect it to
- Prompt now clears two spaces beyond writing
- Don't report new file as modified
- Desired column changes for end-of-line, next word, etc.
- only do gofmt on go files

### Features
- Added syntax colors for html, javascript, coffeescript
- go-to line number
- if search answer is past last line, scroll up more than the offset
- "smart" start of line


## [0.1.2] 2015-10-07

### Bugfixes:
- Conditional build for windows, because suspend doesn't work.


## [0.1.1] 2015-10-06

### Bugfixes:
- Don't allow multicursor delete of newlines, otherwise weird stuff happens when
  trying to delete indents.
- Don't let multicursor insert characters in the first column if other rows are
  at col > 0. This way we can indent multiple blocks of code separated by
  blank lines.
- Allow alternate alt-keys.  Some terminals don't make alt keys equal to esc.
  If the character code is in the right range, use c-128 as the alt key code.
- Fix opening of subdirectory files in menu selector.  I wasn't prepending
  the directory path to the file name.
- Toggling off auto indent now toggles off autoindent...  I had done everything
  except the actually turning off of autoindent...
- Multicursor in first column does not add whitespace to blank lines.  Now
  it will add whitespace to blank line only if all lines are blank.
- When selecting a file from the file selection menu, open the file by relative path,
  not absolute.
- Screen refresh will work now.  I now write garbage to the screen, then
  write spaces to the screen.  This should clear it completely...
- Make syntax coloring work on full line, not just the part of the line in the
  current view.


## [0.1] "Almost Usable" 2015-09-30

I am have been using sith to edit sith, most of the time. I'm starting
to consider using it as my main editor. Almost. This release fixes a few
bugs and introduces a bunch of new features.

### Bugfixes:
- Was crashing in prompt, when navigating empty history
- Search starting at end of line caused crash from slice index out-of-bounds.
- Don't add an empty line to the bottom of the file

### Features:
- Ctrl-L/K/U: cut line, to end-of-line, to start-of-line in prompt.
- Select file window from menu
- Open new file (with file browser menu)
- go fmt
- syntax coloring
- justify
- Autoindent (toggle-able)
- undo/redo should keep cursor on changed line



## [0.0] "Bootstrap" 2015-09-25

Basic text editor functionality is working. Probably many bugs, and an
occasional crash.  Definitely many, many rough edges.  Starting to use
sith to write sith, hence the name "bootstrap".


