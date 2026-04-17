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
	"github.com/orayew2002/db/src/shared"
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

type Arg interface {
	Raw() []byte
	Vals() []string
}

func (w *Wal) Append(a T, table string, arg Arg) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errors.New("wal closed")
	}

	if a == CT {
		if err := w.catalog.Write(table); err != nil {
			return 0, fmt.Errorf("error add field to wal catalog: %w", err)
		}

		for _, v := range arg.Vals() {
			if err := w.catalog.Write(v); err != nil {
				return 0, fmt.Errorf("error add field to wal catalog: %w", err)
			}
		}
	}

	w.lsn++
	entry := &lp.WalRecord{
		Lsn:     w.lsn,
		Op:      uint32(a),
		TableId: w.catalog.GetID(table),
		Data:    arg.Raw(),
	}

	b, err := proto.Marshal(entry)
	if err != nil {
		return 0, fmt.Errorf("error marhsal proto file: %w", err)
	}

	if err := binary.Write(w.buf, binary.LittleEndian, uint32(len(b))); err != nil {
		return 0, fmt.Errorf("error write marshaled data to file: %w", err)
	}

	if _, err := w.buf.Write(b); err != nil {
		return 0, err
	}

	return w.lsn, nil
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
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for {
		var size uint32
		if err := binary.Read(w.f, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		buf := make([]byte, size)
		if _, err := io.ReadFull(w.f, buf); err != nil {
			return err
		}

		var rec lp.WalRecord
		if err := proto.Unmarshal(buf, &rec); err != nil {
			return err
		}

		// 4. build action from record
		action, err := w.buildAction(&rec)
		if err != nil {
			return err
		}

		handler(action)
	}

	_, err := w.f.Seek(0, io.SeekEnd)
	return err
}

func (w *Wal) buildAction(rec *lp.WalRecord) (Action, error) {
	a := Action{
		LSN:   rec.Lsn,
		T:     T(rec.Op),
		Table: w.catalog.GetName(rec.TableId),
	}

	switch T(rec.Op) {
	case CT:
		var ct lp.CreateTable
		if err := proto.Unmarshal(rec.GetData(), &ct); err != nil {
			return Action{}, err
		}
		a.Val = ct.GetVals()

	case I:
		var d lp.Insert
		if err := proto.Unmarshal(rec.GetData(), &d); err != nil {
			return Action{}, err
		}
		a.Val = shared.UnmarshalMap(d.Val)

	case D:
		var d lp.Delete
		if err := proto.Unmarshal(rec.Data, &d); err != nil {
			return Action{}, err
		}

		a.Val = shared.Unmarshal(d.GetVal())
		a.Col = d.GetCol()

		// TODO
		// case UPDATE:
	}

	return a, nil
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
