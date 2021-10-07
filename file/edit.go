package file

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/wx13/sith/autocomplete"
	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
)

// Same as fmt.Sprintf, but ignores extra arguments.
func sprintf(format string, a ...interface{}) string {
	s := fmt.Sprintf(format, a...)
	r := regexp.MustCompile(`%!\(EXTRA`)
	return r.Split(s, 2)[0]
}

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
	hasLineBounds := file.fmtCmd != sprintf(file.fmtCmd, 0, 1)

	// Grab the text for formatting.
	startRow := 0
	endRow := 0
	if len(selection) > 0 {
		file.MultiCursor.OuterMost()
		startRow, endRow = file.MultiCursor.MinMaxRow()
		if (ext == "go") || (!hasLineBounds) {
			subBuffer := file.buffer.InclSlice(startRow, endRow)
			contents = subBuffer.ToString(file.newline)
		} else {
			contents = file.ToString()
		}
	} else {
		contents = file.ToString()
		startRow = 0
		endRow = file.buffer.Length() - 1
	}

	// Format the text.
	var err error
	if ext == "go" {
		contents, err = file.goFmt(contents)
	} else {
		contents, err = file.runFmt(contents, startRow, endRow)
	}
	if err != nil {
		return err
	}

	stringBuf := strings.Split(contents, file.newline)
	newBuffer := buffer.MakeBuffer(stringBuf)

	if (len(selection) > 0) && (ext == "go" || !hasLineBounds) {
		file.buffer.ReplaceLines(newBuffer.Lines(), startRow, endRow)
	} else {
		file.buffer.ReplaceBuffer(newBuffer)
	}

	file.Snapshot()
	return nil
}

// runFmt runs the fmt command on the input string. It returns the formatted text.
func (file *File) runFmt(contents string, startRow, endRow int) (string, error) {

	if file.fmtCmd == "" {
		return contents, nil
	}

	data := struct {
		Filename  string
		FirstLine int
		LastLine  int
	}{
		file.Name,
		startRow + 1,
		endRow + 1,
	}
	tmpl := template.New("fmtCmd")
	tmpl, err := tmpl.Parse(file.fmtCmd)
	if err != nil {
		return "", err
	}
	var tmplOut bytes.Buffer
	err = tmpl.Execute(&tmplOut, data)
	if err != nil {
		return "", err
	}
	cmdStr := tmplOut.String()

	args := regexp.MustCompile(`\s+`).Split(cmdStr, -1)

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
	if len(prefix) == 0 || prefix[len(prefix)-1] == ' ' ||
		prefix[len(prefix)-1] == '\t' || len(strings.Fields(prefix)) == 0 {
		return false
	}

	// Get the completion suggestion.
	results := file.AutoComplete(prefix)

	// If there are no results, just return.
	if len(results) == 0 {
		return true
	}

	// Default is the first result.
	answer := results[0]

	// If there is a prefix extension suggestion, use that.
	common := autocomplete.GetCommonPrefix(results)
	if len(common) > 0 {
		answer = common
	} else if len(results) > 1 {
		if time.Since(file.lastTab) > file.doubleTab {
			file.lastTab = time.Now()
			return true
		}
		// If there are multiple matches, let the user choose from a menu.
		menu := ui.NewMenu(file.screen, terminal.NewKeyboard())
		cmd, r := menu.ShowOnly(results)
		answer = common
		if cmd == "char" {
			answer = common + string(r)
		}
		file.lastTab = time.Now()
	}

	// Insert only the new characters.
	file.InsertStr(answer)
	return true
}

// InsertChar insters a character (rune) into the current cursor position.
func (file *File) InsertChar(ch rune) {

	rate := file.timer.Tick()
	// Don't even try autocomplete if text is being pasted.
	if rate < file.maxRate {
		// Possibly do auto completion.
		if file.complete(ch) {
			file.Snapshot()
			return
		}
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

	// For a single cursor, do autoindent.
	if len(file.MultiCursor.Cursors()) == 1 {
		cursor := file.MultiCursor.Cursors()[0]
		row, col := cursor.RowCol()
		lineStart := file.buffer.RowSlice(row, 0, col)
		lineEnd := file.buffer.RowSlice(row, col, -1)
		newLines := []buffer.Line{lineStart, lineEnd}

		file.buffer.ReplaceLines(newLines, row, row)

		file.MultiCursor.SetCursor(0, row+1, 0, 0)

		// Turn off autoindent for fast entry (probably pasting text).
		rate := file.timer.Tick()
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

	// Non-whitespace indent.
	nonWS := ""
	if row >= 2 {
		indent := file.buffer.GetRow(row - 1).CommonStart(file.buffer.GetRow(row - 2))
		nonWS = indent.RemoveLeadingWhitespace().ToString()
	}
	// A single character will only autoindent if more than two lines share it.
	if len(nonWS) == 1 {
		if row < 3 {
			nonWS = ""
		} else {
			prev := strings.TrimPrefix(file.buffer.GetRow(row-3).ToString(), ws)
			if len(prev) == 0 || prev[0] != nonWS[0] {
				nonWS = ""
			}
		}
	}

	indent := ws + nonWS
	if len(indent) == 0 {
		return
	}

	file.ForceSnapshot()

	newLineStr := indent + origLine.ToString()

	// Split the line on whitespace so we can undo parts of the indent.
	newLineParts := chunkString(newLineStr)
	newLineStr = ""
	for i, part := range newLineParts {
		newLineStr += part
		file.buffer.SetRow(row, buffer.MakeLine(newLineStr))
		col := file.MultiCursor.GetCol(idx) + len(indent)
		file.MultiCursor.SetCursor(idx, row, col, col)
		if i+1 < len(newLineParts) {
			file.ForceSnapshot()
		}
	}

}

func chunkString(s string) []string {
	if len(s) <= 1 {
		return []string{s}
	}
	chunks := []string{}
	chunk := s[:1]
	for _, r := range s[1:] {
		chunk_is_space := chunk[0] == ' ' || chunk[0] == '\t'
		rune_is_space := r == ' ' || r == '\t'
		if chunk_is_space == rune_is_space {
			chunk += string(r)
		} else {
			chunks = append(chunks, chunk)
			chunk = string(r)
		}
	}
	chunks = append(chunks, chunk)
	for i := len(chunks) - 1; i > 0; i-- {
		chunk = chunks[i]
		if chunk == " " || chunk == "\t" {
			if i == len(chunks)-1 {
				continue
			}
			chunks[i+1] = chunk + chunks[i+1]
			chunks = append(chunks[:i], chunks[i+1:]...)
		}
	}
	return chunks
}

// Justify justifies the marked text.
func (file *File) Justify() {
	file.justify(file.lineLen)
}
func (file *File) UnJustify() {
	file.justify(0)
}

func (file *File) justify(lineLen int) {
	minRow, maxRow := file.MultiCursor.MinMaxRow()
	file.buffer.Justify(minRow, maxRow, lineLen,
		[]string{"//", "#", "%", ";", "\\*"})
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

// CutToStartOfWord cuts the text from the cursor to the start of the word.
func (file *File) CutWord(mode int) {
	for idx := range file.MultiCursor.Cursors() {
		row, col := file.MultiCursor.GetRowCol(idx)
		newCol := file.buffer.CutWord(row, col, mode)
		file.MultiCursor.SetCursor(idx, row, newCol, newCol)
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
// each cursor position. Effectively, it undoes a CursorAlign.
func (file *File) CursorUnalign() {
	rows := file.MultiCursor.GetRowsCols()
	rows = file.buffer.Unalign(rows)
	file.MultiCursor.ResetCursors(rows)
	file.Snapshot()
}
