package cursor_test

import (
	"github.com/wx13/sith/file/cursor"
	"testing"
)

func TestCursorDup(t *testing.T) {
	cur1 := cursor.MakeCursor(10, 22)
	cur2 := cur1.Dup()
	if cur1.Row() != cur2.Row() {
		t.Error("Cursor Dup broken", cur1, cur2)
	}
	cur1.Set(11, 23, 23)
	if cur1.Row() == cur2.Row() {
		t.Error("Cursor Dup broken", cur1, cur2)
	}
}

func TestMCClear(t *testing.T) {
	mc := cursor.MakeMultiCursor()

	mc.Snapshot()
	mc.Snapshot()
	if mc.Length() != 3 {
		t.Error("Wrong MC length:", mc.Length())
	}

	mc.Clear()
	if mc.Length() != 1 {
		t.Error("Wrong MC length:", mc.Length())
	}

	mc.Snapshot()
	mc.Snapshot()
	mc.Snapshot()
	mc.Snapshot()
	mc.OuterMost()
	if mc.Length() != 2 {
		t.Error("Outermost failed:", mc.Length())
	}

	mc.Clear()
	mc.Set(5, 10, 10)
	mc.Append(cursor.MakeCursor(10, 12))
	mc.SetColumn()
	if mc.Length() != 6 {
		t.Error("SetColumn() failed:", mc.Length())
	}
}

func TestSortedRowsCols(t *testing.T) {
	mc := cursor.MakeMultiCursor()
	mc.Set(10, 12, 12)
	mc.Snapshot()
	mc.Set(10, 5, 5)
	mc.Snapshot()
	mc.Set(10, 15, 15)
	mc.Snapshot()
	mc.Set(2, 15, 15)
	mc.Snapshot()
	mc.Set(2, 1, 1)
	rows, cols := mc.SortedRowsCols()
	if len(rows) != 2 || rows[0] != 2 || rows[1] != 10 {
		t.Error("Rows:", rows)
	}
	if len(cols[10]) != 3 || cols[10][0] != 5 || cols[10][2] != 15 {
		t.Error("Cols:", cols)
	}
}
