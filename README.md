# GoDB

A lightweight embedded database written in Go. It stores data in memory, persists it through a Write-Ahead Log (WAL) for crash recovery, and checkpoints state to a JSON snapshot file. Comes with an interactive CLI.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                      Database                       │
│                                                     │
│   ┌──────────────┐     ┌──────────────────────────┐ │
│   │  In-Memory   │     │   WAL (Write-Ahead Log)  │ │
│   │   Tables     │◄────│  binary framed JSON log  │ │
│   │  map[string] │     │  async flush every 5ms   │ │
│   │   *Table     │     └──────────────────────────┘ │
│   └──────┬───────┘                                  │
│          │                                          │
│   ┌──────▼───────┐                                  │
│   │ File Manager │                                  │
│   │  (FM / full  │                                  │
│   │    dump)     │                                  │
│   └──────────────┘                                  │
└─────────────────────────────────────────────────────┘
```

### Components

| Component | File | Role |
|-----------|------|------|
| `Database` | `src/db/db.go` | Top-level coordinator. Owns tables, WAL, and file manager |
| `WAL` | `src/db/wal.go` | Append-only binary log. Guarantees durability before in-memory apply |
| `FmFullDump` | `src/fm/full_dump.go` | Persistence layer. Atomically snapshots all tables to a JSON file |
| `Table` | `src/db/table.go` | In-memory table. Supports insert, delete (filter-copy), and update |
| `CLI` | `src/cli/` | Interactive shell with SQL-like syntax |

---

## How It Works

### Startup sequence

Every time `Create()` is called, three steps run in order:

```
1. Load snapshot  →  2. Replay WAL  →  3. Flush checkpoint
```

1. **Load snapshot** — the file manager reads the last JSON snapshot into memory. If no snapshot exists, the database starts empty.
2. **Replay WAL** — every committed entry in the WAL is re-applied on top of the snapshot. This recovers any writes that happened after the last checkpoint.
3. **Flush checkpoint** — the recovered in-memory state is immediately written back to the snapshot file, and the WAL can be considered replayed.

### Write path

Every write (Insert / Delete / Update / CreateTable) follows this sequence:

```
Append to WAL buffer
        │
        ▼
Apply change in memory
        │
        ▼
Commit(lsn)  ←── blocks until WAL is flushed to disk
```

The WAL uses **group commit**: a background goroutine flushes the buffer to disk every 5 ms and calls `fsync`. All writers that appended during that window are unblocked together. This trades a tiny latency (≤ 5 ms) for much higher write throughput under concurrent load.

### WAL format

Each entry is length-prefixed:

```
┌──────────────┬──────────────────────────────┐
│ size: uint32 │ JSON payload (action struct)  │
│  (4 bytes)   │  { lsn, action, table, col,  │
│ little-endian│    val }                      │
└──────────────┴──────────────────────────────┘
```

Actions: `insert` · `delete` · `update` · `create_table` · `update_table` · `delete_table`

### Crash recovery guarantee

- **After a committed write** — the WAL entry is on disk before `Commit()` returns. On restart the entry is replayed. No data loss.
- **During a write** — if the process crashes between `Append` and `Commit`, the entry may or may not be in the WAL. On replay it will either be present (applied) or absent (not applied). Either way the database is consistent.
- **Snapshot write** — `FmFullDump` writes to a `_tmp` file then atomically renames it. A crash mid-snapshot leaves the previous snapshot intact.

---

## Usage

### Programmatic (Go)

```go
import "github.com/orayew2002/db/src/db"

// Open (or create) a database.
// First argument  = WAL file path
// Second argument = snapshot file path
database := db.Create("database/wal.json", "database/db.json")
defer database.Close()

// Create a table
database.CreateTable("users", []string{"id", "name", "email"})

// Insert a row
database.Insert("users", map[string]any{
    "id":    "u1",
    "name":  "Alice",
    "email": "alice@example.com",
})

// Query all rows
rows := database.Get("users")

// Delete by column value
database.Delete("users", "id", "u1")

// Update rows matching a condition
database.Update("users", "id", "u1", map[string]any{
    "id":    "u1",
    "name":  "Alice Smith",
    "email": "alice@example.com",
})

// List all table names
tables := database.GetTables()
```

### CLI

```bash
go run ./cmd/cli
```

```
db> SHOW TABLES
db> CREATE TABLE users (id,name,email)
db> INSERT INTO users (id=u1,name=Alice,email=alice@example.com)
db> SELECT * FROM users
db> DELETE FROM users WHERE id=u1
db> exit
```

### Seed tool

Generates 100,000 rows into `database/db.json` for load testing:

```bash
go run ./cmd/seed
```

---

## Running Tests & Benchmarks

```bash
# unit tests
go test ./src/...

# benchmark
go test ./src/db/... -bench=BenchmarkInsert -benchtime=5s -count=5

# capture a CPU profile
go test ./src/db/... -bench=BenchmarkInsert -cpuprofile=bench/cpu.prof

# inspect the profile
go tool pprof bench/cpu.prof

# compare two benchmark runs
benchstat bench/v1.txt bench/v2.txt
```

---

## File Manager interface

The `FM` interface is the only thing standing between the database and storage. Swap it out to change how data is persisted:

```go
type FM interface {
    Flush(data any) error
    Load(data any) error
}
```

The built-in implementation (`FmFullDump`) serializes all tables to a single JSON file on every checkpoint. Future implementations can use different formats or storage backends.

---

## Roadmap

### Max-performance write mode

When absolute throughput matters more than per-write durability, the WAL can be configured to skip `fsync` and only flush the OS buffer. Writes return as soon as the entry lands in the kernel page cache instead of waiting for the disk. A crash can lose the last few milliseconds of writes, but throughput increases significantly.

Planned API:
```go
db.Create("wal.json", "db.json", db.WithMaxPerformance())
```

Under the hood this changes `writerLoop` to call `buf.Flush()` without `f.Sync()`, and `Commit` to wait only for the buffer flush, not the fsync.

### Read-only mode

A read-only database opens the snapshot file once, loads everything into memory, and never touches the WAL. With no locking overhead on reads and the entire dataset in memory, query throughput is bounded only by CPU and allocations.

Planned API:
```go
db.CreateReadOnly("db.json")
```

Planned optimisations inside read-only mode:
- **Immutable row storage** — rows are stored as `[]byte` slices (raw JSON) and deserialized lazily only when accessed, keeping the hot path allocation-free for filtered queries.
- **Column indexing** — an optional `map[colValue][]int` index per column for O(1) point lookups instead of full table scans.
- **Zero-copy Get** — `Get` returns a read-only view of the internal slice rather than copying rows.
