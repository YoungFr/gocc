package main

import (
	"fmt"
	"os"
)

// This file contains a recursive descent parser for C.
//
// Most functions in this file are named after the symbols they are
// supposed to read from an input token list. For example, stmt() is
// responsible for reading a statement from a token list. The function
// then construct an AST node representing a statement.
//
// Each function conceptually returns two values, an AST node and
// remaining part of the input tokens. The remaining tokens are returned
// to the caller via a pointer argument.
//
// Input tokens are represented by a linked list. Unlike many recursive
// descent parsers, we don't have the notion of the "input token stream".
// Most parsing functions don't change the global state of the parser.
// So it is very easy to lookahead arbitrary number of tokens in this
// parser.

type NodeKind int

const (
	NodeAdd      NodeKind = iota // lhs + rhs
	NodeSub                      // lhs - rhs
	NodeMul                      // lhs * rhs
	NodeDiv                      // lhs / rhs
	NodeEql                      // lhs == rhs
	NodeNeq                      // lhs != rhs
	NodeLss                      // lhs < rhs
	NodeLeq                      // lhs <= rhs
	NodeAsg                      // lhs = rhs
	NodeNeg                      // - lhs
	NodeAddr                     // & lhs
	NodeDeref                    // * lhs
	NodeVar                      // variable
	NodeNum                      // number
	NodeExprStmt                 // expression statement
	NodeReturn                   // return statement
	NodeBlock                    // block statement
	NodeIf                       // if statement
	NodeFor                      // for or while statement
)

// Object represents a local variable.
type Object struct {
	next   *Object // Next variable
	name   string  // Variable's name
	tp     *Type   // Variable's type
	offset int     // Offset from RBP
}

// All local variable instances created during
// parsing are accumulated to this linked list.
var locals *Object

// NewLvar creates a new local variable instance and
// inserts it into the head of the `locals` linked list.
func NewLvar(name string, tp *Type) *Object {
	variable := &Object{
		next: locals,
		name: name,
		tp:   tp,
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

type Node struct {
	kind NodeKind // Node kind
	lhs  *Node    // Left-hand side
	rhs  *Node    // Right-hand side

	// int, pointer to int, ...
	tp *Type

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
	next *Node

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

// NewAdd In C, `+` operator is overloaded to perform the pointer arithmetic.
// If p is a pointer, p+n adds not n but sizeof(*p)*n to the value of p,
// so that p+n points to the location n elements (not bytes) ahead of p.
// In other words, we need to scale an integer value before adding to a
// pointer value.
func NewAdd(lhs *Node, rhs *Node, token *Token) *Node {
	addtype(lhs)
	addtype(rhs)
	// num + num
	if isint(lhs.tp) && isint(rhs.tp) {
		return NewBinary(NodeAdd, lhs, rhs, token)
	}
	// ptr + ptr
	if lhs.tp.base != nil && rhs.tp.base != nil {
		locate(token.begin, token.length)
		fmt.Fprintln(os.Stderr, "\033[31minvalid opreands\033[0m")
		os.Exit(1)
	}
	// num + ptr -> ptr + num
	if lhs.tp.base == nil && rhs.tp.base != nil {
		lhs, rhs = rhs, lhs
	}
	// ptr + num
	rhs = NewBinary(NodeMul, rhs, NewNumber(8, token), token)
	return NewBinary(NodeAdd, lhs, rhs, token)
}

// NewSub `-` operator is also overloaded to perform the pointer arithmetic.
func NewSub(lhs *Node, rhs *Node, token *Token) *Node {
	addtype(lhs)
	addtype(rhs)
	// num - num
	if isint(lhs.tp) && isint(rhs.tp) {
		return NewBinary(NodeSub, lhs, rhs, token)
	}
	// ptr - num
	if lhs.tp.base != nil && isint(rhs.tp) {
		rhs = NewBinary(NodeMul, rhs, NewNumber(8, token), token)
		addtype(rhs)
		node := NewBinary(NodeSub, lhs, rhs, token)
		node.tp = lhs.tp
		return node
	}
	// num - ptr
	if isint(lhs.tp) && rhs.tp.base != nil {
		locate(token.begin, token.length)
		fmt.Fprintln(os.Stderr, "\033[31minvalid opreands\033[0m")
		os.Exit(1)
	}
	// ptr - ptr
	node := NewBinary(NodeSub, lhs, rhs, token)
	node.tp = tpint
	return NewBinary(NodeDiv, node, NewNumber(8, token), token)
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

type Function struct {
	body      *Node
	locals    *Object
	stackSize int
}

// program -> stmt* EOF
func parse(token *Token) *Function {
	head := Node{}
	curr := &head
	for token.kind != EOF {
		curr.next = stmt(&token, token)
		curr = curr.next
		addtype(curr)
	}
	program := &Function{
		body:   head.next,
		locals: locals,
	}
	return program
}

func equal(token *Token, lexeme string) bool {
	return token.lexeme == lexeme
}

func skip(token *Token, lexeme string) *Token {
	if !equal(token, lexeme) {
		locate(token.begin, token.length)
		fmt.Fprintf(os.Stderr, "\033[31mexpected \"%s\"\n\033[0m", lexeme)
		os.Exit(1)
	}
	return token.next
}

// stmt -> "return" expr ";"
// -->   | "{" block
// -->   | "if" "(" expr ")" stmt ( "else" stmt )?
// -->   | "for" "(" exprStmt expr? ";" expr? ")" stmt
// -->   | "while" "(" expr ")" stmt
// -->   | exprStmt
// -->   | declaration
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
	if equal(token, "int") {
		return declaration(rest, token)
	}
	return exprStmt(rest, token)
}

// declspec -> "int"
func declspec(rest **Token, token *Token) *Type {
	*rest = skip(token, "int")
	return tpint
}

func consume(rest **Token, token *Token, lexeme string) bool {
	if equal(token, lexeme) {
		*rest = token.next
		return true
	}
	*rest = token
	return false
}

// declarator -> "*"* ident
func declarator(rest **Token, token *Token, tp *Type) *Type {
	for consume(&token, token, "*") {
		tp = ptrto(tp)
	}
	if token.kind != IDENT {
		locate(token.begin, token.length)
		fmt.Fprintln(os.Stderr, "\033[31mexpected a variable name\033[0m")
		os.Exit(1)
	}
	tp.name = token
	*rest = token.next
	return tp
}

func getIdent(token *Token) string {
	if token.kind != IDENT {
		locate(token.begin, token.length)
		fmt.Fprintln(os.Stderr, "\033[31mexpected an identifier\033[0m")
		os.Exit(1)
	}
	return token.lexeme
}

// declaration -> declspec (declarator ( "=" expr )?) ( "," declarator ( "=" expr )?)* ";"
func declaration(rest **Token, token *Token) *Node {
	baseType := declspec(&token, token)
	head := Node{}
	curr := &head
	var tp *Type
	var init *Node
	var variable *Object
	tp = declarator(&token, token, baseType)
	variable = NewLvar(getIdent(tp.name), tp)
	if equal(token, "=") {
		token = skip(token, "=")
		init = expr(&token, token)
	}
	if init == nil {
		curr.next = NewUnary(NodeExprStmt, NewVar(variable, tp.name), token)
		curr = curr.next
	} else {
		curr.next = NewUnary(NodeExprStmt, NewBinary(NodeAsg, NewVar(variable, tp.name), init, token), token)
		curr = curr.next
	}
	for token.kind != EOF && !equal(token, ";") {
		token = skip(token, ",")
		tp = declarator(&token, token, baseType)
		variable = NewLvar(getIdent(tp.name), tp)
		if !equal(token, "=") {
			init = nil
		} else {
			token = skip(token, "=")
			init = expr(&token, token)
		}
		if init == nil {
			curr.next = NewUnary(NodeExprStmt, NewVar(variable, tp.name), token)
			curr = curr.next
		} else {
			curr.next = NewUnary(NodeExprStmt, NewBinary(NodeAsg, NewVar(variable, tp.name), init, token), token)
			curr = curr.next
		}
	}
	node := NewNode(NodeBlock, token)
	node.body = head.next
	*rest = skip(token, ";")
	return node
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

// addsub -> muldiv ( "+" muldiv | "-" muldiv )*
func addsub(rest **Token, token *Token) (node *Node) {
	node = muldiv(&token, token)
	for {
		start := token
		if equal(token, "+") {
			node = NewAdd(node, muldiv(&token, token.next), start)
			continue
		}
		if equal(token, "-") {
			node = NewSub(node, muldiv(&token, token.next), start)
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

// unary -> ( "+" | "-" | "*" | "&" ) unary
// -->    | primary
func unary(rest **Token, token *Token) *Node {
	if equal(token, "+") {
		return unary(rest, token.next)
	}
	if equal(token, "-") {
		return NewUnary(NodeNeg, unary(rest, token.next), token)
	}
	if equal(token, "*") {
		return NewUnary(NodeDeref, unary(rest, token.next), token)
	}
	if equal(token, "&") {
		return NewUnary(NodeAddr, unary(rest, token.next), token)
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
			locate(token.begin, token.length)
			fmt.Fprintln(os.Stderr, "\033[31mundefined variable\033[0m")
			os.Exit(1)
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
