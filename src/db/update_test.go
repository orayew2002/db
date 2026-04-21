package db

import (
	"path/filepath"
	"testing"
)

func TestDatabaseUpdatePreservesUnchangedColumnsAndRecovers(t *testing.T) {
	dir := t.TempDir()
	walPath := filepath.Join(dir, "wal")
	snapshotPath := filepath.Join(dir, "db")

	database := Create(Options{
		WFP: walPath,
		FFP: snapshotPath,
	})

	cols := []ColDef{}
	cols = append(cols, ColDef{Name: "id", Type: "int"})
	cols = append(cols, ColDef{Name: "name", Type: "text"})
	cols = append(cols, ColDef{Name: "email", Type: "text"})

	if err := database.CreateTable("users", cols); err != nil {
		t.Fatal(err)
	}

	if err := database.Insert("users", map[string]any{
		"id":    "u1",
		"name":  "Alice",
		"email": "alice@example.com",
	}); err != nil {
		t.Fatal(err)
	}

	if err := database.Update("users", "id", "u1", map[string]any{
		"name": "Alice Smith",
	}); err != nil {
		t.Fatal(err)
	}

	rows, err := database.Get("users")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice Smith" {
		t.Fatalf("expected updated name, got %#v", rows[0]["name"])
	}
	if rows[0]["email"] != "alice@example.com" {
		t.Fatalf("expected unchanged email to be preserved, got %#v", rows[0]["email"])
	}

	database.Close()

	recovered := Create(Options{
		WFP: walPath,
		FFP: snapshotPath,
	})
	defer recovered.Close()

	rows, err = recovered.Get("users")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 recovered row, got %d", len(rows))
	}
	if rows[0]["name"] != "Alice Smith" {
		t.Fatalf("expected recovered update, got %#v", rows[0]["name"])
	}
	if rows[0]["email"] != "alice@example.com" {
		t.Fatalf("expected recovered email to be preserved, got %#v", rows[0]["email"])
	}
}

func TestDatabaseRejectsUnknownUpdateColumns(t *testing.T) {
	dir := t.TempDir()
	database := Create(Options{
		WFP: filepath.Join(dir, "wal"),
		FFP: filepath.Join(dir, "db"),
	})
	defer database.Close()

	cols := []ColDef{}
	cols = append(cols, ColDef{Name: "id", Type: "int"})
	cols = append(cols, ColDef{Name: "name", Type: "text"})
	cols = append(cols, ColDef{Name: "email", Type: "text"})
	if err := database.CreateTable("users", cols); err != nil {
		t.Fatal(err)
	}

	if err := database.Insert("users", map[string]any{
		"id":   "u1",
		"name": "Alice",
	}); err != nil {
		t.Fatal(err)
	}

	if err := database.Update("users", "id", "u1", map[string]any{
		"email": "alice@example.com",
	}); err == nil {
		t.Fatal("expected update with unknown column to fail")
	}
}
