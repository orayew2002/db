package parser

import (
	"strings"
	"testing"

	"github.com/orayew2002/db/src/lexer"
)

func TestParser(t *testing.T) {
	t.Run("testing parser v1", func(t *testing.T) {
		runTest(t, []string{"SELECT", "*", "FROM", "users"})
	})

	t.Run("testing parser v2", func(t *testing.T) {
		runTest(t, []string{"SELECT", "id,name,email", "FROM", "users"})
	})
}

func runTest(t *testing.T, q []string) {
	query := strings.Join(q, " ")
	l := lexer.New(query)
	p := New(l)

	if _, err := p.Parse(); err != nil {
		t.Error(err.Error())
	}
}
