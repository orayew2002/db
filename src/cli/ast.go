package cli

import (
	"fmt"

	"github.com/orayew2002/db/src/parser"
	"github.com/orayew2002/db/src/ui"
)

func (c *CLI) runStmt(s parser.Statement) {
	if stmt, ok := s.(*parser.DropTableStmt); ok {
		c.checkTable(stmt.Table)

		// TODO need complete db function's for drop table there
	}

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
			return
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
			panic("this table already exists")
		}

		if err := c.db.CreateTable(stmt.Table, stmt.Columns); err != nil {
			panic(fmt.Errorf("error creating table: %w", err))
		}
	}

	if stmt, ok := s.(*parser.UpdateStmt); ok {
		c.checkTable(stmt.Table)

		rows, err := c.db.Get(stmt.Table)
		if err != nil {
			panic(fmt.Errorf("error getting data: %w", err))
		}

		if stmt.Where != nil {
			nv := make([]map[string]any, 0)

			for _, r := range rows {
				if r[stmt.Where.Left].(string) == stmt.Where.Right {
					for v, k := range stmt.Set {
						r[v] = k
					}

					nv = append(nv, r)
				}
			}

			for _, n := range nv {
				c.db.Update(stmt.Table, stmt.Where.Left, stmt.Where.Right, n)
			}

			return
		}

		// TODO
		// need write logic for update all elements if where clauser not detected
	}
}

func (c *CLI) checkTable(table string) {
	if err := c.db.CheckTable(table); err != nil {
		panic(fmt.Errorf("table not exists"))
	}
}
