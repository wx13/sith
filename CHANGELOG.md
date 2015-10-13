Change Log
==========

## To Do

features:
 - ensure file is writable
 - truncate prompt histories (search/replace)
 - multi-file search or at least shared search history
 - check on unicode support
 - row cut/paste
 - linewrap
 - remove trailing whitespace
 - Is there a way to prevent trailing whitespace on multicursor indent?
 - if search answer is on last few lines, scroll up the screen
 - Share search history among files/buffers
 - Restrict search / replace to marked lines
 - Revert to last saved copy
   - also should be able to reload from file

bugs:
 - suspend still sometimes causes problems with keyboard (hard to reproduce)
 - menus should expand to include long lines.  Horizontal scroll?
 - Sometimes thinks the file has been modified when it hasn't (undo/redo)


## [unreleased]

### Bugfixes

### Features
 - Added syntax colors for html, javascript, coffeescript
 - go-to line number


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


