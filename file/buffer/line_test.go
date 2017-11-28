package buffer_test

import (
	"testing"

	"github.com/wx13/sith/file/buffer"
)

func TestDup(t *testing.T) {
	line1 := buffer.MakeLine("hello world")
	line2 := line1.Dup()
	if line1.ToString() != line2.ToString() {
		t.Error("Duped line content does not match.", line1, line2)
	}
	line1.SetChar(0, 'H')
	if line1.ToString() == line2.ToString() {
		t.Error("Duped line contains same memory.", line1, line2)
	}
}

func TestCommonStart(t *testing.T) {
	line1 := buffer.MakeLine("hello world")
	line2 := buffer.MakeLine("hello bob")
	cs := line1.CommonStart(line2)
	if cs.ToString() != "hello " {
		t.Error("Common start is wrong", line1, line2, cs)
	}
	line3 := buffer.MakeLine("hello world")
	cs = line1.CommonStart(line3)
	if cs.ToString() != "hello world" {
		t.Error("Common start is wrong", line1, line3, cs)
	}
}

func TestTabs2spaces(t *testing.T) {
	line1 := buffer.MakeLine("\t\thello world")
	line2 := line1.Tabs2spaces(4)
	if line2.ToString() != "        hello world" {
		t.Error("Tabs2spaces:", line2)
	}
}

func TestSearch(t *testing.T) {

	line := buffer.MakeLine("hello world")

	a, b := line.Search("llo", 0, -1)
	if a != 2 || b != 5 {
		t.Error("search:", line.ToString(), "llo", a, b)
	}

	a, b = line.Search("llo", 2, -1)
	if a != 2 || b != 5 {
		t.Error("search:", line.ToString(), "llo", a, b)
	}

	a, b = line.Search("llo", 0, 2)
	if a != -1 {
		t.Error("search:", line.ToString(), "llo", a, b)
	}

	a, b = line.Search("/.o/", 0, -1)
	if a != 3 || b != 5 {
		t.Error("search:", line.ToString(), "/.o/", a, b)
	}

}

func TestRemoveTrailingWhitespace(t *testing.T) {

	tests := [][]string{
		{"  foo", "  foo"},
		{"  foo  ", "  foo"},
		{"  foo\t", "  foo"},
	}

	for _, test := range tests {
		line1 := buffer.MakeLine(test[0])
		line2 := line1.RemoveTrailingWhitespace()
		if line2.ToString() != test[1] {
			t.Errorf("remove trailing whitespace: --%s-- => --%s--", line1.ToString(), line2.ToString())
		}
	}

}

func TestSlice(t *testing.T) {
	line1 := buffer.MakeLine("012345678")
	var line2 buffer.Line

	line2 = line1.Slice(0, 4)
	if line2.ToString() != "0123" {
		t.Error("slice:", line1.ToString(), line2.ToString())
	}

	line2 = line1.Slice(0, -1)
	if line2.ToString() != "012345678" {
		t.Error("slice:", line1.ToString(), line2.ToString())
	}

	line2 = line1.Slice(4, 6)
	if line2.ToString() != "45" {
		t.Error("slice:", line1.ToString(), line2.ToString())
	}

	line2 = line1.Slice(4, 20)
	if line2.ToString() != "45678" {
		t.Error("slice:", line1.ToString(), line2.ToString())
	}

}

func TestBracketMatch(t *testing.T) {

	var line buffer.Line
	var idx, count int

	line = buffer.MakeLine("foo(bar)")
	idx, count = line.BracketMatch('(', ')', 4, 1, 1)
	if count != 0 || idx != 7 {
		t.Error("Simple match", idx, count)
	}

	line = buffer.MakeLine("blah)")
	idx, count = line.BracketMatch('(', ')', 0, 1, 1)
	if count != 0 || idx != 4 {
		t.Error("Continued match, level 1", idx, count)
	}

	line = buffer.MakeLine("blah(foo), blah) {")
	idx, count = line.BracketMatch('(', ')', 0, 1, 1)
	if count != 0 || idx != 15 {
		t.Error("Continued match, level 1, decoy", idx, count)
	}

	line = buffer.MakeLine("def foo(bar(")
	idx, count = line.BracketMatch('(', ')', 8, 1, 1)
	if count != 2 {
		t.Error("No match", idx, count)
	}

	line = buffer.MakeLine("def foo(bar, baz()) {")
	idx, count = line.BracketMatch(')', '(', 17, -1, 1)
	if count != 0 || idx != 7 {
		t.Error("Reverse match", idx, count)
	}

}

func intSliceEq(a []int, b ...int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestInsertStr(t *testing.T) {

	var line buffer.Line
	var cols []int

	line = buffer.MakeLine("")
	cols = line.InsertStr("foo", 1)
	if line.ToString() != "" {
		t.Error("Insert past end of string:", line.ToString())
	}
	if !intSliceEq(cols, 1) {
		t.Error("Insert past end of string:", cols)
	}

	line = buffer.MakeLine("")
	cols = line.InsertStr("foo", 0)
	if line.ToString() != "foo" {
		t.Error("Insert into empty line", line.ToString())
	}
	if !intSliceEq(cols, 3) {
		t.Error("Insert into empty line", cols)
	}

	line = buffer.MakeLine("")
	cols = line.InsertStr("foo", 0, 1)
	if line.ToString() != "foo" {
		t.Error("Insert twice into empty line", line.ToString())
	}
	if !intSliceEq(cols, 3, 4) {
		t.Error("Insert twice into empty line", cols)
	}

	line = buffer.MakeLine("hi bob")
	cols = line.InsertStr(" there", 2)
	if line.ToString() != "hi there bob" {
		t.Error("Insert once", line.ToString())
	}
	if !intSliceEq(cols, 8) {
		t.Error("Insert once", cols)
	}

	line = buffer.MakeLine("hi bob")
	cols = line.InsertStr(" there", 2, 2, 2)
	if line.ToString() != "hi there bob" {
		t.Error("Duplicate column", line.ToString())
	}
	if !intSliceEq(cols, 8) {
		t.Error("Duplicate column", cols)
	}

	line = buffer.MakeLine("a_ a_ a_")
	cols = line.InsertStr("b", 5, 2, 8)
	if line.ToString() != "a_b a_b a_b" {
		t.Error("Insert out of order", line.ToString())
	}
	if !intSliceEq(cols, 3, 7, 11) {
		t.Error("Insert out of order", cols)
	}

}

func TestDeleteFwd(t *testing.T) {

	var line buffer.Line
	var cols []int

	line = buffer.MakeLine("hello")
	cols = line.DeleteFwd(5, 12)
	if line.ToString() != "hello" {
		t.Error("Delete past end of line", line.ToString())
	}
	if !intSliceEq(cols, 0) {
		t.Error("Delete past end of line", cols)
	}

	line = buffer.MakeLine("hello")
	cols = line.DeleteFwd(2, 1)
	if line.ToString() != "hlo" {
		t.Error("Delete in one place", line.ToString())
	}
	if !intSliceEq(cols, 1) {
		t.Error("Delete in one place", cols)
	}

	line = buffer.MakeLine("01234")
	cols = line.DeleteFwd(2, 0)
	if line.ToString() != "234" {
		t.Error("Delete at start of line", line.ToString())
	}
	if !intSliceEq(cols, 0) {
		t.Error("Delete at start of line", cols)
	}

	line = buffer.MakeLine("012345678")
	cols = line.DeleteFwd(2, 0, 5)
	if line.ToString() != "23478" {
		t.Error("Delete at two places", line.ToString())
	}
	if !intSliceEq(cols, 0, 3) {
		t.Error("Delete at two places", cols)
	}

	line = buffer.MakeLine("012345678")
	cols = line.DeleteFwd(4, 0, 3)
	if line.ToString() != "78" {
		t.Error("Overlapping delete", line.ToString())
	}
	if !intSliceEq(cols, 0) {
		t.Error("Overlapping delete", cols)
	}

}

func TestDeleteBkwd(t *testing.T) {

	var line buffer.Line
	var cols []int

	line = buffer.MakeLine("hello")
	cols = line.DeleteBkwd(5, 12)
	if line.ToString() != "hello" {
		t.Error("Delete past end of line", line.ToString())
	}
	if !intSliceEq(cols, 0) {
		t.Error("Delete past end of line", cols)
	}

	line = buffer.MakeLine("0123456789")
	cols = line.DeleteBkwd(3, 7)
	if line.ToString() != "0123789" {
		t.Error("Delete at one col", line.ToString())
	}
	if !intSliceEq(cols, 4) {
		t.Error("Delete at one col", cols)
	}

	line = buffer.MakeLine("0123456789")
	cols = line.DeleteBkwd(2, 3, 8)
	if line.ToString() != "034589" {
		t.Error("Delete at two columns", line.ToString())
	}
	if !intSliceEq(cols, 1, 4) {
		t.Error("Delete at two columns", cols)
	}

	line = buffer.MakeLine("abc/abc/abc/.")
	cols = line.DeleteBkwd(1, 4, 8, 12)
	if line.ToString() != "abcabcabc." {
		t.Error("Delete at three columns", line.ToString())
	}
	if !intSliceEq(cols, 3, 6, 9) {
		t.Error("Delete at three columns", cols)
	}

	line = buffer.MakeLine("0123456789")
	cols = line.DeleteBkwd(5, 2)
	if line.ToString() != "23456789" {
		t.Error("Delete to start of line", line.ToString())
	}
	if !intSliceEq(cols, 0) {
		t.Error("Delete to start of line", cols)
	}

	line = buffer.MakeLine("0")
	cols = line.DeleteBkwd(1, 1)
	if line.ToString() != "" {
		t.Error("Delete only char:", line.ToString())
	}
	if !intSliceEq(cols, 0) {
		t.Error("Delete only char:", cols)
	}

}

func TestCompressPriorSpaces(t *testing.T) {

	var line buffer.Line
	var cols []int

	line = buffer.MakeLine("abc     def")
	cols = line.CompressPriorSpaces([]int{8})
	if line.ToString() != "abc def" {
		t.Error("Single cursor", line.ToString())
	}
	if !intSliceEq(cols, 4) {
		t.Error("Single cursor", cols)
	}

	line = buffer.MakeLine("abc     def     ghi")
	cols = line.CompressPriorSpaces([]int{8, 16})
	if line.ToString() != "abc def ghi" {
		t.Error("Single cursor", line.ToString())
	}
	if !intSliceEq(cols, 4, 8) {
		t.Error("Single cursor", cols)
	}

}

func TestToCorpus(t *testing.T) {
	var line buffer.Line
	var str string

	line = buffer.MakeLine("abc def ghi")
	str = line.ToCorpus(0)
	if str != "def ghi" {
		t.Error("ToCorpus:", str)
	}

	line = buffer.MakeLine("abc def ghi")
	str = line.ToCorpus(2)
	if str != "def ghi" {
		t.Error("ToCorpus:", str)
	}

	line = buffer.MakeLine("abc def ghi")
	str = line.ToCorpus(3)
	if str != "def ghi" {
		t.Error("ToCorpus:", str)
	}

	line = buffer.MakeLine("abc def ghi")
	str = line.ToCorpus(5)
	if str != "abc ghi" {
		t.Error("ToCorpus:", str)
	}

}
