package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatbase(t *testing.T) {
	dir := t.TempDir()
	db := Create(Options{
		WFP: filepath.Join(dir, "wal"),
		FFP: filepath.Join(dir, "db"),
		UWC: false,
	})
	defer db.Close()

	t.Run("run database", func(t *testing.T) {
		cols := []ColDef{}
		cols = append(cols, ColDef{Name: "id", Type: "int"})
		cols = append(cols, ColDef{Name: "name", Type: "text"})
		cols = append(cols, ColDef{Name: "email", Type: "text"})

		if err := db.CreateTable("users", cols); err != nil {
			t.Error(err.Error())
		}

		if err := db.CreateTable("products", cols); err != nil {
			t.Error(err.Error())
		}

		if err := db.CreateTable("markets", cols); err != nil {
			t.Error(err.Error())
		}
	})
}

func TestDatabaseRecoveryDoesNotDuplicateOrReappendWAL(t *testing.T) {
	dir := t.TempDir()
	walPath := filepath.Join(dir, "wal")
	snapshotPath := filepath.Join(dir, "db")

	db := Create(Options{
		WFP: walPath,
		FFP: snapshotPath,
	})
	if err := db.CreateTable("users", nil); err != nil {
		t.Fatal(err)
	}

	err := db.Insert("users", map[string]any{
		"id":    "u1",
		"name":  "Alice",
		"email": "alice@example.com",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	db.Close()

	infoBeforeRecovery, err := os.Stat(walPath)
	if err != nil {
		t.Fatal(err)
	}
	if infoBeforeRecovery.Size() == 0 {
		t.Fatal("expected WAL to contain uncheckpointed writes before recovery")
	}

	recovered := Create(Options{
		WFP: walPath,
		FFP: snapshotPath,
	})
	rows, err := recovered.Get("users")
	if err != nil {
		t.Fatal(err.Error())
	}
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

	restarted := Create(Options{
		WFP: walPath,
		FFP: snapshotPath,
	})
	defer restarted.Close()

	rows, err = restarted.Get("users")
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after second restart, got %d", len(rows))
	}
	if rows[0]["email"] != "alice@example.com" {
		t.Fatalf("expected recovered email to match, got %#v", rows[0]["email"])
	}
}
