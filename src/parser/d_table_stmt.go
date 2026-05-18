package parser

import (
	"errors"
	"fmt"

	"github.com/orayew2002/db/src/lexer"
)

type DropTableStmt struct {
	Table string
}

func (DropTableStmt) isStatement() {}

func (p *Parser) parseDropTable() (*DropTableStmt, error) {
	stmt := &DropTableStmt{}

	t, ok := p.l.NextToken()
	if !ok || t.Type != lexer.KEYWORD || t.Val != "TABLE" {
		return nil, errors.New("error sql query command")
	}

	t, ok = p.l.NextToken()
	if !ok || t.Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("wrong table naming: %s", t.Val)
	}

	stmt.Table = t.Val
	return stmt, nil
}
