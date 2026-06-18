package syntaxcolor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/wx13/sith/config"
)

// LineState represents the syntax state at the start/end of a line.
// For embedded languages (like code blocks in markdown), the state encodes
// both the block type and the embedded language.
type LineState int

const (
	StateNormal LineState = iota
	StateBlockComment
	StateString
	StateRawString
	StateBlockEquation
	StateCodeBlockBase // Code blocks use StateCodeBlockBase + language index
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

	// For markdown embedded language support
	isMarkdown      bool
	codeBlockStart  *regexp.Regexp          // matches ```lang or ```{lang}
	codeBlockEnd    *regexp.Regexp          // matches ```
	equationDelim   *regexp.Regexp          // matches $$
	fullConfig      config.Config           // full config for looking up embedded languages
	embeddedRules   map[string]*SyntaxRules // cached syntax rules for embedded languages
	languageIndex   map[string]int          // maps language name to state offset
	indexToLanguage map[int]string          // reverse mapping
}

// NewSyntaxRules creates a new SyntaxRules object, and initializes
// the syntax rule sets.
func NewSyntaxRules(cfg config.Config) *SyntaxRules {
	rules := SyntaxRules{
		embeddedRules:   make(map[string]*SyntaxRules),
		languageIndex:   make(map[string]int),
		indexToLanguage: make(map[int]string),
	}
	rules.ingestConfig(cfg)
	rules.addWhitespaceRule()
	return &rules
}

// NewSyntaxRulesWithFullConfig creates a SyntaxRules object that has access
// to the full config for looking up embedded language rules.
func NewSyntaxRulesWithFullConfig(cfg config.Config, fullConfig config.Config) *SyntaxRules {
	rules := NewSyntaxRules(cfg)
	rules.fullConfig = fullConfig
	return rules
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
func (rules *SyntaxRules) nextMatch(str string, idx int) (LineColor, error) {
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
func (rules *SyntaxRules) ColorizeWithState(str string, startState LineState) ColorResult {
	// Handle markdown embedded code blocks specially
	if rules.isMarkdown {
		return rules.colorizeMarkdown(str, startState)
	}

	return rules.colorizeNormal(str, startState)
}

// colorizeMarkdown handles syntax highlighting for markdown with embedded code blocks.
func (rules *SyntaxRules) colorizeMarkdown(str string, startState LineState) ColorResult {
	result := ColorResult{
		Colors:   []LineColor{},
		EndState: startState,
	}

	// Check if we're inside a code block
	if startState >= StateCodeBlockBase {
		lang := rules.getLanguageFromState(startState)

		// Check if this line ends the code block
		if rules.codeBlockEnd.MatchString(str) {
			// Color the closing ``` in a distinct color
			result.Colors = append(result.Colors, LineColor{
				Fg:    tcell.ColorBlue,
				Bg:    tcell.ColorDefault,
				Start: 0,
				End:   len(str),
			})
			result.EndState = StateNormal
			return result
		}

		// Apply the embedded language's syntax rules
		embeddedRules := rules.getEmbeddedRules(lang)
		if embeddedRules != nil {
			embeddedResult := embeddedRules.colorizeNormal(str, StateNormal)
			result.Colors = embeddedResult.Colors
		}
		result.EndState = startState
		return result
	}

	// Check if we're inside a block equation
	if startState == StateBlockEquation {
		// Check if this line ends the equation
		if rules.equationDelim.MatchString(str) {
			result.Colors = append(result.Colors, LineColor{
				Fg:    tcell.ColorPurple,
				Bg:    tcell.ColorDefault,
				Start: 0,
				End:   len(str),
			})
			result.EndState = StateNormal
			return result
		}
		// Color the entire line as equation
		result.Colors = append(result.Colors, LineColor{
			Fg:    tcell.ColorPurple,
			Bg:    tcell.ColorDefault,
			Start: 0,
			End:   len(str),
		})
		result.EndState = StateBlockEquation
		return result
	}

	// Normal markdown state - check for code block or equation start
	if match := rules.codeBlockStart.FindStringSubmatch(str); match != nil {
		lang := strings.ToLower(match[1])
		// Color the opening line (```python) in a distinct color
		result.Colors = append(result.Colors, LineColor{
			Fg:    tcell.ColorBlue,
			Bg:    tcell.ColorDefault,
			Start: 0,
			End:   len(str),
		})
		result.EndState = rules.getStateForLanguage(lang)
		return result
	}

	// Check for block equation start
	if rules.equationDelim.MatchString(str) {
		// Check if equation ends on same line ($$...$$ on one line)
		if len(str) > 2 && strings.HasSuffix(strings.TrimSpace(str), "$$") && strings.Count(str, "$$") >= 2 {
			// Single-line equation
			result.Colors = append(result.Colors, LineColor{
				Fg:    tcell.ColorPurple,
				Bg:    tcell.ColorDefault,
				Start: 0,
				End:   len(str),
			})
			result.EndState = StateNormal
			return result
		}
		// Multi-line equation starts
		result.Colors = append(result.Colors, LineColor{
			Fg:    tcell.ColorPurple,
			Bg:    tcell.ColorDefault,
			Start: 0,
			End:   len(str),
		})
		result.EndState = StateBlockEquation
		return result
	}

	// Regular markdown line - use normal colorization
	return rules.colorizeNormal(str, StateNormal)
}

// colorizeNormal is the original colorization logic for non-markdown files.
func (rules *SyntaxRules) colorizeNormal(str string, startState LineState) ColorResult {
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
func (rules *SyntaxRules) Colorize(str string) []LineColor {
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

	// Markdown and Quarto: embedded code blocks and equations
	if ext == "md" || ext == "qmd" || ext == "markdown" || ext == "rmd" {
		rules.setupMarkdownEmbedded()
	}
}

// setupMarkdownEmbedded configures syntax rules for embedded code blocks in markdown.
func (rules *SyntaxRules) setupMarkdownEmbedded() {
	rules.isMarkdown = true
	// Match ```python, ```{python}, ```{r setup, include=FALSE}, etc.
	rules.codeBlockStart = regexp.MustCompile("^```\\{?([a-zA-Z][a-zA-Z0-9_+-]*)")
	rules.codeBlockEnd = regexp.MustCompile("^```\\s*$")
	rules.equationDelim = regexp.MustCompile(`^\$\$`)
}

// getEmbeddedRules returns the syntax rules for an embedded language,
// looking them up from the config and caching the result.
func (rules *SyntaxRules) getEmbeddedRules(lang string) *SyntaxRules {
	lang = strings.ToLower(lang)

	// Check cache first
	if cached, ok := rules.embeddedRules[lang]; ok {
		return cached
	}

	// Map common aliases to file extensions
	extMap := map[string]string{
		"python":     "py",
		"javascript": "js",
		"typescript": "ts",
		"bash":       "sh",
		"shell":      "sh",
		"yml":        "yaml",
		"r":          "r",
	}
	ext := lang
	if mapped, ok := extMap[lang]; ok {
		ext = mapped
	}

	// Look up the config for this extension
	langCfg := rules.fullConfig.ForExt(ext)

	// Create syntax rules from the config
	embeddedRules := NewSyntaxRules(langCfg)
	embeddedRules.SetupForLanguage(ext)

	// Cache and return
	rules.embeddedRules[lang] = embeddedRules
	return embeddedRules
}

// registerLanguage assigns a state index to a language name.
func (rules *SyntaxRules) registerLanguage(lang string) int {
	lang = strings.ToLower(lang)
	if idx, ok := rules.languageIndex[lang]; ok {
		return idx
	}
	idx := len(rules.languageIndex)
	rules.languageIndex[lang] = idx
	rules.indexToLanguage[idx] = lang
	return idx
}

// getLanguageFromState extracts the language name from a code block state.
func (rules *SyntaxRules) getLanguageFromState(state LineState) string {
	if state < StateCodeBlockBase {
		return ""
	}
	idx := int(state - StateCodeBlockBase)
	return rules.indexToLanguage[idx]
}

// getStateForLanguage returns the LineState for a given language.
func (rules *SyntaxRules) getStateForLanguage(lang string) LineState {
	idx := rules.registerLanguage(lang)
	return StateCodeBlockBase + LineState(idx)
}
