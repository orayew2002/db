package db

import (
	"encoding/json"
	"reflect"
)

type Table struct {
	Name    string           `json:"name"`
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

func (t *Table) Insert(row Row) {
	t.Rows = append(t.Rows, row)
}

func equalValues(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}

	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			return af == bf
		}
	}

	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	}

	return reflect.DeepEqual(a, b)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case json.Number:
		if f, err := n.Float64(); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (t *Table) Delete(col string, val any) {
	n := 0
	for _, row := range t.Rows {
		if rv, ext := row[col]; ext && equalValues(rv, val) {
			continue
		}

		t.Rows[n] = row
		n++
	}

	t.Rows = t.Rows[:n]
}

func (t *Table) Update(col string, val any, changes map[string]any) {
	for i, row := range t.Rows {
		if rv, ext := row[col]; ext && equalValues(rv, val) {
			next := make(Row, len(row))
			for key, current := range row {
				next[key] = current
			}

			for key, current := range changes {
				next[key] = current
			}

			t.Rows[i] = next
		}
	}
}
