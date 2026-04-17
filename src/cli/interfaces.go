package cli

type DB interface {
	GetTables() []string
	Get(table string) ([]map[string]any, error)
	GetColumns(table string) ([]string, error)
	CreateTable(name string, rows []string) error
	Insert(table string, vals map[string]any) error
	Update(table string, col string, val any, vals map[string]any) error
	Delete(table string, col string, val any) error
}
