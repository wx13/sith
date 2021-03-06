package terminal

import (
	"strings"

	"github.com/nsf/termbox-go"
)

// Keyboard acts as an interface to the termbox keyboard.
type Keyboard struct {
	KeyMap map[termbox.Key]string
}

// NewKeyboard defines a map from termbox key to a
// string representation.
func NewKeyboard() *Keyboard {
	termbox.SetInputMode(termbox.InputAlt)
	kb := Keyboard{}
	kb.KeyMap = map[termbox.Key]string{
		termbox.KeyBackspace:  "backspace",
		termbox.KeyBackspace2: "backspace",
		termbox.KeyDelete:     "delete",
		termbox.KeyArrowUp:    "arrowUp",
		termbox.KeyArrowDown:  "arrowDown",
		termbox.KeyArrowLeft:  "arrowLeft",
		termbox.KeyArrowRight: "arrowRight",
		termbox.KeySpace:      "space",
		termbox.KeyEnter:      "enter",
		termbox.KeyPgup:       "pageUp",
		termbox.KeyPgdn:       "pageDown",
		termbox.KeyHome:       "home",
		termbox.KeyEnd:        "end",
		termbox.KeyTab:        "tab",
		termbox.KeyCtrl6:      "ctrl6",
		termbox.KeyCtrlA:      "ctrlA",
		termbox.KeyCtrlB:      "ctrlB",
		termbox.KeyCtrlC:      "ctrlC",
		termbox.KeyCtrlD:      "ctrlD",
		termbox.KeyCtrlE:      "ctrlE",
		termbox.KeyCtrlF:      "ctrlF",
		termbox.KeyCtrlG:      "ctrlG",
		termbox.KeyCtrlJ:      "ctrlJ",
		termbox.KeyCtrlK:      "ctrlK",
		termbox.KeyCtrlL:      "ctrlL",
		termbox.KeyCtrlN:      "ctrlN",
		termbox.KeyCtrlO:      "ctrlO",
		termbox.KeyCtrlP:      "ctrlP",
		termbox.KeyCtrlQ:      "ctrlQ",
		termbox.KeyCtrlR:      "ctrlR",
		termbox.KeyCtrlS:      "ctrlS",
		termbox.KeyCtrlT:      "ctrlT",
		termbox.KeyCtrlU:      "ctrlU",
		termbox.KeyCtrlV:      "ctrlV",
		termbox.KeyCtrlW:      "ctrlW",
		termbox.KeyCtrlX:      "ctrlX",
		termbox.KeyCtrlY:      "ctrlY",
		termbox.KeyCtrlZ:      "ctrlZ",
		termbox.KeyCtrlSlash:  "ctrlSlash",
	}
	return &kb
}

func (kb *Keyboard) altKeyToCmd(ev termbox.Event) (string, rune) {
	return "alt" + strings.ToUpper(string(ev.Ch)), 0
}

func (kb *Keyboard) keyToCmd(ev termbox.Event) (string, rune) {

	cmd, ok := kb.KeyMap[ev.Key]

	if ok {
		return cmd, 0
	} else if (ev.Mod & termbox.ModAlt) != 0 {
		return kb.altKeyToCmd(ev)
	} else if ev.Ch > 160 && ev.Ch < 256 {
		ev.Ch -= 128
		return kb.altKeyToCmd(ev)
	} else if ev.Ch >= 32 && ev.Ch < 128 {
		return "char", ev.Ch
	} else {
		return "unknown", 0
	}
}

// GetCmdString turns termbox keyboard input into a string representation
// of the keypress. If the result is "char", then it also returns the rune.
func (kb *Keyboard) GetCmdString(ev termbox.Event) (string, rune) {

	if ev.Type == termbox.EventKey {
		return kb.keyToCmd(ev)
	}
	return "unknown", 0
}

// GetKey returns the human-readable name for a keypress,
// or the rune if it is character.
func (kb *Keyboard) GetKey() (string, rune) {
	ev := termbox.PollEvent()
	return kb.GetCmdString(ev)
}

// Mock keyboard for testing.
type MockKeyboard struct {
	keys  []string
	runes []rune
	idx   int
}

func NewMockKeyboard(keys []string, runes []rune) *MockKeyboard {
	return &MockKeyboard{
		keys:  keys,
		runes: runes,
		idx:   0,
	}
}

func (mkb *MockKeyboard) GetKey() (string, rune) {
	idx := mkb.idx
	key := "unknown"
	var r rune
	if mkb.idx < len(mkb.keys) {
		key = mkb.keys[idx]
	}
	if mkb.idx < len(mkb.runes) {
		r = mkb.runes[idx]
	}
	mkb.idx++
	return key, r
}
