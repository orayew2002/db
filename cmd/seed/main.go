package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Row struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Table struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Rows    []Row    `json:"rows"`
}

type DB struct {
	Users Table `json:"users"`
}

func main() {
	const N = 100000

	rows := make([]Row, 0, N)

	for i := 1; i <= N; i++ {
		rows = append(rows, Row{
			ID:    fmt.Sprintf("%d", i),
			Name:  fmt.Sprintf("User%d", i),
			Email: fmt.Sprintf("user%d@gmail.com", i),
		})
	}

	db := DB{
		Users: Table{
			Name:    "users",
			Columns: []string{"id", "name", "email"},
			Rows:    rows,
		},
	}

	file, _ := os.Create("database/db.json")
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "")

	encoder.Encode(db)

	fmt.Println("✅ Generated", N, "rows")
}
