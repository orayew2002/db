package lexer

import (
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	t.Run("test tokenizer v1", func(t *testing.T) {
		q := []string{"SELECT", "*", "FROM", "users"}
		qt := []TokenType{KEYWORD, SYMBOL, KEYWORD, IDENTIFIER}

		runTest(t, q, qt)
	})

	t.Run("test tokenizer v2", func(t *testing.T) {
		q := []string{"SELECT", "id,name,email", "FROM", "users"}
		qt := []TokenType{KEYWORD, IDENTIFIER, SYMBOL, IDENTIFIER, SYMBOL, IDENTIFIER, KEYWORD, IDENTIFIER}

		runTest(t, q, qt)
	})
}

func runTest(t *testing.T, q []string, qt []TokenType) {
	query := strings.Join(q, " ")
	l := New(query)

	tokens := l.GetTokens()

	if len(tokens) != len(qt) {
		t.Fatalf("expected %d tokens, got %d", len(qt), len(tokens))
	}

	for i, token := range tokens {
		if token.Type != qt[i] {
			t.Errorf("wrong token type at %d: expected %d, got %d (%s)", i,
				qt[i],
				token.Type,
				token.Val,
			)
		}
	}
}
