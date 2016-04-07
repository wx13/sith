package file_test

import (
	"testing"

	"github.com/wx13/sith/file"
)

func TestDup(t *testing.T) {
	line1 := file.Line("hello world")
	line2 := line1.Dup()
	if line1.ToString() != line2.ToString() {
		t.Error("Duped line content does not match.", line1, line2)
	}
	line1[0] = 'H'
	if line1.ToString() == line2.ToString() {
		t.Error("Duped line contains same memory.", line1, line2)
	}
}

func TestCommonStart(t *testing.T) {
	line1 := file.Line("hello world")
	line2 := file.Line("hello bob")
	cs := line1.CommonStart(line2)
	if cs.ToString() != "hello " {
		t.Error("Common start is wrong", line1, line2, cs)
	}
}

func TestTabs2spaces(t *testing.T) {
	line1 := file.Line("\t\thello world")
	line2 := line1.Tabs2spaces()
	if line2.ToString() != "        hello world" {
		t.Error("Tabs2spaces:", line2)
	}
}

func TestSearch(t *testing.T) {

	line := file.Line("hello world")

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
