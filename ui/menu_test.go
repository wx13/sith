package ui_test

import (
	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
	"testing"
)

type MockScreen struct{}

func (ms MockScreen) Flush()                                                                      {}
func (ms MockScreen) Highlight(row, col int)                                                      {}
func (ms MockScreen) HighlightRange(startRow, endRow, starCol, endCol int)                        {}
func (ms MockScreen) ColorRange(startRow, endRow, starCol, endCol int, fg, bg terminal.Attribute) {}
func (ms MockScreen) WriteString(row, col int, text string)                                       {}
func (ms MockScreen) WriteStringColor(row, col int, text string, fg, bg terminal.Attribute)       {}
func (ms MockScreen) Size() (int, int) {
	return 80, 24
}
func (ms MockScreen) Row() int                { return 0 }
func (ms MockScreen) Col() int                { return 0 }
func (ms MockScreen) SetCursor(int, int)      {}
func (ms MockScreen) WriteMessage(msg string) {}

func TestEmptyMenuCancel(t *testing.T) {

	kb := terminal.NewMockKeyboard([]string{"ctrlC"}, []rune{})
	screen := MockScreen{}
	menu := ui.NewMenu(screen, kb)
	_, ans := menu.Choose([]string{}, 0, "")
	if ans != "cancel" {
		t.Error("Expected cancel, got", ans)
	}

}

func TestMenuChoose(t *testing.T) {

	screen := MockScreen{}

	// Choose 0th element.
	kb := terminal.NewMockKeyboard([]string{"enter"}, []rune{})
	menu := ui.NewMenu(screen, kb)
	idx, ans := menu.Choose([]string{"zero", "one"}, 0, "")
	if ans != "" || idx != 0 {
		t.Error("Expected 0, '', got", idx, ans)
	}

	// Choose 1st element.
	kb = terminal.NewMockKeyboard(
		[]string{"arrowDown", "enter"},
		[]rune{},
	)
	menu = ui.NewMenu(screen, kb)
	idx, ans = menu.Choose([]string{"zero", "one"}, 0, "")
	if ans != "" || idx != 1 {
		t.Error("Expected 1, '', got", idx, ans)
	}

	// Search.
	kb = terminal.NewMockKeyboard(
		[]string{"char", "char", "enter"},
		[]rune{'o', 'n'},
	)
	menu = ui.NewMenu(screen, kb)
	idx, ans = menu.Choose([]string{"zero", "one", "two", "three"}, 0, "")
	if ans != "" || idx != 1 {
		t.Error("Expected 1, '', got", idx, ans)
	}

	// Search next.
	kb = terminal.NewMockKeyboard(
		[]string{"char", "ctrlN", "ctrlN", "enter"},
		[]rune{'e'},
	)
	menu = ui.NewMenu(screen, kb)
	idx, ans = menu.Choose([]string{"zero", "one", "two", "three"}, 0, "")
	if ans != "" || idx != 3 {
		t.Error("Expected 3, '', got", idx, ans)
	}

}
