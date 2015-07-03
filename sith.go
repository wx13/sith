package main

import "fmt"


type TextBuffer struct {
	lines []string
}

func (text TextBuffer) print() {
	for j, line := range(text.lines) {
		fmt.Println(j, line)
	}
}


func main() {
	textBuffer := TextBuffer{[]string{"hello","bye"}}
	textBuffer.print()
}
