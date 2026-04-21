package cli

import "github.com/orayew2002/db/src/db"

type DB interface {
	GetTables() []string
	Get(table string) ([]map[string]any, error)
	GetColumns(table string) ([]db.ColDef, error)
	CreateTable(name string, rows []db.ColDef) error
	Insert(table string, vals map[string]any) error
	Update(table string, col string, val any, vals map[string]any) error
	Delete(table string, col string, val any) error
}
