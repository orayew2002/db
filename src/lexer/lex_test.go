package lexer

import (
	"testing"
)

type testCase struct {
	expectedType  TokenType
	expectedValue string
}

func runTests(t *testing.T, input string, tests []testCase) {
	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] wrong token type. expected=%s, got=%s",
				i, tt.expectedType, tok.Type)
		}

		if tok.Value != tt.expectedValue {
			t.Fatalf("test[%d] wrong token value. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}

func TestSelectBasic(t *testing.T) {
	input := "SELECT name FROM users;"

	tests := []testCase{
		{SELECT, "SELECT"},
		{IDENT, "name"},
		{FROM, "FROM"},
		{IDENT, "users"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	runTests(t, input, tests)
}

func TestComma(t *testing.T) {
	input := "SELECT name, age FROM users;"

	tests := []testCase{
		{SELECT, "SELECT"},
		{IDENT, "name"},
		{COMMA, ","},
		{IDENT, "age"},
		{FROM, "FROM"},
		{IDENT, "users"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	runTests(t, input, tests)
}

func TestWhereCondition(t *testing.T) {
	input := "SELECT age FROM users WHERE age > 18;"

	tests := []testCase{
		{SELECT, "SELECT"},
		{IDENT, "age"},
		{FROM, "FROM"},
		{IDENT, "users"},
		{WHERE, "WHERE"},
		{IDENT, "age"},
		{GT, ">"},
		{NUMBER, "18"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	runTests(t, input, tests)
}

func TestStringLiteral(t *testing.T) {
	input := "SELECT name FROM users WHERE city = 'Paris';"

	tests := []testCase{
		{SELECT, "SELECT"},
		{IDENT, "name"},
		{FROM, "FROM"},
		{IDENT, "users"},
		{WHERE, "WHERE"},
		{IDENT, "city"},
		{EQ, "="},
		{STRING, "Paris"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	runTests(t, input, tests)
}

func TestMultipleConditions(t *testing.T) {
	input := "SELECT name FROM users WHERE age > 18 AND city = 'Paris';"

	tests := []testCase{
		{SELECT, "SELECT"},
		{IDENT, "name"},
		{FROM, "FROM"},
		{IDENT, "users"},
		{WHERE, "WHERE"},
		{IDENT, "age"},
		{GT, ">"},
		{NUMBER, "18"},
		{AND, "AND"},
		{IDENT, "city"},
		{EQ, "="},
		{STRING, "Paris"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	runTests(t, input, tests)
}

func TestWhitespaceHandling(t *testing.T) {
	input := "   SELECT    name   FROM   users   ;  "

	tests := []testCase{
		{SELECT, "SELECT"},
		{IDENT, "name"},
		{FROM, "FROM"},
		{IDENT, "users"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	runTests(t, input, tests)
}

func TestIllegalToken(t *testing.T) {
	input := "SELECT @ FROM users;"

	tests := []testCase{
		{SELECT, "SELECT"},
		{ILLEGAL, "@"},
	}

	runTests(t, input, tests)
}
