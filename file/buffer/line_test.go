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
	line2 := line1.Tabs2spaces()
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
		t.Error("search:", line.ToString(), "llo", a, b)
	}

}

func TestRemoveTrailingWhitespace(t *testing.T) {

	tests := [][]string{
		[]string{"  foo", "  foo"},
		[]string{"  foo  ", "  foo"},
		[]string{"  foo\t", "  foo"},
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
