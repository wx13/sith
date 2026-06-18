package file_test

import (
	"sync"
	"testing"

	"github.com/wx13/sith/config"
	"github.com/wx13/sith/file"
)

func TestNewFile(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	f := file.NewFile("", make(chan struct{}), nil, config.Config{}, &wg)
	if f == nil {
		t.Error("bad")
	}
}

func CheckBuffer(t *testing.T, f *file.File, s, msg string) {
	if f.ToString() != s {
		t.Errorf("Error: %s; Expected %q but got %q\n", msg, s, f.ToString())
	}
}

func TestInsertChar(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	f := file.NewFile("", make(chan struct{}), nil, config.Config{}, &wg)
	wg.Wait()
	f.InsertChar('h')
	f.InsertChar('e')
	f.InsertChar('l')
	f.InsertChar('l')
	f.InsertChar('o')
	CheckBuffer(t, f, "hello", "InsertChar")
}

func TestInsertStr(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	f := file.NewFile("", make(chan struct{}), nil, config.Config{}, &wg)
	wg.Wait()
	f.InsertStr("line 1")
	f.Newline()
	f.InsertStr("line 2")
	f.Newline()
	f.InsertStr("line 3")
	f.Newline()
	CheckBuffer(t, f, "line 1\nline 2\nline 3\n", "InsertStr")
}

func TestEditing(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	f := file.NewFile("", make(chan struct{}), nil, config.Config{}, &wg)
	wg.Wait()
	f.InsertChar('a')
	f.InsertChar('b')
	f.InsertChar('c')
	f.Newline()
	f.InsertChar('d')
	f.InsertChar('e')
	CheckBuffer(t, f, "abc\nde", "InsertChar")
	f.Backspace()
	CheckBuffer(t, f, "abc\nd", "Backspace")
	f.Backspace()
	f.Backspace()
	CheckBuffer(t, f, "abc", "Backspace (over newline)")
}

func TestFindCodeBlockBounds(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	cfg := config.Config{
		FileConfigs: map[string]config.Config{
			"md": {},
		},
	}

	f := file.NewFile("test.md", make(chan struct{}), nil, cfg, &wg)
	wg.Wait()

	// Build a markdown file with a code block
	// Line 0: # Heading
	// Line 1: ```python
	// Line 2: def foo():
	// Line 3:     pass
	// Line 4: ```
	// Line 5: More text
	f.InsertStr("# Heading")
	f.Newline()
	f.InsertStr("```python")
	f.Newline()
	f.InsertStr("def foo():")
	f.Newline()
	f.InsertStr("    pass")
	f.Newline()
	f.InsertStr("```")
	f.Newline()
	f.InsertStr("More text")

	// Test finding bounds when inside code block (row 2 or 3)
	start, end, lang := f.FindCodeBlockBounds(3)
	if start != 1 || end != 4 || lang != "python" {
		t.Errorf("Expected (1, 4, python), got (%d, %d, %s)", start, end, lang)
	}

	// Test when on the opening line
	start, end, lang = f.FindCodeBlockBounds(1)
	if start != 1 || end != 4 || lang != "python" {
		t.Errorf("On opening: Expected (1, 4, python), got (%d, %d, %s)", start, end, lang)
	}

	// Test when on the closing line
	start, end, lang = f.FindCodeBlockBounds(4)
	if start != 1 || end != 4 || lang != "python" {
		t.Errorf("On closing: Expected (1, 4, python), got (%d, %d, %s)", start, end, lang)
	}

	// Test when outside code block
	start, end, lang = f.FindCodeBlockBounds(0) // On "# Heading"
	if start != -1 {
		t.Errorf("Outside block: Expected -1, got %d", start)
	}

	start, end, lang = f.FindCodeBlockBounds(5) // On "More text"
	if start != -1 {
		t.Errorf("After block: Expected -1, got %d", start)
	}
}

func TestFindCodeBlockBoundsQuarto(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	cfg := config.Config{
		FileConfigs: map[string]config.Config{
			"qmd": {},
		},
	}

	f := file.NewFile("test.qmd", make(chan struct{}), nil, cfg, &wg)
	wg.Wait()

	// Build a Quarto file with a code block using {python} syntax
	// Line 0: # Analysis
	// Line 1: ```{python}
	// Line 2: import numpy as np
	// Line 3: ```
	f.InsertStr("# Analysis")
	f.Newline()
	f.InsertStr("```{python}")
	f.Newline()
	f.InsertStr("import numpy as np")
	f.Newline()
	f.InsertStr("```")

	start, end, lang := f.FindCodeBlockBounds(2)
	if start != 1 || end != 3 || lang != "python" {
		t.Errorf("Quarto syntax: Expected (1, 3, python), got (%d, %d, %s)", start, end, lang)
	}
}
