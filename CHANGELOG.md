Change Log
==========

## To Do

 - undo/redo should keep cursor on changed line
 - Autoindent (toggle-able)
 - ensure file is writable
 - truncate prompt histories (search/replace)
 - multi-file search (?)
 - check on unicode support
 - row cut/paste
 - linewrap
 - remove trailing whitespace
 - Is there a way to prevent trailing whitespace on multicursor indent?
 - if search answer is on last few lines, scroll up the screen
 - Share search history among files/buffers
 - Restrict search / replace to marked lines


## [Unreleased]

### Bugfixes:
 - Was crashing in prompt, when navigating empty history
 - Search starting at end of line caused crash from slice index out-of-bounds.

### Features:
 - Ctrl-L/K/U: cut line, to end-of-line, to start-of-line in prompt.
 - Select file window from menu
 - Open new file (with file browser menu)
 - go fmt
 - syntax coloring
 - justify


## [0.0] "Bootstrap" 2015-09-25

Basic text editor functionality is working. Probably many bugs, and an
occasional crash.  Definitely many, many rough edges.  Starting to use
sith to write sith, hence the name "bootstrap".



