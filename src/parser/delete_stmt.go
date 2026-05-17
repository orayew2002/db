package parser

import (
	"errors"

	"github.com/orayew2002/db/src/lexer"
)

type DeleteStmt struct {
	Table string
	Where *WhereClause
}

func (DeleteStmt) isStatement() {}

func (p *Parser) parseDelete() (*DeleteStmt, error) {
	stmt := &DeleteStmt{}

	// FROM
	t, ok := p.l.NextToken()
	if !ok || t.Type != lexer.KEYWORD || t.Val != "FROM" {
		return nil, errors.New("wrong query: expected FROM")
	}

	// table name
	t, ok = p.l.NextToken()
	if !ok || t.Type != lexer.IDENTIFIER {
		return nil, errors.New("wrong query: invalid table name")
	}
	stmt.Table = t.Val

	// optional WHERE
	t, ok = p.l.NextToken()
	if ok && t.Type == lexer.KEYWORD && t.Val == "WHERE" {

		clause := &WhereClause{}

		// left side
		t, ok = p.l.NextToken()
		if !ok || t.Type != lexer.IDENTIFIER {
			return nil, errors.New("wrong query: invalid where left side")
		}
		clause.Left = t.Val

		// operator
		t, ok = p.l.NextToken()
		if !ok || t.Type != lexer.SYMBOL {
			return nil, errors.New("wrong query: invalid where operator")
		}
		clause.Operator = t.Val

		// right side
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("wrong query: invalid where right side")
		}
		clause.Right = t.Val

		stmt.Where = clause
	}

	return stmt, nil
}
