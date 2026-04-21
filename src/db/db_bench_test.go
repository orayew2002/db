package db

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func init() {
	go func() {
		log.Println("pprof HTTP server started on :6060")
		http.ListenAndServe("localhost:6060", nil)
	}()
}

func BenchmarkInsert(b *testing.B) {
	db := Create(Options{
		WFP: "../../database/wal",
		FFP: "../../database/db",
	})

	cols := []ColDef{}
	cols = append(cols, ColDef{Name: "id", Type: "int"})
	cols = append(cols, ColDef{Name: "name", Type: "text"})
	cols = append(cols, ColDef{Name: "email", Type: "text"})

	db.CreateTable("users", cols)

	row := map[string]any{
		"id":    "bench-id",
		"name":  "user name",
		"email": "user email",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Insert("users", row)
	}

	db.Close()
}
