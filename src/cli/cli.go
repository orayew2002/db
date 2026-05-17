package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/orayew2002/db/src/db"
)

type CLI struct {
	db *db.Database
}

func NewCli(db *db.Database) *CLI {
	return &CLI{db: db}
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

		c.runStmt(parseCMD(line))
	}
}

func printPref() {
	fmt.Print("db>")
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
