package main

import (
	"github.com/orayew2002/db/src/cli"
	"github.com/orayew2002/db/src/db"
)

func main() {
	db := db.Create(db.Options{
		WFP: "database/wal.json",
		FFP: "database/db.json",
		UWC: false,
	})

	c := cli.NewCli(db)
	c.Run()
}
