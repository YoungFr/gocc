package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// token -> 词法单元

type TokenKind int

const (
	TokenOperator TokenKind = iota
	TokenNum
	TokenEof
)

type Token struct {
	kind   TokenKind // 词法单元的类型
	next   *Token    // 下一个词法单元
	value  int       // 类型为 TokenNum 时词素代表的值
	begin  int       // 词素的起始索引
	length int       // 词素的长度
	lexeme string    // 词素
}

func locateError(offset int) {
	fmt.Fprintln(os.Stderr, source)
	fmt.Fprintf(os.Stderr, "%*s^ ", offset, "")
}

func equal(token *Token, lexeme string) bool {
	return token.lexeme == lexeme
}

func skip(token *Token, lexeme string) *Token {
	if !equal(token, lexeme) {
		locateError(token.begin)
		fmt.Fprintf(os.Stderr, "expected \"%s\"\n", lexeme)
		os.Exit(1)
	}
	return token.next
}

func getNumber(token *Token) int {
	if token.kind != TokenNum {
		locateError(token.begin)
		fmt.Fprintln(os.Stderr, "expected a number")
		os.Exit(1)
	}
	return token.value
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

func tokenize() *Token {
	head := Token{}
	curr := &head
	p := 0
	for p < len(source) {
		if unicode.IsSpace(rune(source[p])) {
			p++
			continue
		}
		if unicode.IsDigit(rune(source[p])) {
			q := p
			for p < len(source) && unicode.IsDigit(rune(source[p])) {
				p++
			}
			curr.next = NewToken(TokenNum, q, p)
			curr = curr.next
			value, _ := strconv.Atoi(curr.lexeme)
			curr.value = value
			continue
		}
		if source[p] == '+' || source[p] == '-' {
			curr.next = NewToken(TokenOperator, p, p+1)
			curr = curr.next
			p++
			continue
		}
		locateError(p)
		fmt.Fprintln(os.Stderr, "invalid token")
		os.Exit(1)
	}
	curr.next = NewToken(TokenEof, p, p)
	curr = curr.next
	curr.lexeme = "EOF"
	return head.next
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "expected 2 arguments but got %d", len(os.Args))
		os.Exit(1)
	}

	source = os.Args[1]
	token := tokenize()

	fmt.Printf("  .globl main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov $%d, %%rax\n", getNumber(token))
	token = token.next

	for token.kind != TokenEof {
		if equal(token, "+") {
			fmt.Printf("  add $%d, %%rax\n", getNumber(token.next))
			token = token.next.next
			continue
		}
		token = skip(token, "-")
		fmt.Printf("  sub $%d, %%rax\n", getNumber(token))
		token = token.next
	}

	fmt.Printf("  ret\n")
}
