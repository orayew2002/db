package db

import (
	"errors"
	"fmt"

	"github.com/orayew2002/db/src/fm"
	"github.com/orayew2002/db/src/wal"
)

type Options struct {
	WFP string
	FFP string
	UWC bool
}
type Database struct {
	tables map[string]*Table
	fm     FM
	w      *wal.Wal

	// UWC (Use WAL Commit) indicates whether each request waits for WAL commit.
	// When enabled, every operation calls Commit() and waits for fsync,
	// providing strong durability guarantees but significantly reducing performance.
	uwc bool
}

// Create initializes a new Database instance.
//
// WFP (WAL File Path) specifies the location of the write-ahead log file.
// FFP (Full File Path) specifies the location of the full database dump file.
//
// Initialization flow:
// 1. Load previously persisted data from storage into memory.
// 2. Replay WAL (Write-Ahead Log) entries to restore the latest state.
// 3. Flush the updated in-memory state back to persistent storage.
func Create(o Options) *Database {
	db := Database{
		tables: make(map[string]*Table),
		w:      wal.NewWal(o.WFP),
		fm:     fm.NewFmFullDump(o.FFP),
		uwc:    o.UWC,
	}

	// Step 1: Load existing data from persistent storage.
	if err := db.fm.Load(&db.tables); err != nil {
		fmt.Println(err.Error())
	}

	// Step 2: Replay WAL entries to reconstruct the latest state.
	if err := db.w.Replay(db.walFunction()); err != nil {
		fmt.Println(err.Error())
	}

	// Step 3: Persist the updated state to storage.
	if err := db.fm.Flush(db.tables); err != nil {
		fmt.Println(err.Error())
	}

	// Step 4: Clear WAL after successful synchronization.
	if err := db.w.Reset(); err != nil {
		fmt.Println(err.Error())
	}

	return &db
}

func (db *Database) walFunction() func(a wal.Action) {
	return func(a wal.Action) {
		if err := db.apply(a); err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (d *Database) Setfm(fm FM) error {
	d.fm = fm
	return d.fm.Load(&d.tables)
}

func (d *Database) CreateTable(name string, columns []string) error {
	if _, ext := d.tables[name]; ext {
		return nil
	}

	lsn, err := d.w.Append(wal.CT, name, "", columns)
	if err != nil {
		return err
	}

	d.applyCreateTable(name, columns)

	if d.uwc {
		return d.w.Commit(lsn)
	}

	return nil
}

func (d *Database) Delete(t string, col string, val any) error {
	if err := d.checkTable(t); err != nil {
		return err
	}

	lsn, err := d.w.Append(wal.D, t, col, val)
	if err != nil {
		return err
	}

	d.applyDelete(t, col, val)
	d.w.Commit(lsn)

	return nil
}

func (d *Database) Insert(t string, v map[string]any) error {
	if err := d.checkTable(t); err != nil {
		return err
	}

	table := d.tables[t]
	if len(table.Columns) != len(v) {
		return errors.New("error vals count mismatching")
	}

	lsn, err := d.w.Append(wal.I, t, "", v)
	if err != nil {
		return errors.New("error append to wall file action")
	}
	d.applyInsert(t, v)
	_ = d.w.Commit(lsn)

	return nil
}

// Update item structure
type us struct {
	val  any
	vals vals
}

func (d *Database) Update(name string, col string, val any, v vals) error {
	if err := d.checkTable(name); err != nil {
		return err
	}

	table := d.tables[name]
	if len(table.Columns) != len(v) {
		return errors.New("error vals count mismatching")
	}

	lsn, err := d.w.Append(wal.U, name, col, us{
		val:  val,
		vals: v,
	})
	if err != nil {
		return err
	}

	d.applyUpdate(name, col, val, v)
	d.w.Commit(lsn)

	return nil
}

func (d *Database) Get(name string) ([]map[string]any, error) {
	if err := d.checkTable(name); err != nil {
		return nil, err
	}

	return d.tables[name].Rows, nil
}

func (d *Database) GetColumns(name string) ([]string, error) {
	if err := d.checkTable(name); err != nil {
		return nil, err
	}

	return append([]string(nil), d.tables[name].Columns...), nil
}

func (d *Database) checkTable(name string) error {
	if _, ext := d.tables[name]; !ext {
		return errors.New("table not exists")
	}

	return nil
}

func (d *Database) GetTables() []string {
	var t []string
	for table := range d.tables {
		t = append(t, table)
	}

	return t
}

func (d *Database) Close() {
	d.w.Close()
}

func (d *Database) apply(a wal.Action) error {
	switch a.T {
	case wal.CT:
		d.applyCreateTable(a.Table, toStringSlice(a.Val))
		return nil
	case wal.I:
		v, ok := a.Val.(map[string]any)
		if !ok {
			return errors.New("invalid insert payload")
		}
		if err := d.checkTable(a.Table); err != nil {
			return err
		}

		d.applyInsert(a.Table, v)
		return nil
	case wal.D:
		if err := d.checkTable(a.Table); err != nil {
			return err
		}

		d.applyDelete(a.Table, a.Col, a.Val)
		return nil
	case wal.U:
		m, ok := a.Val.(map[string]any)
		if !ok {
			return errors.New("invalid update payload")
		}

		v, ok := m["vals"].(map[string]any)
		if !ok {
			return errors.New("invalid update values")
		}

		if err := d.checkTable(a.Table); err != nil {
			return err
		}
		d.applyUpdate(a.Table, a.Col, m["val"], vals(v))
		return nil
	default:
		return fmt.Errorf("unsupported action %q", a.T)
	}
}

func (d *Database) applyCreateTable(name string, columns []string) {
	if _, ext := d.tables[name]; ext {
		return
	}

	d.tables[name] = &Table{
		Name:    name,
		Columns: append([]string(nil), columns...),
	}
}

func (d *Database) applyDelete(table string, col string, val any) {
	d.tables[table].Delete(col, val)
}

func (d *Database) applyInsert(table string, v map[string]any) {
	d.tables[table].Insert(mapToRow(v, d.tables[table].Columns))
}

func (d *Database) applyUpdate(table string, col string, val any, v vals) {
	d.tables[table].Update(col, val, v.toRow(d.tables[table].Columns))
}

type vals map[string]any

func (v vals) toRow(cols []string) Row {
	row := make(Row, len(cols))

	for _, key := range cols {
		var val any

		if v, ext := v[key]; ext {
			val = v
		}

		row[key] = val
	}

	return row
}

func mapToRow(v map[string]any, cols []string) Row {
	row := make(Row, len(cols))

	for _, key := range cols {
		var val any

		if v, ext := v[key]; ext {
			val = v
		}

		row[key] = val
	}

	return row
}

func toStringSlice(v any) []string {
	switch val := v.(type) {

	case []string:
		return val

	case []any:
		res := make([]string, 0, len(val))
		for _, item := range val {
			switch s := item.(type) {
			case string:
				res = append(res, s)
			case []byte:
				res = append(res, string(s))
			default:
				res = append(res, fmt.Sprintf("%v", s))
			}
		}
		return res

	case nil:
		return nil

	default:
		return []string{fmt.Sprintf("%v", val)}
	}
}
