package parser

import (
	"fmt"

	"github.com/orayew2002/db/src/db"
	"github.com/orayew2002/db/src/lexer"
)

type AlterTableStmt struct {
	Table   string
	Columns []db.ColDef
}

func (AlterTableStmt) isStatement() {}

func (p *Parser) parseAlterTable() (*AlterTableStmt, error) {
	stmt := &AlterTableStmt{
		Columns: make([]db.ColDef, 0),
	}

	// ALTER TABLE
	t, ok := p.l.NextToken()
	if !ok || t.Type != lexer.KEYWORD || t.Val != "TABLE" {
		return nil, fmt.Errorf("expected TABLE after ALTER")
	}

	// table name
	t, ok = p.l.NextToken()
	if !ok {
		return nil, fmt.Errorf("expected table name")
	}

	if t.Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("invalid table name: %s", t.Val)
	}

	stmt.Table = t.Val

	// at least one ADD COLUMN required
	foundAdd := false

	for {
		t, ok = p.l.NextToken()
		if !ok {
			break
		}

		// ADD
		if t.Type != lexer.KEYWORD || t.Val != "ADD" {
			return nil, fmt.Errorf("expected ADD")
		}

		foundAdd = true

		// COLUMN
		t, ok = p.l.NextToken()
		if !ok {
			return nil, fmt.Errorf("expected COLUMN after ADD")
		}

		if t.Type != lexer.KEYWORD || t.Val != "COLUMN" {
			return nil, fmt.Errorf("expected COLUMN after ADD")
		}

		// column name
		colName, ok := p.l.NextToken()
		if !ok {
			return nil, fmt.Errorf("expected column name")
		}

		if colName.Type != lexer.IDENTIFIER {
			return nil, fmt.Errorf("invalid column name: %s", colName.Val)
		}

		// column type
		colType, ok := p.l.NextToken()
		if !ok {
			return nil, fmt.Errorf("expected type for column %s", colName.Val)
		}

		if colType.Type != lexer.KEYWORD && colType.Type != lexer.IDENTIFIER {
			return nil, fmt.Errorf("invalid type for column %s", colName.Val)
		}

		stmt.Columns = append(stmt.Columns, db.ColDef{
			Name: colName.Val,
			Type: colType.Val,
		})

		p.l.NextToken()
	}

	if !foundAdd {
		return nil, fmt.Errorf("expected ADD COLUMN statement")
	}

	return stmt, nil
}
