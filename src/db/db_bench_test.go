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
	db := Create(Options{
		WFP: "../../database/wal.json",
		FFP: "../../database/db.json",
		UWC: false,
	})

	db.CreateTable("users", []string{"id", "name", "email"})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Insert("users", map[string]any{
			"id":    fmt.Sprintf("%d", i),
			"name":  "user name",
			"email": "user email",
		})
	}

	db.Close()
}
