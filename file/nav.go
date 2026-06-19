package file

import (
	"regexp"
	"sort"
)

func (file *File) enforceColBounds(indexes ...int) {
	if len(indexes) == 0 {
		for idx, _ := range file.MultiCursor.Cursors() {
			indexes = append(indexes, idx)
		}
	}
	for _, idx := range indexes {
		cursor := file.MultiCursor.GetCursor(idx)
		if cursor.Col() > file.buffer.RowLength(cursor.Row()) {
			file.MultiCursor.SetCol(idx, file.buffer.RowLength(cursor.Row()))
		}
		if cursor.Col() < 0 {
			file.MultiCursor.SetCol(idx, 0)
		}
	}
}

func (file *File) enforceRowBounds(indexes ...int) {
	if len(indexes) == 0 {
		for idx, _ := range file.MultiCursor.Cursors() {
			indexes = append(indexes, idx)
		}
	}
	for _, idx := range indexes {
		cursor := file.MultiCursor.GetCursor(idx)
		if cursor.Row() >= file.buffer.Length() {
			file.MultiCursor.SetRow(idx, file.buffer.Length()-1)
		}
		if cursor.Row() < 0 {
			file.MultiCursor.SetRow(idx, 0)
		}
	}
}

func (file *File) makeCursorNotAtTopBottom() {
	row := file.MultiCursor.GetRow(0)
	_, rows := file.screen.Size()
	bottom := file.rowOffset + rows - 1
	if row >= bottom {
		file.rowOffset += (row - bottom) + rows/8
	}
}

// CursorGoTo moves the cursor to a row, col position. If row is negative, then
// it specifies from the end of the file.
func (file *File) CursorGoTo(row, col int) {
	if row < 0 {
		row = file.Length() + row
	}
	file.MultiCursor.Set(row, col, col)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.makeCursorNotAtTopBottom()
}

// PageDown moves the cursor half a screen down.
func (file *File) PageDown() {
	_, rows := file.screen.Size()
	file.CursorDown(rows/2 - 1)
}

// PageUp moves the cursor have a screen up.
func (file *File) PageUp() {
	_, rows := file.screen.Size()
	file.CursorUp(rows/2 - 1)
}

// CursorUp moves the cursor up n rows.
func (file *File) CursorUp(n int) {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() || file.MultiCursor.NavModeIsColumn() {
		cursors = cursors[:1]
	}
	for idx := range cursors {
		row, _, colwant := file.MultiCursor.GetCursorRCC(idx)
		row -= n
		if row < 0 {
			row = 0
		}
		file.MultiCursor.SetCursor(idx, row, colwant, colwant)
	}
	file.enforceRowBounds()
	file.enforceColBounds()
}

// CursorDown moves the cursor down n rows.
func (file *File) CursorDown(n int) {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() || file.MultiCursor.NavModeIsColumn() {
		cursors = cursors[:1]
	}
	for idx := range cursors {
		row, _, colwant := file.MultiCursor.GetCursorRCC(idx)
		row += n
		if row >= file.buffer.Length() {
			row = file.buffer.Length() - 1
		}
		file.MultiCursor.SetCursor(idx, row, colwant, colwant)
	}
	file.enforceRowBounds()
	file.enforceColBounds()
}

// CursorRight moves the cursor one column to the right.
func (file *File) CursorRight() {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx, cursor := range cursors {
		row, col := cursor.RowCol()
		if col < file.buffer.RowLength(row) {
			file.MultiCursor.SetCol(idx, col+1)
		} else {
			if file.MultiCursor.Length() > 1 {
				continue
			}
			if row < file.buffer.Length()-1 {
				file.MultiCursor.SetRow(idx, row+1)
				file.MultiCursor.SetCol(idx, 0)
			}
		}
		file.MultiCursor.SetColwant(idx, -1)
	}
	file.enforceRowBounds()
	file.enforceColBounds()
}

// CursorLeft moves the cursor one column to the left.
func (file *File) CursorLeft() {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx, cursor := range cursors {
		row, col := cursor.RowCol()
		if col > 0 {
			file.MultiCursor.SetCol(idx, col-1)
		} else {
			if file.MultiCursor.Length() > 1 {
				continue
			}
			if row > 0 {
				row--
				col = file.buffer.RowLength(row)
				file.MultiCursor.SetCursor(idx, row, col, col)
			}
		}
		file.MultiCursor.SetColwant(idx, -1)
	}
}

// GetCursor returns the row, col position for the specified multi-cursor index.
func (file *File) GetCursor(idx int) (int, int) {
	file.enforceRowBounds(idx)
	file.enforceColBounds(idx)
	row, col, _ := file.MultiCursor.GetCursorRCC(idx)
	line := file.buffer.GetRowDirect(row).Slice(0, col).Tabs2spaces(file.tabWidth)
	n := file.screen.StringDispLen(line.ToString())
	return row - file.rowOffset, n - file.colOffset
}

func (file *File) GetRowCol(idx int) (int, int) {
	file.enforceRowBounds(idx)
	file.enforceColBounds(idx)
	row, col, _ := file.MultiCursor.GetCursorRCC(idx)
	return row, col
}

func (file *File) GetRowsCols() map[int][]int {
	return file.MultiCursor.GetRowsCols()
}

// ScrollLeft shifts the view screen to the left.
func (file *File) ScrollLeft() {
	file.colOffset++
}

// ScrollRight shifts the view screen to the right.
func (file *File) ScrollRight() {
	if file.colOffset > 0 {
		file.colOffset--
	}
}

// ScrollUp shifts the screen up one row.
func (file *File) ScrollUp() {
	if file.rowOffset < file.buffer.Length()-1 {
		file.rowOffset++
	}
}

// ScrollDown shifts the screen down one row.
func (file *File) ScrollDown() {
	if file.rowOffset > 0 {
		file.rowOffset--
	}
}

func (file *File) updateOffsets(nRows, nCols int) {

	row := file.MultiCursor.GetRow(0)
	if row < file.rowOffset {
		file.rowOffset = row
	}
	if row >= file.rowOffset+nRows-1 {
		file.rowOffset = row - nRows + 1
	}

	_, col := file.GetCursor(0)
	col += file.colOffset
	if col < file.colOffset {
		file.colOffset = col
	}
	if col >= file.colOffset+nCols-1 {
		file.colOffset = col - nCols + 1
	}

}

// StartOfLine moves the cursors to the start of the line.
// If they are already at the start, the moves them to the first
// non-whitespace character.
func (file *File) StartOfLine() {
	allAtZero := true
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for _, cursor := range cursors {
		if cursor.Col() != 0 {
			allAtZero = false
			break
		}
	}
	if allAtZero {
		re := regexp.MustCompile("^[ \t]*")
		for idx, cursor := range cursors {
			row := cursor.Row()
			line := file.buffer.GetRowDirect(row)
			match := re.FindStringIndex(line.ToString())
			file.MultiCursor.SetCol(idx, match[1])
			file.MultiCursor.SetColwant(idx, -1)
		}
	} else {
		for idx := range cursors {
			file.MultiCursor.SetCol(idx, 0)
			file.MultiCursor.SetColwant(idx, -1)
		}
	}
}

// EndOfLine moves the cursors to the end of the line.
func (file *File) EndOfLine() {
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}
	for idx := range cursors {
		row := file.MultiCursor.GetRow(idx)
		line := file.buffer.GetRowDirect(row)
		file.MultiCursor.SetCol(idx, line.Length())
		file.MultiCursor.SetColwant(idx, -1)
	}
}

// NextWord moves the cursor to the next word.
func (file *File) NextWord() {
	file.prevNextWord(1)
}

// PrevWord moves the cursor to the previous word.
func (file *File) PrevWord() {
	file.prevNextWord(-1)
}

func (file *File) prevNextWord(incr int) {
	// If in detached mode, only one cursor moves.
	cursors := file.MultiCursor.Cursors()
	if file.MultiCursor.NavModeIsDetached() {
		cursors = cursors[:1]
	}

	// If the cursors columns don't change, then we are at the start/end of the
	// line and should wrap.
	unchanged := true

	// Move each cursor.
	for idx, cursor := range cursors {
		row := cursor.Row()
		col := cursor.Col()
		line := file.buffer.GetRowDirect(row)

		// Store the old cursor, compute the new, and check for changes.
		old_col := col
		col = line.PrevNextWord(col, incr)
		if old_col != col {
			unchanged = false
		}

		// Move the cursor.
		file.MultiCursor.SetCol(idx, col)
		file.MultiCursor.SetColwant(idx, -1)
	}

	// Move up/down if at start/end of line.
	if unchanged && (file.MultiCursor.NavModeIsAllTogether() ||
		file.MultiCursor.Length() == 1) {
		if incr > 0 {
			file.CursorDown(1)
			file.StartOfLine()
		} else {
			file.CursorUp(1)
			file.EndOfLine()
		}

	}
}

// NextChange moves the cursor to the next modified or added line.
func (file *File) NextChange() {
	file.gotoChange(1)
}

// PrevChange moves the cursor to the previous modified or added line.
func (file *File) PrevChange() {
	file.gotoChange(-1)
}

func (file *File) gotoChange(direction int) {
	diffResult := file.buffer.DiffLinesFull(&file.savedBuffer)

	// Build sorted list of all change points:
	// - Changed/added lines (navigate to that line)
	// - Deletion points (navigate to the line after the deletion, or line 0 if deletion at start)
	changePoints := make([]int, 0)

	for lineNum := range diffResult.Changes {
		changePoints = append(changePoints, lineNum)
	}

	for _, delPoint := range diffResult.DeletionPoints {
		// Deletion after line N means we navigate to line N+1 (or 0 if N is -1)
		targetLine := delPoint + 1
		if targetLine < 0 {
			targetLine = 0
		}
		// Only add if not already in the list
		found := false
		for _, cp := range changePoints {
			if cp == targetLine {
				found = true
				break
			}
		}
		if !found {
			changePoints = append(changePoints, targetLine)
		}
	}

	sort.Ints(changePoints)

	if len(changePoints) == 0 {
		file.NotifyUser("No changes")
		return
	}

	currentRow := file.MultiCursor.GetRow(0)

	var targetRow int
	found := false

	if direction > 0 {
		// Find next change after current row
		for _, line := range changePoints {
			if line > currentRow {
				targetRow = line
				found = true
				break
			}
		}
		// Wrap around to first change
		if !found {
			targetRow = changePoints[0]
			found = true
		}
	} else {
		// Find previous change before current row
		for i := len(changePoints) - 1; i >= 0; i-- {
			if changePoints[i] < currentRow {
				targetRow = changePoints[i]
				found = true
				break
			}
		}
		// Wrap around to last change
		if !found {
			targetRow = changePoints[len(changePoints)-1]
			found = true
		}
	}

	if found {
		file.CursorGoTo(targetRow, 0)
	}
}
