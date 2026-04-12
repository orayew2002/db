package db

import (
	"fmt"
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
	db := Create("../../database/bench_wal.json", "../../database/bench_db.json")
	db.CreateTable("users", []string{"id", "name", "email"})

	b.ResetTimer()

	for b.Loop() {
		db.Insert("users", map[string]any{
			"id":    fmt.Sprintf("%d", b.N),
			"name":  "user name",
			"email": "user email",
		})
	}
}
