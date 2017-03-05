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

func makeMC() cursor.MultiCursor {
	mc := cursor.MakeMultiCursor()
	mc.ResetCursors(map[int][]int{
		10: {12, 12, 15},
		2:  {15, 15},
	})
	return mc
}

func TestGetRows(t *testing.T) {
	mc := makeMC()
	rows := mc.GetRows()
	if len(rows) != 2 {
		t.Errorf("GetRows failed: %#v %#v\n", mc, rows)
	}
}

func TestMCOnePerLine(t *testing.T) {
	mc := makeMC()
	mc.OnePerLine()
	cursors := mc.Cursors()
	if len(cursors) != 2 {
		t.Errorf("OnePerLine failed: %#v\n", cursors)
	}
}

func TestMCClear(t *testing.T) {
	mc := cursor.MakeMultiCursor()

	mc.Snapshot()
	mc.Snapshot()
	if mc.Length() != 1 {
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
	if mc.Length() != 1 {
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
