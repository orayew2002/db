package db

// FM -> File Manager for saving data to storage
type FM interface {
	Flush(data any) error
	Load(data any) error
}
