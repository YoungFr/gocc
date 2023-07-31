package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "expect 2 arguments but got %d", len(os.Args))
		os.Exit(1)
	}
	// 算术表达式
	s := os.Args[1]

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl main\n")
	fmt.Printf("main:\n")

	// 操作数
	nums := strings.FieldsFunc(s, func(r rune) bool {
		return r == '+' || r == '-'
	})
	// 操作符
	ops := make([]rune, 0)
	for _, r := range s {
		if r == '+' || r == '-' {
			ops = append(ops, r)
		}
	}

	num, _ := strconv.Atoi(nums[0])
	fmt.Printf("  mov rax, %d\n", num)

	for i, op := range ops {
		// len(nums) = len(ops) + 1
		num, _ = strconv.Atoi(nums[i+1])
		if op == '+' {
			fmt.Printf("  add rax, %d\n", num)
		} else {
			fmt.Printf("  sub rax, %d\n", num)
		}
	}

	fmt.Printf("  ret\n")
}
