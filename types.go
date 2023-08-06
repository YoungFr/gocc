package main

type TypeKind int

const (
	TPINT TypeKind = iota // int
	TPPTR                 // pointer
)

type Type struct {
	kind TypeKind // Type kind
	base *Type    // Used if kind == TPPTR
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
	case NodeEql, NodeNeq, NodeLss, NodeLeq, NodeVar, NodeNum:
		node.tp = tpint
		return
	case NodeAddr:
		node.tp = ptrto(node.lhs.tp)
		return
	case NodeDeref:
		if node.lhs.tp.kind == TPPTR {
			node.tp = node.lhs.tp.base
		} else {
			node.tp = tpint
		}
		return
	}
}
