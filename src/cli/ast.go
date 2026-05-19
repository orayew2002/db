package cli

import (
	"fmt"
	"maps"

	"github.com/orayew2002/db/src/parser"
	"github.com/orayew2002/db/src/ui"
)

func (c *CLI) runStmt(s parser.Statement) error {
	if stmt, ok := s.(*parser.DropTableStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table)
		}

		if err := c.db.DropTable(stmt.Table); err != nil {
			return fmt.Errorf("error drop table: %w", err)
		}
	}

	if stmt, ok := s.(*parser.SelectStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table)
		}

		date, err := c.db.Get(stmt.Table)
		if err != nil {
			return fmt.Errorf("error featching database data : %w", err)
		}

		ui.ShowTable(stmt.Table, date)
	}

	if stmt, ok := s.(*parser.InsertStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table)
		}

		r := make(map[string]any, len(stmt.Values))
		for i, c := range stmt.Columns {
			r[c] = stmt.Values[i]
		}

		if err := c.db.Insert(stmt.Table, r); err != nil {
			return fmt.Errorf("error inserting data to table: %w", err)
		}
	}

	if stmt, ok := s.(*parser.DeleteStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table)
		}

		rows, err := c.db.Get(stmt.Table)
		if err != nil {
			return fmt.Errorf("error featching database data : %w", err)
		}

		if stmt.Where != nil {
			var i int
			for _, r := range rows {
				if r[stmt.Where.Left] == stmt.Where.Right {
					for k, v := range r {
						c.db.Delete(stmt.Table, k, v)
						i++
						break
					}
				}
			}

			fmt.Printf("delete data: %d \n", i)
			return nil
		}

		var i int
		for _, r := range rows {
			for k, v := range r {
				c.db.Delete(stmt.Table, k, v)
				i++
				break
			}
		}

		fmt.Printf("delete data: %d \n", i)
	}

	if stmt, ok := s.(*parser.CreateTableStmt); ok {
		if err := c.db.CheckTable(stmt.Table); err == nil {
			return fmt.Errorf("%s table already exists", stmt.Table)
		}

		if err := c.db.CreateTable(stmt.Table, stmt.Columns); err != nil {
			return fmt.Errorf("error creating table: %w", err)
		}
	}

	if stmt, ok := s.(*parser.UpdateStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table)
		}

		rows, err := c.db.Get(stmt.Table)
		if err != nil {
			return fmt.Errorf("error getting data: %w", err)
		}

		if stmt.Where != nil {
			nv := make([]map[string]any, 0)

			for _, r := range rows {
				if r[stmt.Where.Left].(string) == stmt.Where.Right {
					maps.Copy(r, stmt.Set)
					nv = append(nv, r)
				}
			}

			for _, n := range nv {
				c.db.Update(stmt.Table, stmt.Where.Left, stmt.Where.Right, n)
			}

			return nil
		}

		for _, r := range rows {
			maps.Copy(r, stmt.Set)

			for k, v := range r {
				c.db.Update(stmt.Table, k, v, r)
				break
			}
		}
	}

	return nil
}

func (c *CLI) checkTable(table string) bool {
	return c.db.CheckTable(table) == nil
}
