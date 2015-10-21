package syntaxcolor

import "regexp"
import "github.com/nsf/termbox-go"
import "path"
import "fmt"

type Color struct {
	fg, bg termbox.Attribute
}

// SyntaxRule simply maps a regexp to a color.
type SyntaxRule struct {
	re    *regexp.Regexp
	color Color
}

type SyntaxRules []SyntaxRule

func NewSyntaxRules(filename string) *SyntaxRules {

	rules := SyntaxRules{}

	filetype := rules.getFileType(filename)

	// Filetype specific rules
	switch filetype {
	case "sh", "py", "yaml", "js", "c", "coffee", "html", "rb":
		rules.addSingleQuoteRule(termbox.ColorYellow)
		rules.addDoubleQuoteRule(termbox.ColorYellow)
	case "go":
		rules.addSingleQuoteRule(termbox.ColorRed)
		rules.addDoubleQuoteRule(termbox.ColorYellow)
	}
	switch filetype {
	case "go":
		rules.addLineCommentRule("//", termbox.ColorCyan)
	case "rb", "sh", "py", "yaml", "coffee", "git":
		rules.addLineCommentRule("#", termbox.ColorCyan)
	case "c":
		rules.addLineCommentRule("//", termbox.ColorCyan)
	case "md":
		rules.addRule("^#+.*$", Color{fg: termbox.ColorGreen})
		rules.addRule("^===*$", Color{fg: termbox.ColorGreen})
		rules.addRule("^---*$", Color{fg: termbox.ColorGreen})
	case "js":
		rules.addLineCommentRule("//", termbox.ColorCyan)
	case "html":
		rules.addRule("<!--.*?-->", Color{fg: termbox.ColorCyan})
	}

	// Trailing whitespace
	rules.addRule("[ \t]+$", Color{bg: termbox.ColorYellow})

	return &rules

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

func (rules SyntaxRules) getFileType(filename string) string {
	ext := path.Ext(filename)
	if len(ext) == 0 {
		basename := path.Base(filename)
		if basename == "COMMIT_EDITMSG" {
			return "git"
		}
		return ""
	}
	ext = ext[1:]
	switch ext {
	case ".c", ".C", ".cpp", ".c++":
		return "c"
	case ".sh", ".csh":
		return "sh"
	case "":
		return ""
	default:
		return ext
	}
}

type LineColor struct {
	Fg, Bg     termbox.Attribute
	Start, End int
}

// Colorize takes in a string and outputs an array of LineColor objects.
// A linecolor object maps a color pair (bg, fg) with start/end indices
// within the string.
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
