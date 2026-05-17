package cli

import (
	"fmt"

	"github.com/orayew2002/db/src/parser"
	"github.com/orayew2002/db/src/ui"
)

func (c *CLI) runStmt(s parser.Statement) {
	if selectStmt, ok := s.(*parser.SelectStmt); ok {

		// First check table is exists
		if err := c.db.CheckTable(selectStmt.Table); err != nil {
			panic(fmt.Errorf("table not exists"))
		}

		date, err := c.db.Get(selectStmt.Table)
		if err != nil {
			panic(fmt.Errorf("error featching database data : %w", err))
		}

		ui.ShowTable(selectStmt.Table, date)
	}
}
