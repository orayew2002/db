package parser

import (
	"fmt"

	"github.com/orayew2002/db/src/lexer"
)

func New(l *lexer.Lexer) *Parser {
	return &Parser{l: l}
}

func (p *Parser) Parse() (Statement, error) {
	tok, ok := p.l.NextToken()
	if !ok {
		return nil, fmt.Errorf("empty query")
	}

	if tok.Type == lexer.KEYWORD && tok.Val == "SELECT" {
		return p.parseSelect()
	}

	if tok.Type == lexer.KEYWORD && tok.Val == "INSERT" {
		return p.parseInsert()
	}

	return nil, fmt.Errorf("unknown statement: %s", tok.Val)
}
