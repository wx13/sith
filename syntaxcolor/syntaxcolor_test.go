package syntaxcolor_test

import (
	"fmt"
	"testing"

	"github.com/wx13/sith/syntaxcolor"
)

func TestGetFileType(t *testing.T) {

	sr := syntaxcolor.NewSyntaxRules("foo.go")

	ft := sr.GetFileType("foo.go")
	if ft != "go" {
		t.Error("Error getting file type. Expected 'go', got", ft)
	}

	ft = sr.GetFileType("foo.cpp")
	if ft != "c" {
		t.Error("Error getting file type. Expected 'c', got", ft)
	}

}

func ExampleColorize() {

	sr := syntaxcolor.NewSyntaxRules("foo.go")

	lc := sr.Colorize("package main")
	fmt.Println(lc)

	lc = sr.Colorize("var x ")
	fmt.Println(lc[0].Start, lc[0].End)

	// Output:
	// []
	// 5 6

}
