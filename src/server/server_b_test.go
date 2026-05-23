package server

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
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
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			_ = conn.Close()
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	tb.Fatalf("server not accepting connections on %s", addr)
}

func execute(tb testing.TB, conn net.Conn, query string) {
	tb.Helper()

	if _, err := conn.Write([]byte(query)); err != nil {
		tb.Fatalf("write: %v", err)
	}

	buf := make([]byte, 4096)

	if _, err := conn.Read(buf); err != nil {
		tb.Fatalf("read: %v", err)
	}
}

func roundTrip(tb testing.TB, conn net.Conn, query []byte, buf []byte) {
	tb.Helper()

	if _, err := conn.Write(query); err != nil {
		tb.Fatalf("write: %v", err)
	}

	if _, err := conn.Read(buf); err != nil {
		tb.Fatalf("read: %v", err)
	}
}

func primeTable(tb testing.TB) {
	tb.Helper()

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		tb.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	execute(tb, conn, "CREATE TABLE users (id text)")
}

func cleanupTable(tb testing.TB) {
	tb.Helper()

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		tb.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	execute(tb, conn, "DELETE FROM users")
}

type Metrics struct {
	Inserts uint64
	Reads   uint64
	Errors  uint64
}

func printMetrics(start time.Time, m *Metrics) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fmt.Println("========================================")
	fmt.Printf("runtime: %s\n", time.Since(start))
	fmt.Printf("goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("heap alloc: %.2f MB\n", float64(mem.HeapAlloc)/1024/1024)
	fmt.Printf("total alloc: %.2f MB\n", float64(mem.TotalAlloc)/1024/1024)
	fmt.Printf("sys memory: %.2f MB\n", float64(mem.Sys)/1024/1024)
	fmt.Printf("gc count: %d\n", mem.NumGC)

	fmt.Printf("inserts: %d\n", atomic.LoadUint64(&m.Inserts))
	fmt.Printf("reads: %d\n", atomic.LoadUint64(&m.Reads))
	fmt.Printf("errors: %d\n", atomic.LoadUint64(&m.Errors))
	fmt.Println("========================================")
}

func BenchmarkInsert(b *testing.B) {
	setup(b)
	primeTable(b)
	cleanupTable(b)

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	query := []byte("INSERT INTO users (id) VALUES ('test')")
	buf := make([]byte, 4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		roundTrip(b, conn, query, buf)
	}

	b.StopTimer()

	cleanupTable(b)
}

func BenchmarkSelect(b *testing.B) {
	setup(b)
	primeTable(b)
	cleanupTable(b)

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		b.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	insertQuery := []byte("INSERT INTO users (id) VALUES ('seed')")
	buf := make([]byte, 4096)

	for i := 0; i < 1000; i++ {
		roundTrip(b, conn, insertQuery, buf)
	}

	query := []byte("SELECT * FROM users")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		roundTrip(b, conn, query, buf)
	}

	b.StopTimer()

	cleanupTable(b)
}

func TestStressInsert(t *testing.T) {
	setup(t)
	primeTable(t)
	cleanupTable(t)

	const workers = 50
	const insertsPerWorker = 20000

	var wg sync.WaitGroup

	metrics := &Metrics{}
	start := time.Now()

	// metrics printer
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				printMetrics(start, metrics)

			case <-done:
				return
			}
		}
	}()

	for w := 0; w < workers; w++ {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", dialAddr)
			if err != nil {
				atomic.AddUint64(&metrics.Errors, 1)
				t.Errorf("worker=%d dial: %v", worker, err)
				return
			}
			defer conn.Close()

			query := []byte(
				fmt.Sprintf(
					"INSERT INTO users (id) VALUES ('worker-%d')",
					worker,
				),
			)

			buf := make([]byte, 4096)

			for i := 0; i < insertsPerWorker; i++ {
				if _, err := conn.Write(query); err != nil {
					atomic.AddUint64(&metrics.Errors, 1)
					t.Errorf("worker=%d write: %v", worker, err)
					return
				}

				if _, err := conn.Read(buf); err != nil {
					atomic.AddUint64(&metrics.Errors, 1)
					t.Errorf("worker=%d read: %v", worker, err)
					return
				}

				atomic.AddUint64(&metrics.Inserts, 1)

				// print progress every 5000 inserts
				if i%5000 == 0 {
					fmt.Printf(
						"[worker=%d] inserted=%d\n",
						worker,
						i,
					)
				}
			}
		}(w)
	}

	wg.Wait()

	close(done)

	printMetrics(start, metrics)

	cleanupTable(t)
}

func TestStressSelect(t *testing.T) {
	setup(t)
	primeTable(t)
	cleanupTable(t)

	conn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// seed data
	buf := make([]byte, 4096)

	for i := 0; i < 10000; i++ {
		query := fmt.Sprintf(
			"INSERT INTO users (id) VALUES ('seed-%d')",
			i,
		)

		roundTrip(t, conn, []byte(query), buf)
	}

	_ = conn.Close()

	const workers = 100
	const readsPerWorker = 10000

	var wg sync.WaitGroup

	metrics := &Metrics{}
	start := time.Now()

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				printMetrics(start, metrics)

			case <-done:
				return
			}
		}
	}()

	for w := 0; w < workers; w++ {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", dialAddr)
			if err != nil {
				atomic.AddUint64(&metrics.Errors, 1)
				return
			}
			defer conn.Close()

			query := []byte("SELECT * FROM users")
			buf := make([]byte, 1024*1024)

			for i := 0; i < readsPerWorker; i++ {
				if _, err := conn.Write(query); err != nil {
					atomic.AddUint64(&metrics.Errors, 1)
					return
				}

				if _, err := conn.Read(buf); err != nil {
					atomic.AddUint64(&metrics.Errors, 1)
					return
				}

				atomic.AddUint64(&metrics.Reads, 1)
			}
		}(w)
	}

	wg.Wait()

	close(done)

	printMetrics(start, metrics)

	cleanupTable(t)
}
