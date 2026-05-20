package cli

import (
	"bufio"
	"fmt"
	"os"

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

		if line == "clear" {
			clearScreen()
			continue
		}

		if err := c.RunStmt(ParseCMD(line)); err != nil {
			fmt.Printf("error: %s \n", err.Error())
		}
	}
}

func printPref() {
	fmt.Print("db>")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
