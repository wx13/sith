package buffer

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/wx13/sith/file/cursor"
)

type Buffer struct {
	lines []Line
	mutex *sync.Mutex
}

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

func (buffer *Buffer) Lines() []Line {
	lines := buffer.DeepDup().lines
	return lines
}

// Dup creates a new buffer with the same lines.
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

// ReplaceBuffer replaces the content (lines) with the
// content from another buffer.
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

func (buffer *Buffer) Append(line ...Line) {
	buffer.mutex.Lock()
	buffer.lines = append(buffer.lines, line...)
	buffer.mutex.Unlock()
}

// MakeSplitBuffer creates a buffer from a long string by splitting
// the string at a certain length.
func MakeSplitBuffer(bigString string, lineLen int) Buffer {
	words := strings.Fields(bigString)
	lines := []Line{}
	lineStr := words[0]
	for _, word := range words[1:] {
		if lineLen > 0 && len(lineStr)+len(word) > lineLen {
			lines = append(lines, MakeLine(lineStr))
			lineStr = word
		} else {
			lineStr += " " + word
		}
	}
	lines = append(lines, MakeLine(lineStr))
	return Buffer{
		lines: lines,
		mutex: &sync.Mutex{},
	}
}

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

func (buffer *Buffer) RowSlice(row, startCol, endCol int) Line {
	buffer.mutex.Lock()
	line := buffer.lines[row].Slice(startCol, endCol)
	buffer.mutex.Unlock()
	return line
}

func (buffer *Buffer) StrSlab(row1, row2, col1, col2 int) []string {
	lines := buffer.Lines()[row1:row2]
	strs := make([]string, len(lines))
	for idx, line := range lines {
		strs[idx] = line.StrSlice(col1, col2)
	}
	return strs
}

// ToString concatenates the buffer into one long string.
func (buffer *Buffer) ToString(newline string) string {
	if buffer.Length() == 0 {
		return ""
	}
	str := ""
	for _, line := range buffer.Lines() {
		str += line.ToString() + newline
	}
	return str[:len(str)-1]
}

func (buffer *Buffer) InsertAfter(row int, lines ...Line) {
	buffer.mutex.Lock()
	buffer.lines = append(buffer.lines[:row+1], append(lines, buffer.lines[row+1:]...)...)
	buffer.mutex.Unlock()
}

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

// Search searches for a string within the buffer.
func (buffer *Buffer) Search(searchTerm string, cursor cursor.Cursor, loop bool) (int, int, error) {
	var col int
	col, _ = buffer.GetRow(cursor.Row()).Search(searchTerm, cursor.Col()+1, -1)
	if col >= 0 {
		return cursor.Row(), col, nil
	}
	for row := cursor.Row() + 1; row < buffer.Length(); row++ {
		col, _ = buffer.GetRow(row).Search(searchTerm, 0, -1)
		if col >= 0 {
			return row, col, nil
		}
	}
	if !loop {
		return cursor.Row(), cursor.Col(), errors.New("Not Found")
	}
	for row := 0; row < cursor.Row(); row++ {
		col, _ = buffer.GetRow(row).Search(searchTerm, 0, -1)
		if col >= 0 {
			return row, col, nil
		}
	}
	col, _ = buffer.GetRow(cursor.Row()).Search(searchTerm, 0, col)
	if col >= 0 {
		return cursor.Row(), col, nil
	}
	return cursor.Row(), cursor.Col(), errors.New("Not Found")
}

// Replace replaces occurrences of a string within a line.
func (buffer *Buffer) ReplaceWord(searchTerm, replaceTerm string, row, col int) {
	startCol, endCol := buffer.GetRow(row).Search(searchTerm, col, -1)
	strLine := buffer.GetRow(row).ToString()
	newStrLine := strLine[:startCol] + replaceTerm + strLine[endCol:]
	buffer.lines[row] = MakeLine(newStrLine)
}

func (buffer *Buffer) GetRow(row int) Line {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	if row < 0 || row > len(buffer.lines) {
		return MakeLine("")
	}
	line := buffer.lines[row]
	return MakeLine(line.ToString())
}

func (buffer *Buffer) SetRow(row int, line Line) error {
	if row >= buffer.Length() {
		return errors.New("index exceeds buffer length")
	}
	buffer.mutex.Lock()
	buffer.lines[row] = line
	buffer.mutex.Unlock()
	return nil
}

func (buffer *Buffer) RowLength(row int) int {
	return buffer.GetRow(row).Length()
}

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

func (buffer *Buffer) CompressPriorSpaces(row, col int) int {
	line := buffer.GetRow(row)
	line, col = line.CompressPriorSpaces(col)
	buffer.SetRow(row, line)
	return col
}

// BracketMatch looks for matching partner rune in a set of lines.
//
//   row, col         where to start the search from
//   end_row          last row to search
func (buffer *Buffer) BracketMatch(row, col, end_row int) (int, int, error) {

	// Get the rune under the cursor.
	current_line := buffer.GetRow(row)
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
		line := buffer.GetRow(row)
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
func (buffer *Buffer) DeleteChars(count int, rows map[int][]int) map[int][]int {
	for row, cols := range rows {
		if row > buffer.Length() || row < 0 {
			continue
		}
		line := buffer.GetRow(row)
		if count >= 0 {
			cols = line.DeleteFwd(count, cols...)
		} else {
			cols = line.DeleteBkwd(-count, cols...)
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
	for row, _ := range rowsMap {
		rows = append(rows, row)
		cols = append(cols, 0)
	}
	sort.Ints(rows)

	// Loop over rows and merge.
	for k, _ := range rows {
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
