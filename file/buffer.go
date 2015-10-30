package file

import "errors"
import "strings"
import "regexp"

type Buffer []Line

func (buffer Buffer) Dup() Buffer {
	bufCopy := make(Buffer, len(buffer))
	for row, line := range buffer {
		bufCopy[row] = line
	}
	return bufCopy
}

func (buffer Buffer) DeepDup() Buffer {
	bufCopy := make(Buffer, len(buffer))
	for row, line := range buffer {
		bufCopy[row] = line.Dup()
	}
	return bufCopy
}

func MakeBuffer(stringBuf []string) Buffer {
	buffer := make(Buffer, len(stringBuf))
	for row, str := range stringBuf {
		buffer[row] = Line(str)
	}
	return buffer
}

func (buffer Buffer) ToString(newline string) string {
	str := ""
	for _, line := range buffer {
		str += string(line) + newline
	}
	return str[:len(str)-1]
}

func (buffer Buffer) Search(searchTerm string, cursor Cursor, loop bool) (int, int, error) {
	var col int
	col, _ = buffer[cursor.row].Search(searchTerm, cursor.col+1, -1)
	if col >= 0 {
		return cursor.row, col, nil
	}
	for row := cursor.row + 1; row < len(buffer); row++ {
		col, _ = buffer[row].Search(searchTerm, 0, -1)
		if col >= 0 {
			return row, col, nil
		}
	}
	if !loop {
		return cursor.row, cursor.col, errors.New("Not Found")
	}
	for row := 0; row < cursor.row; row++ {
		col, _ = buffer[row].Search(searchTerm, 0, -1)
		if col >= 0 {
			return row, col, nil
		}
	}
	col, _ = buffer[cursor.row].Search(searchTerm, 0, col)
	if col >= 0 {
		return cursor.row, col, nil
	}
	return cursor.row, cursor.col, errors.New("Not Found")
}

func (buffer *Buffer) Replace(searchTerm, replaceTerm string, row, col int) {
	startCol, endCol := (*buffer)[row].Search(searchTerm, col, -1)
	strLine := string((*buffer)[row])
	newStrLine := strLine[:startCol] + replaceTerm + strLine[endCol:]
	(*buffer)[row] = Line(newStrLine)
}

func (buffer Buffer) slice(startRow, endRow, startCol, endCol int) []string {
	slice := make([]string, endRow-startRow)
	for row := startRow; row < endRow; row++ {
		line := buffer[row].tabs2spaces()
		rowEndCol := endCol
		if rowEndCol > len(line) {
			rowEndCol = len(line)
		}
		if rowEndCol <= startCol {
			slice[row-startRow] = ""
		} else {
			slice[row-startRow] = string(line[startCol:rowEndCol])
		}
	}
	return slice
}

func (buffer Buffer) GetIndent() (string, bool) {
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

func (buffer Buffer) countLeadingSpacesAndTabs() []int {
	spaceHisto := make([]int, 33)
	re := regexp.MustCompile("^[ \t]*")
	for _, line := range buffer {
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

func (buffer Buffer) scoreIndents(spaceHisto []int) (int, int) {
	count := 0
	nSpaces := 0
	for indentSize := 1; indentSize < 9; indentSize++ {
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
