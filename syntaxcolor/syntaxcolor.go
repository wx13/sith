package syntaxcolor

import (
	"fmt"
	"regexp"

	"github.com/nsf/termbox-go"
	"github.com/wx13/sith/config"
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
func NewSyntaxRules(cfg config.Config) *SyntaxRules {
	rules := SyntaxRules{}
	rules.ingestConfig(cfg)
	rules.addWhitespaceRule()
	return &rules
}

func (rules *SyntaxRules) ingestConfig(cfg config.Config) {
	for pattern, color := range cfg.SyntaxRules {
		rules.addRule(pattern, Color{
			fg: toColor(color.FG),
			bg: toColor(color.BG),
		})
	}
}

func toColor(name string) termbox.Attribute {
	colormap := map[string]termbox.Attribute{
		"green":   termbox.ColorGreen,
		"red":     termbox.ColorRed,
		"blue":    termbox.ColorBlue,
		"cyan":    termbox.ColorCyan,
		"magenta": termbox.ColorMagenta,
		"yellow":  termbox.ColorYellow,
		"white":   termbox.ColorWhite,
		"black":   termbox.ColorBlack,
		"default": termbox.ColorDefault,
	}
	color, ok := colormap[name]
	if ok {
		return color
	}
	return termbox.ColorDefault
}

func (rules *SyntaxRules) addWhitespaceRule() {
	rules.whitespace = regexp.MustCompile("[ \t]+$")
}

func (rules *SyntaxRules) addRule(reStr string, color Color) {
	re, _ := regexp.Compile(reStr)
	rules.list = append(rules.list, SyntaxRule{re, color})
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
