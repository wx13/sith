package editor

import (
	"errors"
)

type Action struct {
	Func func()
	Name string
}

type KeyMap map[string]Action

func (editor *Editor) MakeKeyMap() KeyMap {
	km := make(KeyMap)
	km.Add("backspace", editor.file.Backspace, "Backspace")
	return km
}

func (km KeyMap) Add(key string, f func(), name string) {
	km[key] = Action{f, name}
}

func (km KeyMap) Run(key string) error {
	action, ok := km[key]
	if !ok {
		return errors.New("Unknown keypress")
	}
	action.Func()
	return nil
}
