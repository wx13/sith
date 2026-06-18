package syntaxcolor_test

import (
	"fmt"
	"testing"

	"github.com/wx13/sith/config"
	"github.com/wx13/sith/syntaxcolor"
)

func ExampleSyntaxRules_Colorize() {

	cfg := config.Config{
		SyntaxRules: map[string]config.Color{
			"abc": {FG: "green"},
		},
	}

	sr := syntaxcolor.NewSyntaxRules(cfg)

	lc := sr.Colorize("package main")
	fmt.Println(lc)

	lc = sr.Colorize("var abc ")
	fmt.Println(lc[0].Start, lc[0].End)

	// Output:
	// []
	// 4 7

}

func TestMarkdownCodeBlock(t *testing.T) {
	// Set up a config with Python syntax rules
	fullCfg := config.Config{
		FileConfigs: map[string]config.Config{
			"py": {
				SyntaxRules: map[string]config.Color{
					"#.*$":   {FG: "cyan"},
					`".*?"`:  {FG: "yellow"},
					"'.*?'":  {FG: "yellow"},
					"\\bdef\\b": {FG: "green"},
				},
			},
			"md": {
				SyntaxRules: map[string]config.Color{
					"^#+.*$": {FG: "green"},
				},
			},
		},
	}

	// Create markdown syntax rules with full config for embedded language lookup
	mdCfg := fullCfg.ForExt("md")
	sr := syntaxcolor.NewSyntaxRulesWithFullConfig(mdCfg, fullCfg)
	sr.SetupForLanguage("md")

	// Test: markdown heading should be colored
	result := sr.ColorizeWithState("# Heading", syntaxcolor.StateNormal)
	if len(result.Colors) == 0 {
		t.Error("Expected markdown heading to be colored")
	}
	if result.EndState != syntaxcolor.StateNormal {
		t.Errorf("Expected end state to be Normal, got %d", result.EndState)
	}

	// Test: code block start should transition to code block state
	result = sr.ColorizeWithState("```python", syntaxcolor.StateNormal)
	if result.EndState < syntaxcolor.StateCodeBlockBase {
		t.Errorf("Expected end state to be a code block state, got %d", result.EndState)
	}
	codeBlockState := result.EndState

	// Test: inside code block, Python syntax should be applied
	result = sr.ColorizeWithState("def foo():  # comment", codeBlockState)
	if result.EndState != codeBlockState {
		t.Errorf("Expected to stay in code block state")
	}
	// Should have colors for 'def' and '# comment'
	if len(result.Colors) < 2 {
		t.Errorf("Expected at least 2 color regions for Python code, got %d", len(result.Colors))
	}

	// Test: code block end should return to normal state
	result = sr.ColorizeWithState("```", codeBlockState)
	if result.EndState != syntaxcolor.StateNormal {
		t.Errorf("Expected end state to be Normal after code block end, got %d", result.EndState)
	}
}

func TestMarkdownEquationBlock(t *testing.T) {
	fullCfg := config.Config{
		FileConfigs: map[string]config.Config{
			"md": {
				SyntaxRules: map[string]config.Color{
					"^#+.*$": {FG: "green"},
				},
			},
		},
	}

	mdCfg := fullCfg.ForExt("md")
	sr := syntaxcolor.NewSyntaxRulesWithFullConfig(mdCfg, fullCfg)
	sr.SetupForLanguage("md")

	// Test: equation block start
	result := sr.ColorizeWithState("$$", syntaxcolor.StateNormal)
	if result.EndState != syntaxcolor.StateBlockEquation {
		t.Errorf("Expected end state to be BlockEquation, got %d", result.EndState)
	}

	// Test: inside equation block
	result = sr.ColorizeWithState("x^2 + y^2 = z^2", syntaxcolor.StateBlockEquation)
	if result.EndState != syntaxcolor.StateBlockEquation {
		t.Errorf("Expected to stay in equation state")
	}
	if len(result.Colors) == 0 {
		t.Error("Expected equation content to be colored")
	}

	// Test: equation block end
	result = sr.ColorizeWithState("$$", syntaxcolor.StateBlockEquation)
	if result.EndState != syntaxcolor.StateNormal {
		t.Errorf("Expected end state to be Normal after equation end, got %d", result.EndState)
	}
}

func TestQuartoCodeBlock(t *testing.T) {
	fullCfg := config.Config{
		FileConfigs: map[string]config.Config{
			"py": {
				SyntaxRules: map[string]config.Color{
					"#.*$": {FG: "cyan"},
				},
			},
			"qmd": {
				Parent: "md",
			},
			"md": {
				SyntaxRules: map[string]config.Color{
					"^#+.*$": {FG: "green"},
				},
			},
		},
	}

	// Test with Quarto-style code block: ```{python}
	qmdCfg := fullCfg.ForExt("qmd")
	sr := syntaxcolor.NewSyntaxRulesWithFullConfig(qmdCfg, fullCfg)
	sr.SetupForLanguage("qmd")

	result := sr.ColorizeWithState("```{python}", syntaxcolor.StateNormal)
	if result.EndState < syntaxcolor.StateCodeBlockBase {
		t.Errorf("Expected end state to be a code block state for Quarto syntax, got %d", result.EndState)
	}
}
