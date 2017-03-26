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
type SyntaxRules []SyntaxRule

// NewSyntaxRules creates a new SyntaxRules object, and initializes
// the syntax rule sets.
func NewSyntaxRules(filename string) *SyntaxRules {
	rules := SyntaxRules{}
	filetype := rules.GetFileType(filename)
	rules.addQuoteRules(filetype)
	rules.addCommentRules(filetype)
	rules.addMiscRules(filetype)
	rules.addWhitespaceRules(filetype)
	return &rules
}

func (rules *SyntaxRules) addQuoteRules(filetype string) {
	switch filetype {
	case "sh", "py", "yaml", "js", "c", "coffee", "html", "rb":
		rules.addSingleQuoteRule(termbox.ColorYellow)
		rules.addDoubleQuoteRule(termbox.ColorYellow)
	case "go":
		rules.addSingleQuoteRule(termbox.ColorRed)
		rules.addDoubleQuoteRule(termbox.ColorYellow)
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

func (rules *SyntaxRules) addWhitespaceRules(filetype string) {
	rules.addRule("[ \t]+$", Color{bg: termbox.ColorYellow})
}

func (rules *SyntaxRules) addRule(reStr string, color Color) {
	re, _ := regexp.Compile(reStr)
	*rules = append(*rules, SyntaxRule{re, color})
}

func (rules *SyntaxRules) addLineCommentRule(commStr string, fg termbox.Attribute) {
	reStr := fmt.Sprintf("%s.*$", commStr)
	rules.addRule(reStr, Color{fg: fg})
}

func (rules *SyntaxRules) addSingleQuoteRule(fg termbox.Attribute) {
	rules.addRule("'.*?'", Color{fg: fg})
}

func (rules *SyntaxRules) addDoubleQuoteRule(fg termbox.Attribute) {
	rules.addRule("\".*?\"", Color{fg: fg})
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

// Colorize takes in a string and outputs an array of LineColor objects.
func (rules SyntaxRules) Colorize(str string) []LineColor {
	lc := []LineColor{}
	for _, rule := range rules {
		matches := rule.re.FindAllStringIndex(str, -1)
		if matches == nil {
			continue
		}
		for _, startEnd := range matches {
			lc = append(lc, LineColor{Fg: rule.color.fg, Bg: rule.color.bg, Start: startEnd[0], End: startEnd[1]})
		}
	}
	return lc
}
