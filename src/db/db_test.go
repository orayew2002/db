package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatbase(t *testing.T) {
	dir := t.TempDir()
	db := Create(filepath.Join(dir, "wal.log"), filepath.Join(dir, "db.json"))
	defer db.Close()

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

func TestDatabaseRecoveryDoesNotDuplicateOrReappendWAL(t *testing.T) {
	dir := t.TempDir()
	walPath := filepath.Join(dir, "wal.log")
	snapshotPath := filepath.Join(dir, "db.json")

	db := Create(walPath, snapshotPath)
	if err := db.CreateTable("users", []string{"id", "name", "email"}); err != nil {
		t.Fatal(err)
	}

	db.Insert("users", map[string]any{
		"id":    "u1",
		"name":  "Alice",
		"email": "alice@example.com",
	})
	db.Close()

	infoBeforeRecovery, err := os.Stat(walPath)
	if err != nil {
		t.Fatal(err)
	}
	if infoBeforeRecovery.Size() == 0 {
		t.Fatal("expected WAL to contain uncheckpointed writes before recovery")
	}

	recovered := Create(walPath, snapshotPath)
	rows := recovered.Get("users")
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after recovery, got %d", len(rows))
	}
	if rows[0]["id"] != "u1" {
		t.Fatalf("expected recovered row id to be u1, got %#v", rows[0]["id"])
	}
	recovered.Close()

	infoAfterRecovery, err := os.Stat(walPath)
	if err != nil {
		t.Fatal(err)
	}
	if infoAfterRecovery.Size() != 0 {
		t.Fatalf("expected replayed WAL to be truncated, size=%d", infoAfterRecovery.Size())
	}

	restarted := Create(walPath, snapshotPath)
	defer restarted.Close()

	rows = restarted.Get("users")
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after second restart, got %d", len(rows))
	}
	if rows[0]["email"] != "alice@example.com" {
		t.Fatalf("expected recovered email to match, got %#v", rows[0]["email"])
	}
}
