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
	NodeAsg                      // lhs = rhs
	NodeExprStmt                 // expression statement
	NodeReturn                   // return statement
	NodeBlock                    // block statement
	NodeIf                       // if statement
	NodeFor                      // for or while statement
	NodeVar                      // variable
	NodeNum                      // number
)

// Object represents a local variable.
type Object struct {
	next   *Object // Next object
	name   string  // Variable name
	offset int     // Offset from RBP
}

// All local variable instances created during
// parsing are accumulated to this list.
var locals *Object

// NewLvar creates a new local variable instance
// and inserts it into the head of the `locals` list.
func NewLvar(name string) *Object {
	variable := &Object{
		next: locals,
		name: name,
	}
	locals = variable
	return variable
}

// Find a local variable by name.
func findVar(token *Token) *Object {
	for v := locals; v != nil; v = v.next {
		if v.name == token.lexeme {
			return v
		}
	}
	return nil
}

type Function struct {
	body      *Node
	locals    *Object
	stackSize int
}

type Node struct {
	kind NodeKind // Node kind
	next *Node    // Next node
	lhs  *Node    // Left-hand side
	rhs  *Node    // Right-hand side

	// Representative token
	token *Token

	// Used if kind == NodeIf | NodeFor
	condition  *Node
	thenBranch *Node

	// Used if kind == NodeIf
	elseBranch *Node

	// Used if kind == NodeFor
	initializer *Node
	increment   *Node

	// Used if kind == NodeBlock
	// The list of statements within the block
	body *Node

	// Used if kind == NodeVar
	// Variable's struct representation
	variable *Object

	// Used if kind == NodeNum
	value int
}

func NewNode(kind NodeKind, token *Token) *Node {
	return &Node{
		kind:  kind,
		token: token,
	}
}

func NewBinary(kind NodeKind, lhs *Node, rhs *Node, token *Token) *Node {
	node := NewNode(kind, token)
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func NewNumber(value int, token *Token) *Node {
	node := NewNode(NodeNum, token)
	node.value = value
	return node
}

func NewUnary(kind NodeKind, expr *Node, token *Token) *Node {
	node := NewNode(kind, token)
	node.lhs = expr
	return node
}

func NewVar(variable *Object, token *Token) *Node {
	node := NewNode(NodeVar, token)
	node.variable = variable
	return node
}

// program -> stmt* EOF
func parse(token *Token) *Function {
	head := Node{}
	curr := &head
	for token.kind != EOF {
		curr.next = stmt(&token, token)
		curr = curr.next
	}
	program := &Function{
		body:   head.next,
		locals: locals,
	}
	return program
}

// stmt -> "return" expr ";"
// -->   | "{" block
// -->   | "if" "(" expr ")" stmt ( "else" stmt )?
// -->   | "for" "(" exprStmt expr? ";" expr? ")" stmt
// -->   | "while" "(" expr ")" stmt
// -->   | exprStmt
func stmt(rest **Token, token *Token) *Node {
	if equal(token, "return") {
		node := NewUnary(NodeReturn, expr(&token, token.next), token)
		*rest = skip(token, ";")
		return node
	}
	if equal(token, "{") {
		return block(rest, token.next)
	}
	if equal(token, "if") {
		node := NewNode(NodeIf, token)
		token = skip(token.next, "(")
		node.condition = expr(&token, token)
		token = skip(token, ")")
		node.thenBranch = stmt(&token, token)
		if equal(token, "else") {
			node.elseBranch = stmt(&token, token.next)
		}
		*rest = token
		return node
	}
	if equal(token, "for") {
		node := NewNode(NodeFor, token)
		token = skip(token.next, "(")
		node.initializer = exprStmt(&token, token)
		if !equal(token, ";") {
			node.condition = expr(&token, token)
		}
		token = skip(token, ";")
		if !equal(token, ")") {
			node.increment = expr(&token, token)
		}
		token = skip(token, ")")
		node.thenBranch = stmt(&token, token)
		*rest = token
		return node
	}
	if equal(token, "while") {
		node := NewNode(NodeFor, token)
		token = skip(token.next, "(")
		node.condition = expr(&token, token)
		token = skip(token, ")")
		node.thenBranch = stmt(&token, token)
		*rest = token
		return node
	}
	return exprStmt(rest, token)
}

// block -> stmt* "}"
func block(rest **Token, token *Token) *Node {
	node := NewNode(NodeBlock, token)
	// statements' linked list
	head := Node{}
	curr := &head
	for token.kind != EOF && !equal(token, "}") {
		curr.next = stmt(&token, token)
		curr = curr.next
	}
	node.body = head.next
	*rest = skip(token, "}")
	return node
}

// exprStmt -> expr? ";"
func exprStmt(rest **Token, token *Token) *Node {
	if equal(token, ";") {
		*rest = token.next
		return NewNode(NodeBlock, token)
	}
	node := NewUnary(NodeExprStmt, expr(&token, token), token)
	*rest = skip(token, ";")
	return node
}

// expr -> assign
func expr(rest **Token, token *Token) *Node {
	return assign(rest, token)
}

// assign -> equality ( "=" assign )?
func assign(rest **Token, token *Token) (node *Node) {
	node = equality(&token, token)
	if equal(token, "=") {
		node = NewBinary(NodeAsg, node, assign(&token, token.next), token)
	}
	*rest = token
	return
}

// equality -> relational ( "==" relational | "!=" relational )*
func equality(rest **Token, token *Token) (node *Node) {
	node = relational(&token, token)
	for {
		start := token
		if equal(token, "==") {
			node = NewBinary(NodeEql, node, relational(&token, token.next), start)
			continue
		}
		if equal(token, "!=") {
			node = NewBinary(NodeNeq, node, relational(&token, token.next), start)
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
		start := token
		if equal(token, "<") {
			node = NewBinary(NodeLss, node, addsub(&token, token.next), start)
			continue
		}
		if equal(token, "<=") {
			node = NewBinary(NodeLeq, node, addsub(&token, token.next), start)
			continue
		}
		if equal(token, ">") {
			node = NewBinary(NodeLss, addsub(&token, token.next), node, start)
			continue
		}
		if equal(token, ">=") {
			node = NewBinary(NodeLeq, addsub(&token, token.next), node, start)
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
		start := token
		if equal(token, "+") {
			node = NewBinary(NodeAdd, node, muldiv(&token, token.next), start)
			continue
		}
		if equal(token, "-") {
			node = NewBinary(NodeSub, node, muldiv(&token, token.next), start)
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
		start := token
		if equal(token, "*") {
			node = NewBinary(NodeMul, node, unary(&token, token.next), start)
			continue
		}
		if equal(token, "/") {
			node = NewBinary(NodeDiv, node, unary(&token, token.next), start)
			continue
		}
		*rest = token
		return
	}
}

// unary -> ( "+" | "-" ) unary
// -->    | primary
func unary(rest **Token, token *Token) *Node {
	if equal(token, "+") {
		return unary(rest, token.next)
	}
	if equal(token, "-") {
		return NewUnary(NodeNeg, unary(rest, token.next), token)
	}
	return primary(rest, token)
}

// primary -> "(" expr ")"
// -->      | number
// -->      | ident
func primary(rest **Token, token *Token) (node *Node) {
	if equal(token, "(") {
		node = expr(&token, token.next)
		*rest = skip(token, ")")
		return
	}
	if token.kind == NUM {
		node = NewNumber(token.value, token)
		*rest = token.next
		return
	}
	if token.kind == IDENT {
		variable := findVar(token)
		if variable == nil {
			variable = NewLvar(token.lexeme)
		}
		*rest = token.next
		node = NewVar(variable, token)
		return
	}
	locate(token.begin, token.length)
	fmt.Fprintln(os.Stderr, "\033[31mexpected an expression\033[0m")
	os.Exit(1)
	return
}
