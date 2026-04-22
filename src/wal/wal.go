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
	dataBuf    []byte        // reusable buffer for arg payload (protected by mu)
	recBuf     []byte        // reusable buffer for marshaled WalRecord (protected by mu)
	entry      lp.WalRecord  // reusable record struct (protected by mu)
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
		dataBuf: make([]byte, 0, 256),
		recBuf:  make([]byte, 0, 512),
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
	AppendRaw([]byte) []byte
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
	w.dataBuf = arg.AppendRaw(w.dataBuf[:0])

	w.entry.Lsn = w.lsn
	w.entry.Op = uint32(a)
	w.entry.TableId = w.catalog.GetID(table)
	w.entry.Data = w.dataBuf

	// Reserve 4 bytes for the length prefix, then marshal the record after them.
	// Filling both into a single heap-resident slice avoids a stack-escaping
	// [4]byte that would otherwise be forced onto the heap by the bufio.Writer's
	// internal io.Writer interface call.
	w.recBuf = append(w.recBuf[:0], 0, 0, 0, 0)
	var err error
	w.recBuf, err = proto.MarshalOptions{}.MarshalAppend(w.recBuf, &w.entry)
	if err != nil {
		return 0, fmt.Errorf("error marshal proto file: %w", err)
	}
	binary.LittleEndian.PutUint32(w.recBuf[:4], uint32(len(w.recBuf)-4))
	if _, err := w.buf.Write(w.recBuf); err != nil {
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
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}

		buf := make([]byte, size)
		if _, err := io.ReadFull(w.f, buf); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}

		var rec lp.WalRecord
		if err := proto.Unmarshal(buf, &rec); err != nil {
			break
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
		a.Val = ct.GetCols()

	case I:
		var m map[string]any
		if err := json.Unmarshal(rec.GetData(), &m); err != nil {
			return Action{}, err
		}
		a.Val = m

	case D:
		var d lp.Delete
		if err := proto.Unmarshal(rec.Data, &d); err != nil {
			return Action{}, err
		}

		a.Val = shared.Unmarshal(d.GetVal())
		a.Col = d.GetCol()

	case U:
		var u lp.Update
		if err := proto.Unmarshal(rec.Data, &u); err != nil {
			return Action{}, err
		}

		a.Val = map[string]any{
			"val":  shared.Unmarshal(u.GetVal()),
			"vals": shared.UnmarshalMap(u.GetArgs()),
		}

		a.Col = u.GetCol()
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

		var rec lp.WalRecord
		if err := proto.Unmarshal(data, &rec); err != nil {
			continue
		}

		if rec.GetLsn() > last {
			last = rec.GetLsn()
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
	if err := w.f.Sync(); err != nil {
		return err
	}
	if err := w.catalog.Close(); err != nil {
		return err
	}

	return w.f.Close()
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
