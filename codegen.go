package main

import (
	"fmt"
	"os"
)

// Code generator

func init() {
	counter = makecounter()
}

var counter func() int

func makecounter() func() int {
	i := 1
	return func() int {
		t := i
		i++
		return t
	}
}

func push() {
	fmt.Println("  push %rax")
}

func pop(arg string) {
	fmt.Printf("  pop %s\n", arg)
}

// Assign offsets to local variables.
func assignLvarOffsets(program *Function) {
	offset := 0
	for v := program.locals; v != nil; v = v.next {
		offset += 8
		v.offset = -offset
	}
	program.stackSize = alignTo(offset, 16)
}

// Round up `n` to the nearest multiple of `align`.
// For instance, alignTo(5, 8) == 8 && alignTo(11, 8) == 16
func alignTo(n, align int) int {
	return (n + align - 1) / align * align
}

func gen(program *Function) {
	assignLvarOffsets(program)
	fmt.Println("  .globl main")
	fmt.Println("main:")
	fmt.Println("  push %rbp")
	fmt.Println("  mov %rsp, %rbp")
	fmt.Printf("  sub $%d, %%rsp\n", program.stackSize)
	for n := program.body; n != nil; n = n.next {
		genStmt(n)
	}
	fmt.Println(".L.return:")
	fmt.Println("  mov %rbp, %rsp")
	fmt.Println("  pop %rbp")
	fmt.Println("  ret")
}

func genStmt(node *Node) {
	switch node.kind {
	case NodeExprStmt:
		genExpr(node.lhs)
		return
	case NodeBlock:
		for n := node.body; n != nil; n = n.next {
			genStmt(n)
		}
		return
	case NodeReturn:
		genExpr(node.lhs)
		fmt.Println("  jmp .L.return")
		return
	case NodeIf:
		c := counter()
		genExpr(node.condition)
		fmt.Printf("  cmp $0, %%rax\n")
		fmt.Printf("  je  .L.else.%d\n", c)
		genStmt(node.thenBranch)
		fmt.Printf("  jmp .L.end.%d\n", c)
		fmt.Printf(".L.else.%d:\n", c)
		if node.elseBranch != nil {
			genStmt(node.elseBranch)
		}
		fmt.Printf(".L.end.%d:\n", c)
		return
	case NodeFor:
		c := counter()
		if node.initializer != nil {
			genStmt(node.initializer)
		}
		fmt.Printf(".L.begin.%d:\n", c)
		if node.condition != nil {
			genExpr(node.condition)
			fmt.Printf("  cmp $0, %%rax\n")
			fmt.Printf("  je  .L.end.%d\n", c)
		}
		genStmt(node.thenBranch)
		if node.increment != nil {
			genExpr(node.increment)
		}
		fmt.Printf("  jmp .L.begin.%d\n", c)
		fmt.Printf(".L.end.%d:\n", c)
		return
	}
}

// Compute the absolute address of a given node.
func genAddr(node *Node) {
	switch node.kind {
	case NodeVar:
		fmt.Printf("  lea %d(%%rbp), %%rax\n", node.variable.offset)
		return
	case NodeDeref:
		genExpr(node.lhs)
		return
	}
	locate(node.token.begin, node.token.length)
	fmt.Fprintln(os.Stderr, "\033[31mnot addressable\033[0m")
	os.Exit(1)
}

func genExpr(node *Node) {
	switch node.kind {
	case NodeNum:
		fmt.Printf("  mov $%d, %%rax\n", node.value)
		return
	case NodeNeg:
		genExpr(node.lhs)
		fmt.Println("  neg %rax")
		return
	case NodeDeref:
		genExpr(node.lhs)
		fmt.Println("  mov (%rax), %rax")
		return
	case NodeAddr:
		genAddr(node.lhs)
		return
	case NodeVar:
		genAddr(node)
		fmt.Println("  mov (%rax), %rax")
		return
	case NodeAsg:
		genAddr(node.lhs)
		push()
		genExpr(node.rhs)
		pop("%rdi")
		fmt.Println("  mov %rax, (%rdi)")
		return
	}
	genExpr(node.rhs)
	push()
	genExpr(node.lhs)
	pop("%rdi")
	switch node.kind {
	case NodeAdd:
		fmt.Println("  add %rdi, %rax")
		return
	case NodeSub:
		fmt.Println("  sub %rdi, %rax")
		return
	case NodeMul:
		fmt.Println("  imul %rdi, %rax")
		return
	case NodeDiv:
		fmt.Println("  cqo")
		fmt.Println("  idiv %rdi")
		return
	case NodeEql, NodeNeq, NodeLss, NodeLeq:
		fmt.Println("  cmp %rdi, %rax")
		switch node.kind {
		case NodeEql:
			fmt.Println("  sete %al")
		case NodeNeq:
			fmt.Println("  setne %al")
		case NodeLss:
			fmt.Println("  setl %al")
		case NodeLeq:
			fmt.Println("  setle %al")
		}
		fmt.Println("  movzb %al, %rax")
		return
	}
}
