package editor_test

import (
	"github.com/wx13/sith/editor"
	"testing"
)

func TestBookmarks(t *testing.T) {
	b := editor.NewBookmarks()

	// Empty bookmarks should return empty results.
	file, line := b.Get("foo")
	if file != "" {
		t.Error(file, line)
	}

	// Add some entries.
	b.Add("foo", "bar.c", 71)
	b.Add("foo 2", "bar.c", 100)

	// Should be successful
	file, line = b.Get("foo")
	if file != "bar.c" || line != 71 {
		t.Error(file, line)
	}

	// Set a maximum.
	b.Max = 4
	b.Add("abc", "abc.go", 0)
	b.Add("123", "abc.go", 10)

	// Should still be ok.
	file, line = b.Get("foo")
	if file != "bar.c" || line != 71 {
		t.Error(file, line)
	}

	// Should start expiring stuff.
	b.Add("def", "abc.go", 20)
	b.Add("ghi", "abc.go", 30)
	file, line = b.Get("foo")
	if file != "" {
		t.Error(file, line)
	}

}
