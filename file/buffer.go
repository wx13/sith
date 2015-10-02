package file

import "errors"

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

func (buffer Buffer) ToString() string {
	str := ""
	for _, line := range buffer {
		str += string(line) + "\n"
	}
	return str[:len(str)-1]
}

func (buffer Buffer) Search(searchTerm string, cursor Cursor) (int, int, error) {
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

