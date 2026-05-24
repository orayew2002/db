package ui

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/orayew2002/db/src/db"
)

func ShowTable(title string, rows db.Rows) {
	if len(rows) == 0 {
		fmt.Fprintln(os.Stdout, "empty table")
		return
	}

	headers := make([]string, 0)
	headerIndex := make(map[string]int)

	for k := range rows[0] {
		headerIndex[k] = len(headers)
		headers = append(headers, k)
	}

	table := tablewriter.NewWriter(os.Stdout)

	table.Header(headers)
	for _, rowMap := range rows {
		row := make([]string, len(headers))

		for k, v := range rowMap {
			if idx, ok := headerIndex[k]; ok {
				row[idx] = fmt.Sprintf("%v", v)
			}
		}

		table.Append(row)
	}

	table.Render()
}
