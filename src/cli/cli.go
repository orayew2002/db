package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type CLI struct {
	db DB
}

func NewCli(db DB) *CLI {
	return &CLI{db: db}
}

func (c *CLI) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	clearScreen()

	for {
		fmt.Print("db>")

		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		if line == "exit" {
			clearScreen()
			break
		}

		switch parseCMD(line) {
		case ShowTables:
			c.showTables()

		case CreateTable:
			c.createTable(line)

		case Insert:
			c.insert(line)

		case Delete:
			c.delete(line)

		case Get:
			c.get(line)

		case CommandNotFound:
			fmt.Print("command not found")
		}
	}
}

func (c *CLI) get(cmd string) {
	flags := strings.Split(cmd, " ")
	rows := c.db.Get(flags[3])
	if len(rows) == 0 {
		return
	}

	heads := c.db.GetColumns(flags[3])

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(heads)

	for _, row := range rows {
		table.Append(rowToStrings(row, heads))
	}

	table.Render()
}

func (c *CLI) delete(cmd string) {
	flags := strings.Split(cmd, " ")

	table := flags[2]
	condition := strings.Split(flags[4], "=")

	c.db.Delete(table, condition[0], condition[1])
}

func (c *CLI) insert(cmd string) {
	flags := strings.Split(cmd, " ")

	table := flags[2]
	rows := strings.Split(strings.Trim(flags[3], "()"), ",")

	vals := make(map[string]any)
	for _, r := range rows {
		k := strings.Split(r, "=")
		vals[k[0]] = k[1]
	}

	c.db.Insert(table, vals)
}

func (c *CLI) showTables() {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Id", "Table"})

	tables := c.db.GetTables()
	data := make([][]string, len(tables))

	for i, t := range tables {
		data[i] = append(data[i], strconv.Itoa(i+1))
		data[i] = append(data[i], t)
	}

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}

func (c *CLI) createTable(cmd string) {
	flags := strings.Split(cmd, " ")

	table := flags[2]
	rows := strings.Split(strings.Trim(flags[3], "()"), ",")

	if err := c.db.CreateTable(table, rows); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("%s table created \n", table)
}

func rowToStrings(row map[string]any, columns []string) []string {
	values := make([]string, 0, len(columns))

	for _, column := range columns {
		value, ok := row[column]
		if !ok || value == nil {
			values = append(values, "")
			continue
		}

		values = append(values, fmt.Sprint(value))
	}

	return values
}
