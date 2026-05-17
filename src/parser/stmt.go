package parser

import (
	"github.com/orayew2002/db/src/lexer"
)

type Statement interface {
	isStatement()
}

type Parser struct {
	l *lexer.Lexer
}
