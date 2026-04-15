package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	lp "github.com/orayew2002/db/src/proto"
	"google.golang.org/protobuf/proto"
)

// Catalog is a simple in-memory + file-backed key-value registry.
//
// It is used to store metadata mappings such as:
//   - table name  -> table id
//   - column name -> column id
//
// Why this exists:
// Databases should NOT store strings inside WAL or internal execution logs.
// Instead, they store small numeric IDs for performance and consistency.
// Catalog provides a way to convert human-readable values into stable IDs.
//
// Storage model:
// This implementation uses a SNAPSHOT-based approach:
//   - entire catalog is stored as a single protobuf file
//   - every update rewrites the full file
//   - on startup, file is loaded into memory
//
// This makes it:
//
//	✔ simple
//	✔ fast to read
//	✔ easy to recover
//
// but:
//
//	❌ not optimal for very large datasets
type Catalog struct {
	m  *sync.Mutex       // protects concurrent access
	f  *os.File          // underlying catalog file
	vi map[string]uint32 // in-memory index: value -> id
	iv map[uint32]string // in-memory index: id -> value
	cs *lp.Catalogs      // full snapshot stored in memory
	i  uint32            // next available ID
}

// OpenCatalog opens existing catalog file or creates a new one.
//
// It also loads existing data from disk into memory.
func OpenCatalog(path string) (*Catalog, error) {
	if _, err := os.Stat(path); err != nil {
		fDir := filepath.Dir(path)
		_ = os.MkdirAll(fDir, os.ModePerm)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	c := Catalog{f: f, m: &sync.Mutex{}}
	if err := c.Load(); err != nil {
		return nil, fmt.Errorf("error opening catalog file: %w", err)
	}

	return &c, nil
}

// Load reads catalog file from disk and restores it into memory.
//
// Steps:
//  1. Read full protobuf file
//  2. Unmarshal into Catalogs struct
//  3. Rebuild in-memory map (value -> id)
//  4. Restore next ID counter
//
// This is called automatically on startup.
func (c *Catalog) Load() error {
	c.m.Lock()
	defer c.m.Unlock()

	stat, err := c.f.Stat()
	if err != nil {
		return err
	}

	// If file is empty, initialize fresh state
	if stat.Size() == 0 {
		c.cs = &lp.Catalogs{}
		c.vi = make(map[string]uint32)
		c.iv = make(map[uint32]string)
		c.i = 1
		return nil
	}

	// Read full file into memory
	buf := make([]byte, stat.Size())
	_, err = c.f.ReadAt(buf, 0)
	if err != nil {
		return err
	}

	// Decode protobuf snapshot
	var data lp.Catalogs
	if err := proto.Unmarshal(buf, &data); err != nil {
		return err
	}

	// Restore in-memory structures
	c.cs = &data
	c.vi = make(map[string]uint32)
	c.iv = make(map[uint32]string)

	var maxID uint32 = 0

	for _, item := range data.Catalogs {
		c.vi[item.Value] = item.Id
		c.iv[item.Id] = item.Value

		if item.Id > maxID {
			maxID = item.Id
		}
	}

	// Next ID must continue from highest existing ID
	c.i = maxID + 1

	return nil
}

// Write inserts a new value into the catalog.
//
// Behavior:
//   - If value already exists → returns immediately
//   - Otherwise assigns a new ID
//   - Updates in-memory structures
//   - Writes FULL snapshot back to disk
//
// Important:
// This is a snapshot-based system (not WAL append).
// Each write rewrites the entire file.
func (c *Catalog) Write(val string) error {
	c.m.Lock()
	defer c.m.Unlock()

	// Avoid duplicates
	if _, exists := c.vi[val]; exists {
		return nil
	}

	// Add new entry to snapshot in memory
	c.cs.Catalogs = append(c.cs.Catalogs, &lp.Catalog{
		Id:    c.i,
		Value: val,
	})

	// Serialize full snapshot
	fb, err := proto.Marshal(&lp.Catalogs{
		Catalogs: c.cs.Catalogs,
	})
	if err != nil {
		return err
	}

	// Reset file before rewriting
	if _, err := c.f.Seek(0, 0); err != nil {
		return err
	}

	if err := c.f.Truncate(0); err != nil {
		return err
	}

	// Write new snapshot
	if _, err := c.f.Write(fb); err != nil {
		return err
	}

	if err := c.f.Sync(); err != nil {
		return err
	}

	// Update memory state only after successful disk write
	c.vi[val] = c.i
	c.incrementID()

	return nil
}

// incrementID increases internal ID counter.
func (c *Catalog) incrementID() {
	atomic.AddUint32(&c.i, 1)
}

// GetID returns ID for a given value.
//
// Returns:
//   - id if exists
//   - 0 if not found
//
// Note: 0 is considered "not found".
func (c *Catalog) GetID(v string) uint32 {
	for range 2 {
		if id, exists := c.vi[v]; exists {
			return id
		}

		_ = c.Write(v)
	}

	return 0
}

// GetName returns the name for a given ID.
//
// Returns:
//   - name if exists
//   - empty string if not found
//
// Note: empty string ("") is considered "not found".
func (c *Catalog) GetName(id uint32) string {
	if v, exists := c.iv[id]; exists {
		return v
	}

	return ""
}
