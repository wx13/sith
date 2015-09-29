package main

import "strings"
import "errors"
import "regexp"

type Line []rune

func (line Line) Dup() Line {
	strLine := string(line)
	return Line(strLine)
}

func (line Line) toString() string {
	return string(line)
}

func (line Line) Search(term string, start, end int) (int, int) {
	if end < 0 || end >= len(line) {
		end = len(line) + end
		if end < 0 || end < start {
			return -1, -1
		}
	}
	n := len(term)
	var startCol, endCol int
	target := string(line[start:end])
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
	return str
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

type Cursor struct {
	row, col, colwant int
}

func (cursor Cursor) Dup() Cursor {
	return Cursor{
		row:     cursor.row,
		col:     cursor.col,
		colwant: cursor.colwant,
	}
}

type MultiCursor []Cursor

func MakeMultiCursor() MultiCursor {
	return MultiCursor{Cursor{}}
}

func (mc MultiCursor) Dup() MultiCursor {
	mcCopy := make(MultiCursor, len(mc))
	for k, cursor := range mc {
		mcCopy[k] = cursor.Dup()
	}
	return mcCopy
}

func (mc MultiCursor) Clear() MultiCursor {
	return mc[0:1]
}

func (mc MultiCursor) Add() MultiCursor {
	cursor := Cursor{
		row:     mc[0].row,
		col:     mc[0].col,
		colwant: mc[0].colwant,
	}
	mc = append(mc, cursor)
	return mc
}

func (mc MultiCursor) MinMaxRow() (minRow, maxRow int) {
	minRow = mc[0].row
	maxRow = mc[0].row
	for _, cursor := range mc {
		if cursor.row < minRow {
			minRow = cursor.row
		}
		if cursor.row > maxRow {
			maxRow = cursor.row
		}
	}
	return
}

func (mc MultiCursor) SetColumn() MultiCursor {
	col := mc[0].col
	minRow, maxRow := mc.MinMaxRow()
	mc = MultiCursor{}
	for row := minRow; row <= maxRow; row++ {
		cursor := Cursor{row: row, col: col, colwant: col}
		mc = append(mc, cursor)
	}
	return mc
}

type BufferHist struct {
	buffers []Buffer
	cursors []MultiCursor
	idx     int
}

func NewBufferHist(buffer Buffer, cursor MultiCursor) *BufferHist {
	bh := BufferHist{}
	bh.buffers = append(bh.buffers, buffer)
	bh.cursors = append(bh.cursors, cursor.Dup())
	return &bh
}

func (bh *BufferHist) Snapshot(buffer Buffer, mc MultiCursor) {

	bh.idx = bh.idx + 1

	buffers := append(bh.buffers[:bh.idx], buffer.Dup())
	bh.buffers = append(buffers, bh.buffers[bh.idx:]...)

	cursors := append(bh.cursors[:bh.idx], mc.Dup())
	bh.cursors = append(cursors, bh.cursors[bh.idx:]...)

	bh.Trim(100)

}

func (bh *BufferHist) Trim(n int) {
	if bh.idx+n < len(bh.buffers) {
		bh.buffers = bh.buffers[:(bh.idx + n)]
		bh.cursors = bh.cursors[:(bh.idx + n)]
	}
	if bh.idx >= n {
		bh.buffers = bh.buffers[(bh.idx - n):]
		bh.cursors = bh.cursors[(bh.idx - n):]
		bh.idx -= bh.idx - n
	}
}

func (bh *BufferHist) Current() (Buffer, MultiCursor) {
	return bh.buffers[bh.idx], bh.cursors[bh.idx].Dup()
}

func (bh *BufferHist) Next() (Buffer, MultiCursor) {
	return bh.Increment(1)
}

func (bh *BufferHist) Prev() (Buffer, MultiCursor) {
	return bh.Increment(-1)
}

func (bh *BufferHist) Increment(n int) (Buffer, MultiCursor) {
	bh.idx += n
	if bh.idx >= len(bh.buffers) {
		bh.idx = len(bh.buffers) - 1
	}
	if bh.idx < 0 {
		bh.idx = 0
	}
	return bh.Current()
}
