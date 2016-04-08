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
