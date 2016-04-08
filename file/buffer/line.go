package buffer

import (
	"errors"
	"regexp"
	"strings"
	"sync"
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

func isRegex(term string) bool {
	n := len(term)
	return (term[0:1] == "/" && term[n-1:n] == "/")
}

func (line Line) search(term string, start, end int) (int, int) {

	n := len(term)
	var startCol, endCol int
	target := string(line.chars[start : end+1])

	if isRegex(term) {
		re, err := regexp.Compile(term[1 : n-1])
		if err != nil {
			return -1, -1
		}
		cols := re.FindStringIndex(target)
		if cols == nil {
			return -1, -1
		}
		startCol = cols[0]
		endCol = cols[1]
	} else {
		strLine := strings.ToLower(target)
		term = strings.ToLower(term)
		startCol = strings.Index(strLine, term)
		endCol = startCol + len(term)
	}
	if startCol < 0 {
		return startCol, endCol
	} else {
		return startCol + start, endCol + start
	}
}

func (line Line) RemoveTrailingWhitespace() Line {
	re := regexp.MustCompile("[\t ]*$")
	str := re.ReplaceAllString(string(line.chars), "")
	return MakeLine(str)
}

func (line Line) Tabs2spaces() Line {
	strLine := string(line.chars)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return MakeLine(strLine)
}

func (line Line) StrSlice(startCol, endCol int) string {
	strLine := line.Tabs2spaces().ToString()
	if startCol >= len(strLine) {
		return ""
	}
	if endCol >= len(strLine) {
		endCol = len(strLine)
	}
	if endCol < 0 {
		endCol += len(strLine) + 1
	}
	return strLine[startCol:endCol]
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
