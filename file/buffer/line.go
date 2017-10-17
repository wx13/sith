package buffer

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
)

type Line struct {
	chars []rune
	mutex *sync.Mutex
}

func MakeLine(str string) Line {
	return Line{
		chars: []rune(str),
		mutex: &sync.Mutex{},
	}
}

// Dup returns a new Line with the same content.
func (line Line) Dup() Line {
	newChars := []rune(string(line.chars))
	return Line{
		chars: newChars,
		mutex: &sync.Mutex{},
	}
}

// ToString converts a Line into a string.
func (line Line) ToString() string {
	return string(line.chars)
}

// CommonStart returns the sub-line that is common between
// two lines.
func (line Line) CommonStart(other Line) Line {
	for k, r := range line.chars {
		if k >= other.Length() || other.chars[k] != r {
			subLine := line.Dup()
			subLine.chars = subLine.chars[:k]
			return subLine
		}
	}
	return line.Dup()
}

// Search returns the start/end positions for a search term.
// A -1 indicates no match.
func (line Line) Search(term string, start, end int) (int, int) {

	// if line is empty, there is no match.
	if line.Length() == 0 {
		return -1, -1
	}

	// Negative end indicates "from end of line".
	if end < 0 || end >= line.Length() {
		end = line.Length() + end
		if end < 0 || end < start {
			return -1, -1
		}
	}

	return line.search(term, start, end)

}

// SearchAll returns a list of the start positions for search term matches.
func (line Line) SearchAll(term string, start, end int) []int {
	matches := []int{}
	dir := 1
	if end >= 0 && end < start {
		dir = -1
	}
	for {
		start, _ = line.Search(term, start, end)
		if start < 0 {
			break
		}
		matches = append(matches, start)
		start += dir
	}
	return matches
}

func isRegex(term string) bool {
	n := len(term)
	return (term[0:1] == "/" && term[n-1:n] == "/" && len(term) > 1)
}

func (line Line) search(term string, start, end int) (int, int) {

	forward := end >= start
	if !forward {
		start, end = end, start
	}
	if end >= line.Length() {
		end -= 1
	}

	n := len(term)
	var startCol, endCol int
	if end == line.Length() {
		end--
	}
	target := string(line.chars[start : end+1])

	if isRegex(term) {
		re, err := regexp.Compile(term[1 : n-1])
		if err != nil {
			return -1, -1
		}
		var cols []int
		if forward {
			cols = re.FindStringIndex(target)
		} else {
			colses := re.FindAllStringIndex(target, -1)
			cols = colses[len(colses)-1]
		}
		if cols == nil {
			return -1, -1
		}
		startCol = cols[0]
		endCol = cols[1]
	} else {
		strLine := strings.ToLower(target)
		term = strings.ToLower(term)
		if forward {
			startCol = strings.Index(strLine, term)
		} else {
			startCol = strings.LastIndex(strLine, term)
		}
		endCol = startCol + len(term)
	}
	// Ignore zero-length metches.
	if startCol == endCol {
		return -1, -1
	}
	if startCol < 0 {
		return startCol, endCol
	}
	return startCol + start, endCol + start
}

func (line Line) RemoveTrailingWhitespace() Line {
	re := regexp.MustCompile("[\t ]*$")
	str := re.ReplaceAllString(string(line.chars), "")
	return MakeLine(str)
}

func (line Line) RemoveLeadingWhitespace() Line {
	re := regexp.MustCompile("^[\t ]*")
	str := re.ReplaceAllString(string(line.chars), "")
	return MakeLine(str)
}

func IsWhitespace(r rune) bool {
	if r == ' ' || r == '\t' {
		return true
	}
	return false
}

func (line *Line) CompressPriorSpaces(cols []int) []int {

	line.mutex.Lock()
	defer line.mutex.Unlock()

	sort.Ints(cols)

	var c int
	for k, col := range cols {
		for c = col - 1; c > 0; c-- {
			if !IsWhitespace(line.chars[c]) {
				break
			}
			if IsWhitespace(line.chars[c-1]) {
				line.chars = append(line.chars[:c], line.chars[c+1:]...)
			}
		}
		delta := col - c - 2
		for j := k; j < len(cols); j++ {
			cols[j] -= delta
		}
	}

	return cols
}

func (line Line) Tabs2spaces(tabwidth int) Line {
	strLine := string(line.chars)
	strLine = strings.Replace(strLine, "\t", strings.Repeat(" ", tabwidth), -1)
	return MakeLine(strLine)
}

// Expand tabs to spaces and return the cursor position.
func (line Line) TabCursorPos(col int, tabwidth int) int {
	strLine := string(line.chars[:col])
	strLine = strings.Replace(strLine, "\t", strings.Repeat(" ", tabwidth), -1)
	return len(strLine)
}

func (line Line) StrSlice(startCol, endCol int, tabwidth int) string {
	pline := line.Tabs2spaces(tabwidth)
	return pline.Slice(startCol, endCol).ToString()
}

func (line Line) Slice(startCol, endCol int) Line {
	if startCol >= line.Length() {
		return MakeLine("")
	}
	if endCol >= line.Length() {
		endCol = line.Length()
	}
	if endCol < 0 {
		endCol += line.Length() + 1
	}
	newLine := line.Dup()
	newLine.chars = newLine.chars[startCol:endCol]
	return newLine
}

func (line Line) Length() int {
	return len(line.chars)
}

func (line *Line) SetChar(k int, c rune) error {
	if k > line.Length() {
		return errors.New("index exceeds line length")
	}
	line.chars[k] = c
	return nil
}

func (line *Line) GetChar(k int) rune {
	if k < 0 || k >= line.Length() {
		return 0
	}
	return line.chars[k]
}

func uniqCols(cols []int) []int {
	sort.Ints(cols)
	prev_col := -1
	newCols := []int{}
	for _, col := range cols {
		if col == prev_col {
			continue
		}
		prev_col = col
		newCols = append(newCols, col)
	}
	return newCols
}

// InsertStr inserts a string into a set of positions within a line. Return the
// new cursor positions.
func (line *Line) InsertStr(str string, cols ...int) []int {

	line.mutex.Lock()
	defer line.mutex.Unlock()

	cols = uniqCols(cols)
	runes := []rune(str)

	for i, col := range cols {

		// Enforce bounds.
		if col > len(line.chars) || col < 0 {
			continue
		}

		if col == len(line.chars) {
			// Special case: append.
			line.chars = append(line.chars, runes...)
		} else {
			// Insert in middle.
			line.chars = append(line.chars[:col], append(runes, line.chars[col:]...)...)
		}
		for j := range cols[i:] {
			cols[i+j] += len(runes)
		}

	}

	return cols
}

// DeleteFwd deletes n characters starting at the cursor and going to the right.
func (line *Line) DeleteFwd(count int, cols ...int) []int {

	line.mutex.Lock()
	defer line.mutex.Unlock()

	// Zero-out to-be-deleted elements.
	for _, col := range cols {
		for c := col; c < col+count && c < len(line.chars); c++ {
			line.chars[c] = 0
		}
	}

	// Set column positions.
	cols = []int{}
	n := 0
	newChars := []rune{}
	for _, ch := range line.chars {
		if ch != 0 {
			n++
			newChars = append(newChars, ch)
		} else {
			if len(cols) == 0 || cols[len(cols)-1] != n {
				cols = append(cols, n)
			}
		}
	}
	if len(cols) == 0 {
		cols = []int{0}
	}

	line.chars = newChars

	return cols
}

// DeleteBkwd deletes n characters starting at the cursor and going to the left.
func (line *Line) DeleteBkwd(count int, cols ...int) []int {

	line.mutex.Lock()
	defer line.mutex.Unlock()

	// Zero-out to-be-deleted elements.
	for _, col := range cols {
		for c := col - 1; c >= col-count && c >= 0 && c < len(line.chars); c-- {
			line.chars[c] = 0
		}
	}

	// Set column positions.
	cols = []int{}
	n := 0
	newChars := []rune{}
	for _, ch := range line.chars {
		if ch != 0 {
			n++
			newChars = append(newChars, ch)
		} else {
			if len(cols) == 0 || cols[len(cols)-1] != n {
				cols = append(cols, n)
			}
		}
	}
	if len(cols) == 0 {
		cols = []int{0}
	}

	line.chars = newChars

	return cols

}

func (line *Line) Chars() []rune {
	line.mutex.Lock()
	defer line.mutex.Unlock()
	return line.chars
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isPunct(r rune) bool {
	return unicode.IsPunct(r)
}

func isLetter(r rune) bool {
	return !(unicode.IsPunct(r) || unicode.IsSpace(r))
}

// PrevNextWord returns the column position of the next/previous
// word from the current column position.
func (line *Line) PrevNextWord(col, incr int) int {
	r := line.GetChar(col)
	var charCheck func(r rune) bool
	if isLetter(r) {
		charCheck = isLetter
	} else if isSpace(r) {
		charCheck = isSpace
	} else {
		charCheck = isPunct
	}
	for ; col <= line.Length() && col >= 0; col += incr {
		r = line.GetChar(col)
		if !charCheck(r) {
			return col
		}
	}
	return col
}

// BracketMatch looks for matching partner rune in a line.
//
//   start, end       pair of runes, such as '(' and ')'
//   idx              where to start the search from
//   dir              1 or -1
//   count            current level of bracketing (for continuation lines)
//
// Returns (idx, count); count == 0 means the closing bracket has been found.
func (line *Line) BracketMatch(start, end rune, idx, dir, count int) (int, int) {
	if idx < 0 {
		idx += line.Length()
	}
	for ; idx >= 0 && idx < line.Length(); idx += dir {
		r := line.GetChar(idx)
		if r == end {
			count--
		}
		if r == start {
			count++
		}
		if count == 0 {
			return idx, count
		}
	}
	return idx, count
}

// Check if a line matches a regex pattern.
func (line *Line) RegexMatch(pattern string) (bool, error) {
	return regexp.MatchString(pattern, line.ToString())
}
