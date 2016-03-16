package file

import "strings"
import "regexp"

type Line []rune

// Dup returns a new Line with the same content.
func (line Line) Dup() Line {
	strLine := string(line)
	return Line(strLine)
}

// ToString converts a Line into a string.
func (line Line) ToString() string {
	return string(line)
}

// CommonStart returns the sub-line that is common between
// two lines.
func (line Line) CommonStart(other Line) Line {
	for k, r := range line {
		if k >= len(other) || other[k] != r {
			return line[:k].Dup()
		}
	}
	return line.Dup()
}

// Search returns the start/end positions for a search term.
// A -1 indicates no match.
func (line Line) Search(term string, start, end int) (int, int) {

	// if line is empty, there is no match.
	if len(line) == 0 {
		return -1, -1
	}

	// Negative end indicates "from end of line".
	if end < 0 || end >= len(line) {
		end = len(line) + end
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
	target := string(line[start : end+1])

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
	return Line(re.ReplaceAllString(string(line), ""))
}

func (line Line) Tabs2spaces() Line {
	strLine := string(line)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return Line(strLine)
}
