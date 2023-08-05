package main

import "fmt"

// Code generator

func push() {
	fmt.Println("  push %rax")
}

func pop(arg string) {
	fmt.Printf("  pop %s\n", arg)
}

func gen(node *Node) {
	switch node.kind {
	case NodeNum:
		fmt.Printf("  mov $%d, %%rax\n", node.value)
		return
	case NodeNeg:
		gen(node.lhs)
		fmt.Println("  neg %rax")
		return
	}
	gen(node.rhs)
	push()
	gen(node.lhs)
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
