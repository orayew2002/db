package fm

import (
	"path/filepath"
	"testing"
)

type User struct {
	Id   string
	Name string
}

func TestFullDump(t *testing.T) {
	fmFullDump := NewFmFullDump(filepath.Join(t.TempDir(), "db.json"))

	t.Run("testing dumping", func(t *testing.T) {
		fmFullDump.Flush(User{Id: "123", Name: "Orayew"})
	})
}
