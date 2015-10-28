package terminal

import "github.com/nsf/termbox-go"

type Keyboard struct {
	KeyMap    map[termbox.Key]string
	AltKeyMap map[string]string
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
		termbox.KeyCtrlU:      "ctrlU",
		termbox.KeyCtrlV:      "ctrlV",
		termbox.KeyCtrlW:      "ctrlW",
		termbox.KeyCtrlX:      "ctrlX",
		termbox.KeyCtrlY:      "ctrlY",
		termbox.KeyCtrlZ:      "ctrlZ",
	}
	kb.AltKeyMap = map[string]string{
		"6": "alt6",
		"b": "altB",
		"c": "altC",
		"f": "altF",
		"g": "altG",
		"i": "altI",
		"j": "altJ",
		"l": "altL",
		"m": "altM",
		"n": "altN",
		"o": "altO",
		"p": "altP",
		"q": "altQ",
		"t": "altT",
		"u": "altU",
		"v": "altV",
		"w": "altW",
		"x": "altX",
		"z": "altZ",
	}
	return &kb
}

// GetCmdString turns termbox keyboard input into a string representation
// of the keypress.  If the result is "char", then it also returns the rune.
func (kb *Keyboard) GetCmdString(ev termbox.Event) (string, rune) {
	// handle a keypress event
	if ev.Type == termbox.EventKey {

		// Convert termbox.Key to string description of key.
		cmd, ok := kb.KeyMap[ev.Key]
		if ok {
			return cmd, 0
		} else if (ev.Mod & termbox.ModAlt) != 0 {
			// Handle alt/esc sequences
			cmd, ok := kb.AltKeyMap[string(ev.Ch)]
			if ok {
				return cmd, 0
			} else {
				return "unknown", 0
			}
		} else if ev.Ch > 160 && ev.Ch < 256 {
			// Allow for alternate alt keys which are off by 128
			// on some computers.
			cmd, ok := kb.AltKeyMap[string(ev.Ch-128)]
			if ok {
				return cmd, 0
			} else {
				return "unknown", 0
			}
		} else if ev.Ch >= 32 && ev.Ch < 128 {
			// Handle regular characters.
			return "char", ev.Ch
		} else {
			return "unknown", 0
		}

	} else {
		return "unknown", 0
	}
}

func (kb *Keyboard) GetKey() (string, rune) {
	ev := termbox.PollEvent()
	return kb.GetCmdString(ev)
}
