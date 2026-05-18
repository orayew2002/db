package parser

import (
	"fmt"

	"github.com/orayew2002/db/src/lexer"
)

type WhereClause struct {
	Left     string
	Operator string
	Right    string
}

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

	if tok.Type == lexer.KEYWORD && tok.Val == "DELETE" {
		return p.parseDelete()
	}

	if tok.Type == lexer.KEYWORD && tok.Val == "CREATE" {
		return p.parseCreateTable()
	}

	if tok.Type == lexer.KEYWORD && tok.Val == "UPDATE" {
		return p.parseUpdate()
	}

	if tok.Type == lexer.KEYWORD && tok.Val == "DROP" {
		return p.parseDropTable()
	}

	return nil, fmt.Errorf("unknown statement: %s", tok.Val)
}
