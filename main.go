package main

import (
	"fmt"
	"os"
	"strings"
)

func locate(begin int, length int) {
	fmt.Fprintln(os.Stderr, source)
	if length == 0 {
		length = 1
	}
	fmt.Fprintf(os.Stderr, "%*s\033[31m%s \033[0m", begin, "", strings.Repeat("^", length))
}

var source string

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "\033[31mexpected 2 arguments but got %d\n\033[0m", len(os.Args))
		os.Exit(1)
	}

	source = os.Args[1]
	token := tokenize()
	program := parse(token)
	gen(program)
}
