package parser

import (
	"fmt"

	"github.com/orayew2002/db/src/lexer"
)

type SelectStmt struct {
	Table   string
	Columns []string
	Where   *WhereClause
}

func (SelectStmt) isStatement() {}

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

	if tok, ok = p.l.NextToken(); ok && tok.Type == lexer.KEYWORD && tok.Val == "WHERE" {
		var whereClause WhereClause

		token, ok := p.l.NextToken()
		if !ok || token.Type != lexer.IDENTIFIER {
			return nil, fmt.Errorf("error query left side")
		}
		whereClause.Left = token.Val

		token, ok = p.l.NextToken()
		if !ok || token.Type != lexer.SYMBOL {
			return nil, fmt.Errorf("error query where clause symbol")
		}
		whereClause.Operator = token.Val

		token, ok = p.l.NextToken()
		if !ok || (token.Type != lexer.STRING && token.Type != lexer.NUMBER) {
			return nil, fmt.Errorf("error query where clause right side")
		}
		whereClause.Right = token.Val

		stmt.Where = &whereClause
	}

	return stmt, nil
}
