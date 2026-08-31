package file

import (
	"strings"
	"sync"
	"testing"

	"github.com/wx13/sith/config"
)

func fmtTable(t *testing.T, input string) (string, error) {
	t.Helper()
	table, err := parseGridTable(strings.Split(input, "\n"))
	if err != nil {
		return "", err
	}
	return strings.Join(renderGridTable(table), "\n"), nil
}

func checkTable(t *testing.T, input, expected string) {
	t.Helper()
	out, err := fmtTable(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != expected {
		t.Errorf("got:\n%s\nexpected:\n%s", out, expected)
	}
	// Reformatting must be idempotent.
	again, err := fmtTable(t, out)
	if err != nil {
		t.Fatalf("error on second pass: %v", err)
	}
	if again != out {
		t.Errorf("not idempotent; second pass gave:\n%s", again)
	}
}

func TestFmtTableSloppyRows(t *testing.T) {
	input := strings.Join([]string{
		"+----------------+----------+",
		"| A              | B        |",
		"+================+==========+",
		"| something short | 27 |",
		"+-+",
		"| something too long for the width | also too long, I am pretty sure |",
		"+-+-+",
	}, "\n")
	expected := strings.Join([]string{
		"+----------------+----------+",
		"| A              | B        |",
		"+================+==========+",
		"| something      | 27       |",
		"| short          |          |",
		"+----------------+----------+",
		"| something too  | also too |",
		"| long for the   | long, I  |",
		"| width          | am       |",
		"|                | pretty   |",
		"|                | sure     |",
		"+----------------+----------+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableAlreadyFormatted(t *testing.T) {
	input := strings.Join([]string{
		"+-----+-------------+",
		"| A   | B           |",
		"+=====+=============+",
		"| one | two three   |",
		"|     | four        |",
		"+-----+-------------+",
	}, "\n")
	expected := strings.Join([]string{
		"+-----+-------------+",
		"| A   | B           |",
		"+=====+=============+",
		"| one | two three   |",
		"|     | four        |",
		"+-----+-------------+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableInconsistentHeaderBorders(t *testing.T) {
	// Widest border wins, per column.
	input := strings.Join([]string{
		"+---+---+---+",
		"| A | B |  C  |",
		"+===========+========+=================+",
		"| a | b | c |",
		"+-+",
	}, "\n")
	expected := strings.Join([]string{
		"+-----------+--------+-----------------+",
		"| A         | B      | C               |",
		"+===========+========+=================+",
		"| a         | b      | c               |",
		"+-----------+--------+-----------------+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableLongWordGrowsColumn(t *testing.T) {
	// An unbreakable word wider than the column grows the column.
	input := strings.Join([]string{
		"+-----+-----+",
		"| A   | B   |",
		"+=====+=====+",
		"| supercalifragilistic | b |",
		"+-+",
	}, "\n")
	expected := strings.Join([]string{
		"+----------------------+-----+",
		"| A                    | B   |",
		"+======================+=====+",
		"| supercalifragilistic | b   |",
		"+----------------------+-----+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTablePipesInContent(t *testing.T) {
	// Escaped pipes and pipes in code spans are not cell boundaries.
	input := strings.Join([]string{
		"+----------+----------+",
		"| A        | B        |",
		"+==========+==========+",
		"| a \\| b | `x | y` |",
		"+-+",
	}, "\n")
	expected := strings.Join([]string{
		"+----------+----------+",
		"| A        | B        |",
		"+==========+==========+",
		"| a \\| b   | `x | y`  |",
		"+----------+----------+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableHeaderless(t *testing.T) {
	input := strings.Join([]string{
		"+--------+--------+",
		"| this is too wide | b |",
		"+-+",
		"| c | d |",
		"+-+",
	}, "\n")
	expected := strings.Join([]string{
		"+--------+--------+",
		"| this   | b      |",
		"| is too |        |",
		"| wide   |        |",
		"+--------+--------+",
		"| c      | d      |",
		"+--------+--------+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableIndented(t *testing.T) {
	input := strings.Join([]string{
		"  +-----+-----+",
		"  | A   | B   |",
		"  +=====+=====+",
		"  | a | b |",
		"  +-+",
	}, "\n")
	expected := strings.Join([]string{
		"  +-----+-----+",
		"  | A   | B   |",
		"  +=====+=====+",
		"  | a   | b   |",
		"  +-----+-----+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableEmptyCells(t *testing.T) {
	input := strings.Join([]string{
		"+-----+-----+",
		"| A   | B   |",
		"+=====+=====+",
		"| | b |",
		"+-+",
	}, "\n")
	expected := strings.Join([]string{
		"+-----+-----+",
		"| A   | B   |",
		"+=====+=====+",
		"|     | b   |",
		"+-----+-----+",
	}, "\n")
	checkTable(t, input, expected)
}

func TestFmtTableErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			"pipe table",
			"| A | B |\n|---|---|\n| a | b |",
			"pipe table",
		},
		{
			"wrong cell count",
			"+-----+-----+\n| A | B |\n+=====+=====+\n| a | b | c |\n+-+",
			"expected 2",
		},
		{
			"multiple header separators",
			"+---+\n| A |\n+===+\n| a |\n+===+\n| b |\n+---+",
			"multiple header separators",
		},
		{
			"header separator first",
			"+===+\n| a |\n+---+",
			"no header row",
		},
	}
	for _, c := range cases {
		_, err := fmtTable(t, c.input)
		if err == nil {
			t.Errorf("%s: expected error, got none", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %q does not contain %q", c.name, err.Error(), c.want)
		}
	}
}

func TestFmtTableEndToEnd(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	f := NewFile("", make(chan struct{}), nil, config.Config{}, &wg)
	wg.Wait()
	f.InsertStr("Some text before.")
	f.Newline()
	for _, line := range []string{
		"+-----+----------+",
		"| A   | B        |",
		"+=====+==========+",
		"| a | this is long enough to wrap |",
		"+-+",
	} {
		f.InsertStr(line)
		f.Newline()
	}
	f.InsertStr("Some text after.")

	// Cursor outside the table should error and leave the buffer alone.
	f.MultiCursor.SetCursor(0, 0, 0, 0)
	if err := f.FmtTable(); err == nil {
		t.Error("expected error for cursor outside table")
	}

	// Cursor on the sloppy row.
	f.MultiCursor.SetCursor(0, 4, 0, 0)
	if err := f.FmtTable(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := strings.Join([]string{
		"Some text before.",
		"+-----+----------+",
		"| A   | B        |",
		"+=====+==========+",
		"| a   | this is  |",
		"|     | long     |",
		"|     | enough   |",
		"|     | to wrap  |",
		"+-----+----------+",
		"Some text after.",
	}, "\n")
	if f.ToString() != expected {
		t.Errorf("got:\n%s\nexpected:\n%s", f.ToString(), expected)
	}
}

func TestSplitTableRow(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"| a | b |", []string{" a ", " b "}},
		{"| a | b", []string{" a ", " b"}},
		{"| a \\| b | c |", []string{" a \\| b ", " c "}},
		{"| `a | b` | c |", []string{" `a | b` ", " c "}},
		{"| ``a | b`` | c |", []string{" ``a | b`` ", " c "}},
		{"|  | b |", []string{"  ", " b "}},
	}
	for _, c := range cases {
		got := splitTableRow(c.input)
		if len(got) != len(c.want) {
			t.Errorf("splitTableRow(%q) = %q; want %q", c.input, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitTableRow(%q)[%d] = %q; want %q", c.input, i, got[i], c.want[i])
			}
		}
	}
}
