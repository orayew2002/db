package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/orayew2002/db/src/db"
	"github.com/orayew2002/db/src/executor"
	"github.com/orayew2002/db/src/ui"
)

type CLI struct {
	exec *executor.Exec
}

func NewCli(db *db.Database) *CLI {
	return &CLI{exec: &executor.Exec{DB: db}}
}

func (c *CLI) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	clearScreen()

	for {
		printPref()

		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		if line == "exit" {
			clearScreen()
			break
		}

		if line == "clear" {
			clearScreen()
			continue
		}

		err, res := c.exec.ExecStmt(ParseCMD(line))
		if err != nil {
			fmt.Printf("error: %s \n", err.Error())
			continue
		}

		ui.ShowTable(res.Name, res.Rows)
	}
}

func printPref() {
	fmt.Print("db>")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
