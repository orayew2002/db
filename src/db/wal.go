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
)

type Action string

const (
	I  Action = "insert"
	D  Action = "delete"
	U  Action = "update"
	CT Action = "create_table"
	UT Action = "update_table"
	DT Action = "delete_table"
)

type Wal struct {
	mu         sync.Mutex
	cond       *sync.Cond
	lsn        uint64
	flushedLSN uint64
	f          *os.File
	buf        *bufio.Writer
	closed     bool
}

type action struct {
	LSN   uint64 `json:"lsn"`
	A     Action `json:"a"`
	Table string `json:"table"`
	Col   string `json:"col"`
	Val   any    `json:"val"`
}

func NewWal(path string) *Wal {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	w := &Wal{
		f:   f,
		buf: bufio.NewWriterSize(f, 1<<20),
	}

	w.cond = sync.NewCond(&w.mu)

	w.recoverLSN()

	go w.writerLoop()

	return w
}

func (w *Wal) writerLoop() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		w.flush()
	}
}

func (w *Wal) Append(a Action, table, col string, val any) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errors.New("wal closed")
	}

	lsn := w.lsn + 1

	entry := action{
		LSN:   lsn,
		A:     a,
		Table: table,
		Col:   col,
		Val:   val,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return 0, err
	}

	// length prefix (чтобы избежать проблем Scanner)
	if err := binary.Write(w.buf, binary.LittleEndian, uint32(len(b))); err != nil {
		return 0, err
	}

	if _, err := w.buf.Write(b); err != nil {
		return 0, err
	}

	w.lsn = lsn

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

	if w.buf.Buffered() == 0 {
		w.mu.Unlock()
		return
	}

	// flush buffer в OS
	if err := w.buf.Flush(); err != nil {
		w.mu.Unlock()
		return
	}

	// fsync (самое дорогое, но важное)
	if err := w.f.Sync(); err != nil {
		w.mu.Unlock()
		return
	}

	w.flushedLSN = w.lsn
	w.cond.Broadcast()

	w.mu.Unlock()
}

func (w *Wal) Replay(handler func(a action)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.buf.Flush(); err != nil {
		return err
	}

	if _, err := w.f.Seek(0, 0); err != nil {
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
		if err := json.Unmarshal(data, &act); err != nil {
			continue
		}

		handler(act)
	}

	_, err := w.f.Seek(0, io.SeekEnd)
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

	w.f.Seek(0, io.SeekEnd)
}

func (w *Wal) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.closed = true

	if err := w.buf.Flush(); err != nil {
		return err
	}

	return w.f.Sync()
}
