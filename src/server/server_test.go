package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/orayew2002/db/src/db"
)

func TestServer(t *testing.T) {
	db := db.Create(db.Options{
		WFP: "../../database/wal",
		FFP: "../../database/db",
	})

	s, err := Default(db)
	if err != nil {
		t.Error(err)
	}

	go s.Run()
	time.Sleep(time.Second * 2)

	nc, err := net.Dial("tcp", "0.0.0.0:9696")
	if err != nil {
		t.Error(err)
	}

	t.Run("create new table", func(t *testing.T) {
		nc.Write([]byte("CREATE TABLE users (id text)"))

		buf := make([]byte, 1024)
		n, err := nc.Read(buf)
		if err != nil {
			t.Error(err)
		}

		if string(buf[:n]) == "s" {
			t.Log("table success created")
		}
	})

	t.Run("insert to table", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			q := fmt.Sprintf(
				"INSERT INTO users (id) VALUES ('%d')",
				i,
			)

			_, err := nc.Write([]byte(q))
			if err != nil {
				t.Fatal(err)
			}

			buf := make([]byte, 1024)

			_, err = nc.Read(buf)
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("get table data", func(t *testing.T) {
		nc.Write([]byte("SELECT * FROM users"))

		buf := make([]byte, 1024)
		n, err := nc.Read(buf)
		if err != nil {
			t.Error(err)
		}

		fmt.Printf("%s \n", string(buf[:n]))
	})
}
