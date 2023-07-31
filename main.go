package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "expect 2 arguments but got %d", len(os.Args))
		os.Exit(1)
	}
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl main\n")
	fmt.Printf("main:\n")
	num, _ := strconv.Atoi(os.Args[1])
	fmt.Printf("  mov rax, %d\n", num)
	fmt.Printf("  ret\n")
}
