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
