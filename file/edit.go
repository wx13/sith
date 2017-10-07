package file

import (
	"go/format"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/wx13/sith/autocomplete"
	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
)

// Fmt runs a code formatter on the text buffer and updates the buffer.
// For Go code, this calls the go format library. For all else, it runs an
// external command. If 'selection' is specified, then formatting is done
// only on selected lines.
func (file *File) Fmt(selection ...bool) error {

	ext := GetFileExt(file.Name)
	if file.fmtCmd == "" && ext != "go" {
		return nil
	}

	contents := ""

	// Grab the text for formatting.
	startRow := 0
	endRow := 0
	if len(selection) > 0 {
		file.MultiCursor.OuterMost()
		startRow, endRow = file.MultiCursor.MinMaxRow()
		subBuffer := file.buffer.InclSlice(startRow, endRow)
		contents = subBuffer.ToString(file.newline)
	} else {
		contents = file.ToString()
	}

	// Format the text.
	var err error
	if ext == "go" {
		contents, err = file.goFmt(contents)
	} else {
		contents, err = file.runFmt(contents)
	}
	if err != nil {
		return err
	}

	stringBuf := strings.Split(contents, file.newline)
	newBuffer := buffer.MakeBuffer(stringBuf)

	if len(selection) > 0 {
		file.buffer.ReplaceLines(newBuffer.Lines(), startRow, endRow)
	} else {
		file.buffer.ReplaceBuffer(newBuffer)
	}

	file.Snapshot()
	return nil
}

// runFmt runs the fmt command on the input string. It returns the formatted text.
func (file *File) runFmt(contents string) (string, error) {

	if file.fmtCmd == "" {
		return contents, nil
	}

	args := regexp.MustCompile(`\s+`).Split(file.fmtCmd, -1)

	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return contents, err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, contents)
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return contents, err
	}
	return string(out), nil

}

// goFmt runs the internal gofmt library on the string.
func (file *File) goFmt(contents string) (string, error) {
	bytes, err := format.Source([]byte(contents))
	return string(bytes), err
}

// getMaxCol returns the right-most column index of all the rows.
func getMaxCol(rows map[int][]int) int {
	maxCol := 0
	for _, cols := range rows {
		for _, col := range cols {
			if col > maxCol {
				maxCol = col
			}
		}
	}
	return maxCol
}

// allBlankLines returns true if all the specified rows are empty.
func (file File) allBlankLines(rows map[int][]int) bool {
	for row := range rows {
		if file.buffer.GetRow(row).Length() > 0 {
			return false
		}
	}
	return true
}

// removeBlankLineCursors removes the cursor index of any rows which are empty.
func (file File) removeBlankLineCursors(rows map[int][]int) (map[int][]int, []int) {

	if file.allBlankLines(rows) {
		return rows, []int{}
	}

	blankRows := []int{}
	for row := range rows {
		if file.buffer.GetRow(row).Length() == 0 {
			blankRows = append(blankRows, row)
			delete(rows, row)
			continue
		}
	}

	return rows, blankRows
}

// Complete possibly runs autocompletion, depending on situation.
func (file *File) complete(ch rune) bool {

	// Only run autocompletion if user pressed tab.
	if ch != '\t' {
		return false
	}

	// Only run autocompletion if there is a word to complete (before the cursor).
	row, col := file.MultiCursor.GetRowCol(0)
	prefix := file.buffer.RowSlice(row, 0, col).ToString()
	if len(prefix) == 0 || prefix[len(prefix)-1] == ' ' || prefix[len(prefix)-1] == '\t' {
		return false
	}

	// Get the completion suggestion.
	words := autocomplete.Split(prefix)
	prefix = words[len(words)-1]
	more_prefix, results := file.completer.Complete(prefix)

	// If there are no results, just return.
	if len(results) == 0 {
		return true
	}

	// Default is the first result.
	answer := results[0]

	// If there is a prefix extension suggestion, use that.
	if len(more_prefix) > len(prefix) {
		answer = more_prefix
	} else if len(results) > 1 {
		// If there are multiple matches, let the user choose from a menu.
		menu := ui.NewMenu(file.screen, terminal.NewKeyboard())
		idx, str := menu.Choose(results, 0, prefix, "tab")
		file.Flush()
		if idx < 0 || str == "cancel" || str == "tab" {
			return true
		}
		answer = results[idx]
	}

	// Insert only the new characters.
	diff := answer[len(prefix):]
	file.InsertStr(diff)
	return true
}

// InsertChar insters a character (rune) into the current cursor position.
func (file *File) InsertChar(ch rune) {

	// Possibly do auto completion.
	if file.complete(ch) {
		file.Snapshot()
		return
	}

	str := string(ch)
	if ch == '\t' && file.autoTab && file.tabString != "\t" {
		str = file.tabString
	}

	file.InsertStr(str)

	file.Snapshot()
}

func (file *File) InsertStr(str string) {
	rows := file.MultiCursor.GetRowsCols()
	var blankRows []int
	rows, blankRows = file.removeBlankLineCursors(rows)
	rows = file.buffer.InsertStr(str, rows)
	for _, row := range blankRows {
		rows[row] = []int{0}
	}
	file.MultiCursor.ResetCursors(rows)
}

func allColsZero(rows map[int][]int) bool {
	for _, cols := range rows {
		for _, col := range cols {
			if col > 0 {
				return false
			}
		}
	}
	return true
}

// Backspace removes the character before the cursor.
func (file *File) Backspace() {

	indent := 0
	if file.autoTab {
		indent = len(file.tabString)
	}

	rows := file.MultiCursor.GetRowsCols()
	if allColsZero(rows) {
		rows = file.buffer.DeleteNewlines(rows)
	} else {
		rows = file.buffer.DeleteChars(-1, rows, indent)
	}
	file.MultiCursor.ResetCursors(rows)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
}

// Delete deletes the character under the cursor.
func (file *File) Delete() {
	file.CursorRight()
	file.Backspace()
}

// Newline breaks the current line into two.
func (file *File) Newline() {

	if len(file.MultiCursor.Cursors()) == 1 {
		rate := file.timer.Tick()
		cursor := file.MultiCursor.Cursors()[0]
		row, col := cursor.RowCol()
		lineStart := file.buffer.RowSlice(row, 0, col)
		lineEnd := file.buffer.RowSlice(row, col, -1)
		newLines := []buffer.Line{lineStart, lineEnd}

		file.buffer.ReplaceLines(newLines, row, row)

		file.MultiCursor.SetCursor(0, row+1, 0, 0)

		if file.autoIndent && rate < file.maxRate && lineEnd.Length() == 0 {
			file.doAutoIndent(0)
		}

		file.buffer.SetRow(row, lineStart.RemoveTrailingWhitespace())

	} else {
		rows := file.MultiCursor.GetRowsCols()
		rows = file.buffer.InsertNewlines(rows)
		file.MultiCursor.ResetCursors(rows)
	}

	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
}

func (file *File) doAutoIndent(idx int) {

	row := file.MultiCursor.GetRow(idx)
	if row == 0 {
		return
	}

	origLine := file.buffer.GetRow(row).Dup()

	// Whitespace-only indent.
	re, _ := regexp.Compile("^[ \t]+")
	prevLineStr := file.buffer.GetRow(row - 1).ToString()
	ws := re.FindString(prevLineStr)
	if len(ws) > 0 {
		newLineStr := ws + file.buffer.GetRow(row).ToString()
		file.buffer.SetRow(row, buffer.MakeLine(newLineStr))
		col := file.MultiCursor.GetCol(idx) + len(ws)
		file.MultiCursor.SetCursor(idx, row, col, col)
		if file.buffer.GetRow(row-1).Length() == len(ws) {
			file.buffer.SetRow(row-1, buffer.MakeLine(""))
		}
	}

	if row < 2 {
		return
	}

	// Non-whitespace indent.
	indent := file.buffer.GetRow(row - 1).CommonStart(file.buffer.GetRow(row - 2))
	if indent.Length() > len(ws)+3 {
		file.ForceSnapshot()
		newLineStr := indent.ToString() + origLine.ToString()
		file.buffer.SetRow(row, buffer.MakeLine(newLineStr))
		col := file.MultiCursor.GetCol(idx) + indent.Length() - len(ws)
		file.MultiCursor.SetCursor(idx, row, col, col)
	}

}

// Justify justifies the marked text.
func (file *File) Justify(lineLen int) {
	minRow, maxRow := file.MultiCursor.MinMaxRow()
	bigString := file.buffer.InclSlice(minRow, maxRow).ToString(" ")
	splitBuf := buffer.MakeSplitBuffer(bigString, lineLen)
	file.buffer.ReplaceLines(splitBuf.Lines(), minRow, maxRow)
	file.MultiCursor.Clear()
	file.Snapshot()
}

// Cut cuts the current line and adds to the copy buffer.
func (file *File) Cut() []string {
	row := file.MultiCursor.GetRow(0)
	cutLines := file.buffer.InclSlice(row, row).Dup()
	strs := make([]string, cutLines.Length())
	for idx, line := range cutLines.Lines() {
		strs[idx] = line.ToString()
	}
	file.buffer.DeleteRow(row)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
	return strs
}

// Paste inserts the copy buffer into buffer at the current line.
func (file *File) Paste(strs []string) {
	row := file.MultiCursor.GetRow(0)
	pasteLines := make([]buffer.Line, len(strs))
	for idx, str := range strs {
		pasteLines[idx] = buffer.MakeLine(str)
	}
	file.buffer.InsertAfter(row-1, pasteLines...)
	file.CursorDown(len(pasteLines))
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()
}

// CutToStartOfLine cuts the text from the cursor to the start of the line.
func (file *File) CutToStartOfLine() {
	for idx := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		line := file.buffer.GetRow(row).Slice(col, -1)
		file.buffer.SetRow(row, line)
		file.MultiCursor.SetCursor(idx, row, 0, 0)
	}
	file.Snapshot()
}

// CutToEndOfLine cuts the text from the cursor to the end of the line.
func (file *File) CutToEndOfLine() {
	for idx := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		line := file.buffer.GetRow(row).Slice(0, col)
		file.buffer.SetRow(row, line)
	}
	file.Snapshot()
}

// CursorAlign inserts spaces into each cursor position, in order to
// align the cursors vertically.
func (file *File) CursorAlign() {
	rows := file.MultiCursor.GetRowsCols()
	rows = file.buffer.Align(rows)
	file.MultiCursor.ResetCursors(rows)
	file.Snapshot()
}

// CursorUnalign removes whitespace (except for 1 space) immediately preceding
// each cursor position.  Effectively, it undoes a CursorAlign.
func (file *File) CursorUnalign() {
	rows := file.MultiCursor.GetRowsCols()
	rows = file.buffer.Unalign(rows)
	file.MultiCursor.ResetCursors(rows)
	file.Snapshot()
}
