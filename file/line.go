package file

import "strings"
import "regexp"

type Line []rune

func (line Line) Dup() Line {
	strLine := string(line)
	return Line(strLine)
}

func (line Line) toString() string {
	return string(line)
}

func (line Line) CommonStart(other Line) Line {
	for k, r := range line {
		if k >= len(other) || other[k] != r {
			return line[:k].Dup()
		}
	}
	return line.Dup()
}

func (line Line) Search(term string, start, end int) (int, int) {
	if len(line) == 0 {
		return -1, -1
	}
	if end < 0 || end >= len(line) {
		end = len(line) + end
		if end < 0 || end < start {
			return -1, -1
		}
	}
	n := len(term)
	var startCol, endCol int
	target := string(line[start:end+1])
	if term[0:1] == "/" && term[n-1:n] == "/" {
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

func (line Line) tabs2spaces() Line {
	strLine := string(line)
	strLine = strings.Replace(strLine, "\t", "    ", -1)
	return Line(strLine)
}

