package db

import (
	"errors"
	"fmt"

	"github.com/orayew2002/db/src/fm"
)

type Options struct {
	WFP string
	FFP string
	UWC bool
}

type Database struct {
	tables map[string]*Table
	fm     FM
	w      *Wal

	// UWC (Use WAL Commit) indicates whether each request waits for WAL commit.
	// When enabled, every operation calls Commit() and waits for fsync,
	// providing strong durability guarantees but significantly reducing performance.
	uwc bool
}

// WFP -> WAl file path , FFP -> DB file path
func Create(o Options) *Database {
	db := Database{
		tables: make(map[string]*Table),
		w:      NewWal(o.WFP),
		fm:     fm.NewFmFullDump(o.FFP),
		uwc:    o.UWC,
	}

	if err := db.fm.Load(&db.tables); err != nil {
		fmt.Println(err.Error())
	}

	if err := db.w.Replay(db.walFunction()); err != nil {
		fmt.Println(err.Error())
	}

	if err := db.fm.Flush(db.tables); err != nil {
		fmt.Println(err.Error())
	}

	return &db
}

func (db *Database) walFunction() func(a action) {
	return func(a action) {
		switch a.A {
		case CT:
			db.CreateTable(a.Table, toStringSlice(a.Val))

		case I:
			v, ok := a.Val.(map[string]any)
			if ok {
				db.Insert(a.Table, v)
			}

		case D:
			db.Delete(a.Table, a.Col, a.Val)

		case U:
			m, ok := a.Val.(map[string]any)
			if !ok {
				return
			}
			v, ok := m["vals"].(map[string]any)
			if !ok {
				return
			}
			db.Update(a.Table, a.Col, m["val"], vals(v))
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

	lsn, err := d.w.Append(CT, name, "", columns)
	if err != nil {
		return err
	}

	d.tables[name] = &Table{
		Name:    name,
		Columns: columns,
	}

	if d.uwc {
		return d.w.Commit(lsn)
	}

	return nil
}

func (d *Database) Delete(t string, col string, val any) {
	d.checkTable(t)
	lsn, err := d.w.Append(D, t, col, val)
	if err != nil {
		panic(err.Error())
	}

	d.tables[t].Delete(col, val)
	if d.uwc {
		d.w.Commit(lsn)
	}
}

func (d *Database) Insert(t string, v map[string]any) {
	d.checkTable(t)

	table := d.tables[t]
	if len(table.Columns) != len(v) {
		panic(errors.New("error vals count mismatching"))
	}

	lsn, err := d.w.Append(I, t, "", v)
	if err != nil {
		panic(err.Error())
	}
	table.Insert(mapToRow(v, table.Columns))

	if d.uwc {
		_ = d.w.Commit(lsn)
	}
}

// Update item structure
type us struct {
	val  any
	vals vals
}

func (d *Database) Update(name string, col string, val any, v vals) {
	d.checkTable(name)

	table := d.tables[name]
	if len(table.Columns) != len(v) {
		panic(errors.New("error vals count mismatching"))
	}

	lsn, err := d.w.Append(U, name, col, us{
		val:  val,
		vals: v,
	})
	if err != nil {
		panic(err.Error())
	}

	table.Update(col, val, v.toRow(table.Columns))
	if d.uwc {
		_ = d.w.Commit(lsn)
	}
}

func (d *Database) Get(name string) []map[string]any {
	d.checkTable(name)
	return d.tables[name].Rows
}

func (d *Database) checkTable(name string) {
	if _, ext := d.tables[name]; !ext {
		panic(errors.New("table not exists"))
	}
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
