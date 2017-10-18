package editor

import (
	"fmt"
	"strconv"

	"github.com/wx13/sith/ui"
)

func (editor *Editor) Bookmark() {
	p := ui.MakePrompt(editor.screen, editor.keyboard)
	name, err := p.Ask("bookmark:", nil)
	if err != nil {
		return
	}

	if editor.bookmarks == nil {
		editor.bookmarks = NewBookmarks()
	}
	file := editor.file.Name
	row, _ := editor.file.GetRowCol(0)
	editor.bookmarks.Add(name, file, row)
}

func (editor *Editor) BookmarkMenu() {
	if editor.bookmarks == nil {
		return
	}
	names := editor.bookmarks.Names()
	menu := ui.NewMenu(editor.screen, editor.keyboard)
	idx, key := menu.Choose(names, 0, "")
	editor.Flush()
	if idx < 0 || key == "cancel" {
		return
	}
	editor.GoToBookmark(names[idx])
}

func (editor *Editor) GoToBookmark(name string) error {
	filename, line := editor.bookmarks.Get(name)
	if filename == "" {
		return fmt.Errorf("no such bookmark")
	}
	err := editor.SwitchFileByName(filename)
	if err != nil {
		return err
	}
	editor.file.CursorGoTo(line, 0)
	return nil
}

func (editor *Editor) GoToLine() {
	prompt := ui.MakePrompt(editor.screen, editor.keyboard)
	ans := prompt.GetAnswer("goto:", &editor.gotoHist)
	if ans == "" {
		return
	}

	if editor.bookmarks != nil {
		err := editor.GoToBookmark(ans)
		if err == nil {
			return
		}
	}
	row, err := strconv.Atoi(ans)
	if err == nil {
		editor.file.CursorGoTo(row, 0)
	}
}

type Bookmarks struct {
	Max   int
	newer map[string]Bookmark
	older map[string]Bookmark
}

type Bookmark struct {
	filename string
	line     int
}

func NewBookmarks() *Bookmarks {
	b := Bookmarks{
		newer: map[string]Bookmark{},
		older: map[string]Bookmark{},
		Max:   10000,
	}
	return &b
}

func (b *Bookmarks) Names() []string {
	names := []string{}
	for name, _ := range b.newer {
		names = append(names, name)
	}
	for name, _ := range b.older {
		names = append(names, name)
	}
	return names
}

func (b *Bookmarks) Add(name, filename string, line int) {
	b.newer[name] = Bookmark{
		filename: filename,
		line:     line,
	}
	if len(b.newer) > b.Max/2 {
		b.older = b.newer
		b.newer = map[string]Bookmark{}
	}
}

func (b *Bookmarks) Get(name string) (string, int) {
	bookmark, ok := b.newer[name]
	if ok {
		return bookmark.filename, bookmark.line
	}
	bookmark, ok = b.older[name]
	if ok {
		return bookmark.filename, bookmark.line
	}
	return "", 0
}
