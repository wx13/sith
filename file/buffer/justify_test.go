package buffer

import (
	"testing"
)

func strSliceEq(a []string, b ...string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func intSliceEq(a []int, b ...int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetIndentsRemainders(t *testing.T) {
	var b Buffer
	var indents, remainders []string
	var err error

	// Empty buffer
	b = MakeBuffer([]string{})
	indents, remainders, err = b.getIndentsRemainders([]string{})
	if err != nil {
		t.Error(err)
	}
	if len(indents) != 0 {
		t.Error("Expecting empty indents, got:", indents)
	}
	if len(remainders) != 0 {
		t.Error("Expecting empty remainders, got:", remainders)
	}

	// Single blank line
	b = MakeBuffer([]string{""})
	indents, remainders, err = b.getIndentsRemainders([]string{})
	if err != nil {
		t.Error(err)
	}
	if !strSliceEq(indents, "") {
		t.Error("Expecting single blank indent, got:", indents)
	}
	if !strSliceEq(remainders, "") {
		t.Error("Expecting single blank remainder, got:", remainders)
	}

	// Multiple non-blank lines.
	b = MakeBuffer([]string{"abc", "def", "ghi"})
	indents, remainders, err = b.getIndentsRemainders([]string{})
	if err != nil {
		t.Error(err)
	}
	if !strSliceEq(indents, "", "", "") {
		t.Error("Expecting three blank indents, got:", indents)
	}
	if !strSliceEq(remainders, "abc", "def", "ghi") {
		t.Error("Wrong remainders:", remainders)
	}

	// Whitespace only
	b = MakeBuffer([]string{"abc", "  def", "  ", "    ghi"})
	indents, remainders, err = b.getIndentsRemainders([]string{})
	if err != nil {
		t.Error(err)
	}
	if !strSliceEq(indents, "", "  ", "  ", "    ") {
		t.Error("Wrong indents:", indents)
	}
	if !strSliceEq(remainders, "abc", "def", "", "ghi") {
		t.Error("Wrong remainders:", remainders)
	}

	// Whitespace and comments
	b = MakeBuffer([]string{
		"abc",
		"def",
		"",
		"  // 123",
		"  // 456",
		"  //",
		"  // abc",
	})
	indents, remainders, err = b.getIndentsRemainders([]string{"//"})
	if err != nil {
		t.Error(err)
	}
	if !strSliceEq(indents, "", "", "", "  // ", "  // ", "  //", "  // ") {
		t.Error("Wrong indents:", indents)
	}
	if !strSliceEq(remainders, "abc", "def", "", "123", "456", "", "abc") {
		t.Error("Wrong remainders:", remainders)
	}

}

func TestFindParagraphs(t *testing.T) {
	var indents, remainders []string
	var paragraphs []int
	var err error

	// Empty indents
	indents = []string{}
	remainders = []string{}
	paragraphs, err = findParagraphs(indents, remainders)
	if err != nil {
		t.Error(err)
	}
	if !intSliceEq(paragraphs, 0) {
		t.Error("Wrong paragraphs:", paragraphs)
	}

	// Simple indents
	indents = []string{"  ", "  ", "", "  "}
	remainders = []string{"abc", "def", "", "ghi"}
	paragraphs, err = findParagraphs(indents, remainders)
	if err != nil {
		t.Error(err)
	}
	if !intSliceEq(paragraphs, 0, 2, 3) {
		t.Error("Wrong paragraphs:", paragraphs)
	}

	// More complex indents
	indents = []string{"", "", "", ""}
	remainders = []string{"abc", "def", "", "ghi"}
	paragraphs, err = findParagraphs(indents, remainders)
	if err != nil {
		t.Error(err)
	}
	if !intSliceEq(paragraphs, 0, 2, 3) {
		t.Error("Wrong paragraphs:", paragraphs)
	}

	// One multi-line paragraph
	indents = []string{"", "", ""}
	remainders = []string{"abc", "def", "ghi"}
	paragraphs, err = findParagraphs(indents, remainders)
	if err != nil {
		t.Error(err)
	}
	if !intSliceEq(paragraphs, 0) {
		t.Error("Wrong paragraphs:", paragraphs)
	}

}

func TestJustify(t *testing.T) {

	// One line, infinte line length
	b := MakeBuffer([]string{"123 456 789"})
	b.Justify(0, 0, 0, []string{})
	if b.ToString("\n") != "123 456 789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// One line, finte line length
	b = MakeBuffer([]string{"123 456 789"})
	b.Justify(0, 0, 9, []string{})
	if b.ToString("\n") != "123 456\n789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Multiline, infinte line length
	b = MakeBuffer([]string{"123", "456", "789"})
	b.Justify(0, 2, 0, []string{})
	if b.ToString("\n") != "123 456 789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Multiline, finite line length
	b = MakeBuffer([]string{"123", "456", "789"})
	b.Justify(0, 2, 9, []string{})
	if b.ToString("\n") != "123 456\n789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Single indented line
	b = MakeBuffer([]string{"  123 456 789"})
	b.Justify(0, 0, 0, []string{})
	if b.ToString("\n") != "  123 456 789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Single indented line, wrapped
	b = MakeBuffer([]string{"  123 456 789"})
	b.Justify(0, 0, 9, []string{})
	if b.ToString("\n") != "  123 456\n  789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Multiple indented lines become one
	b = MakeBuffer([]string{"  123", "  456", "  789"})
	b.Justify(0, 2, 0, []string{})
	if b.ToString("\n") != "  123 456 789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Multiline comment becomes one line.
	b = MakeBuffer([]string{"  // 123", "  // 456", "  // 789"})
	b.Justify(0, 2, 0, []string{"//", "#"})
	if b.ToString("\n") != "  // 123 456 789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// One-line comment becomes multiline.
	b = MakeBuffer([]string{"    // 123 456 789"})
	b.Justify(0, 0, 12, []string{"//", "#"})
	if b.ToString("\n") != "    // 123\n    // 456\n    // 789" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

	// Multiple paragraphs.
	b = MakeBuffer([]string{"abc", "def", "", "ghi", "jkl"})
	b.Justify(0, 4, 80, []string{})
	if b.ToString("\n") != "abc def\n\nghi jkl" {
		t.Error("Screwed up the buffer:", b.ToString("\n"))
	}

}
