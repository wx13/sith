package ui_test

import (
	"github.com/wx13/sith/terminal"
	"github.com/wx13/sith/ui"
	"testing"
)

func TestPromptAskYesNo(t *testing.T) {
	screen := MockScreen{}

	// Lowercase y
	kb := terminal.NewMockKeyboard(
		[]string{"char"},
		[]rune{'y'},
	)
	prompt := ui.MakePrompt(screen, kb)
	yes, err := prompt.AskYesNo("Well?")
	if err != nil {
		t.Error("Expected yes, got", err)
	}
	if !yes {
		t.Error("Expected yes, got no")
	}

	// Uppercase y
	kb = terminal.NewMockKeyboard(
		[]string{"char"},
		[]rune{'Y'},
	)
	prompt = ui.MakePrompt(screen, kb)
	yes, err = prompt.AskYesNo("Well?")
	if err != nil {
		t.Error("Expected yes, got", err)
	}
	if !yes {
		t.Error("Expected yes, got no")
	}

	// Cancel
	kb = terminal.NewMockKeyboard(
		[]string{"ctrlC"},
		[]rune{},
	)
	prompt = ui.MakePrompt(screen, kb)
	yes, err = prompt.AskYesNo("Well?")
	if err == nil {
		t.Error("Expected err, got nil")
	}

}
