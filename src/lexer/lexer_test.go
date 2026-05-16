package lexer

import (
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	t.Run("test tokenizer", func(t *testing.T) {
		q := []string{"SELECT", "*", "FROM", "users"}
		qt := []TokenType{KEYWORD, SYMBOL, KEYWORD, IDENTIFIER}

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
	})
}
