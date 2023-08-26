package main

import (
	"fmt"
	"os"
)

type TypeKind int

const (
	TPINT TypeKind = iota // int
	TPPTR                 // pointer
)

type Type struct {
	kind TypeKind // Type kind
	base *Type    // Used if kind == TPPTR
	name *Token   // Declaration
}

func isint(t *Type) bool {
	return t.kind == TPINT
}

func ptrto(base *Type) *Type {
	return &Type{
		kind: TPPTR,
		base: base,
	}
}

var tpint = &Type{kind: TPINT}

func addtype(node *Node) {
	if node == nil || node.tp != nil {
		return
	}
	addtype(node.lhs)
	addtype(node.rhs)
	addtype(node.condition)
	addtype(node.thenBranch)
	addtype(node.elseBranch)
	addtype(node.initializer)
	addtype(node.increment)
	for n := node.body; n != nil; n = n.next {
		addtype(n)
	}
	switch node.kind {
	case NodeAdd, NodeSub, NodeMul, NodeDiv, NodeNeg, NodeAsg:
		node.tp = node.lhs.tp
		return
	case NodeEql, NodeNeq, NodeLss, NodeLeq, NodeNum:
		node.tp = tpint
		return
	case NodeVar:
		node.tp = node.variable.tp
	case NodeAddr:
		node.tp = ptrto(node.lhs.tp)
		return
	case NodeDeref:
		if node.lhs.tp.kind != TPPTR {
			locate(node.token.begin, node.token.length)
			fmt.Fprintln(os.Stderr, "\033[31minvalid pointer dereference\033[0m")
			os.Exit(1)
		}
		node.tp = node.lhs.tp.base
		return
	}
}
