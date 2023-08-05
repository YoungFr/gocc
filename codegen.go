package main

import (
	"fmt"
	"os"
)

// Code generator

func push() {
	fmt.Println("  push %rax")
}

func pop(arg string) {
	fmt.Printf("  pop %s\n", arg)
}

func gen(node *Node) {
	fmt.Println("  .globl main")
	fmt.Println("main:")
	fmt.Println("  push %rbp")
	fmt.Println("  mov %rsp, %rbp")
	// 208 == ('z' - 'a' + 1) * 8, it's the stack size for
	// all possible single-letter 64 bit integer variables.
	fmt.Println("  sub $208, %rsp")
	for n := node; n != nil; n = n.next {
		genStmt(n)
	}
	fmt.Println("  mov %rbp, %rsp")
	fmt.Println("  pop %rbp")
	fmt.Println("  ret")
}

func genStmt(node *Node) {
	if node.kind == NodeExprStmt {
		genExpr(node.lhs)
		return
	}
	fmt.Fprintln(os.Stderr, "invalid statement")
	os.Exit(1)
}

func genAddr(node *Node) {
	if node.kind == NodeVar {
		offset := (node.name[0] - 'a' + 1) * 8
		fmt.Printf("  lea %d(%%rbp), %%rax\n", -offset)
		return
	}
	fmt.Fprintln(os.Stderr, "not a lvalue")
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
