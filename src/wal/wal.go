package wal

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	lp "github.com/orayew2002/db/src/proto"
	"google.golang.org/protobuf/proto"
)

// T represents the type of WAL operation.
type T uint32

const (
	I  T = iota // Insert a new row into a table
	D           // Delete an existing row from a table
	U           // Update an existing row in a table
	CT          // Create a new table
	UT          // Modify table structure (e.g., add/drop column)
	DT          // Drop (remove) a table from the database
)

type Wal struct {
	mu         sync.Mutex    // protects concurrent access to WAL state and buffer writes
	cond       *sync.Cond    // used to coordinate goroutines waiting for WAL flush (sync signaling)
	lsn        uint64        // Log Sequence Number: monotonically increasing ID for each WAL record
	flushedLSN uint64        // last LSN that has been safely flushed to disk
	f          *os.File      // underlying WAL file on disk
	buf        *bufio.Writer // buffered writer to reduce disk I/O and improve performance
	closed     bool          // indicates whether WAL is closed and no further writes are allowed
	done       chan struct{} // signals shutdown/termination of WAL background workers (if any)
	catalog    *Catalog      // schema registry used to resolve table/column names into numeric IDs
}

type Action struct {
	LSN   uint64
	T     T
	Table string
	Col   string
	Val   any
}

func NewWal(path string) *Wal {
	_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	ctlPath := filepath.Join(filepath.Dir(path), "catalog")
	ctl, err := OpenCatalog(ctlPath)
	if err != nil {
		panic(err.Error())
	}

	w := &Wal{
		f:       f,
		buf:     bufio.NewWriterSize(f, 1<<20),
		catalog: ctl,
		done:    make(chan struct{}),
	}

	w.cond = sync.NewCond(&w.mu)

	w.recoverLSN()
	go w.writerLoop()

	return w
}

func (w *Wal) writerLoop() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.flush()
		case <-w.done:
			return
		}
	}
}

func (w *Wal) Append(a T, table, col string, val any) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errors.New("wal closed")
	}

	fmt.Println(a, table, col, val)
	w.lsn++
	lsn := w.lsn

	entry := &lp.WalRecord{
		Lsn:     lsn,
		Op:      uint32(a),
		TableId: w.catalog.GetID(table),
		ColId:   1,
		Value:   nil,
	}

	b, err := proto.Marshal(entry)
	if err != nil {
		return 0, err
	}

	if err := binary.Write(w.buf, binary.LittleEndian, uint32(len(b))); err != nil {
		return 0, err
	}

	if _, err := w.buf.Write(b); err != nil {
		return 0, err
	}

	return lsn, nil
}

func (w *Wal) Commit(lsn uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for w.flushedLSN < lsn {
		w.cond.Wait()
	}
	return nil
}

func (w *Wal) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Buffered() == 0 {
		return
	}

	_ = w.buf.Flush()
	_ = w.f.Sync()

	w.flushedLSN = w.lsn
	w.cond.Broadcast()
}

func (w *Wal) Replay(handler func(a Action)) error {
	// Flush and seek under the lock to avoid racing with writerLoop.
	w.mu.Lock()
	_ = w.buf.Flush()
	_, err := w.f.Seek(0, 0)
	w.mu.Unlock()

	if err != nil {
		return err
	}

	reader := bufio.NewReader(w.f)

	for {
		var size uint32
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		data := make([]byte, size)
		if _, err := io.ReadFull(reader, data); err != nil {
			return err
		}

		var act Action
		if err := json.Unmarshal(data, &act); err == nil {
			handler(act) // SAFE: no mutex held
		}
	}

	w.mu.Lock()
	_, err = w.f.Seek(0, io.SeekEnd)
	w.mu.Unlock()

	return err
}

func (w *Wal) recoverLSN() {
	_, err := w.f.Seek(0, 0)
	if err != nil {
		return
	}

	reader := bufio.NewReader(w.f)

	var last uint64

	for {
		var size uint32
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			break
		}

		data := make([]byte, size)
		if _, err := io.ReadFull(reader, data); err != nil {
			break
		}

		var act Action
		if err := json.Unmarshal(data, &act); err != nil {
			continue
		}

		if act.LSN > last {
			last = act.LSN
		}
	}

	w.lsn = last
	w.flushedLSN = last

	_, _ = w.f.Seek(0, io.SeekEnd)
}

func (w *Wal) Close() error {
	close(w.done) // signal writerLoop to exit

	w.mu.Lock()
	defer w.mu.Unlock()

	w.closed = true

	_ = w.buf.Flush()
	return w.f.Sync()
}

func (w *Wal) Reset() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.buf.Flush(); err != nil {
		return err
	}

	if err := w.f.Truncate(0); err != nil {
		return err
	}

	if _, err := w.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := w.f.Sync(); err != nil {
		return err
	}

	w.lsn = 0
	w.flushedLSN = 0

	return nil
}
