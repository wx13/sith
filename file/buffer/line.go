package buffer

import (
	"errors"
	"regexp"
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

func IsWhitespace(r rune) bool {
	if r == ' ' || r == '\t' {
		return true
	}
	return false
}

func (line Line) CompressPriorSpaces(col int) (Line, int) {
	line = line.Dup()
	for ; col > 1; col-- {
		if !IsWhitespace(line.chars[col-1]) {
			break
		}
		if IsWhitespace(line.chars[col-2]) {
			line.chars = append(line.chars[:col-1], line.chars[col:]...)
		}
	}
	return line, col + 1
}

func (line Line) Tabs2spaces() Line {
	strLine := string(line.chars)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return MakeLine(strLine)
}

// Expand tabs to spaces and return the cursor position.
func (line Line) TabCursorPos(col int) int {
	strLine := string(line.chars[:col])
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return len(strLine)
}

func (line Line) StrSlice(startCol, endCol int) string {
	pline := line.Tabs2spaces()
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

func (line *Line) Chars() []rune {
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
