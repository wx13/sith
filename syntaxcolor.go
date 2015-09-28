package main

import "regexp"
import "github.com/nsf/termbox-go"

type Color struct {
	fg, bg termbox.Attribute
}

type SyntaxRules map[*regexp.Regexp]Color

func MakeSyntaxRules(filetype string) SyntaxRules {

	rules := make(SyntaxRules)

	// Trailing whitespace
	re, _ := regexp.Compile("[ \t]*$")
	rules[re] = Color{bg: termbox.ColorYellow}

	// Filetype specific rules
	switch filetype {
	case "go":
	case "python":
	case "ruby":
	case "c":
	case "markdown":
	}

	return rules

}

type LineColor struct {
	fg, bg termbox.Attribute
	start, end int
}

func (rules SyntaxRules) Colorize(str string) []LineColor {
	lc := []LineColor{}
	//lc = append(lc, LineColor{fg:termbox.ColorBlue, bg:termbox.ColorYellow, start: 3, end: 6})
	for re, color := range rules {
		startEnd := re.FindStringIndex(str)
		lc = append(lc, LineColor{fg: color.fg, bg: color.bg, start: startEnd[0], end: startEnd[1]})
	}
	return lc
}




