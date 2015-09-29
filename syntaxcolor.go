package main

import "regexp"
import "github.com/nsf/termbox-go"
import "path"
import "fmt"

type Color struct {
	fg, bg termbox.Attribute
}

type SyntaxRule struct {
	re    *regexp.Regexp
	color Color
}

type SyntaxRules []SyntaxRule

func NewSyntaxRules(filename string) *SyntaxRules {

	rules := SyntaxRules{}

	filetype := rules.GetFileType(filename)

	// Filetype specific rules
	switch filetype {
	case "go":
		rules.AddSingleQuoteRule(termbox.ColorRed)
		rules.AddDoubleQuoteRule(termbox.ColorYellow)
		rules.AddLineCommentRule("//", termbox.ColorCyan)
	case "python":
		rules.AddSingleQuoteRule(termbox.ColorYellow)
		rules.AddDoubleQuoteRule(termbox.ColorYellow)
		rules.AddLineCommentRule("#", termbox.ColorCyan)
	case "ruby":
		rules.AddSingleQuoteRule(termbox.ColorYellow)
		rules.AddDoubleQuoteRule(termbox.ColorYellow)
		rules.AddLineCommentRule("#", termbox.ColorCyan)
	case "c":
		rules.AddSingleQuoteRule(termbox.ColorYellow)
		rules.AddDoubleQuoteRule(termbox.ColorYellow)
		rules.AddLineCommentRule("//", termbox.ColorCyan)
	case "markdown":
		rules.AddRule("^#+.*$", Color{fg: termbox.ColorGreen})
		rules.AddRule("^===*$", Color{fg: termbox.ColorGreen})
		rules.AddRule("^---*$", Color{fg: termbox.ColorGreen})
	}

	// Trailing whitespace
	rules.AddRule("[ \t]+$", Color{bg: termbox.ColorYellow})

	return &rules

}

func (rules *SyntaxRules) AddRule(reStr string, color Color) {
	re, _ := regexp.Compile(reStr)
	*rules = append(*rules, SyntaxRule{re, color})
}

func (rules *SyntaxRules) AddLineCommentRule(commStr string, fg termbox.Attribute) {
	reStr := fmt.Sprintf("%s.*$", commStr)
	rules.AddRule(reStr, Color{fg: fg})
}

func (rules *SyntaxRules) AddSingleQuoteRule(fg termbox.Attribute) {
	rules.AddRule("'.*?'", Color{fg: fg})
}

func (rules *SyntaxRules) AddDoubleQuoteRule(fg termbox.Attribute) {
	rules.AddRule("\".*?\"", Color{fg: fg})
}

func (rules SyntaxRules) GetFileType(filename string) string {
	switch path.Ext(filename){
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".c", ".C", ".cpp", ".c++":
		return "c"
	case ".rb":
		return "ruby"
	case ".md":
		return "markdown"
	default:
		return ""
	}
}

type LineColor struct {
	fg, bg termbox.Attribute
	start, end int
}

func (rules SyntaxRules) Colorize(str string) []LineColor {
	lc := []LineColor{}
	for _, rule := range rules {
		matches := rule.re.FindAllStringIndex(str, -1)
		if matches == nil {
			continue
		}
		for _, startEnd := range matches {
			lc = append(lc, LineColor{fg: rule.color.fg, bg: rule.color.bg, start: startEnd[0], end: startEnd[1]})
		}
	}
	return lc
}




