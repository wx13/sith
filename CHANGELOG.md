Change Log
==========


## [Unreleased]

### Bugfixes

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


