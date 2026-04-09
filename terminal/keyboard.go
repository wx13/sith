package terminal

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Keyboard acts as an interface to the tcell keyboard.
type Keyboard struct {
	KeyMap map[tcell.Key]string
	screen tcell.Screen
}

// NewKeyboard defines a map from tcell key to a
// string representation.
func NewKeyboard() *Keyboard {
	kb := Keyboard{}
	kb.KeyMap = map[tcell.Key]string{
		tcell.KeyBackspace:  "backspace",
		tcell.KeyBackspace2: "backspace",
		tcell.KeyDelete:     "delete",
		tcell.KeyUp:         "arrowUp",
		tcell.KeyDown:       "arrowDown",
		tcell.KeyLeft:       "arrowLeft",
		tcell.KeyRight:      "arrowRight",
		tcell.KeyEnter:      "enter",
		tcell.KeyPgUp:       "pageUp",
		tcell.KeyPgDn:       "pageDown",
		tcell.KeyHome:       "home",
		tcell.KeyEnd:        "end",
		tcell.KeyTab:        "tab",
		tcell.KeyCtrlA:      "ctrlA",
		tcell.KeyCtrlB:      "ctrlB",
		tcell.KeyCtrlC:      "ctrlC",
		tcell.KeyCtrlD:      "ctrlD",
		tcell.KeyCtrlE:      "ctrlE",
		tcell.KeyCtrlF:      "ctrlF",
		tcell.KeyCtrlG:      "ctrlG",
		tcell.KeyCtrlJ:      "ctrlJ",
		tcell.KeyCtrlK:      "ctrlK",
		tcell.KeyCtrlL:      "ctrlL",
		tcell.KeyCtrlN:      "ctrlN",
		tcell.KeyCtrlO:      "ctrlO",
		tcell.KeyCtrlP:      "ctrlP",
		tcell.KeyCtrlQ:      "ctrlQ",
		tcell.KeyCtrlR:      "ctrlR",
		tcell.KeyCtrlS:      "ctrlS",
		tcell.KeyCtrlT:      "ctrlT",
		tcell.KeyCtrlU:      "ctrlU",
		tcell.KeyCtrlV:      "ctrlV",
		tcell.KeyCtrlW:      "ctrlW",
		tcell.KeyCtrlX:      "ctrlX",
		tcell.KeyCtrlY:      "ctrlY",
		tcell.KeyCtrlZ:      "ctrlZ",
		tcell.KeyCtrlBackslash: "ctrlSlash",
	}
	return &kb
}

// SetScreen sets the tcell screen for polling events.
func (kb *Keyboard) SetScreen(screen tcell.Screen) {
	kb.screen = screen
}

func (kb *Keyboard) altKeyToCmd(r rune) (string, rune) {
	return "alt" + strings.ToUpper(string(r)), 0
}

func (kb *Keyboard) keyToCmd(ev *tcell.EventKey) (string, rune) {
	key := ev.Key()
	r := ev.Rune()
	mod := ev.Modifiers()

	cmd, ok := kb.KeyMap[key]
	if ok {
		return cmd, 0
	}

	if mod&tcell.ModAlt != 0 {
		return kb.altKeyToCmd(r)
	}

	// Handle Ctrl+6 (ctrl6) - tcell represents this differently
	if key == tcell.KeyCtrlCarat {
		return "ctrl6", 0
	}

	// Space is a rune in tcell, not a special key
	if key == tcell.KeyRune && r == ' ' {
		return "space", 0
	}

	if key == tcell.KeyRune && r >= 32 && r < 128 {
		return "char", r
	}

	// Handle extended characters (alt+char in some terminals)
	if key == tcell.KeyRune && r > 160 && r < 256 {
		return kb.altKeyToCmd(r - 128)
	}

	if key == tcell.KeyRune {
		return "char", r
	}

	return "unknown", 0
}

// GetCmdString turns tcell keyboard input into a string representation
// of the keypress. If the result is "char", then it also returns the rune.
func (kb *Keyboard) GetCmdString(ev *tcell.EventKey) (string, rune) {
	return kb.keyToCmd(ev)
}

// GetKey returns the human-readable name for a keypress,
// or the rune if it is character.
func (kb *Keyboard) GetKey() (string, rune) {
	for {
		ev := kb.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			return kb.GetCmdString(ev)
		case *tcell.EventResize:
			kb.screen.Sync()
		}
	}
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
