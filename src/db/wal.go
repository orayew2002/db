package db

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Action string

const (
	// Use this key for inserting data
	I Action = "insert"

	// Use this key for delete data
	D Action = "delete"

	// Use this key for update data
	U Action = "update"

	// Use this key for create table
	CT Action = "create_table"

	// Use this key for update table
	UT Action = "update_table"

	// Use this key for delete table
	DT Action = "delete_table"
)

type Wal struct {
	lsn uint64
	f   *os.File
	buf *bufio.Writer
}

type action struct {
	LSN   uint64 `json:"lsn"`
	A     Action `json:"a"`
	Table string `json:"table"`
	Col   string `json:"col"`
	Val   any    `json:"val"`
}

type update struct {
	val  any
	vals any
}

func NewWal(path string) *Wal {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	w := &Wal{
		f:   f,
		buf: bufio.NewWriter(f),
	}

	w.recoverLSN()

	return w
}

func (w *Wal) Append(a Action, table, col string, val any) error {
	entry := action{
		LSN:   w.lsn + 1,
		A:     a,
		Table: table,
		Col:   col,
		Val:   val,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	b = append(b, '\n')

	n, err := w.buf.Write(b)
	if err != nil || n != len(b) {
		return errors.New("write failed")
	}

	w.lsn++

	return nil
}

func (w *Wal) Sync() error {
	if err := w.buf.Flush(); err != nil {
		return err
	}
	return w.f.Sync()
}

func (w *Wal) AS(a Action, table, col string, val any) error {
	if err := w.Append(a, table, col, val); err != nil {
		return err
	}
	return w.Sync()
}

func (w *Wal) Replay(handler func(a action)) error {
	_, err := w.f.Seek(0, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(w.f)

	for scanner.Scan() {
		line := scanner.Bytes()
		var act action
		if err := json.Unmarshal(line, &act); err != nil {
			continue
		}

		handler(act)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if w.buf != nil {
		if err := w.buf.Flush(); err != nil {
			return err
		}
	}
	if err := w.f.Sync(); err != nil {
		return err
	}

	if err := w.f.Truncate(0); err != nil {
		return err
	}

	w.lsn = 0
	if _, err := w.f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	w.buf = bufio.NewWriter(w.f)
	return nil
}

func (w *Wal) recoverLSN() {
	_, err := w.f.Seek(0, 0)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(w.f)

	var last uint64

	for scanner.Scan() {
		var act action
		if err := json.Unmarshal(scanner.Bytes(), &act); err != nil {
			continue
		}

		if act.LSN > last {
			last = act.LSN
		}
	}

	w.lsn = last
	w.f.Seek(0, io.SeekEnd)
}

func (w *Wal) Close() error {
	if err := w.buf.Flush(); err != nil {
		return err
	}
	return w.f.Close()
}

func toStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}

	res := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			res = append(res, s)
		}
	}
	return res
}
