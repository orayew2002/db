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

	t.Run("delete without where", func(t *testing.T) {
		runTest(t, []string{
			"DELETE",
			"FROM",
			"users",
		})
	})

	t.Run("delete with where number", func(t *testing.T) {
		runTest(t, []string{
			"DELETE",
			"FROM",
			"users",
			"WHERE",
			"id",
			"=",
			"1",
		})
	})

	t.Run("delete with where string", func(t *testing.T) {
		runTest(t, []string{
			"DELETE",
			"FROM",
			"users",
			"WHERE",
			"name",
			"=",
			"'John'",
		})
	})

	t.Run("delete missing from", func(t *testing.T) {
		runTestFail(t, []string{
			"DELETE",
			"users",
		})
	})

	t.Run("delete missing table", func(t *testing.T) {
		runTestFail(t, []string{
			"DELETE",
			"FROM",
		})
	})

	t.Run("delete invalid where", func(t *testing.T) {
		runTestFail(t, []string{
			"DELETE",
			"FROM",
			"users",
			"WHERE",
			"=",
			"1",
		})
	})

	t.Run("delete missing operator", func(t *testing.T) {
		runTestFail(t, []string{
			"DELETE",
			"FROM",
			"users",
			"WHERE",
			"id",
			"1",
		})
	})

	t.Run("delete missing right side", func(t *testing.T) {
		runTestFail(t, []string{
			"DELETE",
			"FROM",
			"users",
			"WHERE",
			"id",
			"=",
		})
	})

	t.Run("create table basic", func(t *testing.T) {
		runTest(t, []string{
			"CREATE", "TABLE", "users",
			"(", "id", "INT", ",", "name", "TEXT", ")",
		})
	})

	t.Run("create table single column", func(t *testing.T) {
		runTest(t, []string{
			"CREATE", "TABLE", "users",
			"(", "id", "INT", ")",
		})
	})

	t.Run("create table multiple columns", func(t *testing.T) {
		runTest(t, []string{
			"CREATE", "TABLE", "users",
			"(", "id", "INT", ",", "name", "TEXT", ",", "email", "TEXT", ")",
		})
	})

	t.Run("create table missing table name", func(t *testing.T) {
		runTestFail(t, []string{
			"CREATE", "TABLE",
			"(", "id", "INT", ")",
		})
	})

	t.Run("create table missing opening paren", func(t *testing.T) {
		runTestFail(t, []string{
			"CREATE", "TABLE", "users",
			"id", "INT", ")",
		})
	})

	t.Run("create table missing column type", func(t *testing.T) {
		runTestFail(t, []string{
			"CREATE", "TABLE", "users",
			"(", "id", ",", "name", "TEXT", ")",
		})
	})

	t.Run("create table invalid column name", func(t *testing.T) {
		runTestFail(t, []string{
			"CREATE", "TABLE", "users",
			"(", ",", "INT", ")",
		})
	})

	t.Run("create table missing closing paren", func(t *testing.T) {
		runTestFail(t, []string{
			"CREATE", "TABLE", "users",
			"(", "id", "INT",
		})
	})

	t.Run("update basic without where", func(t *testing.T) {
		runTest(t, []string{
			"UPDATE", "users",
			"SET", "name", "=", "'John'",
		})
	})

	t.Run("update with where number", func(t *testing.T) {
		runTest(t, []string{
			"UPDATE", "users",
			"SET", "name", "=", "'John'",
			"WHERE", "id", "=", "1",
		})
	})

	t.Run("update with where string", func(t *testing.T) {
		runTest(t, []string{
			"UPDATE", "users",
			"SET", "name", "=", "'John'",
			"WHERE", "name", "=", "'Alice'",
		})
	})

	t.Run("update multiple set columns", func(t *testing.T) {
		runTest(t, []string{
			"UPDATE", "users",
			"SET",
			"name", "=", "'John'", ",",
			"email", "=", "'john@mail.com'",
		})
	})

	t.Run("update multiple set with where", func(t *testing.T) {
		runTest(t, []string{
			"UPDATE", "users",
			"SET",
			"name", "=", "'John'", ",",
			"email", "=", "'john@mail.com'",
			"WHERE", "id", "=", "1",
		})
	})

	t.Run("update missing table", func(t *testing.T) {
		runTestFail(t, []string{
			"UPDATE",
			"SET", "name", "=", "'John'",
		})
	})

	t.Run("update missing set", func(t *testing.T) {
		runTestFail(t, []string{
			"UPDATE", "users",
			"name", "=", "'John'",
		})
	})

	t.Run("update invalid set syntax", func(t *testing.T) {
		runTestFail(t, []string{
			"UPDATE", "users",
			"SET", "name", "'John'",
		})
	})

	t.Run("update invalid where", func(t *testing.T) {
		runTestFail(t, []string{
			"UPDATE", "users",
			"SET", "name", "=", "'John'",
			"WHERE", "=", "1",
		})
	})

	t.Run("update missing value", func(t *testing.T) {
		runTestFail(t, []string{
			"UPDATE", "users",
			"SET", "name", "=",
		})
	})

	t.Run("drop table basic", func(t *testing.T) {
		runTest(t, []string{
			"DROP", "TABLE", "users",
		})
	})

	t.Run("drop table missing table keyword", func(t *testing.T) {
		runTestFail(t, []string{
			"DROP", "users",
		})
	})

	t.Run("drop table missing table name", func(t *testing.T) {
		runTestFail(t, []string{
			"DROP", "TABLE",
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
