package db

// FM (File Manager) defines methods for persisting and restoring data
// to and from storage.
type FM interface {
	// Flush writes the provided data to persistent storage.
	Flush(data any) error

	// Load reads data from storage and populates the provided target.
	Load(data any) error
}
