package server

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/orayew2002/db/src/db"
)

const dialAddr = "127.0.0.1:9696"

var startOnce sync.Once

func setup(tb testing.TB) {
	tb.Helper()
	startOnce.Do(func() {
		database := db.Create(db.Options{
			WFP: "../../database/wal",
			FFP: "../../database/db",
		})

		s, err := Default(database)
		if err != nil {
			tb.Fatalf("create server: %v", err)
		}

		go s.Run()
		waitForServer(tb, dialAddr, 5*time.Second)
	})
}

func waitForServer(tb testing.TB, addr string, timeout time.Duration) {
	tb.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c, err := net.Dial("tcp", addr); err == nil {
			_ = c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	tb.Fatalf("server not accepting connections on %s within %s", addr, timeout)
}

func roundTrip(b *testing.B, conn net.Conn, query, buf []byte) {
	if _, err := conn.Write(query); err != nil {
		b.Fatalf("write: %v", err)
	}
	if _, err := conn.Read(buf); err != nil {
		b.Fatalf("read: %v", err)
	}
}

func primeTable(b *testing.B) {
	b.Helper()

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	_, _ = conn.Write([]byte("CREATE TABLE users (id text)"))
	_, _ = conn.Read(make([]byte, 4096))
}

// -------------------- SELECT BENCHMARK --------------------

func BenchmarkSelect(b *testing.B) {
	setup(b)
	primeTable(b)

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	query := []byte("SELECT * FROM users")
	buf := make([]byte, 4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		roundTrip(b, conn, query, buf)
	}
}

// -------------------- INSERT BENCHMARK --------------------

func BenchmarkInsert(b *testing.B) {
	setup(b)
	primeTable(b)

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := []byte("INSERT INTO users (id) VALUES ('t')")
		roundTrip(b, conn, query, buf)
	}
}
