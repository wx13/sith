package file

// Grid-table reformatting for markdown/quarto files. Rows may be typed
// sloppily (unaligned pipes, unwrapped text, shorthand "+-+" separators);
// FmtTable re-wraps all cell text to the column widths declared by the
// header border lines.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/wx13/sith/file/buffer"
)

// gridTable is the parsed form of a grid table: logical rows of cell text,
// plus the column widths declared by the border lines.
type gridTable struct {
	indent     string
	headerRows [][]string
	bodyRows   [][]string
	widths     []int // interior width of each column (the dash count)
}

var dashSepRe = regexp.MustCompile(`^\+[-+]*-[-+]*\+$`)
var eqSepRe = regexp.MustCompile(`^\+[=+]*=[=+]*\+$`)

// FmtTable reformats the grid table containing the cursor. Column widths
// come from the header border lines (widest wins); cell text is re-wrapped
// to fit. Pipe tables are left alone.
func (file *File) FmtTable() error {

	row := file.MultiCursor.GetRow(0)
	startRow, endRow := file.findTableBounds(row)
	if startRow < 0 {
		return fmt.Errorf("cursor is not inside a table")
	}

	lines := make([]string, 0, endRow-startRow+1)
	for r := startRow; r <= endRow; r++ {
		lines = append(lines, file.buffer.GetRow(r).ToString())
	}

	table, err := parseGridTable(lines)
	if err != nil {
		return err
	}

	newLines := renderGridTable(table)
	bufLines := make([]buffer.Line, len(newLines))
	for i, line := range newLines {
		bufLines[i] = buffer.MakeLine(line)
	}

	file.buffer.ReplaceLines(bufLines, startRow, endRow)
	file.InvalidateSyntaxCache(startRow)
	file.enforceRowBounds()
	file.enforceColBounds()
	file.Snapshot()

	return nil
}

// findTableBounds returns the first and last rows of the table containing
// the given row, or (-1, -1) if the row is not part of a table.
func (file *File) findTableBounds(row int) (int, int) {
	isTableLine := func(r int) bool {
		s := strings.TrimSpace(file.buffer.GetRow(r).ToString())
		return len(s) > 0 && (s[0] == '|' || s[0] == '+')
	}
	if row >= file.buffer.Length() || !isTableLine(row) {
		return -1, -1
	}
	startRow, endRow := row, row
	for startRow > 0 && isTableLine(startRow-1) {
		startRow--
	}
	for endRow < file.buffer.Length()-1 && isTableLine(endRow+1) {
		endRow++
	}
	return startRow, endRow
}

// parseGridTable leniently parses grid-table lines. Pipes do not need to
// align with the borders; a line's cells are found by splitting on
// unescaped pipes (ignoring pipes inside backtick code spans). Separator
// lines act only as row delimiters, except the top border and the "+===+"
// header separator, which also declare column widths.
func parseGridTable(lines []string) (*gridTable, error) {

	indent := ""
	if len(lines) > 0 {
		trimmed := strings.TrimLeft(lines[0], " \t")
		indent = lines[0][:len(lines[0])-len(trimmed)]
	}

	// Split lines into row blocks, delimited by separator lines.
	var blocks [][][]string
	var blockStarts []int
	var cur [][]string
	curStart := -1
	headerIdx := -1
	var topBorder, eqBorder []int
	sawBorder := false
	flush := func() {
		if len(cur) > 0 {
			blocks = append(blocks, cur)
			blockStarts = append(blockStarts, curStart)
			cur = nil
		}
	}
	for n, line := range lines {
		s := strings.TrimSpace(line)
		switch {
		case dashSepRe.MatchString(s):
			if !sawBorder && len(blocks) == 0 && len(cur) == 0 {
				topBorder = sepWidths(s)
			}
			sawBorder = true
			flush()
		case eqSepRe.MatchString(s):
			sawBorder = true
			flush()
			if headerIdx >= 0 {
				return nil, fmt.Errorf("table has multiple header separators (line %d of table)", n+1)
			}
			if len(blocks) == 0 {
				return nil, fmt.Errorf("header separator with no header row (line %d of table)", n+1)
			}
			headerIdx = len(blocks)
			eqBorder = sepWidths(s)
		case len(s) > 0 && s[0] == '|':
			if len(cur) == 0 {
				curStart = n
			}
			cur = append(cur, splitTableRow(s))
		default:
			return nil, fmt.Errorf("line %d of table is not a table line", n+1)
		}
	}
	flush()

	if !sawBorder {
		return nil, fmt.Errorf("pipe table: only grid tables are reformatted")
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("table has no content rows")
	}

	// All content lines must have the same number of cells.
	ncols := len(blocks[0][0])
	for b, block := range blocks {
		for l, fields := range block {
			if len(fields) != ncols {
				return nil, fmt.Errorf("line %d of table has %d cells; expected %d",
					blockStarts[b]+l+1, len(fields), ncols)
			}
		}
	}

	// Merge each block's lines into one logical row per block.
	rows := make([][]string, len(blocks))
	for b, block := range blocks {
		row := make([]string, ncols)
		for _, fields := range block {
			for i, field := range fields {
				field = strings.TrimSpace(field)
				if field == "" {
					continue
				}
				if row[i] == "" {
					row[i] = field
				} else {
					row[i] += " " + field
				}
			}
		}
		rows[b] = row
	}

	table := gridTable{indent: indent}
	if headerIdx >= 0 {
		table.headerRows = rows[:headerIdx]
		table.bodyRows = rows[headerIdx:]
	} else {
		table.bodyRows = rows
	}

	table.widths = computeWidths(ncols, [][]int{topBorder, eqBorder}, rows)
	return &table, nil
}

// computeWidths determines column widths: the widest qualifying border line
// wins, floored by the longest unbreakable word in each column. If no
// border line has the right column count, content is left unwrapped.
func computeWidths(ncols int, borders [][]int, rows [][]string) []int {
	widths := make([]int, ncols)
	haveBorder := false
	for _, border := range borders {
		if len(border) != ncols {
			continue
		}
		haveBorder = true
		for i, w := range border {
			widths[i] = max(widths[i], w)
		}
	}
	for _, row := range rows {
		for i, cell := range row {
			if !haveBorder {
				widths[i] = max(widths[i], runewidth.StringWidth(cell)+2)
				continue
			}
			for _, word := range strings.Fields(cell) {
				widths[i] = max(widths[i], runewidth.StringWidth(word)+2)
			}
		}
	}
	for i := range widths {
		widths[i] = max(widths[i], 3)
	}
	return widths
}

// sepWidths returns the dash/equals counts between the plus signs of a
// separator line, e.g. "+---+--+" => [3, 2].
func sepWidths(line string) []int {
	segments := strings.Split(strings.Trim(line, "+"), "+")
	widths := make([]int, len(segments))
	for i, seg := range segments {
		widths[i] = len(seg)
	}
	return widths
}

// splitTableRow splits a content line into cells on unescaped pipes,
// ignoring pipes inside backtick code spans. The leading pipe and trailing
// pipe (if present) are dropped. Cell text is kept verbatim.
func splitTableRow(line string) []string {
	runes := []rune(line)
	i := 0
	if len(runes) > 0 && runes[0] == '|' {
		i = 1
	}
	var fields []string
	var cell []rune
	inCode := 0 // length of the opening backtick run, 0 if not in a code span
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == '\\' && i+1 < len(runes):
			cell = append(cell, r, runes[i+1])
			i += 2
		case r == '`':
			n := 1
			for i+n < len(runes) && runes[i+n] == '`' {
				n++
			}
			if inCode == 0 {
				inCode = n
			} else if inCode == n {
				inCode = 0
			}
			cell = append(cell, runes[i:i+n]...)
			i += n
		case r == '|' && inCode == 0:
			fields = append(fields, string(cell))
			cell = nil
			i++
		default:
			cell = append(cell, r)
			i++
		}
	}
	// Anything after the last pipe is a cell only if non-blank.
	if len(strings.TrimSpace(string(cell))) > 0 {
		fields = append(fields, string(cell))
	}
	return fields
}

// renderGridTable renders the table back to text, wrapping each cell to
// its column width.
func renderGridTable(t *gridTable) []string {
	dashSep := renderSep(t, '-')
	eqSep := renderSep(t, '=')

	lines := []string{dashSep}
	for i, row := range t.headerRows {
		lines = append(lines, renderRow(t, row)...)
		if i < len(t.headerRows)-1 {
			lines = append(lines, dashSep)
		} else {
			lines = append(lines, eqSep)
		}
	}
	for _, row := range t.bodyRows {
		lines = append(lines, renderRow(t, row)...)
		lines = append(lines, dashSep)
	}
	return lines
}

func renderSep(t *gridTable, ch byte) string {
	var b strings.Builder
	b.WriteString(t.indent)
	b.WriteByte('+')
	for _, w := range t.widths {
		b.WriteString(strings.Repeat(string(ch), w))
		b.WriteByte('+')
	}
	return b.String()
}

func renderRow(t *gridTable, row []string) []string {
	wrapped := make([][]string, len(row))
	height := 1
	for i, cell := range row {
		wrapped[i] = wrapCell(cell, t.widths[i]-2)
		height = max(height, len(wrapped[i]))
	}
	lines := make([]string, height)
	for j := 0; j < height; j++ {
		var b strings.Builder
		b.WriteString(t.indent)
		b.WriteByte('|')
		for i := range row {
			frag := ""
			if j < len(wrapped[i]) {
				frag = wrapped[i][j]
			}
			b.WriteByte(' ')
			b.WriteString(frag)
			b.WriteString(strings.Repeat(" ", max(0, t.widths[i]-2-runewidth.StringWidth(frag))))
			b.WriteString(" |")
		}
		lines[j] = b.String()
	}
	return lines
}

// wrapCell greedily word-wraps text to the given display width. Words wider
// than the width get a line to themselves (computeWidths normally prevents
// this by growing the column).
func wrapCell(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	cur := words[0]
	for _, word := range words[1:] {
		if runewidth.StringWidth(cur)+1+runewidth.StringWidth(word) <= width {
			cur += " " + word
		} else {
			lines = append(lines, cur)
			cur = word
		}
	}
	return append(lines, cur)
}
