package buffer

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

func (buffer *Buffer) Justify(startRow, endRow, lineLen int, comStrs []string) error {

	// Get indents and remainders.
	subBuffer := buffer.InclSlice(startRow, endRow)
	indents, remainders, err := subBuffer.getIndentsRemainders(comStrs)
	if err != nil {
		return err
	}

	// Find paragraphs.
	paragraphs, err := findParagraphs(indents, remainders)
	if err != nil {
		return fmt.Errorf("error dividing text into paragraphs: %s", err)
	}

	// Justify each paragraph and append to a buffer.
	fBuffer := MakeBuffer([]string{})
	for i, p := range paragraphs {
		start := p
		end := subBuffer.Length() - 1
		if i+1 < len(paragraphs) {
			end = paragraphs[i+1] - 1
		}
		bigStr := strings.Join(remainders[start:end+1], " ")
		splitBuf := MakeSplitBuffer(bigStr, lineLen-len(indents[p]))

		// Put indentation back
		splitBuf.IndentByStr(indents[p], 0, -1)

		fBuffer.Append(splitBuf.Lines()...)
	}

	// Replace the original buffer lines with the justified ones.
	buffer.ReplaceLines(fBuffer.Lines(), startRow, endRow)

	return nil

}

func (buffer *Buffer) justify(lineLen int) *Buffer {

	// Make one long string out of the remainder of the lines.
	bigStr := buffer.ToString(" ")

	// Justify.
	splitBuf := MakeSplitBuffer(bigStr, lineLen)

	return &splitBuf

}

func (buffer *Buffer) getIndentsRemainders(comStrs []string) (indents, remainders []string, err error) {

	indents = make([]string, buffer.Length())
	remainders = make([]string, buffer.Length())

	// Construct the regexp to find indentation (including comments strings).
	pattern := "^[ \t]*(//|#)*[ \t]*"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return indents, remainders, err
	}

	// Separete lines into indent + remainder.
	for i, line := range buffer.Lines() {
		lineStr := line.ToString()
		indent := re.FindString(lineStr)
		indents[i] = indent
		remainders[i] = lineStr[len(indent):]
	}

	return indents, remainders, nil

}

// Separate lines of text into paragraphs based on indentation.
// Return a list of paragraph starts.
func findParagraphs(indents []string, remainders []string) ([]int, error) {
	paragraphs := []int{0}

	// Sanity check
	if len(indents) != len(remainders) {
		return paragraphs, fmt.Errorf("indents and remainders are different lengths")
	}

	// Trivial case
	if len(indents) == 0 {
		return paragraphs, nil
	}

	lastIndent := indents[0]
	lastRemainder := remainders[0]
	for i, indent := range indents {
		remainder := remainders[i]
		if indent != lastIndent {
			// If indentation changes, then start a new paragraph.
			paragraphs = append(paragraphs, i)
		} else if remainder != "" && lastRemainder == "" {
			paragraphs = append(paragraphs, i)
		} else if remainder == "" && lastRemainder != "" {
			paragraphs = append(paragraphs, i)
		}
		lastRemainder = remainder
		lastIndent = indent
	}
	return paragraphs, nil
}

func (buffer *Buffer) getParagraphs() [][]int {
	offsets := [][]int{}
	start := 0
	prevIsBlank := true
	for i, line := range buffer.Lines() {
		// Blank line after non blanks.
		if line.Length() == 0 && !prevIsBlank && i > 0 {
			offsets = append(offsets, []int{start, i - 1})
		}
		// Non blank line after a blank.
		if line.Length() > 0 && prevIsBlank {
			start = i
		}
		prevIsBlank = line.Length() == 0
	}
	if !prevIsBlank {
		offsets = append(offsets, []int{start, buffer.Length() - 1})
	}
	return offsets
}

// MakeSplitBuffer creates a buffer from a long string by splitting
// the string at a certain length.
func MakeSplitBuffer(bigString string, lineLen int) Buffer {
	words := strings.Fields(bigString)
	if len(words) == 0 {
		return MakeBuffer([]string{""})
	}
	lines := []Line{}
	lineStr := words[0]
	for _, word := range words[1:] {
		if lineLen > 0 && len(lineStr)+len(word) >= lineLen {
			lines = append(lines, MakeLine(lineStr))
			lineStr = word
		} else {
			lineStr += " " + word
		}
	}
	lines = append(lines, MakeLine(lineStr))
	return Buffer{
		lines: lines,
		mutex: &sync.Mutex{},
	}
}
