package cli

type DB interface {
	GetTables() []string
	Get(table string) []map[string]any
	GetColumns(table string) []string
	CreateTable(name string, rows []string) error
	Insert(table string, vals map[string]any)
	Delete(table string, col string, val any)
}
