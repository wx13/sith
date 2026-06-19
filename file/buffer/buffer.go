// Package buffer provides a single editable text buffer.
// The text is stored as a slice of Lines (split on line-endings).
// A Line is a wrapper around a slice of runes.
package buffer

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Buffer is an array of Line objects.
type Buffer struct {
	lines []Line
	mutex *sync.Mutex
}

// Cursor is any object which returns a row/col position.
type Cursor interface {
	Row() int
	Col() int
}

// MakeBuffer takes in a slice os strings and creates a slice of
// Line objects.
func MakeBuffer(stringBuf []string) Buffer {
	lines := make([]Line, len(stringBuf))
	for row, str := range stringBuf {
		lines[row] = MakeLine(str)
	}
	return Buffer{
		lines: lines,
		mutex: &sync.Mutex{},
	}
}

// Lines returns the slice of lines that the buffer contains. The slice
// is a "deep copy" of the buffer's internal Line slice.
func (buffer *Buffer) Lines() []Line {
	lines := buffer.DeepDup().lines
	return lines
}

// Dup creates a new buffer with the same lines. The lines are shallow
// copies of the original lines.
func (buffer *Buffer) Dup() Buffer {
	buffer.mutex.Lock()
	linesCopy := make([]Line, len(buffer.lines))
	for row, line := range buffer.lines {
		linesCopy[row] = line
	}
	buffer.mutex.Unlock()
	return Buffer{
		lines: linesCopy,
		mutex: &sync.Mutex{},
	}
}

// DeepDup creates a new buffer with copies of the lines.
func (buffer *Buffer) DeepDup() Buffer {
	buffer.mutex.Lock()
	linesCopy := make([]Line, len(buffer.lines))
	for row, line := range buffer.lines {
		linesCopy[row] = line.Dup()
	}
	buffer.mutex.Unlock()
	return Buffer{
		lines: linesCopy,
		mutex: &sync.Mutex{},
	}
}

// Length returns the number of lines in the buffer.
func (buffer *Buffer) Length() int {
	if buffer.mutex == nil {
		return 0
	}
	buffer.mutex.Lock()
	n := len(buffer.lines)
	buffer.mutex.Unlock()
	return n
}

// ReplaceBuffer replaces the content (lines) with the content from
// another buffer. If the buffer got shorter, then just copy over the
// lines. Otherwise, check each line for equality, and only replace
// if changed.
func (buffer *Buffer) ReplaceBuffer(newBuffer Buffer) {

	newLen := newBuffer.Length()
	bufLen := buffer.Length()

	if newLen <= bufLen {
		buffer.mutex.Lock()
		buffer.lines = buffer.lines[:newLen]
		buffer.mutex.Unlock()
	}

	for k, line := range newBuffer.Lines() {
		if k >= bufLen {
			buffer.Append(line)
		} else {
			if buffer.GetRow(k).ToString() != line.ToString() {
				buffer.ReplaceLine(line, k)
			}
		}
	}

}

// Append appends a new line on to the buffer.
func (buffer *Buffer) Append(line ...Line) {
	buffer.mutex.Lock()
	buffer.lines = append(buffer.lines, line...)
	buffer.mutex.Unlock()
}

// InclSlice returns a slice of the buffer, inclusive of the endpoints.
func (buffer *Buffer) InclSlice(row1, row2 int) *Buffer {
	if row2 >= buffer.Length() {
		row2 = buffer.Length() - 1
	}
	if row2 < 0 {
		row2 += buffer.Length()
	}
	buffer.mutex.Lock()
	lines := buffer.lines[row1 : row2+1]
	buffer.mutex.Unlock()
	return &Buffer{lines: lines, mutex: &sync.Mutex{}}
}

// RowSlice returns a Line containing a subset of the line at 'row'.
func (buffer *Buffer) RowSlice(row, startCol, endCol int) Line {
	buffer.mutex.Lock()
	line := buffer.lines[row].Slice(startCol, endCol)
	buffer.mutex.Unlock()
	return line
}

// StrSlab returns a slice of strings corresponding to a "slab" of text
// which is an offset subset of the buffer. Specify the start and end rows,
// and start and end columns. Also specify the tab width, because all tabs
// are converted to spaces.
func (buffer *Buffer) StrSlab(row1, row2, col1, col2, tabwidth int) []string {
	lines := buffer.Lines()[row1:row2]
	strs := make([]string, len(lines))
	for idx, line := range lines {
		strs[idx] = line.StrSlice(col1, col2, tabwidth)
	}
	return strs
}

// ToString concatenates the buffer into one long string. Specify the newline
// character to insert between Lines.
func (buffer *Buffer) ToString(newline string) string {
	if buffer.Length() == 0 {
		return ""
	}
	lines := []string{}
	for _, line := range buffer.Lines() {
		lines = append(lines, line.ToString())
	}
	return strings.Join(lines, newline)
}

// ToCorpus concatenates the buffer into one long string. Specify the rows/cols
// of the cursor to remove the current token. Used for autocomplete.
func (buffer *Buffer) ToCorpus(cursors map[int][]int) string {
	if buffer.Length() == 0 {
		return ""
	}
	lines := []string{}
	for i, line := range buffer.Lines() {
		cols, ok := cursors[i]
		if ok {
			lines = append(lines, line.ToCorpus(cols...))
		} else {
			lines = append(lines, line.ToString())
		}
	}
	return strings.Join(lines, " ")
}

// InsertAfter inserts a set of lines after the specified row in the buffer.
func (buffer *Buffer) InsertAfter(row int, lines ...Line) {
	buffer.mutex.Lock()
	buffer.lines = append(buffer.lines[:row+1], append(lines, buffer.lines[row+1:]...)...)
	buffer.mutex.Unlock()
}

// DeleteRow deletes the specified row from the buffer.
func (buffer *Buffer) DeleteRow(row int) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	if len(buffer.lines) == 1 {
		buffer.lines = []Line{MakeLine("")}
	} else if row == 0 {
		buffer.lines = buffer.lines[1:]
	} else if row < len(buffer.lines)-1 {
		buffer.lines = append(buffer.lines[:row], buffer.lines[row+1:]...)
	} else {
		buffer.lines = buffer.lines[:row]
	}
}

// ReplaceLine replaces the line at the specified row.
func (buffer *Buffer) ReplaceLine(line Line, row int) {
	if row >= buffer.Length() {
		return
	}
	buffer.mutex.Lock()
	buffer.lines[row] = line
	defer buffer.mutex.Unlock()
}

// MergeRows merges the current row into the previous.
func (buffer *Buffer) MergeRows(row int) error {
	if row <= 0 || row >= buffer.Length() {
		return errors.New("bad MergeRows index")
	}
	str1 := buffer.GetRow(row - 1).ToString()
	str2 := buffer.GetRow(row).ToString()
	buffer.ReplaceLine(MakeLine(str1+str2), row-1)
	buffer.DeleteRow(row)
	return nil
}

// ReplaceLines replaces the lines from minRow to maxRow with lines.
func (buffer *Buffer) ReplaceLines(lines []Line, minRow, maxRow int) {
	buffer.mutex.Lock()
	buffer.lines = append(buffer.lines[:minRow], append(lines, buffer.lines[maxRow+1:]...)...)
	buffer.mutex.Unlock()
}

// Search searches for a string within the buffer. The 'loop' toggle says
// to loop around to the start of the file when searching.
func (buffer *Buffer) Search(searchTerm string, cursor Cursor, loop bool) (int, int, error) {
	var col int

	// Search the current row, from the current column to the end of the line.
	col, _ = buffer.GetRowDirect(cursor.Row()).Search(searchTerm, cursor.Col()+1, -1)
	if col >= 0 {
		return cursor.Row(), col, nil
	}

	// Search each row, from the next row to the end of the buffer.
	for row := cursor.Row() + 1; row < buffer.Length(); row++ {
		col, _ = buffer.GetRowDirect(row).Search(searchTerm, 0, -1)
		if col >= 0 {
			return row, col, nil
		}
	}
	if !loop {
		return cursor.Row(), cursor.Col(), errors.New("Not Found")
	}

	// Loop around: search from the start of the file to the original row (minus 1).
	for row := 0; row < cursor.Row(); row++ {
		col, _ = buffer.GetRowDirect(row).Search(searchTerm, 0, -1)
		if col >= 0 {
			return row, col, nil
		}
	}

	// Finally, search the original row from the start of the line to the
	// original column position.
	col, _ = buffer.GetRowDirect(cursor.Row()).Search(searchTerm, 0, col)
	if col >= 0 {
		return cursor.Row(), col, nil
	}

	return cursor.Row(), cursor.Col(), errors.New("Not Found")
}

// Replace replaces occurrences of a string within a line.
func (buffer *Buffer) ReplaceWord(searchTerm, replaceTerm string, row, col int) {
	startCol, endCol := buffer.GetRowDirect(row).Search(searchTerm, col, -1)
	strLine := buffer.GetRowDirect(row).ToString()
	newStrLine := strLine[:startCol] + replaceTerm + strLine[endCol:]
	buffer.lines[row] = MakeLine(newStrLine)
}

// GetRow returns a copy of the Line at the specified row index.
// Use this when you need to modify the line.
func (buffer *Buffer) GetRow(row int) Line {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	if row < 0 || row >= len(buffer.lines) {
		return MakeLine("")
	}
	line := buffer.lines[row]
	return MakeLine(line.ToString())
}

// GetRowDirect returns the Line at the specified row index without copying.
// This is more efficient but the caller MUST NOT modify the returned Line.
// Use this for read-only operations like display, search, length checks.
func (buffer *Buffer) GetRowDirect(row int) Line {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	if row < 0 || row >= len(buffer.lines) {
		return MakeLine("")
	}
	return buffer.lines[row]
}

// SetRow replaces the line at the specified row index.
func (buffer *Buffer) SetRow(row int, line Line) error {
	if row >= buffer.Length() {
		return errors.New("index exceeds buffer length")
	}
	buffer.mutex.Lock()
	buffer.lines[row] = line
	buffer.mutex.Unlock()
	return nil
}

// RowLength returns the length of the given row.
func (buffer *Buffer) RowLength(row int) int {
	return buffer.GetRowDirect(row).Length()
}

// GetIndent estimates the indentation string of the buffer.
// It also returns a bool indicating that the indentation is clean (true)
// or mixed (false).
func (buffer *Buffer) GetIndent() (string, bool) {
	spaceHisto := buffer.countLeadingSpacesAndTabs()
	tabCount := spaceHisto[0]
	nSpaces, spaceCount := buffer.scoreIndents(spaceHisto)
	clean := true
	if tabCount > 0 && spaceCount > 0 {
		clean = false
	}
	if tabCount >= spaceCount {
		return "\t", clean
	} else {
		return strings.Repeat(" ", nSpaces), clean
	}
}

func (buffer *Buffer) countLeadingSpacesAndTabs() []int {
	spaceHisto := make([]int, 33)
	re := regexp.MustCompile("^[ \t]*")
	for _, line := range buffer.Lines() {
		indentStr := re.FindString(line.ToString())
		nSpaces := strings.Count(indentStr, " ")
		nTabs := strings.Count(indentStr, "\t")
		if nSpaces > 0 && nSpaces <= 32 {
			spaceHisto[nSpaces]++
		}
		if nTabs > 0 {
			spaceHisto[0]++
		}
	}
	return spaceHisto
}

func (buffer *Buffer) scoreIndents(spaceHisto []int) (int, int) {
	count := 0
	nSpaces := 0
	for indentSize := 2; indentSize < 9; indentSize++ {
		score := 0
		for n := 1; n <= 4; n++ {
			score += spaceHisto[n*indentSize]
		}
		if score > count && spaceHisto[indentSize] > 0 {
			nSpaces = indentSize
			count = score
		}
	}
	return nSpaces, count
}

// Equals returns true if the two buffers are:
// - the same length, and
// - each line has the same string serialization
func (buffer *Buffer) Equals(buffer2 *Buffer) bool {
	if buffer.Length() != buffer2.Length() {
		return false
	}

	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	buffer2.mutex.Lock()
	defer buffer2.mutex.Unlock()

	for idx := range buffer.lines {
		if buffer.lines[idx].ToString() != buffer2.lines[idx].ToString() {
			return false
		}
	}
	return true
}

func (buffer *Buffer) hasLines(lines ...string) bool {
	if len(lines) > buffer.Length() {
		return false
	}
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	for i := range lines {
		if buffer.lines[i].ToString() != lines[i] {
			return false
		}
	}
	return true
}

// BracketMatch looks for matching partner rune in a set of lines.
//
//   row, col         where to start the search from
//   end_row          last row to search
func (buffer *Buffer) BracketMatch(row, col, end_row int) (int, int, error) {

	// Get the rune under the cursor.
	current_line := buffer.GetRowDirect(row)
	start := current_line.GetChar(col)

	// Get the partner rune:
	formap := map[rune]rune{'(': ')', '[': ']', '{': '}', '<': '>'}
	backmap := map[rune]rune{}
	for k, v := range formap {
		backmap[v] = k
	}
	dir := 1
	end, ok := formap[start]
	if !ok {
		end, ok = backmap[start]
		if !ok {
			return row, col, errors.New("not a bracket character")
		}
		dir = -1
	}

	// Start at one level, and move off the current char.
	count := 1
	col += dir
	if col < 0 {
		row += dir
	}
	for ; row >= 0 && row != end_row+dir && row < buffer.Length(); row += dir {
		line := buffer.GetRowDirect(row)
		col, count = line.BracketMatch(start, end, col, dir, count)
		if count == 0 {
			return row, col, nil
		}
		// Reset column.
		col = (dir - 1) / 2
	}

	return row, col, errors.New("could not find bracket match")
}

// DeleteChars deletes count characters at each position in the rows map.
func (buffer *Buffer) DeleteChars(count int, rows map[int][]int, indent ...int) map[int][]int {
	for row, cols := range rows {
		if row > buffer.Length() || row < 0 {
			continue
		}
		line := buffer.GetRow(row)
		if count >= 0 {
			cols = line.DeleteFwd(count, cols...)
		} else {

			// Unindent
			c := -count
			if len(indent) > 0 {
				col := cols[0]
				if line.Slice(0, col).ToString() == strings.Repeat(" ", col) {
					n := indent[0]
					if n*(col/n) == col {
						c = n
					}
				}
			}

			cols = line.DeleteBkwd(c, cols...)
		}
		buffer.SetRow(row, line)
		rows[row] = cols
	}
	return rows
}

// InsertNewlines splits lines at cursors.
func (buffer *Buffer) InsertNewlines(rowMap map[int][]int) map[int][]int {

	// Sort everything.
	rows := []int{}
	for row, cols := range rowMap {
		rows = append(rows, row)
		sort.Ints(cols)
		rowMap[row] = cols
	}
	sort.Ints(rows)

	// Keep track of new lines that we'll insert.
	newLines := map[int][]Line{}

	// Create a new row map, that we'll return.
	newRowMap := map[int][]int{}

	// Loop over rows *in order*.
	total := 0
	for _, row := range rows {

		line := buffer.GetRow(row)

		lines := []Line{}
		c0 := 0
		for i, c := range rowMap[row] {
			lines = append(lines, line.Slice(c0, c))
			c0 = c
			newRowMap[row+i+1+total] = []int{0}
		}
		total += len(lines)
		if c0 <= line.Length() {
			lines = append(lines, line.Slice(c0, -1))
		}
		newLines[row] = lines

	}

	// Insert into buffer.
	total = 0
	for _, row := range rows {
		lines := newLines[row]
		buffer.ReplaceLines(lines, row+total, row+total)
		total += len(lines) - 1
	}

	return newRowMap

}

// DeleteNewlines deletes the newline chars at the start of each row specified.
func (buffer *Buffer) DeleteNewlines(rowsMap map[int][]int) map[int][]int {

	// Create ordered lists of rows and columns. Columns start
	// out at 0, b/c we must be a the start of a line to delete
	// the newline.
	rows := []int{}
	cols := []int{}
	for row := range rowsMap {
		rows = append(rows, row)
		cols = append(cols, 0)
	}
	sort.Ints(rows)

	// Loop over rows and merge.
	for k := range rows {
		row := rows[k]

		// New col position is at end of previous line. Put in temp var
		// because we won't keep it if merge fails.
		col := buffer.GetRow(row - 1).Length()

		// Merge rows.
		err := buffer.MergeRows(row)
		if err != nil {
			continue
		}

		// Record our col position and adjust all subsequent rows.
		cols[k] = col
		for j := k; j < len(rows); j++ {
			rows[j]--
		}
	}

	// Reconstruct the rows map from rows/cols arrays.
	rowsMap = map[int][]int{}
	for _, row := range rows {
		rowsMap[row] = []int{}
	}
	for k, row := range rows {
		rowsMap[row] = append(rowsMap[row], cols[k])
	}
	return rowsMap
}

func (buffer *Buffer) InsertChar(ch rune, rows map[int][]int) map[int][]int {
	return buffer.InsertStr(string(ch), rows)
}

// insert the indentation string at the start of each line.
func (buffer *Buffer) IndentByStr(str string, startRow, endRow int) {
	if endRow < 0 {
		endRow = buffer.Length() - 1
	}
	for row := startRow; row <= endRow; row++ {
		line := buffer.GetRow(row)
		line.InsertStr(str, 0)
		buffer.SetRow(row, line)
	}
}

// InsertStr inserts the specified string into each position in the rows map.
func (buffer *Buffer) InsertStr(str string, rows map[int][]int) map[int][]int {
	for row, cols := range rows {
		if row > buffer.Length() || row < 0 {
			continue
		}
		line := buffer.GetRow(row)
		cols = line.InsertStr(str, cols...)
		buffer.SetRow(row, line)
		rows[row] = cols
	}
	return rows
}

// Align inserts spaces into cursor positions to align them.
func (buffer *Buffer) Align(rows map[int][]int) map[int][]int {

	// Sort all columns, and get the relative column positions.
	numCols := 0
	rowDeltas := map[int][]int{}
	for row, cols := range rows {
		sort.Ints(cols)
		rows[row] = cols
		if len(cols) > numCols {
			numCols = len(cols)
		}
		rowDeltas[row] = []int{cols[0]}
		for i := 1; i < len(cols); i++ {
			rowDeltas[row] = append(rowDeltas[row], cols[i]-cols[i-1])
		}
	}

	// Get desired column positions.
	newCols := []int{}
	for i := 0; i < numCols; i++ {
		maxDelta := 0
		for _, colDeltas := range rowDeltas {
			if len(colDeltas) <= i {
				continue
			}
			if colDeltas[i] > maxDelta {
				maxDelta = colDeltas[i]
			}
		}
		newCols = append(newCols, maxDelta)
	}
	for i := 1; i < len(newCols); i++ {
		newCols[i] += newCols[i-1]
	}

	// Construct the new rows map
	newRows := map[int][]int{}
	for row, cols := range rows {
		newRows[row] = []int{}
		for i := range cols {
			newRows[row] = append(newRows[row], newCols[i])
		}
	}

	// Alter the buffer based on the new rows map.
	for row, cols := range rows {
		lineStr := buffer.GetRow(row).ToString()
		for k := range cols {
			col := cols[k]
			newCol := newRows[row][k]
			n := newCol - col
			lineStr = lineStr[:col] + strings.Repeat(" ", n) + lineStr[col:]
			for j := k + 1; j < len(cols); j++ {
				cols[j] += n
			}
		}
		buffer.SetRow(row, MakeLine(lineStr))
	}

	return newRows

}

// Unalign removes redundant whitespace preceding each cursor..
func (buffer *Buffer) Unalign(rows map[int][]int) map[int][]int {
	for row, cols := range rows {
		line := buffer.GetRow(row)
		rows[row] = line.CompressPriorSpaces(cols)
		buffer.SetRow(row, line)
	}
	return rows
}

// CutWord removes the word under the specified cursor. Returns the new cursor position.
func (buffer *Buffer) CutWord(row, col, mode int) int {
	line := buffer.GetRow(row)
	start, end := line.WordBounds(col)
	newCol := col
	switch mode {
	case -1:
		line = line.Remove(start, col)
		newCol -= col - start
	case 1:
		line = line.Remove(col, end+1)
	case 0:
		line = line.Remove(start, end+1)
		newCol -= col - start
	}
	buffer.SetRow(row, line)
	return newCol
}

// LineStatus represents the diff status of a line.
type LineStatus int

const (
	LineUnchanged LineStatus = iota
	LineModified
	LineAdded
)

// DiffResult contains the full diff information including deletions.
type DiffResult struct {
	// Changes maps line numbers to their status (modified or added)
	Changes map[int]LineStatus
	// DeletionPoints contains line numbers *after which* deletions occurred.
	// A value of -1 means deletions at the very beginning of the file.
	DeletionPoints []int
}

// DiffLines computes the diff between this buffer and another buffer (typically the saved version).
// Returns a map from line number (in this buffer) to its status.
func (buffer *Buffer) DiffLines(saved *Buffer) map[int]LineStatus {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	saved.mutex.Lock()
	defer saved.mutex.Unlock()

	result := make(map[int]LineStatus)

	// Build string slices for comparison
	current := make([]string, len(buffer.lines))
	for i, line := range buffer.lines {
		current[i] = line.ToString()
	}

	savedLines := make([]string, len(saved.lines))
	for i, line := range saved.lines {
		savedLines[i] = line.ToString()
	}

	// Compute LCS to find unchanged lines
	lcs := computeLCS(current, savedLines)

	// Map LCS back to indices in both buffers
	// For current buffer: which lines are unchanged
	currentInLCS := make([]bool, len(current))
	lcsIdx := 0
	for i := 0; i < len(current) && lcsIdx < len(lcs); i++ {
		if current[i] == lcs[lcsIdx] {
			currentInLCS[i] = true
			lcsIdx++
		}
	}

	// For saved buffer: which lines are unchanged
	savedInLCS := make([]bool, len(savedLines))
	lcsIdx = 0
	for i := 0; i < len(savedLines) && lcsIdx < len(lcs); i++ {
		if savedLines[i] == lcs[lcsIdx] {
			savedInLCS[i] = true
			lcsIdx++
		}
	}

	// Walk through both buffers in parallel, matching LCS positions
	// Lines not in LCS in current are either added or modified
	// - Modified: replaces a line from saved that's also not in LCS
	// - Added: extra line with no corresponding deleted line
	curIdx := 0
	savedIdx := 0

	for curIdx < len(current) {
		if currentInLCS[curIdx] {
			// This line is unchanged, advance both pointers past it
			// Skip any non-LCS lines in saved (those were deleted)
			for savedIdx < len(savedLines) && !savedInLCS[savedIdx] {
				savedIdx++
			}
			savedIdx++ // skip the matching LCS line in saved
			curIdx++
			continue
		}

		// Current line is not in LCS - it's new content
		// Check if there's a corresponding deleted line in saved
		if savedIdx < len(savedLines) && !savedInLCS[savedIdx] {
			// There's a deleted line here - this is a modification
			result[curIdx] = LineModified
			savedIdx++
		} else {
			// No deleted line to pair with - this is an addition
			result[curIdx] = LineAdded
		}
		curIdx++
	}

	return result
}

// DiffLinesFull computes the diff including deletion points.
func (buffer *Buffer) DiffLinesFull(saved *Buffer) DiffResult {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	saved.mutex.Lock()
	defer saved.mutex.Unlock()

	result := DiffResult{
		Changes:        make(map[int]LineStatus),
		DeletionPoints: []int{},
	}

	// Build string slices for comparison
	current := make([]string, len(buffer.lines))
	for i, line := range buffer.lines {
		current[i] = line.ToString()
	}

	savedLines := make([]string, len(saved.lines))
	for i, line := range saved.lines {
		savedLines[i] = line.ToString()
	}

	// Compute LCS to find unchanged lines
	lcs := computeLCS(current, savedLines)

	// Map LCS entries to their positions in both buffers
	currentLCSPos := make([]int, len(lcs))
	savedLCSPos := make([]int, len(lcs))

	lcsIdx := 0
	for i := 0; i < len(current) && lcsIdx < len(lcs); i++ {
		if current[i] == lcs[lcsIdx] {
			currentLCSPos[lcsIdx] = i
			lcsIdx++
		}
	}

	lcsIdx = 0
	for i := 0; i < len(savedLines) && lcsIdx < len(lcs); i++ {
		if savedLines[i] == lcs[lcsIdx] {
			savedLCSPos[lcsIdx] = i
			lcsIdx++
		}
	}

	// Build sets of which lines are matched to LCS
	currentInLCS := make(map[int]bool)
	savedInLCS := make(map[int]bool)
	for i := 0; i < len(lcs); i++ {
		currentInLCS[currentLCSPos[i]] = true
		savedInLCS[savedLCSPos[i]] = true
	}

	// Process chunks between LCS anchors
	prevCurAnchor := -1
	prevSavedAnchor := -1

	processChunk := func(curStart, curEnd, savedStart, savedEnd int) {
		curCount := curEnd - curStart
		savedCount := savedEnd - savedStart
		minCount := curCount
		if savedCount < minCount {
			minCount = savedCount
		}

		// Mark changes in current
		for i := 0; i < curCount; i++ {
			if i < minCount {
				result.Changes[curStart+i] = LineModified
			} else {
				result.Changes[curStart+i] = LineAdded
			}
		}

		// If there are more deleted lines than added/modified, record deletion point
		if savedCount > curCount {
			// Deletions occurred after line (curStart - 1) in current buffer
			// If curStart is 0, deletions are at the very beginning (-1)
			deletionPoint := curStart - 1
			result.DeletionPoints = append(result.DeletionPoints, deletionPoint)
		}
	}

	// Walk through LCS anchors and process chunks
	for i := 0; i < len(lcs); i++ {
		curAnchor := currentLCSPos[i]
		savedAnchor := savedLCSPos[i]
		processChunk(prevCurAnchor+1, curAnchor, prevSavedAnchor+1, savedAnchor)
		prevCurAnchor = curAnchor
		prevSavedAnchor = savedAnchor
	}

	// Process final chunk
	processChunk(prevCurAnchor+1, len(current), prevSavedAnchor+1, len(savedLines))

	// Consolidate adjacent deletion points
	// If we have deletion at line N and line N+1, merge them into just line N
	if len(result.DeletionPoints) > 1 {
		consolidated := []int{result.DeletionPoints[0]}
		for i := 1; i < len(result.DeletionPoints); i++ {
			prev := consolidated[len(consolidated)-1]
			curr := result.DeletionPoints[i]
			// If they're adjacent (curr == prev + 1), they're part of the same deletion region
			// Keep only the first one (the one "above")
			if curr != prev+1 {
				consolidated = append(consolidated, curr)
			}
		}
		result.DeletionPoints = consolidated
	}

	return result
}

// computeLCS returns the longest common subsequence of two string slices.
func computeLCS(a, b []string) []string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return nil
	}

	// Build LCS length table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Backtrack to find LCS
	lcsLen := dp[m][n]
	lcs := make([]string, lcsLen)
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs[lcsLen-1] = a[i-1]
			lcsLen--
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

// GetRegionDiff returns a git-style diff for a region of the current buffer compared to saved.
// startRow and endRow define the region in the current buffer (inclusive).
// Returns lines prefixed with "-" (deleted), "+" (added), or " " (context).
func (buffer *Buffer) GetRegionDiff(saved *Buffer, startRow, endRow int) []string {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	saved.mutex.Lock()
	defer saved.mutex.Unlock()

	// Get current lines in region
	current := make([]string, 0)
	for i := startRow; i <= endRow && i < len(buffer.lines); i++ {
		current = append(current, buffer.lines[i].ToString())
	}

	savedLines := make([]string, len(saved.lines))
	for i, line := range saved.lines {
		savedLines[i] = line.ToString()
	}

	// Compute full LCS to find anchor points
	allCurrent := make([]string, len(buffer.lines))
	for i, line := range buffer.lines {
		allCurrent[i] = line.ToString()
	}
	lcs := computeLCS(allCurrent, savedLines)

	// Find LCS positions
	currentLCSPos := make([]int, len(lcs))
	savedLCSPos := make([]int, len(lcs))

	lcsIdx := 0
	for i := 0; i < len(allCurrent) && lcsIdx < len(lcs); i++ {
		if allCurrent[i] == lcs[lcsIdx] {
			currentLCSPos[lcsIdx] = i
			lcsIdx++
		}
	}

	lcsIdx = 0
	for i := 0; i < len(savedLines) && lcsIdx < len(lcs); i++ {
		if savedLines[i] == lcs[lcsIdx] {
			savedLCSPos[lcsIdx] = i
			lcsIdx++
		}
	}

	// Find the saved region that corresponds to our current region
	// Look for LCS anchors just before startRow and just after endRow
	savedStart := 0
	savedEnd := len(savedLines) - 1

	for i := 0; i < len(lcs); i++ {
		if currentLCSPos[i] < startRow {
			savedStart = savedLCSPos[i] + 1
		}
		if currentLCSPos[i] > endRow {
			savedEnd = savedLCSPos[i] - 1
			break
		}
	}

	if savedStart > savedEnd {
		savedEnd = savedStart - 1 // empty range
	}

	// Get saved lines in corresponding region
	savedRegion := make([]string, 0)
	for i := savedStart; i <= savedEnd && i < len(savedLines); i++ {
		savedRegion = append(savedRegion, savedLines[i])
	}

	// Compute diff between current region and saved region
	regionLCS := computeLCS(current, savedRegion)

	// Build diff output
	result := make([]string, 0)

	curIdx := 0
	savedIdx := 0
	lcsIdx = 0

	for curIdx < len(current) || savedIdx < len(savedRegion) {
		// Check if current line is in LCS
		curInLCS := false
		if curIdx < len(current) && lcsIdx < len(regionLCS) && current[curIdx] == regionLCS[lcsIdx] {
			curInLCS = true
		}

		// Check if saved line is in LCS
		savedInLCS := false
		if savedIdx < len(savedRegion) && lcsIdx < len(regionLCS) && savedRegion[savedIdx] == regionLCS[lcsIdx] {
			savedInLCS = true
		}

		if curInLCS && savedInLCS {
			// Both match LCS - context line
			result = append(result, " "+current[curIdx])
			curIdx++
			savedIdx++
			lcsIdx++
		} else if !savedInLCS && savedIdx < len(savedRegion) {
			// Saved line not in LCS - deleted
			result = append(result, "-"+savedRegion[savedIdx])
			savedIdx++
		} else if !curInLCS && curIdx < len(current) {
			// Current line not in LCS - added
			result = append(result, "+"+current[curIdx])
			curIdx++
		} else {
			// Edge case - advance whichever is behind
			if savedIdx < len(savedRegion) {
				result = append(result, "-"+savedRegion[savedIdx])
				savedIdx++
			}
			if curIdx < len(current) {
				result = append(result, "+"+current[curIdx])
				curIdx++
			}
		}
	}

	return result
}
