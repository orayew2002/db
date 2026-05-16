package lexer

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	KEYWORD TokenType = iota
	SYMBOL
	NUMBER
	STRING
	IDENTIFIER
	UNDEFINED
)

var KEYWORDS = map[string]bool{
	"SELECT": true,
	"FROM":   true,
	"WHERE":  true,
	"DELETE": true,
	"UPDATE": true,
	"SET":    true,
	"VALUES": true,
}

type Token struct {
	Type TokenType
	Val  string
}

type Lexer struct {
	str string
	pos int
}

func New(str string) *Lexer {
	return &Lexer{str: str}
}

func (l *Lexer) NextToken() (Token, bool) {
	l.skipSpaces()

	if l.pos >= len(l.str) {
		return Token{}, false
	}

	ch := l.str[l.pos]

	if ch == '\'' {
		l.pos++
		start := l.pos

		for l.pos < len(l.str) && l.str[l.pos] != '\'' {
			l.pos++
		}

		val := l.str[start:l.pos]

		if l.pos < len(l.str) {
			l.pos++
		}

		return Token{Type: STRING, Val: val}, true
	}

	if unicode.IsDigit(rune(ch)) {
		start := l.pos

		for l.pos < len(l.str) && unicode.IsDigit(rune(l.str[l.pos])) {
			l.pos++
		}

		return Token{
			Type: NUMBER,
			Val:  l.str[start:l.pos],
		}, true
	}

	if isSymbol(ch) {
		l.pos++
		return Token{Type: SYMBOL, Val: string(ch)}, true
	}

	start := l.pos

	for l.pos < len(l.str) &&
		!isSpace(l.str[l.pos]) &&
		!isSymbol(l.str[l.pos]) {

		l.pos++
	}

	val := l.str[start:l.pos]
	upper := strings.ToUpper(val)

	if KEYWORDS[upper] {
		return Token{Type: KEYWORD, Val: upper}, true
	}

	return Token{Type: IDENTIFIER, Val: val}, true
}

func (l *Lexer) skipSpaces() {
	for l.pos < len(l.str) && isSpace(l.str[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) GetTokens() []Token {
	var tokens []Token

	for {
		tok, ok := l.NextToken()
		if !ok {
			break
		}
		tokens = append(tokens, tok)
	}

	return tokens
}
