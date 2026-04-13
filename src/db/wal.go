package db

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	lp "github.com/orayew2002/db/src/proto"
	"google.golang.org/protobuf/proto"
)

type Action uint32

const (
	I Action = iota
	D
	U
	CT
	UT
	DT
)

type Wal struct {
	mu         sync.Mutex
	cond       *sync.Cond
	lsn        uint64
	flushedLSN uint64
	f          *os.File
	buf        *bufio.Writer
	closed     bool
	done       chan struct{}
}

type action struct {
	LSN   uint64 `json:"lsn"`
	A     Action `json:"a"`
	Table string `json:"table"`
	Col   string `json:"col"`
	Val   any    `json:"val"`
}

func NewWal(path string) *Wal {
	_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	w := &Wal{
		f:    f,
		buf:  bufio.NewWriterSize(f, 1<<20),
		done: make(chan struct{}),
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

func (w *Wal) Append(a Action, table, col string, val any) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errors.New("wal closed")
	}

	w.lsn++
	lsn := w.lsn

	entry := &lp.WalRecord{
		Lsn:     lsn,
		Op:      uint32(a),
		TableId: 1,
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

func (w *Wal) Replay(handler func(a action)) error {
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

		var act action
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

		var act action
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
