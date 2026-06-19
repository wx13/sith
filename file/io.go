package file

import (
	"crypto/md5"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/wx13/sith/file/buffer"
	"github.com/wx13/sith/syntaxcolor"
)

// Flush writes the buffer contents to the screen.
func (file *File) Flush() {
	file.ComputeIndent()
	cols, rows := file.screen.Size()
	slice := file.Slice(rows-1, cols)
	file.screen.Clear()

	// Compute diff - any line that's changed or adjacent to a deletion gets marked
	diffResult := file.buffer.DiffLinesFull(&file.savedBuffer)

	// Build set of lines that have changes (added, modified, or adjacent to deletion)
	changedLines := make(map[int]bool)
	for lineNum := range diffResult.Changes {
		changedLines[lineNum] = true
	}
	for _, delPoint := range diffResult.DeletionPoints {
		// Mark lines adjacent to deletions
		if delPoint >= 0 {
			changedLines[delPoint] = true // line before deletion
		}
		changedLines[delPoint+1] = true // line after deletion
	}

	// Ensure states are calculated for all lines before the visible area
	file.ensureSyntaxStates(file.rowOffset)

	for row, str := range slice {
		file.screen.WriteString(row, 0, str)
		bufferRow := row + file.rowOffset
		fullStr := file.buffer.GetRowDirect(bufferRow).Tabs2spaces(file.tabWidth).ToString()

		// Get the start state for this line
		startState := file.stateCache.GetState(bufferRow)
		result := file.SyntaxRules.ColorizeWithState(fullStr, startState)

		// Cache the end state
		file.stateCache.SetEndState(bufferRow, result.EndState)

		file.screen.Colorize(row, result.Colors, file.colOffset)

		// Draw change indicator in gutter column 0
		if changedLines[bufferRow] {
			file.screen.DrawGutterSymbol(row, '▸', tcell.ColorYellow)
		}

		// Draw vertical bar for code blocks in gutter column 1 (for markdown files)
		if file.SyntaxRules.IsMarkdown() {
			if startState.IsCodeBlock() || result.EndState.IsCodeBlock() {
				file.screen.DrawLeftBar(row, tcell.ColorBlue)
			} else if startState.IsBlockEquation() || result.EndState.IsBlockEquation() {
				file.screen.DrawLeftBar(row, tcell.ColorPurple)
			}
		}
	}
	for row := len(slice); row < rows-1; row++ {
		file.screen.WriteString(row, 0, "~")
	}
	file.ColorBracketMatch(rows)
	// file.HighlightCurrentWord()
}

// ensureSyntaxStates calculates syntax states for all lines up to (but not including) lineNum.
func (file *File) ensureSyntaxStates(lineNum int) {
	for i := 0; i < lineNum && i < file.buffer.Length(); i++ {
		// Check if we already have the state cached
		if file.stateCache.GetState(i+1) != syntaxcolor.StateNormal || i == 0 {
			// State might be cached, but we need to verify by checking if it was calculated
			// For simplicity, just recalculate if we don't have enough cached states
		}

		startState := file.stateCache.GetState(i)
		fullStr := file.buffer.GetRowDirect(i).Tabs2spaces(file.tabWidth).ToString()
		result := file.SyntaxRules.ColorizeWithState(fullStr, startState)
		stateChanged := file.stateCache.SetEndState(i, result.EndState)

		// If state didn't change, and we have more states cached, we can stop
		if !stateChanged && file.stateCache.GetState(i+1) != syntaxcolor.StateNormal {
			// The rest should still be valid
			break
		}
	}
}

// ColorBracketMatch colorizes a matching bracket character.
func (file *File) ColorBracketMatch(rows int) {
	cursor := file.MultiCursor.GetCursor(0)
	row, col := cursor.RowCol()
	row, col, err := file.buffer.BracketMatch(row, col, row+rows)
	if err != nil {
		return
	}
	col = file.buffer.GetRowDirect(row).TabCursorPos(col, file.tabWidth)
	lc := []syntaxcolor.LineColor{
		{
			Fg:    tcell.ColorRed,
			Start: col,
			End:   col + 1,
		},
	}

	file.screen.Colorize(row-file.rowOffset, lc, file.colOffset)
}

// HightlightCurrentWord highlights the word currently under the cursor.
func (file *File) HighlightCurrentWord() {
	for row, cols := range file.MultiCursor.GetRowsCols() {
		line := file.buffer.GetRowDirect(row)
		for _, col := range cols {
			start, end := line.WordBounds(col)
			if start >= end {
				continue
			}
			start = line.TabCursorPos(start, file.tabWidth)
			end = line.TabCursorPos(end+1, file.tabWidth)
			file.screen.Underline(row-file.rowOffset, start, end, file.colOffset)
		}
	}
}

// setNewline determines (estimates) the newline string that the file uses.
// It simply looks for most used newline string.
func (file *File) setNewline(bufferStr string) {

	// Default to line feed.
	file.newline = "\n"
	count := strings.Count(bufferStr, "\n")

	// Check if carriage return is more popular.
	c := strings.Count(bufferStr, "\r")
	if c > count {
		count = c
		file.newline = "\r"
	}

	// Check for CRLF or LFCR.
	for _, newline := range []string{"\n\r", "\r\n"} {
		c := strings.Count(bufferStr, newline)
		// Factor of two to prevent overcounting.
		if c > count/2 {
			count = c
			file.newline = newline
		}
	}

}

// ReadFile reads in a file (if it exists).
func (file *File) ReadFile(name string, wgs ...*sync.WaitGroup) {

	for _, wg := range wgs {
		defer wg.Done()
	}

	file.md5sum = md5.Sum([]byte(""))

	fileInfo, err := os.Stat(name)
	if err != nil {
		file.buffer.ReplaceBuffer(buffer.MakeBuffer([]string{""}))
		file.modTime = time.Now()
	} else {
		file.fileMode = fileInfo.Mode()
		file.modTime = fileInfo.ModTime()
		stringBuf := []string{""}

		byteBuf, err := os.ReadFile(name)
		if err == nil {
			file.setNewline(string(byteBuf))
			stringBuf = strings.Split(string(byteBuf), file.newline)
			file.md5sum = md5.Sum(byteBuf)
		}

		file.buffer.ReplaceBuffer(buffer.MakeBuffer(stringBuf))
	}

	file.ForceSnapshot()
	file.SnapshotSaved()
	file.savedBuffer.ReplaceBuffer(file.buffer.DeepDup())

	file.RequestFlush()

}

// RequestFlush places a flush request on the flush channel.
func (file *File) RequestFlush() {
	select {
	case file.flushChan <- struct{}{}:
	default:
	}
}

// RequestSave places a save request on the save channel.
func (file *File) RequestSave() {
	select {
	case file.saveChan <- struct{}{}:
	default:
	}
}

func (file *File) processSaveRequests() {
	for {
		<-file.saveChan
		file.Save()
	}
}

// Save saves a file.
func (file *File) Save() {
	if file.autoFmt {
		err := file.Fmt()
		if err != nil {
			file.NotifyUser(err.Error())
		}
	}
	file.SnapshotSaved()
	contents := []byte(file.ToString())
	err := os.WriteFile(file.Name, contents, file.fileMode)
	if err != nil {
		file.NotifyUser("Save Failed: " + err.Error())
	} else {
		file.savedBuffer.ReplaceBuffer(file.buffer.DeepDup())
		file.NotifyUser("Saved.")
		file.modTime = time.Now()
		file.md5sum = md5.Sum(contents)
	}
}
