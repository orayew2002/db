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

	t.Run("send requst to get not exists table data", func(t *testing.T) {
		nc.Write([]byte("SELECT * FROM users"))

		buf := make([]byte, 1024)
		n, err := nc.Read(buf)
		if err != nil {
			t.Error(err)
		}

		fmt.Printf("%s \n", string(buf[:n]))
	})

	t.Run("create new table", func(t *testing.T) {
		nc.Write([]byte("CREATE TABLE users (id text)"))

		buf := make([]byte, 1024)
		n, err := nc.Read(buf)
		if err != nil {
			t.Error(err)
		}

		fmt.Printf("%s \n", string(buf[:n]))
	})
}
