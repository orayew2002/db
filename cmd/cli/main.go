package main

import (
	"github.com/orayew2002/db/src/cli"
	"github.com/orayew2002/db/src/db"
)

func main() {
	db := db.Create("database/wal.json", "database/db.json")
	c := cli.NewCli(db)
	c.Run()
}
