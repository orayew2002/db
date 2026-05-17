package parser

import (
	"fmt"

	"github.com/orayew2002/db/src/lexer"
)

type Statement interface {
	isStatement()
}

type SelectStmt struct {
	Table   string
	Columns []string
}

func (SelectStmt) isStatement() {}

type Parser struct {
	l *lexer.Lexer
}

func (p *Parser) parseSelect() (*SelectStmt, error) {
	stmt := &SelectStmt{}

	for {
		tok, ok := p.l.NextToken()
		if !ok {
			return nil, fmt.Errorf("unexpected eof")
		}

		if tok.Type == lexer.KEYWORD && tok.Val == "FROM" {
			break
		}

		if tok.Type == lexer.SYMBOL && tok.Val == "*" {
			stmt.Columns = append(stmt.Columns, "*")
			continue
		}

		if tok.Type == lexer.IDENTIFIER {
			stmt.Columns = append(stmt.Columns, tok.Val)
			continue
		}

		if tok.Type == lexer.SYMBOL && tok.Val == "," {
			continue
		}

		return nil, fmt.Errorf("unexpected token: %s", tok.Val)
	}

	tok, ok := p.l.NextToken()
	if !ok {
		return nil, fmt.Errorf("missing table name")
	}

	if tok.Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("invalid table name")
	}

	stmt.Table = tok.Val

	return stmt, nil
}
