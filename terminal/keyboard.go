package terminal

import "github.com/nsf/termbox-go"

type Keyboard struct {
	KeyMap    map[termbox.Key]string
	AltKeyMap map[string]string
}

func NewKeyboard() *Keyboard {
	termbox.SetInputMode(termbox.InputAlt)
	kb := Keyboard{}
	kb.KeyMap = map[termbox.Key]string{
		termbox.KeyBackspace:  "backspace",
		termbox.KeyBackspace2: "backspace",
		termbox.KeyDelete:     "delete",
		termbox.KeyCtrlD:      "delete",
		termbox.KeyArrowUp:    "arrowUp",
		termbox.KeyArrowDown:  "arrowDown",
		termbox.KeyArrowLeft:  "arrowLeft",
		termbox.KeyArrowRight: "arrowRight",
		termbox.KeySpace:      "space",
		termbox.KeyEnter:      "enter",
		termbox.KeyPgup:       "pageUp",
		termbox.KeyPgdn:       "pageDown",
		termbox.KeyTab:        "tab",
		termbox.KeyCtrlL:      "ctrlL",
		termbox.KeyCtrlJ:      "ctrlJ",
		termbox.KeyCtrlK:      "ctrlK",
		termbox.KeyCtrlO:      "ctrlO",
		termbox.KeyCtrlN:      "ctrlN",
		termbox.KeyCtrlB:      "ctrlB",
		termbox.KeyCtrl6:      "ctrl6",
		termbox.KeyCtrlX:      "ctrlX",
		termbox.KeyCtrlZ:      "ctrlZ",
		termbox.KeyCtrlY:      "ctrlY",
		termbox.KeyCtrlS:      "ctrlS",
		termbox.KeyCtrlA:      "ctrlA",
		termbox.KeyCtrlE:      "ctrlE",
		termbox.KeyCtrlW:      "ctrlW",
		termbox.KeyCtrlQ:      "ctrlQ",
		termbox.KeyCtrlC:      "ctrlC",
		termbox.KeyCtrlV:      "ctrlV",
		termbox.KeyCtrlF:      "ctrlF",
		termbox.KeyCtrlP:      "ctrlP",
		termbox.KeyCtrlU:      "ctrlU",
	}
	kb.AltKeyMap = map[string]string{
		"q": "altQ",
		"x": "altX",
		"w": "altW",
		"n": "altN",
		"b": "altB",
		"6": "alt6",
		"c": "altC",
		"f": "altF",
		"p": "altP",
		"u": "altU",
		"l": "altL",
		"o": "altO",
		"m": "altM",
		"g": "altG",
		"j": "altJ",
		"i": "altI",
		"z": "altZ",
	}
	return &kb
}

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
