package parser

import (
	"strings"
	"testing"

	"github.com/orayew2002/db/src/lexer"
)

func TestParser(t *testing.T) {
	t.Run("testing parser v1 select stmt", func(t *testing.T) {
		runTest(t, []string{"SELECT", "*", "FROM", "users"})
	})

	t.Run("testing parser v2 select stmt", func(t *testing.T) {
		runTest(t, []string{"SELECT", "id,name,email", "FROM", "users"})
	})

	t.Run("insert with strings", func(t *testing.T) {
		runTest(t, []string{
			"INSERT", "INTO", "users",
			"(", "id", ",", "name", ",", "email", ")",
			"VALUES",
			"(", "1", ",", "'John'", ",", "'john@mail.com'", ")",
		})
	})

	t.Run("insert without columns", func(t *testing.T) {
		runTest(t, []string{
			"INSERT", "INTO", "users",
			"VALUES",
			"(", "1", ",", "'John'", ",", "'john@mail.com'", ")",
		})
	})

	t.Run("insert multiple rows", func(t *testing.T) {
		runTest(t, []string{
			"INSERT", "INTO", "users",
			"(", "id", ",", "name", ")",
			"VALUES",
			"(", "1", ",", "'A'", ")",
			",",
			"(", "2", ",", "'B'", ")",
		})
	})

	t.Run("insert with null", func(t *testing.T) {
		runTest(t, []string{
			"INSERT", "INTO", "users",
			"(", "id", ",", "name", ",", "email", ")",
			"VALUES",
			"(", "1", ",", "NULL", ",", "'mail@test.com'", ")",
		})
	})

	t.Run("insert extra columns", func(t *testing.T) {
		runTest(t, []string{
			"INSERT", "INTO", "users",
			"(", "id", ",", "name", ",", "email", ",", "age", ")",
			"VALUES",
			"(", "1", ",", "'John'", ",", "'mail'", ",", "25", ")",
		})
	})

	t.Run("insert missing values keyword", func(t *testing.T) {
		runTestFail(t, []string{
			"INSERT", "INTO", "users",
			"(", "id", ",", "name", ")",
			"(", "1", ",", "'John'", ")",
		})
	})

	t.Run("insert missing closing paren", func(t *testing.T) {
		runTestFail(t, []string{
			"INSERT", "INTO", "users",
			"(", "id", ",", "name",
			"VALUES",
			"(", "1", ",", "'John'",
		})
	})
}

func runTest(t *testing.T, q []string) {
	query := strings.Join(q, " ")
	l := lexer.New(query)
	p := New(l)

	if _, err := p.Parse(); err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func runTestFail(t *testing.T, q []string) {
	query := strings.Join(q, " ")
	l := lexer.New(query)
	p := New(l)

	if _, err := p.Parse(); err == nil {
		t.Fatal("expected error but got nil")
	}
}
