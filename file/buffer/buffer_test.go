package buffer_test

import (
	"github.com/wx13/sith/file/buffer"
	"testing"
)

func TestMakeBuffer(t *testing.T) {

	var buff buffer.Buffer

	buff = buffer.MakeBuffer([]string{""})
	if buff.Length() != 1 {
		t.Error("MakeBuffer is wrong:", buff)
	}

	buff = buffer.MakeBuffer([]string{"", "", "hello", ""})
	if buff.Length() != 4 {
		t.Error("MakeBuffer is wrong:", buff)
	}
	if buff.GetRow(2).Length() != 5 {
		t.Error("MakeBuffer is wrong:", buff)
	}

}

func TestBufferDup(t *testing.T) {

	buf1 := buffer.MakeBuffer([]string{"hello", "world"})
	buf2 := buf1.Dup()

	if buf1.GetRow(1).ToString() != buf2.GetRow(1).ToString() {
		t.Error("Duped buffers not equal:", buf1, buf2)
	}

	buf2.SetRow(1, buffer.MakeLine("wurld"))
	if buf1.GetRow(1).ToString() == buf2.GetRow(1).ToString() {
		t.Error("Duped buffers too identical:", buf1, buf2)
	}

}

func TestBufferDeepDup(t *testing.T) {

	buf1 := buffer.MakeBuffer([]string{"hello", "world"})
	buf2 := buf1.DeepDup()

	if buf1.GetRow(1).ToString() != buf2.GetRow(1).ToString() {
		t.Error("Duped buffers not equal:", buf1, buf2)
	}

	buf2.SetRow(1, buffer.MakeLine("wurld"))
	if buf1.GetRow(1).ToString() == buf2.GetRow(1).ToString() {
		t.Error("Duped buffers too identical:", buf1, buf2)
	}

}

func TestReplaceBuffer(t *testing.T) {
	buf1 := buffer.MakeBuffer([]string{"hello", "world", "", ""})
	buf2 := buffer.MakeBuffer([]string{"yo", "adrian"})
	buf1.ReplaceBuffer(buf2)
	if buf1.GetRow(0).ToString() != "yo" {
		t.Error("ReplaceBuffer failed:", buf1, buf2)
	}
}

func TestBufferAppend(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"", "hello"})
	line := buffer.MakeLine("world")
	buf.Append(line)
	if buf.Length() != 3 {
		t.Error("Buffer Append failed", buf)
	}
}

func TestMakeSplitBuffer(t *testing.T) {
	str := "This is not a very long line, but it is long enough."
	buf := buffer.MakeSplitBuffer(str, 40)
	if buf.Length() != 2 {
		t.Error("SplitBuffer error:", buf.Length())
	}
	buf = buffer.MakeSplitBuffer(str, 20)
	if buf.Length() != 3 {
		t.Error("SplitBuffer error:", buf.Length())
	}
}

func TestToString(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"hello", "world"})
	str := buf.ToString("\n")
	if str != "hello\nworld" {
		t.Error("ToString is wrong:", str)
	}
}

func TestReplaceLines(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"a", "b", "c", "d", "e"})
	lines := []buffer.Line{
		buffer.MakeLine("hello"),
		buffer.MakeLine("world"),
	}
	buf.ReplaceLines(lines, 1, 3)
	if buf.Length() != 4 {
		t.Error("ReplaceLines failed:", buf.ToString("\n"))
	}
	if buf.GetRow(3).ToString() != "e" {
		t.Error("ReplaceLines failed:", buf.ToString("\n"))
	}
}

func TestRowLength(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"123", "1234"})
	n := buf.RowLength(1)
	if n != 4 {
		t.Error("rowlength is wrong:", n)
	}
}

func TestGetIndent(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"  hello", "  world", "    foo"})
	indent, clean := buf.GetIndent()
	if !clean {
		t.Error("should be clean")
	}
	if indent != "  " {
		t.Error("indent should be two spaces")
	}
}

func TestInclSlice(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"a", "b", "c", "d", "e"})
	buf2 := buf.InclSlice(1, 2)
	if buf2.Length() != 2 {
		t.Error("InclSlice is wrong:", buf2)
	}
	if buf2.GetRow(1).ToString() != "c" {
		t.Error("InclSlice is wrong:", buf2)
	}
}

func TestBufferEdits(t *testing.T) {
	buf := buffer.MakeBuffer([]string{"a", "b", "c"})
	buf.InsertAfter(1, buffer.MakeLine("b2"), buffer.MakeLine("b3"))
	if buf.ToString("-") != "a-b-b2-b3-c" {
		t.Error("InsertAfter is broken", buf.ToString("-"))
	}
	buf.DeleteRow(0)
	if buf.ToString("-") != "b-b2-b3-c" {
		t.Error("InsertAfter is broken", buf.ToString("-"))
	}
	buf.ReplaceLine(buffer.MakeLine("z"), 0)
	if buf.ToString("-") != "z-b2-b3-c" {
		t.Error("InsertAfter is broken", buf.ToString("-"))
	}
	buf.DeleteRow(2)
	if buf.ToString("-") != "z-b2-c" {
		t.Error("InsertAfter is broken", buf.ToString("-"))
	}
}

func TestBufferInsertStr(t *testing.T) {
	var buf buffer.Buffer
	var cols map[int][]int

	buf = buffer.MakeBuffer([]string{"abc", "def", "ghi"})
	cols = buf.InsertStr("//", map[int][]int{0: {0}, 1: {0}})
	if buf.ToString("-") != "//abc-//def-ghi" {
		t.Error("InsertStr at start of lines", buf.ToString("-"))
	}
	if !intSliceEq(cols[0], 2) || !intSliceEq(cols[1], 2) {
		t.Error("InsertStr at start of lines", cols)
	}

	buf = buffer.MakeBuffer([]string{"abcdef", "abcdef"})
	cols = buf.InsertStr("//", map[int][]int{0: {0, 3}, 1: {0}})
	if buf.ToString("-") != "//abc//def-//abcdef" {
		t.Error("InsertStr at multiple places", buf.ToString("-"))
	}
	if !intSliceEq(cols[0], 2, 7) || !intSliceEq(cols[1], 2) {
		t.Error("InsertStr at multiple", cols)
	}

}

func TestDeleteChars(t *testing.T) {
	var buf buffer.Buffer
	var cols map[int][]int

	buf = buffer.MakeBuffer([]string{"abcdef", "abcdef"})
	cols = buf.DeleteChars(2, map[int][]int{0: {3}, 1: {3}})
	if buf.ToString("-") != "abcf-abcf" {
		t.Error("DeleteChars on two lines", buf.ToString("-"))
	}
	if !intSliceEq(cols[0], 3) || !intSliceEq(cols[1], 3) {
		t.Error("DeleteChars on two lines", cols)
	}

	buf = buffer.MakeBuffer([]string{"abcdef", "abcdef"})
	cols = buf.DeleteChars(2, map[int][]int{0: {0, 3}, 1: {3}})
	if buf.ToString("-") != "cf-abcf" {
		t.Error("DeleteChars twice on same row", buf.ToString("-"))
	}
	if !intSliceEq(cols[0], 0, 1) || !intSliceEq(cols[1], 3) {
		t.Error("DeleteChars twice on same row", cols)
	}

	buf = buffer.MakeBuffer([]string{"abcdef", "abcdef"})
	cols = buf.DeleteChars(-2, map[int][]int{0: {3}, 1: {3}})
	if buf.ToString("-") != "adef-adef" {
		t.Error("DeleteChars backwards on two lines", buf.ToString("-"))
	}
	if !intSliceEq(cols[0], 1) || !intSliceEq(cols[1], 1) {
		t.Error("DeleteChars backwards on two lines", cols)
	}

	buf = buffer.MakeBuffer([]string{"0123456789"})
	cols = buf.DeleteChars(-3, map[int][]int{0: {5, 7}})
	if buf.ToString("-") != "01789" {
		t.Error("DeleteChars overlapping backspace", buf.ToString("-"))
	}
	if !intSliceEq(cols[0], 2) {
		t.Error("DeleteChars overlapping backspace", cols)
	}

}

func TestBufferBracketMatch(t *testing.T) {

	var buf buffer.Buffer
	var row, col int
	var err error

	buf = buffer.MakeBuffer([]string{"def foo(a, b) {}"})
	row, col, err = buf.BracketMatch(0, 7, 0)
	if err != nil || row != 0 || col != 12 {
		t.Error("One line", row, col, err)
	}

	buf = buffer.MakeBuffer([]string{"foo {", "blah, blah", "}"})
	row, col, err = buf.BracketMatch(0, 4, 2)
	if err != nil || row != 2 || col != 0 {
		t.Error("Multiline", row, col, err)
	}

	buf = buffer.MakeBuffer([]string{"foo {", "blah, blah", "}"})
	row, col, err = buf.BracketMatch(2, 0, 0)
	if err != nil || row != 0 || col != 4 {
		t.Error("Backward", row, col, err)
	}

	buf = buffer.MakeBuffer([]string{"foo(bar(baz(", "  thing(), thing()))", ")"})
	row, col, err = buf.BracketMatch(2, 0, 0)
	if err != nil || row != 0 || col != 3 {
		t.Error("Backward nested", row, col, err)
	}

}
