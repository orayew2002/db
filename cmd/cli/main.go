package main

import (
	"github.com/orayew2002/db/src/cli"
	"github.com/orayew2002/db/src/db"
)

func main() {
	db := db.Create(db.Options{
		WFP: "database/wal",
		FFP: "database/db",
	})

	c := cli.NewCli(db)
	c.Run()
}
