package syntaxcolor

import (
	"fmt"
	"path"
	"regexp"

	"github.com/nsf/termbox-go"
)

// Color specifies a foreground and background color.
type Color struct {
	fg, bg termbox.Attribute
}

// SyntaxRule simply maps a regexp to a color.
type SyntaxRule struct {
	re    *regexp.Regexp
	color Color
}

// SyntaxRules is the full collection of all syntax rules.
type SyntaxRules struct {
	list       []SyntaxRule
	whitespace *regexp.Regexp
}

// NewSyntaxRules creates a new SyntaxRules object, and initializes
// the syntax rule sets.
func NewSyntaxRules(filename string) *SyntaxRules {
	rules := SyntaxRules{}
	filetype := rules.GetFileType(filename)
	rules.addQuoteRules(filetype)
	rules.addCommentRules(filetype)
	rules.addMiscRules(filetype)
	rules.addWhitespaceRule()
	return &rules
}

func (rules *SyntaxRules) addWhitespaceRule() {
	rules.whitespace = regexp.MustCompile("[ \t]+$")
}

func (rules *SyntaxRules) addQuoteRules(filetype string) {
	switch filetype {
	case "sh", "py", "yaml", "js", "c", "coffee", "html", "rb":
		rules.addSingleQuoteRule(termbox.ColorYellow)
		rules.addDoubleQuoteRule(termbox.ColorYellow)
	case "go":
		rules.addSingleQuoteRule(termbox.ColorRed)
		rules.addDoubleQuoteRule(termbox.ColorYellow)
		rules.addBackticQuoteRule(termbox.ColorYellow)
	}
}

func (rules *SyntaxRules) addCommentRules(filetype string) {
	switch filetype {
	case "go":
		rules.addLineCommentRule("//", termbox.ColorCyan)
	case "rb", "sh", "py", "yaml", "coffee", "git":
		rules.addLineCommentRule("#", termbox.ColorCyan)
	case "c":
		rules.addLineCommentRule("//", termbox.ColorCyan)
		rules.addRule(`/\*.*?\*/`, Color{fg: termbox.ColorCyan})
	case "js":
		rules.addLineCommentRule("//", termbox.ColorCyan)
	case "html":
		rules.addRule("<!--.*?-->", Color{fg: termbox.ColorCyan})
	}
}

func (rules *SyntaxRules) addMiscRules(filetype string) {
	switch filetype {
	case "md":
		rules.addRule("^#+.*$", Color{fg: termbox.ColorGreen})
		rules.addRule("^===*$", Color{fg: termbox.ColorGreen})
		rules.addRule("^---*$", Color{fg: termbox.ColorGreen})
	}
}

func (rules *SyntaxRules) addRule(reStr string, color Color) {
	re, _ := regexp.Compile(reStr)
	rules.list = append(rules.list, SyntaxRule{re, color})
}

func (rules *SyntaxRules) addLineCommentRule(commStr string, fg termbox.Attribute) {
	reStr := fmt.Sprintf("%s.*$", commStr)
	rules.addRule(reStr, Color{fg: fg})
}

func (rules *SyntaxRules) addBackticQuoteRule(fg termbox.Attribute) {
	rules.addRule("`.*?`", Color{fg: fg})
}

func (rules *SyntaxRules) addSingleQuoteRule(fg termbox.Attribute) {
	rules.addRule("'.*?'", Color{fg: fg})
}

func (rules *SyntaxRules) addDoubleQuoteRule(fg termbox.Attribute) {
	rules.addRule(`".*?"`, Color{fg: fg})
}

// GetFileType maps common extensions onto a type (e.g. c, C, cpp, etc
// onto 'c'), but otherwise just returns the extension.
func (rules SyntaxRules) GetFileType(filename string) string {
	ext := rules.getFileExt(filename)
	switch ext {
	case "c", "C", "cpp", "c++", "h", "hpp", "cc", "hh":
		return "c"
	case "sh", "csh":
		return "sh"
	case "":
		return ""
	default:
		return ext
	}
}

func (rules SyntaxRules) getFileExt(filename string) string {
	ext := path.Ext(filename)
	if len(ext) == 0 {
		basename := path.Base(filename)
		if basename == "COMMIT_EDITMSG" {
			return "git"
		}
		return ""
	}
	ext = ext[1:]
	return ext
}

// LineColor object maps a color pair (bg, fg) with start/end indices
// within a string.
type LineColor struct {
	Fg, Bg     termbox.Attribute
	Start, End int
}

// NextMatch finds the next match accross all the rules.
func (rules SyntaxRules) NextMatch(str string, idx int) (LineColor, error) {
	subStr := str[idx:]
	i0 := len(subStr) + 1
	lc := LineColor{}
	for _, rule := range rules.list {
		match := rule.re.FindStringIndex(subStr)
		if match == nil {
			continue
		}
		if match[0] < i0 {
			i0 = match[0]
			lc.Fg, lc.Bg = rule.color.fg, rule.color.bg
			lc.Start, lc.End = idx+match[0], idx+match[1]
		}
	}
	if i0 < len(subStr)+1 {
		return lc, nil
	}
	return lc, fmt.Errorf("no match")
}

// Colorize takes in a string and outputs an array of LineColor objects.
func (rules SyntaxRules) Colorize(str string) []LineColor {
	lcs := []LineColor{}
	idx := 0
	for {
		lc, err := rules.NextMatch(str, idx)
		if err != nil {
			break
		}
		lcs = append(lcs, lc)
		idx = lc.End
	}
	match := rules.whitespace.FindStringIndex(str)
	if match != nil {
		lc := LineColor{}
		lc.Bg = termbox.ColorYellow
		lc.Start, lc.End = match[0], match[1]
		lcs = append(lcs, lc)
	}
	return lcs
}

func (rules *SyntaxRules) addWhitespaceRules(filetype string) {
	rules.addRule("[ \t]+$", Color{bg: termbox.ColorYellow})
}
