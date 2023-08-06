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
	ADD      TokenKind = iota // +
	SUB                       // -
	ASTERISK                  // *
	DIV                       // /
	ASG                       // =
	EQL                       // ==
	NOT                       // !
	NEQ                       // !=
	LSS                       // <
	LEQ                       // <=
	GTR                       // >
	GEQ                       // >=
	LPAREN                    // (
	RPAREN                    // )
	LBRACE                    // {
	RBRACE                    // }
	SEMI                      // ;
	IDENT                     // identifier
	RETURN                    // return
	IF                        // if
	ELSE                      // else
	FOR                       // for
	WHILE                     // while
	NUM                       // number
	EOF                       // EOF
)

type Token struct {
	kind   TokenKind // Token kind
	next   *Token    // Next token
	value  int       // If kind == NUM, its value
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
			curr.next = NewToken(NUM, q, p)
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
				curr.next = NewToken(ADD, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '-':
			switch {
			case lookahead(p, '>') == 2:
			case lookahead(p, '-') == 2:
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(SUB, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '*':
			switch {
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(ASTERISK, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '/':
			switch {
			case lookahead(p, '=') == 2:
			case lookahead(p) == 1:
				curr.next = NewToken(DIV, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '=':
			switch {
			case lookahead(p, '=') == 2:
				curr.next = NewToken(EQL, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(ASG, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '!':
			switch {
			case lookahead(p, '=') == 2:
				curr.next = NewToken(NEQ, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(NOT, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '<':
			switch {
			case lookahead(p, '<', '=') == 3:
			case lookahead(p, '<') == 2:
			case lookahead(p, '=') == 2:
				curr.next = NewToken(LEQ, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(LSS, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '>':
			switch {
			case lookahead(p, '>', '=') == 3:
			case lookahead(p, '>') == 2:
			case lookahead(p, '=') == 2:
				curr.next = NewToken(GEQ, p, p+2)
				p += 2
			case lookahead(p) == 1:
				curr.next = NewToken(GTR, p, p+1)
				p += 1
			}
			curr = curr.next
		case source[p] == '(':
			curr.next = NewToken(LPAREN, p, p+1)
			curr = curr.next
			p++
		case source[p] == ')':
			curr.next = NewToken(RPAREN, p, p+1)
			curr = curr.next
			p++
		case source[p] == '{':
			curr.next = NewToken(LBRACE, p, p+1)
			curr = curr.next
			p++
		case source[p] == '}':
			curr.next = NewToken(RBRACE, p, p+1)
			curr = curr.next
			p++
		case source[p] == ';':
			curr.next = NewToken(SEMI, p, p+1)
			curr = curr.next
			p++
		case isLetter(source[p]):
			q := p
			for p < len(source) && (isLetter(source[p]) || isDigit(source[p])) {
				p++
			}
			if kind, ok := keywords[source[q:p]]; ok {
				curr.next = NewToken(kind, q, p)
			} else {
				curr.next = NewToken(IDENT, q, p)
			}
			curr = curr.next
		default:
			locateError(p)
			fmt.Fprintln(os.Stderr, "\033[31minvalid token\033[0m")
			os.Exit(1)
		}
	}
	curr.next = NewToken(EOF, p, p)
	return head.next
}

var keywords = map[string]TokenKind{
	"return": RETURN,
	"if":     IF,
	"else":   ELSE,
	"for":    FOR,
	"while":  WHILE,
}

func isLetter(c byte) bool {
	return (c == '_') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
