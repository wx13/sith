package syntaxcolor_test

import (
	"fmt"

	"github.com/wx13/sith/config"
	"github.com/wx13/sith/syntaxcolor"
)

func ExampleColorize() {

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
