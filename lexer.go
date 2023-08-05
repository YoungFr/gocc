package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// Tokenizer

type TokenKind int

const (
	TokenAdd    TokenKind = iota // +
	TokenSub                     // -
	TokenMul                     // *
	TokenDiv                     // /
	TokenAsg                     // =
	TokenEql                     // ==
	TokenNot                     // !
	TokenNeq                     // !=
	TokenLss                     // <
	TokenLeq                     // <=
	TokenGtr                     // >
	TokenGeq                     // >=
	TokenLparen                  // (
	TokenRparen                  // )
	TokenSemi                    // ;
	TokenNum                     // number
	TokenEof                     // EOF
)

type Token struct {
	kind   TokenKind // Token kind
	next   *Token    // Next token
	value  int       // If kind == TokenNum, its value
	begin  int       // Starting index of lexeme
	length int       // Length of lexeme
	lexeme string    // A substring in the source that matches the pattern for a token
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

func lookahead(p int, expected ...byte) int {
	n := len(expected)
	if p+n >= len(source) {
		return -1
	}
	toklen := 1
	for i := 1; i <= n; i++ {
		if source[p+i] == expected[i-1] {
			toklen++
		} else {
			return toklen
		}
	}
	return toklen
}

// Create a tokens list
// Return a pointer to the first token
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
		case source[p] == '+':
			switch {
			case lookahead(p, '+') == 2:
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(TokenAdd, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '-':
			switch {
			case lookahead(p, '>') == 2:
			case lookahead(p, '-') == 2:
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(TokenSub, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '*':
			switch {
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(TokenMul, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '/':
			switch {
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(TokenDiv, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '=':
			switch {
			case lookahead(p, '=') == 2:
				curr.next = NewToken(TokenEql, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(TokenAsg, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '!':
			switch {
			case lookahead(p, '=') == 2:
				curr.next = NewToken(TokenNeq, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(TokenNot, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '<':
			switch {
			case lookahead(p, '<', '=') == 3:
			case lookahead(p, '<') == 2:
			case lookahead(p, '=') == 2:
				curr.next = NewToken(TokenLeq, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(TokenLss, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '>':
			switch {
			case lookahead(p, '>', '=') == 3:
			case lookahead(p, '>') == 2:
			case lookahead(p, '=') == 2:
				curr.next = NewToken(TokenGeq, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(TokenGtr, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '(':
			curr.next = NewToken(TokenLparen, p, p+1)
			curr = curr.next
			p++
		case source[p] == ')':
			curr.next = NewToken(TokenRparen, p, p+1)
			curr = curr.next
			p++
		case source[p] == ';':
			curr.next = NewToken(TokenSemi, p, p+1)
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
