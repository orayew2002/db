package db

import (
	"path/filepath"
	"testing"
)

func TestWAL(t *testing.T) {
	wal := NewWal(filepath.Join(t.TempDir(), "wal.log"))

	t.Run("testing create table", func(t *testing.T) {
		lsn, _ := wal.Append(CT, "users", "", []string{"id", "username", "email"})
		wal.Commit(lsn)

		lsn, _ = wal.Append(CT, "products", "", []string{"id", "name", "price"})
		wal.Commit(lsn)

		lsn, _ = wal.Append(CT, "markets", "", []string{"id", "name", "city"})
		wal.Commit(lsn)
	})

	t.Run("testing insert users", func(t *testing.T) {
		users := []map[string]any{
			{"id": "u1001", "username": "john_doe", "email": "john@gmail.com"},
			{"id": "u1002", "username": "alice_smith", "email": "alice@yahoo.com"},
			{"id": "u1003", "username": "mike_ross", "email": "mike@lawfirm.com"},
			{"id": "u1004", "username": "sara_connor", "email": "sara@future.net"},
			{"id": "u1005", "username": "bruce_wayne", "email": "bruce@wayneenterprises.com"},
		}

		for _, u := range users {
			lsn, err := wal.Append(I, "users", "", u)
			if err != nil {
				t.Error(err)
			}
			wal.Commit(lsn)
		}
	})

	t.Run("testing insert products", func(t *testing.T) {
		products := []map[string]any{
			{"id": "p2001", "name": "iPhone 15", "price": 999},
			{"id": "p2002", "name": "Samsung Galaxy S24", "price": 899},
			{"id": "p2003", "name": "MacBook Pro M3", "price": 1999},
			{"id": "p2004", "name": "Dell XPS 13", "price": 1499},
			{"id": "p2005", "name": "Sony WH-1000XM5", "price": 399},
		}

		for _, p := range products {
			lsn, err := wal.Append(I, "products", "", p)
			if err != nil {
				t.Error(err)
			}
			wal.Commit(lsn)
		}
	})

	t.Run("testing insert markets", func(t *testing.T) {
		markets := []map[string]any{
			{"id": "m3001", "name": "Tech Store", "city": "New York"},
			{"id": "m3002", "name": "Gadget World", "city": "San Francisco"},
			{"id": "m3003", "name": "ElectroMart", "city": "Berlin"},
			{"id": "m3004", "name": "Digital Hub", "city": "Tokyo"},
			{"id": "m3005", "name": "Smart Shop", "city": "Dubai"},
		}

		for _, m := range markets {
			lsn, err := wal.Append(I, "markets", "", m)
			if err != nil {
				t.Error(err)
			}
			wal.Commit(lsn)
		}
	})

	t.Run("testing delete", func(t *testing.T) {
		lsn, err := wal.Append(D, "users", "id", "u1001")
		if err != nil {
			t.Error(err)
		}
		wal.Commit(lsn)
	})

	t.Run("testing update", func(t *testing.T) {
		lsn, err := wal.Append(U, "users", "id", us{
			val:  "u1001",
			vals: map[string]any{"id": "u1001", "username": "john_doe", "email": "john@gmail.com"}})

		if err != nil {
			t.Error(err)
		}
		wal.Commit(lsn)
	})

	t.Run("testing replay", func(t *testing.T) {
		db := Create(filepath.Join("database", "db.json"), filepath.Join("database", "wal.log"))

		wal.Replay(func(a action) {
			switch a.A {
			case CT:
				db.CreateTable(a.Table, toStringSlice(a.Val))

			case I:
				v, ok := a.Val.(map[string]any)
				if ok {
					db.Insert(a.Table, v)
				}

			case D:
				db.Delete(a.Table, a.Col, a.Val)
			}
		})

		for table := range db.tables {
			t.Log("TABLE:", table)
			t.Log("DATA:", db.Get(table))
		}
	})

	t.Log("---------------DONE---------------")
}
