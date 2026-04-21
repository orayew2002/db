package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/orayew2002/db/src/db"
)

type CLI struct {
	db DB
}

var (
	createTablePattern = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+([a-zA-Z_][\w]*)\s*\((.+)\)\s*$`)
	insertPattern      = regexp.MustCompile(`(?i)^INSERT\s+INTO\s+([a-zA-Z_][\w]*)\s*\((.+)\)\s*$`)
	updatePattern      = regexp.MustCompile(`(?i)^UPDATE\s+([a-zA-Z_][\w]*)\s+SET\s+(.+?)\s+WHERE\s+([a-zA-Z_][\w]*)\s*=\s*(.+)\s*$`)
	deletePattern      = regexp.MustCompile(`(?i)^DELETE\s+FROM\s+([a-zA-Z_][\w]*)\s+WHERE\s+([a-zA-Z_][\w]*)\s*=\s*(.+)\s*$`)
	selectPattern      = regexp.MustCompile(`(?i)^SELECT\s+\*\s+FROM\s+([a-zA-Z_][\w]*)\s*$`)
)

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

		case Update:
			c.update(line)

		case Delete:
			c.delete(line)

		case Get:
			c.get(line)

		case DescribreTable:
			c.describeTable(line)

		case CommandNotFound:
			fmt.Print("command not found \n")
		}
	}
}

func (c *CLI) describeTable(cmd string) {
	command := strings.Split(cmd, " ")
	if len(command) != 2 {
		fmt.Println("invalid select syntax")
		return
	}

	table := command[1]

	cols, err := c.db.GetColumns(table)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tb := tablewriter.NewWriter(os.Stdout)
	tb.Header("Column", "Type")

	for _, c := range cols {
		tb.Append(c.Name, c.Type)
	}

	tb.Render()
}

func (c *CLI) get(cmd string) {
	matches := selectPattern.FindStringSubmatch(cmd)
	if len(matches) != 2 {
		fmt.Println("invalid select syntax")
		return
	}

	rows, err := c.db.Get(matches[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	heads, err := c.db.GetColumns(matches[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(heads)

	for _, row := range rows {
		table.Append(rowToStrings(row, nil))
	}

	table.Render()
}

func (c *CLI) delete(cmd string) {
	matches := deletePattern.FindStringSubmatch(cmd)
	if len(matches) != 4 {
		fmt.Println("invalid delete syntax")
		return
	}

	if err := c.db.Delete(matches[1], matches[2], parseValue(matches[3])); err != nil {
		fmt.Println(err.Error())
	}
}

func (c *CLI) insert(cmd string) {
	matches := insertPattern.FindStringSubmatch(cmd)
	if len(matches) != 3 {
		fmt.Println("invalid insert syntax")
		return
	}

	vals, err := parseAssignments(matches[2])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := c.db.Insert(matches[1], vals); err != nil {
		fmt.Println(err.Error())
	}
}

func (c *CLI) update(cmd string) {
	matches := updatePattern.FindStringSubmatch(cmd)
	if len(matches) != 5 {
		fmt.Println("invalid update syntax")
		return
	}

	vals, err := parseAssignments(matches[2])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := c.db.Update(matches[1], matches[3], parseValue(matches[4]), vals); err != nil {
		fmt.Println(err.Error())
	}
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

const RowName int = 0
const RowType int = 1

func (c *CLI) createTable(cmd string) {
	matches := createTablePattern.FindStringSubmatch(cmd)
	if len(matches) != 3 {
		fmt.Println("invalid create table syntax")
		return
	}

	table := matches[1]
	rows := splitCSV(matches[2])
	tRows := make([]db.ColDef, len(rows))
	for _, r := range rows {
		rr := strings.Split(r, " ")
		tRows = append(tRows, db.ColDef{
			Type: rr[RowType],
			Name: rr[RowName],
		})
	}

	if err := c.db.CreateTable(table, tRows); err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("%s table created \n", table)
}

func parseAssignments(raw string) (map[string]any, error) {
	parts := splitCSV(raw)
	if len(parts) == 0 {
		return nil, errors.New("missing assignments")
	}

	vals := make(map[string]any, len(parts))
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid assignment %q", strings.TrimSpace(part))
		}

		key := strings.TrimSpace(kv[0])
		if key == "" {
			return nil, errors.New("empty column name")
		}

		vals[key] = parseValue(kv[1])
	}

	return vals, nil
}

func parseValue(raw string) any {
	value := strings.TrimSpace(raw)
	if len(value) >= 2 {
		if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
			return value[1 : len(value)-1]
		}
	}

	switch strings.ToUpper(value) {
	case "NULL":
		return nil
	case "TRUE":
		return true
	case "FALSE":
		return false
	}

	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	return value
}

func splitCSV(raw string) []string {
	var (
		parts   []string
		current strings.Builder
		inQuote rune
		escaped bool
	)

	for _, r := range raw {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case inQuote != 0:
			if r == inQuote {
				inQuote = 0
			}
			current.WriteRune(r)
		case r == '\'' || r == '"':
			inQuote = r
			current.WriteRune(r)
		case r == ',':
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		parts = append(parts, s)
	}

	return parts
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
