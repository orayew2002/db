package parser

import (
	"errors"

	"github.com/orayew2002/db/src/db"
	"github.com/orayew2002/db/src/lexer"
)

type CreateTableStmt struct {
	Table   string
	Columns []db.ColDef
}

func (CreateTableStmt) isStatement() {}

func (p *Parser) parseCreateTable() (*CreateTableStmt, error) {
	stmt := &CreateTableStmt{}

	t, ok := p.l.NextToken()
	if !ok || t.Type != lexer.KEYWORD || t.Val != "TABLE" {
		return nil, errors.New("unexpected end of input creating")
	}

	t, ok = p.l.NextToken()
	if !ok || t.Type != lexer.IDENTIFIER {
		return nil, errors.New("unexpected end of input")
	}
	stmt.Table = t.Val

	t, ok = p.l.NextToken()
	if !ok || t.Type != lexer.SYMBOL || t.Val != "(" {
		return nil, errors.New("expected '(' after table name")
	}

	for {
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("unexpected end of input in column definition")
		}

		// end )
		if t.Type == lexer.SYMBOL && t.Val == ")" {
			break
		}

		// column name
		if t.Type != lexer.IDENTIFIER {
			return nil, errors.New("expected column name")
		}

		col := db.ColDef{
			Name: t.Val,
		}

		// type
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("unexpected end of input (expected column type)")
		}

		if t.Type != lexer.IDENTIFIER {
			return nil, errors.New("expected column type")
		}
		col.Type = t.Val

		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("unexpected end of input after column type")
		}

		if t.Type == lexer.SYMBOL && t.Val == "," {
			stmt.Columns = append(stmt.Columns, col)
			continue
		}

		if t.Type == lexer.SYMBOL && t.Val == ")" {
			stmt.Columns = append(stmt.Columns, col)
			break
		}

		return nil, errors.New("expected ',' or ')' after column definition")
	}

	return stmt, nil
}
