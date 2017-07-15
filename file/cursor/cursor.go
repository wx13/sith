// Package cursor handles all the cursor and multicursor stuff.
package cursor

import ()

// Cursor keeps track of the row, col of a cursor,
// and also the wanted column (colwant) of the cursor.
// The wanted column is the column the cursor would like
// to be in, if the line were long enough.
type Cursor struct {
	row, col, colwant int
}

// MakeCursor creates a new Cursor object.
func MakeCursor(row, col int) Cursor {
	return Cursor{
		row:     row,
		col:     col,
		colwant: col,
	}
}

// CursorFromSlice creates a new Cursor object, from a slice of ints.
// The input slice is if the form [row, col, colwant].
func CursorFromSlice(s []int) Cursor {
	cursor := Cursor{}
	switch len(s) {
	case 3:
		cursor.colwant = s[2]
		fallthrough
	case 2:
		cursor.col = s[1]
		fallthrough
	case 1:
		cursor.row = s[0]
	}
	return cursor
}

// Dup duplicates a cursor object.
func (cursor Cursor) Dup() Cursor {
	return Cursor{
		row:     cursor.row,
		col:     cursor.col,
		colwant: cursor.colwant,
	}
}

func (cursor Cursor) Row() int {
	return cursor.row
}

func (cursor Cursor) Col() int {
	return cursor.col
}

func (cursor Cursor) Colwant() int {
	return cursor.colwant
}

// RowCol returns the row and column of the cursor.
func (cursor Cursor) RowCol() (int, int) {
	return cursor.row, cursor.col
}

func (cursor *Cursor) Set(row, col, colwant int) {
	cursor.row = row
	cursor.col = col
	cursor.colwant = col
}
