package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// 定位错误所在位置
func locateError(offset int) {
	fmt.Fprintln(os.Stderr, source)
	fmt.Fprintf(os.Stderr, "%*s\033[31m^ \033[0m", offset, "")
}

// 词法分析 Tokenizer

// token -> 词法单元

type TokenKind int

const (
	TokenOperator TokenKind = iota // + - * /
	TokenLparen                    // (
	TokenRparen                    // )
	TokenNum                       // number
	TokenEof                       // EOF
)

type Token struct {
	kind   TokenKind // 词法单元的类型
	next   *Token    // 下一个词法单元
	value  int       // 类型为 TokenNum 时词素代表的值
	begin  int       // 词素的起始索引
	length int       // 词素的长度
	lexeme string    // 词素
}

func equal(token *Token, lexeme string) bool {
	return token.lexeme == lexeme
}

func skip(token *Token, lexeme string) *Token {
	if !equal(token, lexeme) {
		locateError(token.begin)
		fmt.Fprintf(os.Stderr, "\033[31mexpected \"%s\"\n\033[0m", lexeme)
		os.Exit(1)
	}
	return token.next
}

func NewToken(kind TokenKind, begin int, end int) *Token {
	return &Token{
		kind:   kind,
		next:   nil,
		value:  0,
		begin:  begin,
		length: end - begin,
		lexeme: source[begin:end],
	}
}

// 输入字符串
var source string

// 将所有词法单元组织成一个单链表并返回指向第一个词法单元的指针
func tokenize() *Token {
	head := Token{}
	curr := &head
	p := 0
	for p < len(source) {
		switch {
		case unicode.IsSpace(rune(source[p])):
			p++
		case unicode.IsDigit(rune(source[p])):
			q := p
			for p < len(source) && unicode.IsDigit(rune(source[p])) {
				p++
			}
			curr.next = NewToken(TokenNum, q, p)
			curr = curr.next
			value, err := strconv.Atoi(curr.lexeme)
			if err != nil {
				locateError((q + p) / 2)
				fmt.Fprintf(os.Stderr, "\033[31m%s\n\033[0m", err.Error()[len("strconv.Atoi: "):])
				os.Exit(1)
			}
			curr.value = value
		case source[p] == '+' || source[p] == '-' || source[p] == '*' || source[p] == '/':
			curr.next = NewToken(TokenOperator, p, p+1)
			curr = curr.next
			p++
		case source[p] == '(':
			curr.next = NewToken(TokenLparen, p, p+1)
			curr = curr.next
			p++
		case source[p] == ')':
			curr.next = NewToken(TokenRparen, p, p+1)
			curr = curr.next
			p++
		default:
			locateError(p)
			fmt.Fprintln(os.Stderr, "\033[31minvalid token\033[0m")
			os.Exit(1)
		}
	}
	curr.next = NewToken(TokenEof, p, p)
	return head.next
}

// 语法分析 Parser

type NodeKind int

const (
	NodeAdd NodeKind = iota
	NodeSub
	NodeMul
	NodeDiv
	NodeNeg
	NodeNum
)

type Node struct {
	kind  NodeKind
	lhs   *Node
	rhs   *Node
	value int // 当 Node 的类型是 NodeNum 时整数的值
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

// expr -> mul ( "+" mul | "-" mul )*
func expr(rest **Token, token *Token) (node *Node) {
	node = mul(&token, token)
	for {
		if equal(token, "+") {
			node = NewBinary(NodeAdd, node, mul(&token, token.next))
			continue
		}
		if equal(token, "-") {
			node = NewBinary(NodeSub, node, mul(&token, token.next))
			continue
		}
		*rest = token
		return
	}
}

// mul -> unary ( "*" unary | "/" unary )*
func mul(rest **Token, token *Token) (node *Node) {
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

// primary -> number | "(" expr ")"
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
	locateError(token.begin)
	fmt.Fprintln(os.Stderr, "\033[31mexpected an expression\033[0m")
	os.Exit(1)
	return
}

// 代码生成 Code generator

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
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "\033[31mexpected 2 arguments but got %d\n\033[0m", len(os.Args))
		os.Exit(1)
	}

	source = os.Args[1]
	token := tokenize()
	node := expr(&token, token)
	if token.kind != TokenEof {
		locateError(token.begin)
		fmt.Fprintln(os.Stderr, "\033[31mextra token\033[0m")
		os.Exit(1)
	}

	fmt.Println("  .globl main\nmain:")
	gen(node)
	fmt.Println("  ret")
}
