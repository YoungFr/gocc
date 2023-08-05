package main

import (
	"fmt"
	"os"
)

func locateError(offset int) {
	fmt.Fprintln(os.Stderr, source)
	fmt.Fprintf(os.Stderr, "%*s\033[31m^ \033[0m", offset, "")
}

var source string

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "\033[31mexpected 2 arguments but got %d\n\033[0m", len(os.Args))
		os.Exit(1)
	}

	source = os.Args[1]
	token := tokenize()
	node := expr(&token, token)
	if token.kind != TokenEof {
		locateError(token.begin)
		fmt.Fprintln(os.Stderr, "\033[31mextra token\033[0m")
		os.Exit(1)
	}

	fmt.Println("  .globl main\nmain:")
	gen(node)
	fmt.Println("  ret")
}
