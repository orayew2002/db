package wal

import (
	"os"
	"testing"
)

func TestCatalog(t *testing.T) {
	f := "../../database/catalog"

	// clean before test
	_ = os.RemoveAll(f)

	// ensure cleanup after test
	t.Cleanup(func() {
		_ = os.RemoveAll(f)
	})

	t.Run("open catalog", func(t *testing.T) {
		if _, err := OpenCatalog(f); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("insert data to catalog", func(t *testing.T) {
		c, err := OpenCatalog(f)
		if err != nil {
			t.Fatal(err)
		}

		if err := c.Write("users"); err != nil {
			t.Fatal(err)
		}

		c2, err := OpenCatalog(f)
		if err != nil {
			t.Fatal(err)
		}

		id := c2.GetID("users")
		if id != 1 {
			t.Fatalf("expected id 1 got %d", id)
		}
	})
}
