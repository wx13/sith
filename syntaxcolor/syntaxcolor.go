package syntaxcolor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/wx13/sith/config"
)

// LineState represents the syntax state at the start/end of a line.
type LineState int

const (
	StateNormal LineState = iota
	StateBlockComment
	StateString
	StateRawString
)

// Color specifies a foreground and background color.
type Color struct {
	fg, bg tcell.Color
}

// SyntaxRule simply maps a regexp to a color.
type SyntaxRule struct {
	re    *regexp.Regexp
	color Color
}

// MultilineRule defines a multiline construct like block comments.
type MultilineRule struct {
	start *regexp.Regexp
	end   *regexp.Regexp
	color Color
	state LineState
}

// SyntaxRules is the full collection of all syntax rules.
type SyntaxRules struct {
	list       []SyntaxRule
	clobber    []SyntaxRule
	multiline  []MultilineRule
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
		if color.Clobber {
			rules.addClobberRule(pattern, Color{
				fg: toColor(color.FG),
				bg: toColor(color.BG),
			})
		} else {
			rules.addRule(pattern, Color{
				fg: toColor(color.FG),
				bg: toColor(color.BG),
			})
		}
	}
}

// AddBlockComment adds a block comment multiline rule.
func (rules *SyntaxRules) AddBlockComment(start, end string, fg, bg tcell.Color) {
	rules.multiline = append(rules.multiline, MultilineRule{
		start: regexp.MustCompile(regexp.QuoteMeta(start)),
		end:   regexp.MustCompile(regexp.QuoteMeta(end)),
		color: Color{fg: fg, bg: bg},
		state: StateBlockComment,
	})
}

// AddMultilineString adds a multiline string rule.
func (rules *SyntaxRules) AddMultilineString(delim string, fg, bg tcell.Color, state LineState) {
	rules.multiline = append(rules.multiline, MultilineRule{
		start: regexp.MustCompile(regexp.QuoteMeta(delim)),
		end:   regexp.MustCompile(regexp.QuoteMeta(delim)),
		color: Color{fg: fg, bg: bg},
		state: state,
	})
}

func toColor(name string) tcell.Color {
	colormap := map[string]tcell.Color{
		"green":   tcell.ColorGreen,
		"red":     tcell.ColorRed,
		"blue":    tcell.ColorBlue,
		"cyan":    tcell.ColorTeal,
		"magenta": tcell.ColorPurple,
		"yellow":  tcell.ColorYellow,
		"white":   tcell.ColorWhite,
		"black":   tcell.ColorBlack,
		"default": tcell.ColorDefault,
	}
	color, ok := colormap[name]
	if ok {
		return color
	}
	return tcell.ColorDefault
}

func (rules *SyntaxRules) addWhitespaceRule() {
	rules.whitespace = regexp.MustCompile("[ \t]+$")
}

func (rules *SyntaxRules) addRule(reStr string, color Color) {
	re, _ := regexp.Compile(reStr)
	rules.list = append(rules.list, SyntaxRule{re, color})
}

func (rules *SyntaxRules) addClobberRule(reStr string, color Color) {
	re, _ := regexp.Compile(reStr)
	rules.clobber = append(rules.clobber, SyntaxRule{re, color})
}

// LineColor object maps a color pair (bg, fg) with start/end indices
// within a string.
type LineColor struct {
	Fg, Bg     tcell.Color
	Start, End int
}

// ColorResult contains the colorization output plus the end state.
type ColorResult struct {
	Colors   []LineColor
	EndState LineState
}

// getMultilineRule returns the multiline rule for the given state.
func (rules *SyntaxRules) getMultilineRule(state LineState) *MultilineRule {
	for i := range rules.multiline {
		if rules.multiline[i].state == state {
			return &rules.multiline[i]
		}
	}
	return nil
}

// nextMatch finds the next match across all the rules.
func (rules SyntaxRules) nextMatch(str string, idx int) (LineColor, error) {
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

// ColorizeWithState colorizes a line taking into account the start state.
// Returns the colors and the end state for this line.
func (rules SyntaxRules) ColorizeWithState(str string, startState LineState) ColorResult {
	result := ColorResult{
		Colors:   []LineColor{},
		EndState: startState,
	}

	idx := 0

	// If we're starting inside a multiline construct, find where it ends
	if startState != StateNormal {
		rule := rules.getMultilineRule(startState)
		if rule != nil {
			match := rule.end.FindStringIndex(str)
			if match == nil {
				// Entire line is in the multiline construct
				result.Colors = append(result.Colors, LineColor{
					Fg:    rule.color.fg,
					Bg:    rule.color.bg,
					Start: 0,
					End:   len(str),
				})
				result.EndState = startState
				return result
			}
			// Color up to and including the end delimiter
			result.Colors = append(result.Colors, LineColor{
				Fg:    rule.color.fg,
				Bg:    rule.color.bg,
				Start: 0,
				End:   match[1],
			})
			idx = match[1]
			result.EndState = StateNormal
		}
	}

	// Process the rest of the line in normal state
	for idx < len(str) {
		// Check for multiline construct starts
		foundMultiline := false
		earliestStart := len(str) + 1
		var earliestRule *MultilineRule

		for i := range rules.multiline {
			rule := &rules.multiline[i]
			match := rule.start.FindStringIndex(str[idx:])
			if match != nil && idx+match[0] < earliestStart {
				earliestStart = idx + match[0]
				earliestRule = rule
			}
		}

		// First, colorize any single-line patterns before the multiline start
		for idx < earliestStart && idx < len(str) {
			lc, err := rules.nextMatch(str, idx)
			if err != nil {
				break
			}
			if earliestRule != nil && lc.Start >= earliestStart {
				break
			}
			if earliestRule != nil && lc.End > earliestStart {
				lc.End = earliestStart
			}
			result.Colors = append(result.Colors, lc)
			idx = lc.End
		}

		if earliestRule != nil && earliestStart < len(str) {
			foundMultiline = true
			// Found a multiline start
			startIdx := earliestStart
			startLen := len(earliestRule.start.FindString(str[startIdx:]))

			// Look for the end on this same line
			endMatch := earliestRule.end.FindStringIndex(str[startIdx+startLen:])
			if endMatch != nil {
				// Multiline construct ends on same line
				endIdx := startIdx + startLen + endMatch[1]
				result.Colors = append(result.Colors, LineColor{
					Fg:    earliestRule.color.fg,
					Bg:    earliestRule.color.bg,
					Start: startIdx,
					End:   endIdx,
				})
				idx = endIdx
			} else {
				// Multiline construct continues to next line
				result.Colors = append(result.Colors, LineColor{
					Fg:    earliestRule.color.fg,
					Bg:    earliestRule.color.bg,
					Start: startIdx,
					End:   len(str),
				})
				result.EndState = earliestRule.state
				return result
			}
		}

		if !foundMultiline {
			// No more multiline constructs, finish with single-line rules
			for {
				lc, err := rules.nextMatch(str, idx)
				if err != nil {
					break
				}
				result.Colors = append(result.Colors, lc)
				idx = lc.End
			}
			break
		}
	}

	// Apply clobber rules
	for _, rule := range rules.clobber {
		match := rule.re.FindStringIndex(str)
		if match != nil {
			lc := LineColor{
				Bg:    rule.color.bg,
				Fg:    rule.color.fg,
				Start: match[0],
				End:   match[1],
			}
			result.Colors = append(result.Colors, lc)
		}
	}

	return result
}

// Colorize takes in a string and outputs an array of LineColor objects.
// This is the legacy method that doesn't track state (for backwards compatibility).
func (rules SyntaxRules) Colorize(str string) []LineColor {
	return rules.ColorizeWithState(str, StateNormal).Colors
}

// StateCache manages line state caching for a file.
type StateCache struct {
	states []LineState
}

// NewStateCache creates a new state cache.
func NewStateCache() *StateCache {
	return &StateCache{
		states: []LineState{},
	}
}

// GetState returns the start state for a given line.
// Line 0 always starts in StateNormal.
func (sc *StateCache) GetState(lineNum int) LineState {
	if lineNum <= 0 {
		return StateNormal
	}
	if lineNum-1 < len(sc.states) {
		return sc.states[lineNum-1]
	}
	return StateNormal
}

// SetEndState sets the end state for a given line.
// Returns true if the state changed (meaning subsequent lines need recalculation).
func (sc *StateCache) SetEndState(lineNum int, state LineState) bool {
	// Grow the slice if needed
	for len(sc.states) <= lineNum {
		sc.states = append(sc.states, StateNormal)
	}

	oldState := sc.states[lineNum]
	sc.states[lineNum] = state
	return oldState != state
}

// Invalidate marks all states from lineNum onwards as needing recalculation.
func (sc *StateCache) Invalidate(lineNum int) {
	if lineNum < len(sc.states) {
		sc.states = sc.states[:lineNum]
	}
}

// Clear resets the entire cache.
func (sc *StateCache) Clear() {
	sc.states = sc.states[:0]
}

// SetupForLanguage configures multiline rules based on file extension.
func (rules *SyntaxRules) SetupForLanguage(ext string) {
	ext = strings.ToLower(ext)

	// C-style languages: block comments
	cStyleLangs := map[string]bool{
		"c": true, "h": true, "cc": true, "cpp": true, "c++": true,
		"go": true, "java": true, "js": true, "ts": true, "tsx": true,
		"jsx": true, "rs": true, "swift": true, "kt": true, "scala": true,
		"cs": true, "php": true, "css": true, "scss": true, "less": true,
	}

	if cStyleLangs[ext] {
		rules.AddBlockComment("/*", "*/", tcell.ColorTeal, tcell.ColorDefault)
	}

	// Python/Ruby: multiline strings with triple quotes
	if ext == "py" {
		rules.AddMultilineString(`"""`, tcell.ColorYellow, tcell.ColorDefault, StateString)
		rules.AddMultilineString(`'''`, tcell.ColorYellow, tcell.ColorDefault, StateRawString)
	}

	// Go: raw strings with backticks
	if ext == "go" {
		rules.AddMultilineString("`", tcell.ColorYellow, tcell.ColorDefault, StateRawString)
	}

	// HTML/XML: comments
	if ext == "html" || ext == "htm" || ext == "xml" || ext == "svg" {
		rules.AddBlockComment("<!--", "-->", tcell.ColorTeal, tcell.ColorDefault)
	}
}
