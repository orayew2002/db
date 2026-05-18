package cli

import (
	"fmt"

	"github.com/orayew2002/db/src/parser"
	"github.com/orayew2002/db/src/ui"
)

func (c *CLI) runStmt(s parser.Statement) {
	if stmt, ok := s.(*parser.SelectStmt); ok {
		c.checkTable(stmt.Table)

		date, err := c.db.Get(stmt.Table)
		if err != nil {
			panic(fmt.Errorf("error featching database data : %w", err))
		}

		ui.ShowTable(stmt.Table, date)
	}

	if stmt, ok := s.(*parser.InsertStmt); ok {
		c.checkTable(stmt.Table)

		r := make(map[string]any, len(stmt.Values))
		for i, c := range stmt.Columns {
			r[c] = stmt.Values[i]
		}

		if err := c.db.Insert(stmt.Table, r); err != nil {
			panic(fmt.Errorf("error inserting data to table: %w", err))
		}
	}

	if stmt, ok := s.(*parser.DeleteStmt); ok {
		c.checkTable(stmt.Table)

		rows, err := c.db.Get(stmt.Table)
		if err != nil {
			panic(fmt.Errorf("error featching database data : %w", err))
		}

		if stmt.Where != nil {
			for _, r := range rows {
				if r[stmt.Where.Left] == stmt.Where.Right {
					for k, v := range r {
						c.db.Delete(stmt.Table, k, v)
					}
				}
			}

			return
		}

		for _, r := range rows {
			for k, v := range r {
				c.db.Delete(stmt.Table, k, v)
			}
		}
	}

	if stmt, ok := s.(*parser.CreateTableStmt); ok {
		if err := c.db.CheckTable(stmt.Table); err == nil {
			panic("this table already exists")
		}

		if err := c.db.CreateTable(stmt.Table, stmt.Columns); err != nil {
			panic(fmt.Errorf("error creating table: %w", err))
		}
	}
}

func (c *CLI) checkTable(table string) {
	if err := c.db.CheckTable(table); err != nil {
		panic(fmt.Errorf("table not exists"))
	}
}
