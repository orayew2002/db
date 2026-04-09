package db

import (
	"path/filepath"
	"testing"
)

func TestDatbase(t *testing.T) {
	db := Create(filepath.Join("database", "db.json"), filepath.Join("database", "wal.log"))

	t.Run("run database", func(t *testing.T) {
		if err := db.CreateTable("users", []string{"id", "name", "email"}); err != nil {
			t.Error(err.Error())
		}

		if err := db.CreateTable("products", []string{"id", "name", "email"}); err != nil {
			t.Error(err.Error())
		}

		if err := db.CreateTable("markets", []string{"id", "name", "email"}); err != nil {
			t.Error(err.Error())
		}
	})
}
