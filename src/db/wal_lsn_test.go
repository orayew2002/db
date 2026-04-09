package db

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWalTruncateAndLSN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")

	w := NewWal(path)

	if err := w.AS(CT, "users", "", []string{"id", "name"}); err != nil {
		t.Fatalf("AS failed: %v", err)
	}

	if err := w.Replay(func(a action) {}); err != nil {
		t.Fatalf("Replay failed: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat wal file: %v", err)
	}
	if fi.Size() != 0 {
		t.Fatalf("expected wal file size 0 after replay, got %d", fi.Size())
	}

	if err := w.AS(CT, "products", "", []string{"id"}); err != nil {
		t.Fatalf("AS failed: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open wal file: %v", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read line: %v", err)
	}

	var act action
	if err := json.Unmarshal(line, &act); err != nil {
		t.Fatalf("unmarshal action: %v", err)
	}

	if act.LSN != 1 {
		t.Fatalf("expected LSN 1 after replay, got %d", act.LSN)
	}
}
