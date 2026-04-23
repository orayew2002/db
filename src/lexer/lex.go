package lexer

import (
	"strings"
	"unicode"
)

type TokenType string

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(i string) *Lexer {
	return &Lexer{input: i}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: EOF}
	}

	ch := l.input[l.pos]

	switch ch {
	case ',':
		l.pos++
		return Token{Type: COMMA, Value: ","}

	case ';':
		l.pos++
		return Token{Type: SEMICOLON, Value: ";"}

	case '>':
		l.pos++
		return Token{Type: GT, Value: ">"}

	case '=':
		l.pos++
		return Token{Type: EQ, Value: "="}

	case '\'':
		return l.readString()
	}

	if isLetter(ch) {
		return l.readIdentifier()
	}

	if isDigit(ch) {
		return l.readNumber()
	}

	l.pos++
	return Token{Type: ILLEGAL, Value: string(ch)}
}

func (l *Lexer) readString() Token {
	l.pos++

	start := l.pos

	for l.pos < len(l.input) && l.input[l.pos] != '\'' {
		l.pos++
	}

	value := l.input[start:l.pos]
	if l.pos < len(l.input) {
		l.pos++
	}

	return Token{
		Type:  STRING,
		Value: value,
	}
}

func (l *Lexer) readIdentifier() Token {
	start := l.pos

	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos])) {
		l.pos++
	}

	w := l.input[start:l.pos]
	word := strings.ToUpper(w)

	switch word {
	case "SELECT":
		return Token{Type: SELECT, Value: word}
	case "FROM":
		return Token{Type: FROM, Value: word}
	case "WHERE":
		return Token{Type: WHERE, Value: word}
	case "AND":
		return Token{Type: AND, Value: word}
	}

	return Token{Type: IDENT, Value: w}
}

func (l *Lexer) readNumber() Token {
	start := l.pos

	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}

	return Token{
		Type:  NUMBER,
		Value: l.input[start:l.pos],
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}
