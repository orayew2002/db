package cli

import "strings"

type Action string

const (
	ShowTables  Action = "show_tables"
	CreateTable Action = "create table"
	Insert      Action = "insert"
	Delete      Action = "delete"
	Get         Action = "get"
)

func parseCMD(cmd string) Action {
	cmd = strings.ToUpper(cmd)

	if strings.HasPrefix(cmd, "SHOW TABLES") {
		return ShowTables
	}

	if strings.HasPrefix(cmd, "CREATE TABLE") {
		return CreateTable
	}

	if strings.HasPrefix(cmd, "INSERT INTO") {
		return Insert
	}

	if strings.HasPrefix(cmd, "DELETE FROM") {
		return Delete
	}

	if strings.HasPrefix(cmd, "SELECT") {
		return Get
	}

	return ShowTables
}
