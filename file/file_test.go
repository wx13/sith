package file_test

import (
	"github.com/wx13/sith/file"
	"testing"
)

func TestNewFile(t *testing.T) {
	f := file.NewFile("", make(chan struct{}), nil)
	if f == nil {
		t.Error("bad")
	}
}

func CheckBuffer(t *testing.T, f *file.File, s, msg string) {
	if f.ToString() != s {
		t.Errorf("Error: %s; Expected %q but got %q\n", msg, s, f.ToString())
	}
}

func TestInserChar(t *testing.T) {
	f := file.NewFile("", make(chan struct{}), nil)
	f.InsertChar('h')
	f.InsertChar('e')
	f.InsertChar('l')
	f.InsertChar('l')
	f.InsertChar('o')
	CheckBuffer(t, f, "hello", "InsertChar")
}

func TestEditing(t *testing.T) {
	f := file.NewFile("", make(chan struct{}), nil)
	f.InsertChar('a')
	f.InsertChar('b')
	f.InsertChar('c')
	f.Newline()
	f.InsertChar('d')
	f.InsertChar('e')
	CheckBuffer(t, f, "abc\nde", "InsertChar")
	f.Backspace()
	CheckBuffer(t, f, "abc\nd", "Backspace")
	f.Backspace()
	f.Backspace()
	CheckBuffer(t, f, "abc", "Backspace (over newline)")
}
