package parser

import (
	"errors"

	"github.com/orayew2002/db/src/lexer"
)

type UpdateStmt struct {
	Table string
	Set   map[string]any
	Where *WhereClause
}

func (UpdateStmt) isStatement() {}

func (p *Parser) parseUpdate() (*UpdateStmt, error) {
	stmt := &UpdateStmt{
		Set: make(map[string]any),
	}

	// table name
	t, ok := p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input")
	}
	if t.Type != lexer.IDENTIFIER {
		return nil, errors.New("expected table name")
	}
	stmt.Table = t.Val

	// SET
	t, ok = p.l.NextToken()
	if !ok || t.Type != lexer.KEYWORD || t.Val != "SET" {
		return nil, errors.New("expected SET")
	}

	// SET clause
	for {
		// column
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("unexpected end of input (column)")
		}
		if t.Type != lexer.IDENTIFIER {
			return nil, errors.New("expected column name")
		}
		col := t.Val

		// =
		t, ok = p.l.NextToken()
		if !ok || t.Type != lexer.SYMBOL || t.Val != "=" {
			return nil, errors.New("expected '='")
		}

		// value
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("unexpected end of input (value)")
		}
		stmt.Set[col] = t.Val

		// next: ',' OR WHERE OR end
		t, ok = p.l.NextToken()
		if !ok {
			return stmt, nil
		}

		if t.Type == lexer.SYMBOL && t.Val == "," {
			continue
		}

		if t.Type == lexer.KEYWORD && t.Val == "WHERE" {
			break
		}

		// no WHERE → finished UPDATE
		return stmt, nil
	}

	// OPTIONAL WHERE clause
	t, ok = p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input in WHERE")
	}
	if t.Type != lexer.IDENTIFIER {
		return nil, errors.New("expected WHERE column")
	}

	where := &WhereClause{
		Left: t.Val,
	}

	// operator
	t, ok = p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input (operator)")
	}
	if t.Type != lexer.SYMBOL {
		return nil, errors.New("expected operator")
	}
	where.Operator = t.Val

	// right side
	t, ok = p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input (right side)")
	}
	where.Right = t.Val

	stmt.Where = where

	return stmt, nil
}
