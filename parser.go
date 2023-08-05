package main

import (
	"fmt"
	"os"
)

// Parser

type NodeKind int

const (
	NodeAdd      NodeKind = iota // lhs + rhs
	NodeSub                      // lhs - rhs
	NodeMul                      // lhs * rhs
	NodeDiv                      // lhs / rhs
	NodeNeg                      // - lhs
	NodeEql                      // lhs == rhs
	NodeNeq                      // lhs != rhs
	NodeLss                      // lhs < rhs
	NodeLeq                      // lhs <= rhs
	NodeExprStmt                 // lhs ; (expression statement)
	NodeAsg                      // lhs = rhs
	NodeVar                      // variable
	NodeNum                      // number
)

type Node struct {
	kind  NodeKind // Node kind
	next  *Node    // Next node
	lhs   *Node    // Left-hand side
	rhs   *Node    // Right-hand side
	name  string   // If kind == NodeVar, variable's name
	value int      // If kind == NodeNum, its value
}

func NewNode(kind NodeKind) *Node {
	return &Node{kind: kind}
}

func NewBinary(kind NodeKind, lhs *Node, rhs *Node) *Node {
	node := NewNode(kind)
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func NewNumber(value int) *Node {
	node := NewNode(NodeNum)
	node.value = value
	return node
}

func NewUnary(kind NodeKind, expr *Node) *Node {
	node := NewNode(kind)
	node.lhs = expr
	return node
}

func NewVar(name string) *Node {
	node := NewNode(NodeVar)
	node.name = name
	return node
}

func parse(token *Token) *Node {
	head := Node{}
	curr := &head
	for token.kind != TokenEof {
		curr.next = stmt(&token, token)
		curr = curr.next
	}
	return head.next
}

// stmt -> exprStmt
func stmt(rest **Token, token *Token) *Node {
	return exprStmt(rest, token)
}

// exprStmt -> expr ";"
func exprStmt(rest **Token, token *Token) (node *Node) {
	node = NewUnary(NodeExprStmt, expr(&token, token))
	*rest = skip(token, ";")
	return
}

// expr -> assign
func expr(rest **Token, token *Token) *Node {
	return assign(rest, token)
}

// assign -> equality ( "=" assign )?
func assign(rest **Token, token *Token) (node *Node) {
	node = equality(&token, token)
	if equal(token, "=") {
		node = NewBinary(NodeAsg, node, assign(&token, token.next))
	}
	*rest = token
	return
}

// equality -> relational ( "==" relational | "!=" relational )*
func equality(rest **Token, token *Token) (node *Node) {
	node = relational(&token, token)
	for {
		if equal(token, "==") {
			node = NewBinary(NodeEql, node, relational(&token, token.next))
			continue
		}
		if equal(token, "!=") {
			node = NewBinary(NodeNeq, node, relational(&token, token.next))
			continue
		}
		*rest = token
		return
	}
}

// relational -> addsub ( "<" addsub | "<=" addsub | ">" addsub | ">=" addsub )*
func relational(rest **Token, token *Token) (node *Node) {
	node = addsub(&token, token)
	for {
		if equal(token, "<") {
			node = NewBinary(NodeLss, node, addsub(&token, token.next))
			continue
		}
		if equal(token, "<=") {
			node = NewBinary(NodeLeq, node, addsub(&token, token.next))
			continue
		}
		if equal(token, ">") {
			node = NewBinary(NodeLss, addsub(&token, token.next), node)
			continue
		}
		if equal(token, ">=") {
			node = NewBinary(NodeLeq, addsub(&token, token.next), node)
			continue
		}
		*rest = token
		return
	}
}

// addsub -> muldiv ( "+" muldiv | "/" muldiv )*
func addsub(rest **Token, token *Token) (node *Node) {
	node = muldiv(&token, token)
	for {
		if equal(token, "+") {
			node = NewBinary(NodeAdd, node, muldiv(&token, token.next))
			continue
		}
		if equal(token, "-") {
			node = NewBinary(NodeSub, node, muldiv(&token, token.next))
			continue
		}
		*rest = token
		return
	}
}

// muldiv -> unary ( "*" unary | "/" unary )*
func muldiv(rest **Token, token *Token) (node *Node) {
	node = unary(&token, token)
	for {
		if equal(token, "*") {
			node = NewBinary(NodeMul, node, unary(&token, token.next))
			continue
		}
		if equal(token, "/") {
			node = NewBinary(NodeDiv, node, unary(&token, token.next))
			continue
		}
		*rest = token
		return
	}
}

// unary -> ( "+" | "-" ) unary | primary
func unary(rest **Token, token *Token) *Node {
	if equal(token, "+") {
		return unary(rest, token.next)
	}
	if equal(token, "-") {
		return NewUnary(NodeNeg, unary(rest, token.next))
	}
	return primary(rest, token)
}

// primary -> number | "(" expr ")" | ident
func primary(rest **Token, token *Token) (node *Node) {
	if equal(token, "(") {
		node = expr(&token, token.next)
		*rest = skip(token, ")")
		return
	}
	if token.kind == TokenNum {
		node = NewNumber(token.value)
		*rest = token.next
		return
	}
	if token.kind == TokenIdent {
		node = NewVar(token.lexeme)
		*rest = token.next
		return
	}
	locateError(token.begin)
	fmt.Fprintln(os.Stderr, "\033[31mexpected an expression\033[0m")
	os.Exit(1)
	return
}
