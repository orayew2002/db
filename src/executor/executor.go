package executor

import (
	"fmt"
	"maps"

	"github.com/orayew2002/db/src/db"
	"github.com/orayew2002/db/src/parser"
)

type Exec struct {
	DB *db.Database
}

func (c *Exec) ExecStmt(s parser.Statement) (error, *db.Table) {
	if stmt, ok := s.(*parser.DropTableStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table), nil
		}

		if err := c.DB.DropTable(stmt.Table); err != nil {
			return fmt.Errorf("error drop table: %w", err), nil
		}
	}

	if stmt, ok := s.(*parser.SelectStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table), nil
		}

		date, err := c.DB.Get(stmt.Table)
		if err != nil {
			return fmt.Errorf("error featching database data : %w", err), nil
		}

		return nil, &db.Table{Name: stmt.Table, Rows: date}
	}

	if stmt, ok := s.(*parser.InsertStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table), nil
		}

		r := make(map[string]any, len(stmt.Values))
		for i, c := range stmt.Columns {
			r[c] = stmt.Values[i]
		}

		if err := c.DB.Insert(stmt.Table, r); err != nil {
			return fmt.Errorf("error inserting data to table: %w", err), nil
		}
	}

	if stmt, ok := s.(*parser.DeleteStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table), nil
		}

		rows, err := c.DB.Get(stmt.Table)
		if err != nil {
			return fmt.Errorf("error featching database data : %w", err), nil
		}

		if stmt.Where != nil {
			var i int
			for _, r := range rows {
				if r[stmt.Where.Left] == stmt.Where.Right {
					for k, v := range r {
						c.DB.Delete(stmt.Table, k, v)
						i++
						break
					}
				}
			}

			return nil, nil
		}

		var i int
		for _, r := range rows {
			for k, v := range r {
				c.DB.Delete(stmt.Table, k, v)
				i++
				break
			}
		}
	}

	if stmt, ok := s.(*parser.CreateTableStmt); ok {
		if err := c.DB.CheckTable(stmt.Table); err == nil {
			return fmt.Errorf("%s table already exists", stmt.Table), nil
		}

		if err := c.DB.CreateTable(stmt.Table, stmt.Columns); err != nil {
			return fmt.Errorf("error creating table: %w", err), nil
		}
	}

	if stmt, ok := s.(*parser.UpdateStmt); ok {
		if !c.checkTable(stmt.Table) {
			return fmt.Errorf("%s table not exists", stmt.Table), nil
		}

		rows, err := c.DB.Get(stmt.Table)
		if err != nil {
			return fmt.Errorf("error getting data: %w", err), nil
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
				c.DB.Update(stmt.Table, stmt.Where.Left, stmt.Where.Right, n)
			}

			return nil, nil
		}

		for _, r := range rows {
			maps.Copy(r, stmt.Set)

			for k, v := range r {
				c.DB.Update(stmt.Table, k, v, r)
				break
			}
		}
	}

	return nil, nil
}

func (c *Exec) checkTable(table string) bool {
	return c.DB.CheckTable(table) == nil
}
