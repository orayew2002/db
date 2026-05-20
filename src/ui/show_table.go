package ui

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func ShowTable(title string, cols []map[string]any) {
	if len(cols) == 0 {
		fmt.Fprintln(os.Stdout, "empty table")
		return
	}

	headers := make([]string, 0)
	headerIndex := make(map[string]int)

	for k := range cols[0] {
		headerIndex[k] = len(headers)
		headers = append(headers, k)
	}

	table := tablewriter.NewWriter(os.Stdout)

	table.Header(headers)
	for _, rowMap := range cols {
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
