package parser

import (
	"errors"

	"github.com/orayew2002/db/src/lexer"
)

type InsertStmt struct {
	Table   string
	Columns []string
	Values  []any
}

func (InsertStmt) isStatement() {}

func (p *Parser) parseInsert() (*InsertStmt, error) {
	stmt := &InsertStmt{}

	// INTO
	t, ok := p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input")
	}
	if t.Type != lexer.KEYWORD || t.Val != "INTO" {
		return nil, errors.New("expected INTO")
	}

	// table name
	t, ok = p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input")
	}
	if t.Type != lexer.IDENTIFIER {
		return nil, errors.New("expected table name")
	}
	stmt.Table = t.Val

	// optional (columns)
	t, ok = p.l.NextToken()
	if !ok {
		return nil, errors.New("unexpected end of input")
	}

	if t.Type == lexer.SYMBOL && t.Val == "(" {
		for {
			t, ok = p.l.NextToken()
			if !ok {
				return nil, errors.New("expected ')', got EOF")
			}

			if t.Type == lexer.SYMBOL && t.Val == ")" {
				break
			}

			if t.Type == lexer.IDENTIFIER {
				stmt.Columns = append(stmt.Columns, t.Val)
			}
		}

		// next token after columns
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("unexpected end of input (expected VALUES)")
		}
	}

	// VALUES
	if t.Type != lexer.KEYWORD || t.Val != "VALUES" {
		return nil, errors.New("expected VALUES")
	}

	// (
	t, ok = p.l.NextToken()
	if !ok || t.Type != lexer.SYMBOL || t.Val != "(" {
		return nil, errors.New("expected '(' after VALUES")
	}

	// values
	for {
		t, ok = p.l.NextToken()
		if !ok {
			return nil, errors.New("expected ')', got EOF")
		}

		if t.Type == lexer.SYMBOL && t.Val == ")" {
			break
		}

		if t.Type == lexer.STRING || t.Type == lexer.NUMBER {
			stmt.Values = append(stmt.Values, t.Val)
		}
	}

	return stmt, nil
}
